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
	roomv1 "github.com/jamjarlabs/jamjar-relay-server/internal/v1/room"
	sessionv1 "github.com/jamjarlabs/jamjar-relay-server/internal/v1/session"
	"github.com/jamjarlabs/jamjar-relay-server/specs/v1/api"
	clientv1 "github.com/jamjarlabs/jamjar-relay-server/specs/v1/client"
	relayv1 "github.com/jamjarlabs/jamjar-relay-server/specs/v1/relay"
	roomspecv1 "github.com/jamjarlabs/jamjar-relay-server/specs/v1/room"
	transportv1 "github.com/jamjarlabs/jamjar-relay-server/specs/v1/transport"

	"google.golang.org/protobuf/proto"
)

// StandardProtocol is the standard implementation of the v1 relay protocol
type StandardProtocol struct {
	RoomManager roomv1.Manager
}

// Connect handles a new client connecting to a room
func (p *StandardProtocol) Connect(payload *transportv1.Payload, connected *sessionv1.Session, currentRoom roomv1.Room) (*sessionv1.Session, roomv1.Room) {
	if currentRoom != nil {
		connected.Write <- Fail(&transportv1.Error{
			Code:    http.StatusBadRequest,
			Message: "Cannot connect to a different room while already connected to another",
		})
		return connected, currentRoom
	}

	joinRequest := &roomspecv1.JoinRoomRequest{}
	err := proto.Unmarshal(payload.Data, joinRequest)
	if err != nil {
		connected.Write <- Fail(&transportv1.Error{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Invalid join request provided, does not conform to spec, %v", err),
		})
		return connected, currentRoom
	}

	rooms, err := p.RoomManager.ListRooms()
	if err != nil {
		connected.Write <- Fail(&transportv1.Error{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("Failed to retrieve room list, %v", err),
		})
		return connected, currentRoom
	}

	for _, matchRoom := range rooms {
		if !matchRoom.RoomMatches(joinRequest.RoomID, joinRequest.RoomSecret) {
			continue
		}
		connected, err = matchRoom.NewClient(connected)
		if err != nil {
			switch v := err.(type) {
			case roomv1.ErrRoomFull:
				connected.Write <- Fail(&transportv1.Error{
					Code:    http.StatusBadRequest,
					Message: v.Message,
				})
				return connected, currentRoom
			default:
				connected.Write <- Fail(&transportv1.Error{
					Code:    http.StatusInternalServerError,
					Message: fmt.Sprintf("Failed to register new client to room, %v", err),
				})
				return connected, currentRoom
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

		connected.Write <- Succeed(&transportv1.Payload{
			Flag: transportv1.Payload_RESPONSE_CONNECT,
			Data: responseData,
		})

		p.setHostIfNone(connected, matchRoom)

		p.sendClientConnectToHost(connected, matchRoom)

		return connected, matchRoom
	}

	connected.Write <- Fail(&transportv1.Error{
		Code:    http.StatusBadRequest,
		Message: fmt.Sprintf("No valid room match found for ID %d", joinRequest.RoomID),
	})

	return connected, currentRoom
}

// Reconnect handles an existing client reconnecting to a room
func (p *StandardProtocol) Reconnect(payload *transportv1.Payload, connected *sessionv1.Session, room roomv1.Room) (*sessionv1.Session, roomv1.Room) {
	if room != nil {
		connected.Write <- Fail(&transportv1.Error{
			Code:    http.StatusBadRequest,
			Message: "Cannot connect to a different room while already connected to another",
		})
		return connected, room
	}

	rejoinRequest := &roomspecv1.RejoinRoomRequest{}
	err := proto.Unmarshal(payload.Data, rejoinRequest)
	if err != nil {
		connected.Write <- Fail(&transportv1.Error{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Invalid join request provided, does not conform to spec, %v", err),
		})
		return connected, room
	}

	rooms, err := p.RoomManager.ListRooms()
	if err != nil {
		connected.Write <- Fail(&transportv1.Error{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("Failed to retrieve room list, %v", err),
		})
		return connected, room
	}

	for _, matchRoom := range rooms {
		if matchRoom.RoomMatches(rejoinRequest.RoomID, rejoinRequest.RoomSecret) {
			connected, err = matchRoom.ExistingClient(connected, rejoinRequest.ClientID, rejoinRequest.ClientSecret)
			if err != nil {
				switch v := err.(type) {
				case roomv1.ErrInvalidSecret:
					connected.Write <- Fail(&transportv1.Error{
						Code:    http.StatusBadRequest,
						Message: v.Message,
					})
					return connected, room
				case roomv1.ErrRoomFull:
					connected.Write <- Fail(&transportv1.Error{
						Code:    http.StatusBadRequest,
						Message: v.Message,
					})
					return connected, room
				default:
					connected.Write <- Fail(&transportv1.Error{
						Code:    http.StatusInternalServerError,
						Message: fmt.Sprintf("Failed to register existing client to room, %v", err),
					})
					return connected, room
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

			connected.Write <- Succeed(&transportv1.Payload{
				Flag: transportv1.Payload_RESPONSE_CONNECT,
				Data: responseData,
			})

			p.setHostIfNone(connected, matchRoom)

			p.sendClientConnectToHost(connected, matchRoom)

			return connected, matchRoom
		}
	}

	connected.Write <- Fail(&transportv1.Error{
		Code:    http.StatusBadRequest,
		Message: fmt.Sprintf("No valid room match found for ID %d", rejoinRequest.RoomID),
	})
	return connected, room
}

// Disconnect handles a client disconnecting from a room and closing the connection
func (p *StandardProtocol) Disconnect(connected *sessionv1.Session, room roomv1.Room) {
	connected.Close()
	if connected.Client == nil || room == nil || room.GetStatus() == roomv1.StatusClosing {
		return
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
	return
}

// List handles a client requesting a list of all clients connected to a room
func (p *StandardProtocol) List(payload *transportv1.Payload, connected *sessionv1.Session, room roomv1.Room) {
	if connected == nil || room == nil {
		connected.Write <- Fail(&transportv1.Error{
			Code:    http.StatusBadRequest,
			Message: "Must be connected to a room to list a room's clients",
		})
		return
	}

	connectedClients, err := room.GetConnected()
	if err != nil {
		connected.Write <- Fail(&transportv1.Error{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("Failed to retrieve room's connected clients, %v", err),
		})
		return
	}

	list := make([]*clientv1.SanitisedClient, 0)
	for i := 0; i < len(connectedClients); i++ {
		connectedClient := connectedClients[i]
		host, err := room.IsHost(connectedClient.Client)
		if err != nil {
			connected.Write <- Fail(&transportv1.Error{
				Code:    http.StatusInternalServerError,
				Message: fmt.Sprintf("Failed to determine if client is host, %v", err),
			})
			return
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

	connected.Write <- Succeed(&transportv1.Payload{
		Flag: transportv1.Payload_RESPONSE_LIST,
		Data: responseData,
	})
	return
}

// RelayMessage handles a client sending a message to the room
func (p *StandardProtocol) RelayMessage(payload *transportv1.Payload, connected *sessionv1.Session, room roomv1.Room) {
	if connected == nil || room == nil {
		connected.Write <- Fail(&transportv1.Error{
			Code:    http.StatusBadRequest,
			Message: "Must be connected to a room to relay a message",
		})
		return
	}

	relayMsg := &relayv1.Relay{}
	err := proto.Unmarshal(payload.Data, relayMsg)
	if err != nil {
		connected.Write <- Fail(&transportv1.Error{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Relayed message does not conform to spec, %v", err),
		})
		return
	}

	connectedClientList, err := room.GetConnected()
	if err != nil {
		connected.Write <- Fail(&transportv1.Error{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("Failed to retrieve room's connected clients, %v", err),
		})
		return
	}

	isHost, err := room.IsHost(connected.Client)
	if err != nil {
		connected.Write <- Fail(&transportv1.Error{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("Failed to determine if client is host, %v", err),
		})
		return
	}

	switch relayMsg.Type {
	case relayv1.Relay_BROADCAST:
		if !isHost {
			connected.Write <- Fail(&transportv1.Error{
				Code:    http.StatusBadRequest,
				Message: "Must be host to broadcast",
			})
			return
		}
		p.broadcast(payload, connected, room, connectedClientList)
		return
	case relayv1.Relay_TARGET:
		if !isHost {
			connected.Write <- Fail(&transportv1.Error{
				Code:    http.StatusBadRequest,
				Message: "Must be host to send targeted messages",
			})
			return
		}

		if relayMsg.Target == nil {
			connected.Write <- Fail(&transportv1.Error{
				Code:    http.StatusBadRequest,
				Message: "Must provide a target ID to send a message to",
			})
			return
		}

		for _, connectedClient := range connectedClientList {
			if *relayMsg.Target != connectedClient.Client.ID {
				continue
			}
			connectedClient.Write <- Succeed(&transportv1.Payload{
				Flag: transportv1.Payload_RESPONSE_RELAY_MESSAGE,
				Data: payload.Data,
			})
			return
		}
		connected.Write <- Fail(&transportv1.Error{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("No target client found with ID %d", *relayMsg.Target),
		})
		return
	case relayv1.Relay_HOST:
		if isHost {
			connected.Write <- Fail(&transportv1.Error{
				Code:    http.StatusBadRequest,
				Message: "Hosts cannot send messages to themselves",
			})
			return
		}

		host, err := room.GetHost()
		if err != nil {
			connected.Write <- Fail(&transportv1.Error{
				Code:    http.StatusInternalServerError,
				Message: fmt.Sprintf("Failed to get host, %v", err),
			})
			return
		}

		host.Write <- Succeed(&transportv1.Payload{
			Flag: transportv1.Payload_RESPONSE_RELAY_MESSAGE,
			Data: payload.Data,
		})
	}
	return
}

// GrantHost handles a client transferring the room's host powers to another client
func (p *StandardProtocol) GrantHost(payload *transportv1.Payload, connected *sessionv1.Session, room roomv1.Room) {
	if connected == nil || room == nil {
		connected.Write <- Fail(&transportv1.Error{
			Code:    http.StatusBadRequest,
			Message: "Must be connected to a room to grant another client host",
		})
		return
	}

	grantHostRequest := &roomspecv1.GrantHostRequest{}
	err := proto.Unmarshal(payload.Data, grantHostRequest)
	if err != nil {
		connected.Write <- Fail(&transportv1.Error{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Invalid grant host request provided, does not conform to spec, %v", err),
		})
		return
	}

	isHost, err := room.IsHost(connected.Client)
	if err != nil {
		connected.Write <- Fail(&transportv1.Error{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("Failed to determine if client is host, %v", err),
		})
		return
	}

	if !isHost {
		connected.Write <- Fail(&transportv1.Error{
			Code:    http.StatusBadRequest,
			Message: "Must be host to grant host to another host",
		})
		return
	}

	if grantHostRequest.HostID == connected.Client.ID {
		connected.Write <- Fail(&transportv1.Error{
			Code:    http.StatusBadRequest,
			Message: "Cannot transfer host powers to yourself",
		})
		return
	}

	host, err := room.GetClient(grantHostRequest.HostID)
	if err != nil {
		switch v := err.(type) {
		case roomv1.ErrNoMatchingClient:
			connected.Write <- Fail(&transportv1.Error{
				Code:    http.StatusBadRequest,
				Message: v.Message,
			})
			return
		default:
			connected.Write <- Fail(&transportv1.Error{
				Code:    http.StatusInternalServerError,
				Message: fmt.Sprintf("Failed to get client with ID %d, %v", grantHostRequest.HostID, err),
			})
			return
		}
	}

	err = p.changeHost(room, host)
	if err != nil {
		connected.Write <- Fail(&transportv1.Error{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("Failed to change host, %v", err),
		})
		return
	}

	return
}

// Kick handles a client removing another client from the room
func (p *StandardProtocol) Kick(payload *transportv1.Payload, connected *sessionv1.Session, room roomv1.Room) {
	if connected == nil || room == nil {
		connected.Write <- Fail(&transportv1.Error{
			Code:    http.StatusBadRequest,
			Message: "Must be connected to a room to kick a client",
		})
		return
	}

	kickRequest := &roomspecv1.KickRequest{}
	err := proto.Unmarshal(payload.Data, kickRequest)
	if err != nil {
		connected.Write <- Fail(&transportv1.Error{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Invalid kick request provided, does not conform to spec, %v", err),
		})
		return
	}

	isHost, err := room.IsHost(connected.Client)
	if err != nil {
		connected.Write <- Fail(&transportv1.Error{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("Failed to determine if client is host, %v", err),
		})
		return
	}

	if !isHost {
		connected.Write <- Fail(&transportv1.Error{
			Code:    http.StatusBadRequest,
			Message: "Must be host to kick",
		})
		return
	}

	if kickRequest.ClientID == connected.Client.ID {
		connected.Write <- Fail(&transportv1.Error{
			Code:    http.StatusBadRequest,
			Message: "Cannot kick yourself",
		})
		return
	}

	kickedClient, err := room.GetClient(kickRequest.ClientID)
	if err != nil {
		switch v := err.(type) {
		case roomv1.ErrNoMatchingClient:
			connected.Write <- Fail(&transportv1.Error{
				Code:    http.StatusBadRequest,
				Message: v.Message,
			})
			return
		default:
			connected.Write <- Fail(&transportv1.Error{
				Code:    http.StatusInternalServerError,
				Message: fmt.Sprintf("Failed to kick client with ID %d, %v", kickRequest.ClientID, err),
			})
			return
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

	connected.Write <- Succeed(&transportv1.Payload{
		Flag: transportv1.Payload_RESPONSE_KICK,
		Data: kickData,
	})
	return
}

// CloseRoom handles a room being closed and all clients disconnecting
func (p *StandardProtocol) CloseRoom(roomID int32) error {
	retrievedRoom, err := p.RoomManager.GetRoom(roomID)
	if err != nil {
		return err
	}

	retrievedRoom.SetStatus(roomv1.StatusClosing)

	connectedClientList, err := retrievedRoom.GetConnected()
	if err != nil {
		return err
	}

	for _, connectedClient := range connectedClientList {
		p.Disconnect(connectedClient, retrievedRoom)
	}

	return p.RoomManager.DeleteRoom(roomID)
}

// CreateRoom creates a new room
func (p *StandardProtocol) CreateRoom(maxClients int32) (roomv1.Room, error) {
	return p.RoomManager.CreateRoom(maxClients)
}

// GetRoom returns any matching room, if no room is found an error is returned
func (p *StandardProtocol) GetRoom(roomID int32) (roomv1.Room, error) {
	return p.RoomManager.GetRoom(roomID)
}

// Summary returns a summary of all rooms
func (p *StandardProtocol) Summary() (*api.RoomsSummary, error) {
	return p.RoomManager.Summary()
}

// ListRooms returns a list of all rooms
func (p *StandardProtocol) ListRooms() ([]roomv1.Room, error) {
	return p.RoomManager.ListRooms()
}

func (p *StandardProtocol) setHostIfNone(connected *sessionv1.Session, room roomv1.Room) {
	host, err := room.GetHost()
	if err != nil {
		connected.Write <- Fail(&transportv1.Error{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("Failed to retrieve the current host, %v", err),
		})
		return
	}

	if host == nil {
		_, err := room.SetHost(&connected.Client.ID)
		if err != nil {
			connected.Write <- Fail(&transportv1.Error{
				Code:    http.StatusInternalServerError,
				Message: fmt.Sprintf("Failed to update host, %v", err),
			})
		} else {
			connected.Write <- Succeed(&transportv1.Payload{
				Flag: transportv1.Payload_RESPONSE_ASSIGN_HOST,
			})
		}
	}
}

func (p *StandardProtocol) broadcast(payload *transportv1.Payload, connected *sessionv1.Session, room roomv1.Room, connectedClientList []*sessionv1.Session) {
	for _, connectedClient := range connectedClientList {
		if connected.Client.ID == connectedClient.Client.ID {
			// Message should only be sent to other clients, not sent back to origin
			continue
		}
		connectedClient.Write <- Succeed(&transportv1.Payload{
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
		connectedClient.Write <- Succeed(&transportv1.Payload{
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

	host.Write <- Succeed(&transportv1.Payload{
		Flag: transportv1.Payload_RESPONSE_ASSIGN_HOST,
		Data: finishMigrationBytes,
	})

	for _, connectedClient := range connectedClients {
		connectedClient.Write <- Succeed(&transportv1.Payload{
			Flag: transportv1.Payload_RESPONSE_FINISH_HOST_MIGRATE,
		})
	}

	return nil
}

func (p *StandardProtocol) sendClientConnectToHost(connected *sessionv1.Session, room roomv1.Room) {
	host, err := room.GetHost()
	if err != nil {
		connected.Write <- Fail(&transportv1.Error{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("Failed to retrieve the current host, %v", err),
		})
		return
	}

	if host == nil {
		return
	}

	responseData, err := proto.Marshal(&clientv1.SanitisedClient{
		ID:   connected.Client.ID,
		Host: connected.Client.ID == host.Client.ID,
	})
	if err != nil {
		// Should not occur, panic
		panic(err)
	}

	host.Write <- Succeed(&transportv1.Payload{
		Flag: transportv1.Payload_RESPONSE_CLIENT_CONNECT,
		Data: responseData,
	})
	return
}
