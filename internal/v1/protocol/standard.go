/*
Copyright 2021 The JamJar Relay Server Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package protocol

import (
	"fmt"
	"net/http"

	"github.com/golang/glog"
	"github.com/jamjarlabs/jamjar-relay-server/internal/api/v1/api"
	"github.com/jamjarlabs/jamjar-relay-server/internal/v1/room"
	"github.com/jamjarlabs/jamjar-relay-server/internal/v1/session"
	"github.com/jamjarlabs/jamjar-relay-server/specs/v1/client"
	"github.com/jamjarlabs/jamjar-relay-server/specs/v1/relay"
	"github.com/jamjarlabs/jamjar-relay-server/specs/v1/transport"
	"google.golang.org/protobuf/proto"

	roomspec "github.com/jamjarlabs/jamjar-relay-server/specs/v1/room"
)

type StandardProtocol struct {
	RoomManager room.RoomManager
}

func (p *StandardProtocol) Connect(payload *transport.Payload, connected *session.Session, currentRoom room.Room) (*session.Session, room.Room, bool) {
	if currentRoom != nil {
		connected.Write <- api.WebSocketFail(&transport.Error{
			Code:    http.StatusBadRequest,
			Message: "Cannot connect to a different room while already connected to another",
		})
		return connected, currentRoom, false
	}

	joinRequest := &roomspec.JoinRoomRequest{}
	err := proto.Unmarshal(payload.Data, joinRequest)
	if err != nil {
		connected.Write <- api.WebSocketFail(&transport.Error{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Invalid join request provided, does not conform to spec, %v", err),
		})
		return connected, currentRoom, false
	}

	rooms, err := p.RoomManager.GetRoomList()
	if err != nil {
		connected.Write <- api.WebSocketFail(&transport.Error{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("Failed to retrieve room list, %v", err),
		})
		return connected, currentRoom, false
	}

	for _, matchRoom := range rooms {
		if !matchRoom.RoomMatches(joinRequest.RoomID, joinRequest.RoomSecret) {
			continue
		}
		connected, err = matchRoom.NewClient(connected)
		if err != nil {
			switch v := err.(type) {
			case room.ErrRoomFull:
				connected.Write <- api.WebSocketFail(&transport.Error{
					Code:    http.StatusBadRequest,
					Message: v.Message,
				})
				return connected, currentRoom, false
			default:
				connected.Write <- api.WebSocketFail(&transport.Error{
					Code:    http.StatusInternalServerError,
					Message: fmt.Sprintf("Failed to register new client to room, %v", err),
				})
				return connected, currentRoom, false
			}
		}

		responseClient := &client.Client{
			ID:     connected.Client.ID,
			Secret: connected.Client.Secret,
		}

		responseData, err := proto.Marshal(responseClient)
		if err != nil {
			// Should not occur, panic
			panic(err)
		}

		connected.Write <- api.WebSocketSucceed(&transport.Payload{
			Flag: transport.Payload_RESPONSE_CONNECT,
			Data: responseData,
		})

		p.setHostIfNone(connected, matchRoom)

		return connected, matchRoom, false
	}

	connected.Write <- api.WebSocketFail(&transport.Error{
		Code:    http.StatusBadRequest,
		Message: fmt.Sprintf("No valid room match found for ID %d", joinRequest.RoomID),
	})

	return connected, currentRoom, false
}

func (p *StandardProtocol) Reconnect(payload *transport.Payload, connected *session.Session, currentRoom room.Room) (*session.Session, room.Room, bool) {
	if currentRoom != nil {
		connected.Write <- api.WebSocketFail(&transport.Error{
			Code:    http.StatusBadRequest,
			Message: "Cannot connect to a different room while already connected to another",
		})
		return connected, currentRoom, false
	}

	rejoinRequest := &roomspec.RejoinRoomRequest{}
	err := proto.Unmarshal(payload.Data, rejoinRequest)
	if err != nil {
		connected.Write <- api.WebSocketFail(&transport.Error{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Invalid join request provided, does not conform to spec, %v", err),
		})
		return connected, currentRoom, false
	}

	rooms, err := p.RoomManager.GetRoomList()
	if err != nil {
		connected.Write <- api.WebSocketFail(&transport.Error{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("Failed to retrieve room list, %v", err),
		})
		return connected, currentRoom, false
	}

	for _, matchRoom := range rooms {
		if matchRoom.RoomMatches(rejoinRequest.RoomID, rejoinRequest.RoomSecret) {
			connected, err = matchRoom.ExistingClient(connected, rejoinRequest.ClientID, rejoinRequest.ClientSecret)
			if err != nil {
				switch v := err.(type) {
				case room.ErrInvalidSecret:
					connected.Write <- api.WebSocketFail(&transport.Error{
						Code:    http.StatusBadRequest,
						Message: v.Message,
					})
					return connected, currentRoom, false
				case room.ErrRoomFull:
					connected.Write <- api.WebSocketFail(&transport.Error{
						Code:    http.StatusBadRequest,
						Message: v.Message,
					})
					return connected, currentRoom, false
				default:
					connected.Write <- api.WebSocketFail(&transport.Error{
						Code:    http.StatusInternalServerError,
						Message: fmt.Sprintf("Failed to register existing client to room, %v", err),
					})
					return connected, currentRoom, false
				}
			}

			responseClient := &client.Client{
				ID:     connected.Client.ID,
				Secret: connected.Client.Secret,
			}

			responseData, err := proto.Marshal(responseClient)
			if err != nil {
				// Should not occur, panic
				panic(err)
			}

			connected.Write <- api.WebSocketSucceed(&transport.Payload{
				Flag: transport.Payload_RESPONSE_CONNECT,
				Data: responseData,
			})

			p.setHostIfNone(connected, matchRoom)

			return connected, matchRoom, false
		}
	}

	connected.Write <- api.WebSocketFail(&transport.Error{
		Code:    http.StatusBadRequest,
		Message: fmt.Sprintf("No valid room match found for ID %d", rejoinRequest.RoomID),
	})
	return connected, currentRoom, false
}

func (p *StandardProtocol) Disconnect(connected *session.Session, room room.Room) bool {
	connected.Close()
	if connected.Client == nil || room == nil {
		return true
	}

	isHost, err := room.IsHost(connected.Client)
	if err != nil {
		glog.Errorf("Failed to determine if disconnecting client with ID %d is host, %v", connected.Client.ID, err)
	}

	err = room.RemoveClient(connected.Client.ID)
	if err != nil {
		glog.Errorf("Failed to disconnect client with ID %d, %v", connected.Client.ID, err)
	}

	if isHost {
		err := p.migrateHost(room)
		if err != nil {
			glog.Errorf("Failed to migrate host, %v", err)
		}
	}
	return true
}

func (p *StandardProtocol) List(payload *transport.Payload, connected *session.Session, room room.Room) bool {
	if connected == nil || room == nil {
		connected.Write <- api.WebSocketFail(&transport.Error{
			Code:    http.StatusBadRequest,
			Message: "Must be connected to a room to list a room's clients",
		})
		return false
	}

	connectedClients, err := room.GetConnected()
	if err != nil {
		connected.Write <- api.WebSocketFail(&transport.Error{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("Failed to retrieve room's connected clients, %v", err),
		})
		return false
	}

	list := make([]*client.SanitisedClient, 0)
	for i := 0; i < len(connectedClients); i++ {
		connectedClient := connectedClients[i]
		host, err := room.IsHost(connectedClient.Client)
		if err != nil {
			connected.Write <- api.WebSocketFail(&transport.Error{
				Code:    http.StatusInternalServerError,
				Message: fmt.Sprintf("Failed to determine if client is host, %v", err),
			})
			return false
		}
		list = append(list, &client.SanitisedClient{
			ID:   connectedClient.Client.ID,
			Host: host,
		})
	}

	responseData, err := proto.Marshal(&client.ClientList{
		List: list,
	})
	if err != nil {
		// Should not occur, panic
		panic(err)
	}

	connected.Write <- api.WebSocketSucceed(&transport.Payload{
		Flag: transport.Payload_RESPONSE_LIST,
		Data: responseData,
	})
	return false

}

func (p *StandardProtocol) RelayMessage(payload *transport.Payload, connected *session.Session, room room.Room) bool {
	if connected == nil || room == nil {
		connected.Write <- api.WebSocketFail(&transport.Error{
			Code:    http.StatusBadRequest,
			Message: "Must be connected to a room to relay a message",
		})
		return false
	}

	relayMsg := &relay.Relay{}
	err := proto.Unmarshal(payload.Data, relayMsg)
	if err != nil {
		connected.Write <- api.WebSocketFail(&transport.Error{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Relayed message does not conform to spec, %v", err),
		})
		return false
	}

	connectedClientList, err := room.GetConnected()
	if err != nil {
		connected.Write <- api.WebSocketFail(&transport.Error{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("Failed to retrieve room's connected clients, %v", err),
		})
		return false
	}

	isHost, err := room.IsHost(connected.Client)
	if err != nil {
		connected.Write <- api.WebSocketFail(&transport.Error{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("Failed to determine if client is host, %v", err),
		})
		return false
	}

	switch relayMsg.Type {
	case relay.Relay_BROADCAST:
		if !isHost {
			connected.Write <- api.WebSocketFail(&transport.Error{
				Code:    http.StatusBadRequest,
				Message: "Must be host to broadcast",
			})
			return false
		}
		p.broadcast(payload, connected, room, connectedClientList)
		return false
	case relay.Relay_TARGET:
		if !isHost {
			connected.Write <- api.WebSocketFail(&transport.Error{
				Code:    http.StatusBadRequest,
				Message: "Must be host to send targeted messages",
			})
			return false
		}

		if relayMsg.Target == nil {
			connected.Write <- api.WebSocketFail(&transport.Error{
				Code:    http.StatusBadRequest,
				Message: "Must provide a target ID to send a message to",
			})
			return false
		}

		for _, connectedClient := range connectedClientList {
			if *relayMsg.Target != connectedClient.Client.ID {
				continue
			}
			connectedClient.Write <- api.WebSocketSucceed(&transport.Payload{
				Flag: transport.Payload_RESPONSE_RELAY_MESSAGE,
				Data: payload.Data,
			})
			return false
		}
		connected.Write <- api.WebSocketFail(&transport.Error{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("No target client found with ID %d", *relayMsg.Target),
		})
		return false
	case relay.Relay_HOST:
		if isHost {
			connected.Write <- api.WebSocketFail(&transport.Error{
				Code:    http.StatusBadRequest,
				Message: "Hosts cannot send messages to themselves",
			})
			return false
		}

		host, err := room.GetHost()
		if err != nil {
			connected.Write <- api.WebSocketFail(&transport.Error{
				Code:    http.StatusInternalServerError,
				Message: fmt.Sprintf("Failed to get host, %v", err),
			})
			return false
		}

		host.Write <- api.WebSocketSucceed(&transport.Payload{
			Flag: transport.Payload_RESPONSE_RELAY_MESSAGE,
			Data: payload.Data,
		})
	}
	return false
}

func (p *StandardProtocol) setHostIfNone(connected *session.Session, matchRoom room.Room) {
	host, err := matchRoom.GetHost()
	if err != nil {
		connected.Write <- api.WebSocketFail(&transport.Error{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("Failed to retrieve the current host, %v", err),
		})
	} else {
		if host == nil {
			_, err := matchRoom.SetHost(&connected.Client.ID)
			if err != nil {
				connected.Write <- api.WebSocketFail(&transport.Error{
					Code:    http.StatusInternalServerError,
					Message: fmt.Sprintf("Failed to update host, %v", err),
				})
			} else {
				connected.Write <- api.WebSocketSucceed(&transport.Payload{
					Flag: transport.Payload_RESPONSE_ASSIGN_HOST,
				})
			}
		}
	}
}

func (p *StandardProtocol) broadcast(payload *transport.Payload, connected *session.Session, room room.Room, connectedClientList []*session.Session) {
	for _, connectedClient := range connectedClientList {
		if connected.Client.ID == connectedClient.Client.ID {
			// Message should only be sent to other clients, not sent back to origin
			continue
		}
		connectedClient.Write <- api.WebSocketSucceed(&transport.Payload{
			Flag: transport.Payload_RESPONSE_RELAY_MESSAGE,
			Data: payload.Data,
		})
	}
}

func (p *StandardProtocol) migrateHost(room room.Room) error {
	connectedClients, err := room.GetConnected()
	if err != nil {
		return err
	}

	if len(connectedClients) <= 0 {
		_, err = room.SetHost(nil)
		return err
	}

	for _, connectedClient := range connectedClients {
		connectedClient.Write <- api.WebSocketSucceed(&transport.Payload{
			Flag: transport.Payload_RESPONSE_BEGIN_HOST_MIGRATE,
		})
	}

	newHost := connectedClients[0]

	_, err = room.SetHost(&newHost.Client.ID)
	if err != nil {
		return err
	}

	finishHostMigrationResponse := &roomspec.FinishHostMigrationResponse{
		HostID: newHost.Client.ID,
	}

	finishMigrationBytes, err := proto.Marshal(finishHostMigrationResponse)
	if err != nil {
		// Should not occur, panic
		panic(err)
	}

	newHost.Write <- api.WebSocketSucceed(&transport.Payload{
		Flag: transport.Payload_RESPONSE_ASSIGN_HOST,
		Data: finishMigrationBytes,
	})

	for _, connectedClient := range connectedClients {
		connectedClient.Write <- api.WebSocketSucceed(&transport.Payload{
			Flag: transport.Payload_RESPONSE_FINISH_HOST_MIGRATE,
		})
	}

	return nil
}
