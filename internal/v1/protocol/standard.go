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
	apiv1 "github.com/jamjarlabs/jamjar-relay-server/internal/api/v1/api"
	roomv1 "github.com/jamjarlabs/jamjar-relay-server/internal/v1/room"
	sessionv1 "github.com/jamjarlabs/jamjar-relay-server/internal/v1/session"
	clientv1 "github.com/jamjarlabs/jamjar-relay-server/specs/v1/client"
	relayv1 "github.com/jamjarlabs/jamjar-relay-server/specs/v1/relay"
	roomspecv1 "github.com/jamjarlabs/jamjar-relay-server/specs/v1/room"
	transportv1 "github.com/jamjarlabs/jamjar-relay-server/specs/v1/transport"

	"google.golang.org/protobuf/proto"
)

type StandardProtocol struct {
	RoomManager roomv1.RoomManager
}

func (p *StandardProtocol) Connect(payload *transportv1.Payload, connected *sessionv1.Session, currentRoom roomv1.Room) (*sessionv1.Session, roomv1.Room, bool) {
	if currentRoom != nil {
		connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
			Code:    http.StatusBadRequest,
			Message: "Cannot connect to a different room while already connected to another",
		})
		return connected, currentRoom, false
	}

	joinRequest := &roomspecv1.JoinRoomRequest{}
	err := proto.Unmarshal(payload.Data, joinRequest)
	if err != nil {
		connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Invalid join request provided, does not conform to spec, %v", err),
		})
		return connected, currentRoom, false
	}

	rooms, err := p.RoomManager.ListRooms()
	if err != nil {
		connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
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
			case roomv1.ErrRoomFull:
				connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
					Code:    http.StatusBadRequest,
					Message: v.Message,
				})
				return connected, currentRoom, false
			default:
				connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
					Code:    http.StatusInternalServerError,
					Message: fmt.Sprintf("Failed to register new client to room, %v", err),
				})
				return connected, currentRoom, false
			}
		}

		responseClient := &clientv1.Client{
			ID:     connected.Client.ID,
			Secret: connected.Client.Secret,
		}

		responseData, err := proto.Marshal(responseClient)
		if err != nil {
			// Should not occur, panic
			panic(err)
		}

		connected.Write <- apiv1.WebSocketSucceed(&transportv1.Payload{
			Flag: transportv1.Payload_RESPONSE_CONNECT,
			Data: responseData,
		})

		p.setHostIfNone(connected, matchRoom)

		return connected, matchRoom, false
	}

	connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
		Code:    http.StatusBadRequest,
		Message: fmt.Sprintf("No valid room match found for ID %d", joinRequest.RoomID),
	})

	return connected, currentRoom, false
}

func (p *StandardProtocol) Reconnect(payload *transportv1.Payload, connected *sessionv1.Session, room roomv1.Room) (*sessionv1.Session, roomv1.Room, bool) {
	if room != nil {
		connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
			Code:    http.StatusBadRequest,
			Message: "Cannot connect to a different room while already connected to another",
		})
		return connected, room, false
	}

	rejoinRequest := &roomspecv1.RejoinRoomRequest{}
	err := proto.Unmarshal(payload.Data, rejoinRequest)
	if err != nil {
		connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Invalid join request provided, does not conform to spec, %v", err),
		})
		return connected, room, false
	}

	rooms, err := p.RoomManager.ListRooms()
	if err != nil {
		connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("Failed to retrieve room list, %v", err),
		})
		return connected, room, false
	}

	for _, matchRoom := range rooms {
		if matchRoom.RoomMatches(rejoinRequest.RoomID, rejoinRequest.RoomSecret) {
			connected, err = matchRoom.ExistingClient(connected, rejoinRequest.ClientID, rejoinRequest.ClientSecret)
			if err != nil {
				switch v := err.(type) {
				case roomv1.ErrInvalidSecret:
					connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
						Code:    http.StatusBadRequest,
						Message: v.Message,
					})
					return connected, room, false
				case roomv1.ErrRoomFull:
					connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
						Code:    http.StatusBadRequest,
						Message: v.Message,
					})
					return connected, room, false
				default:
					connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
						Code:    http.StatusInternalServerError,
						Message: fmt.Sprintf("Failed to register existing client to room, %v", err),
					})
					return connected, room, false
				}
			}

			responseClient := &clientv1.Client{
				ID:     connected.Client.ID,
				Secret: connected.Client.Secret,
			}

			responseData, err := proto.Marshal(responseClient)
			if err != nil {
				// Should not occur, panic
				panic(err)
			}

			connected.Write <- apiv1.WebSocketSucceed(&transportv1.Payload{
				Flag: transportv1.Payload_RESPONSE_CONNECT,
				Data: responseData,
			})

			p.setHostIfNone(connected, matchRoom)

			return connected, matchRoom, false
		}
	}

	connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
		Code:    http.StatusBadRequest,
		Message: fmt.Sprintf("No valid room match found for ID %d", rejoinRequest.RoomID),
	})
	return connected, room, false
}

