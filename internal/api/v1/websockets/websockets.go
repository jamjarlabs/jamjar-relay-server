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

package websockets

import (
	"fmt"
	"net/http"

	"github.com/golang/glog"
	"github.com/gorilla/websocket"
	"github.com/jamjarlabs/jamjar-relay-server/internal/api/v1/api"
	"github.com/jamjarlabs/jamjar-relay-server/internal/v1/protocol"
	"github.com/jamjarlabs/jamjar-relay-server/internal/v1/room"
	"github.com/jamjarlabs/jamjar-relay-server/internal/v1/session"
	"github.com/jamjarlabs/jamjar-relay-server/specs/v1/transport"
	"google.golang.org/protobuf/proto"
)

type Handle struct {
	Protocol protocol.Protocol
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *Handle) Websocket(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		glog.Errorf("upgrade: %v", err)
		return
	}
	defer c.Close()

	connectedClient := &session.Session{
		Write:       make(chan []byte),
		CloseSignal: make(chan struct{}),
		Closed:      false,
	}

	go func() {
		<-connectedClient.CloseSignal
		c.Close()
	}()

	// Set up write loop
	go func() {
		for {
			select {
			case msg := <-connectedClient.Write:
				if connectedClient.Closed {
					return
				}
				err := c.WriteMessage(websocket.BinaryMessage, msg)
				if err != nil {
					if connectedClient.Client == nil {
						glog.Errorf("failed to write message to client, %v", err)
					} else {
						glog.Errorf("failed to write message to client with ID %d, %v", connectedClient.Client.ID, err)
					}
				}
			}
		}
	}()

	var room room.Room

	// Set up listen loop
	for {
		if connectedClient.Closed {
			return
		}

		mt, messageData, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				h.Protocol.Disconnect(connectedClient, room)
				return
			}
			if connectedClient.Client == nil {
				glog.Errorf("failed to read message from client, %v", err)
			} else {
				glog.Errorf("failed to read message from client with ID %d, %v", connectedClient.Client.ID, err)
			}
			continue
		}

		exit := false

		switch mt {
		case websocket.BinaryMessage:
			payload := &transport.Payload{}

			err = proto.Unmarshal(messageData, payload)
			if err != nil {
				glog.Error(err)
				connectedClient.Write <- api.WebSocketFail(&transport.Error{
					Code:    http.StatusBadRequest,
					Message: fmt.Sprintf("Invalid message provided, does not conform to spec, %v", err),
				})
				break
			}
			switch payload.Flag {
			case transport.Payload_REQUEST_CONNECT:
				connectedClient, room, exit = h.Protocol.Connect(payload, connectedClient, room)
			case transport.Payload_REQUEST_RECONNECT:
				connectedClient, room, exit = h.Protocol.Reconnect(payload, connectedClient, room)
			case transport.Payload_REQUEST_LIST:
				exit = h.Protocol.List(payload, connectedClient, room)
			case transport.Payload_REQUEST_RELAY_MESSAGE:
				exit = h.Protocol.RelayMessage(payload, connectedClient, room)
			case transport.Payload_REQUEST_GRANT_HOST:
				exit = h.Protocol.GrantHost(payload, connectedClient, room)
			case transport.Payload_REQUEST_KICK:
				exit = h.Protocol.Kick(payload, connectedClient, room)
			}
		case websocket.CloseMessage:
			exit = h.Protocol.Disconnect(connectedClient, room)
		default:
			connectedClient.Write <- api.WebSocketFail(&transport.Error{
				Code:    http.StatusBadRequest,
				Message: "Invalid message provided, must be in binary format",
			})
		}

		if exit {
			return
		}
	}
}