func (p *StandardProtocol) Disconnect(connected *sessionv1.Session, room roomv1.Room) bool {
	connected.Close()
	if connected.Client == nil || room == nil || room.GetStatus() == roomv1.RoomStatus_CLOSING {
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

func (p *StandardProtocol) List(payload *transportv1.Payload, connected *sessionv1.Session, room roomv1.Room) bool {
	if connected == nil || room == nil {
		connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
			Code:    http.StatusBadRequest,
			Message: "Must be connected to a room to list a room's clients",
		})
		return false
	}

	connectedClients, err := room.GetConnected()
	if err != nil {
		connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("Failed to retrieve room's connected clients, %v", err),
		})
		return false
	}

	list := make([]*clientv1.SanitisedClient, 0)
	for i := 0; i < len(connectedClients); i++ {
		connectedClient := connectedClients[i]
		host, err := room.IsHost(connectedClient.Client)
		if err != nil {
			connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
				Code:    http.StatusInternalServerError,
				Message: fmt.Sprintf("Failed to determine if client is host, %v", err),
			})
			return false
		}
		list = append(list, &clientv1.SanitisedClient{
			ID:   connectedClient.Client.ID,
			Host: host,
		})
	}

	responseData, err := proto.Marshal(&clientv1.ClientList{
		List: list,
	})
	if err != nil {
		// Should not occur, panic
		panic(err)
	}

	connected.Write <- apiv1.WebSocketSucceed(&transportv1.Payload{
		Flag: transportv1.Payload_RESPONSE_LIST,
		Data: responseData,
	})
	return false
}

func (p *StandardProtocol) RelayMessage(payload *transportv1.Payload, connected *sessionv1.Session, room roomv1.Room) bool {
	if connected == nil || room == nil {
		connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
			Code:    http.StatusBadRequest,
			Message: "Must be connected to a room to relay a message",
		})
		return false
	}

	relayMsg := &relayv1.Relay{}
	err := proto.Unmarshal(payload.Data, relayMsg)
	if err != nil {
		connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Relayed message does not conform to spec, %v", err),
		})
		return false
	}

	connectedClientList, err := room.GetConnected()
	if err != nil {
		connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("Failed to retrieve room's connected clients, %v", err),
		})
		return false
	}

	isHost, err := room.IsHost(connected.Client)
	if err != nil {
		connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("Failed to determine if client is host, %v", err),
		})
		return false
	}

	switch relayMsg.Type {
	case relayv1.Relay_BROADCAST:
		if !isHost {
			connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
				Code:    http.StatusBadRequest,
				Message: "Must be host to broadcast",
			})
			return false
		}
		p.broadcast(payload, connected, room, connectedClientList)
		return false
	case relayv1.Relay_TARGET:
		if !isHost {
			connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
				Code:    http.StatusBadRequest,
				Message: "Must be host to send targeted messages",
			})
			return false
		}

		if relayMsg.Target == nil {
			connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
				Code:    http.StatusBadRequest,
				Message: "Must provide a target ID to send a message to",
			})
			return false
		}

		for _, connectedClient := range connectedClientList {
			if *relayMsg.Target != connectedClient.Client.ID {
				continue
			}
			connectedClient.Write <- apiv1.WebSocketSucceed(&transportv1.Payload{
				Flag: transportv1.Payload_RESPONSE_RELAY_MESSAGE,
				Data: payload.Data,
			})
			return false
		}
		connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("No target client found with ID %d", *relayMsg.Target),
		})
		return false
	case relayv1.Relay_HOST:
		if isHost {
			connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
				Code:    http.StatusBadRequest,
				Message: "Hosts cannot send messages to themselves",
			})
			return false
		}

		host, err := room.GetHost()
		if err != nil {
			connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
				Code:    http.StatusInternalServerError,
				Message: fmt.Sprintf("Failed to get host, %v", err),
			})
			return false
		}

		host.Write <- apiv1.WebSocketSucceed(&transportv1.Payload{
			Flag: transportv1.Payload_RESPONSE_RELAY_MESSAGE,
			Data: payload.Data,
		})
	}
	return false
}

func (p *StandardProtocol) GrantHost(payload *transportv1.Payload, connected *sessionv1.Session, room roomv1.Room) bool {
	if connected == nil || room == nil {
		connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
			Code:    http.StatusBadRequest,
			Message: "Must be connected to a room to grant another client host",
		})
		return false
	}

	grantHostRequest := &roomspecv1.GrantHostRequest{}
	err := proto.Unmarshal(payload.Data, grantHostRequest)
	if err != nil {
		connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Invalid grant host request provided, does not conform to spec, %v", err),
		})
		return false
	}

	isHost, err := room.IsHost(connected.Client)
	if err != nil {
		connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("Failed to determine if client is host, %v", err),
		})
		return false
	}

	if !isHost {
		connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
			Code:    http.StatusBadRequest,
			Message: "Must be host to grant host to another host",
		})
		return false
	}

	if grantHostRequest.HostID == connected.Client.ID {
		connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
			Code:    http.StatusBadRequest,
			Message: "Cannot transfer host powers to yourself",
		})
		return false
	}

	host, err := room.GetClient(grantHostRequest.HostID)
	if err != nil {
		switch v := err.(type) {
		case roomv1.ErrNoMatchingClient:
			connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
				Code:    http.StatusBadRequest,
				Message: v.Message,
			})
			return false
		default:
			connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
				Code:    http.StatusInternalServerError,
				Message: fmt.Sprintf("Failed to get client with ID %d, %v", grantHostRequest.HostID, err),
			})
			return false
		}
	}

	err = p.changeHost(room, host)
	if err != nil {
		connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("Failed to change host, %v", err),
		})
		return false
	}

	return false
}

func (p *StandardProtocol) Kick(payload *transportv1.Payload, connected *sessionv1.Session, room roomv1.Room) bool {
	if connected == nil || room == nil {
		connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
			Code:    http.StatusBadRequest,
			Message: "Must be connected to a room to kick a client",
		})
		return false
	}

	kickRequest := &roomspecv1.KickRequest{}
	err := proto.Unmarshal(payload.Data, kickRequest)
	if err != nil {
		connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Invalid kick request provided, does not conform to spec, %v", err),
		})
		return false
	}

	isHost, err := room.IsHost(connected.Client)
	if err != nil {
		connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("Failed to determine if client is host, %v", err),
		})
		return false
	}

	if !isHost {
		connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
			Code:    http.StatusBadRequest,
			Message: "Must be host to kick",
		})
		return false
	}

	if kickRequest.ClientID == connected.Client.ID {
		connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
			Code:    http.StatusBadRequest,
			Message: "Cannot kick yourself",
		})
		return false
	}

	kickedClient, err := room.GetClient(kickRequest.ClientID)
	if err != nil {
		switch v := err.(type) {
		case roomv1.ErrNoMatchingClient:
			connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
				Code:    http.StatusBadRequest,
				Message: v.Message,
			})
			return false
		default:
			connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
				Code:    http.StatusInternalServerError,
				Message: fmt.Sprintf("Failed to kick client with ID %d, %v", kickRequest.ClientID, err),
			})
			return false
		}
	}

	p.Disconnect(kickedClient, room)

	kickData, err := proto.Marshal(&roomspecv1.KickResponse{
		ClientID: kickRequest.ClientID,
	})
	if err != nil {
		// Should not occur, panic
		panic(err)
	}

	connected.Write <- apiv1.WebSocketSucceed(&transportv1.Payload{
		Flag: transportv1.Payload_RESPONSE_KICK,
		Data: kickData,
	})
	return false
}

func (p *StandardProtocol) CloseRoom(roomID int32) error {
	retrievedRoom, err := p.RoomManager.GetRoom(roomID)
	if err != nil {
		return err
	}

	retrievedRoom.SetStatus(roomv1.RoomStatus_CLOSING)

	connectedClientList, err := retrievedRoom.GetConnected()
	if err != nil {
		return err
	}

	for _, connectedClient := range connectedClientList {
		p.Disconnect(connectedClient, retrievedRoom)
	}

	return p.RoomManager.DeleteRoom(roomID)
}

func (p *StandardProtocol) setHostIfNone(connected *sessionv1.Session, room roomv1.Room) {
	host, err := room.GetHost()
	if err != nil {
		connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("Failed to retrieve the current host, %v", err),
		})
	} else {
		if host == nil {
			_, err := room.SetHost(&connected.Client.ID)
			if err != nil {
				connected.Write <- apiv1.WebSocketFail(&transportv1.Error{
					Code:    http.StatusInternalServerError,
					Message: fmt.Sprintf("Failed to update host, %v", err),
				})
			} else {
				connected.Write <- apiv1.WebSocketSucceed(&transportv1.Payload{
					Flag: transportv1.Payload_RESPONSE_ASSIGN_HOST,
				})
			}
		}
	}
}

func (p *StandardProtocol) broadcast(payload *transportv1.Payload, connected *sessionv1.Session, room roomv1.Room, connectedClientList []*sessionv1.Session) {
	for _, connectedClient := range connectedClientList {
		if connected.Client.ID == connectedClient.Client.ID {
			// Message should only be sent to other clients, not sent back to origin
			continue
		}
		connectedClient.Write <- apiv1.WebSocketSucceed(&transportv1.Payload{
			Flag: transportv1.Payload_RESPONSE_RELAY_MESSAGE,
			Data: payload.Data,
		})
	}
}

func (p *StandardProtocol) migrateHost(room roomv1.Room) error {
	connectedClients, err := room.GetConnected()
	if err != nil {
		return err
	}

	if len(connectedClients) <= 0 {
		_, err = room.SetHost(nil)
		return err
	}

	newHost := connectedClients[0]

	return p.changeHost(room, newHost)
}

func (p *StandardProtocol) changeHost(room roomv1.Room, host *sessionv1.Session) error {
	connectedClients, err := room.GetConnected()
	if err != nil {
		return err
	}

	for _, connectedClient := range connectedClients {
		connectedClient.Write <- apiv1.WebSocketSucceed(&transportv1.Payload{
			Flag: transportv1.Payload_RESPONSE_BEGIN_HOST_MIGRATE,
		})
	}

	_, err = room.SetHost(&host.Client.ID)
	if err != nil {
		return err
	}

	finishHostMigrationResponse := &roomspecv1.FinishHostMigrationResponse{
		HostID: host.Client.ID,
	}

	finishMigrationBytes, err := proto.Marshal(finishHostMigrationResponse)
	if err != nil {
		// Should not occur, panic
		panic(err)
	}

	host.Write <- apiv1.WebSocketSucceed(&transportv1.Payload{
		Flag: transportv1.Payload_RESPONSE_ASSIGN_HOST,
		Data: finishMigrationBytes,
	})

	for _, connectedClient := range connectedClients {
		connectedClient.Write <- apiv1.WebSocketSucceed(&transportv1.Payload{
			Flag: transportv1.Payload_RESPONSE_FINISH_HOST_MIGRATE,
		})
	}

	return nil
}
