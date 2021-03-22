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

package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jamjarlabs/jamjar-relay-server/specs/v1/client"
	"github.com/jamjarlabs/jamjar-relay-server/specs/v1/relay"
	"github.com/jamjarlabs/jamjar-relay-server/specs/v1/transport"
	"google.golang.org/protobuf/proto"

	roomspec "github.com/jamjarlabs/jamjar-relay-server/specs/v1/room"
)

const (
	actionConnect     = "c"
	actionDisconnect  = "d"
	actionReconnect   = "r"
	actionSendMessage = "s"
	actionListClients = "l"
)

func main() {
	// Get the first arg as the address of the relay server
	relayAddress := os.Args[1]

	// Connect to relay server

	c, _, err := websocket.DefaultDialer.Dial(relayAddress, nil)

	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	defer c.Close()

	done := make(chan struct{})

	// Set up listening loop
	go func() {
		defer close(done)
		for {
			_, inputBytes, err := c.ReadMessage()
			if err != nil {
				log.Fatalf("read: %v", err)
				return
			}

			payload := &transport.Payload{}

			proto.Unmarshal(inputBytes, payload)

			switch payload.Flag {
			case transport.Payload_RESPONSE_ERROR:
				networkErr := &transport.Error{}

				proto.Unmarshal(payload.Data, networkErr)

				fmt.Printf("\nCode: %d, Message: %s\n", networkErr.Code, networkErr.Message)
			case transport.Payload_RESPONSE_CONNECT:
				clientInfo := &client.Client{}

				proto.Unmarshal(payload.Data, clientInfo)

				fmt.Printf("\nID: %d, Secret: %d\n", clientInfo.ID, clientInfo.Secret)
			case transport.Payload_RESPONSE_LIST:
				clientList := &client.ClientList{}

				proto.Unmarshal(payload.Data, clientList)

				for i, client := range clientList.List {
					if client.Host {
						fmt.Printf("[%d] ID: %d (host)\n", i, client.ID)
					} else {
						fmt.Printf("[%d] ID: %d\n", i, client.ID)
					}
				}
			case transport.Payload_RESPONSE_RELAY_MESSAGE:
				fmt.Printf("\nMessage received: %s\n", string(payload.Data))
			case transport.Payload_RESPONSE_ASSIGN_HOST:
				fmt.Println("\nAssigned Host")
			}
		}
	}()

	// Begin loop action
	for {
		fmt.Println("=========================")
		fmt.Printf("%s - Connect\n", actionConnect)
		fmt.Printf("%s - Disconnect\n", actionDisconnect)
		fmt.Printf("%s - Reconnect\n", actionReconnect)
		fmt.Printf("%s - Send message\n", actionSendMessage)
		fmt.Printf("%s - List clients\n", actionListClients)
		fmt.Printf("Action: ")
		var action string
		fmt.Scanln(&action)

		switch action {
		case actionConnect:
			fmt.Printf("ID of the room to connect to: ")
			var idStr string
			fmt.Scanln(&idStr)

			fmt.Printf("Secret of the room to connect to: ")
			var secretStr string
			fmt.Scanln(&secretStr)

			id, err := strconv.ParseInt(idStr, 10, 32)
			if err != nil {
				fmt.Printf("Invalid room ID, %v", err)
				break
			}

			secret, err := strconv.ParseInt(secretStr, 10, 32)
			if err != nil {
				fmt.Printf("Invalid secret, %v", err)
				break
			}

			joinRequest := &roomspec.JoinRoomRequest{
				RoomID:     int32(id),
				RoomSecret: int32(secret),
			}

			joinReq, _ := proto.Marshal(joinRequest)

			payload := &transport.Payload{
				Flag: transport.Payload_REQUEST_CONNECT,
				Data: joinReq,
			}

			payloadByes, _ := proto.Marshal(payload)

			c.WriteMessage(websocket.BinaryMessage, payloadByes)
		case actionReconnect:
			fmt.Printf("ID of the room to connect to: ")
			var idStr string
			fmt.Scanln(&idStr)

			fmt.Printf("Secret of the room to connect to: ")
			var secretStr string
			fmt.Scanln(&secretStr)

			fmt.Printf("Reconnecting client ID: ")
			var clientIDStr string
			fmt.Scanln(&clientIDStr)

			fmt.Printf("Reconnecting client secret: ")
			var clientSecretStr string
			fmt.Scanln(&clientSecretStr)

			id, err := strconv.ParseInt(idStr, 10, 32)
			if err != nil {
				fmt.Printf("Invalid room ID, %v", err)
				break
			}

			secret, err := strconv.ParseInt(secretStr, 10, 32)
			if err != nil {
				fmt.Printf("Invalid room secret, %v", err)
				break
			}

			clientID, err := strconv.ParseInt(clientIDStr, 10, 32)
			if err != nil {
				fmt.Printf("Invalid client ID, %v", err)
				break
			}

			clientSecret, err := strconv.ParseInt(clientSecretStr, 10, 32)
			if err != nil {
				fmt.Printf("Invalid client secret, %v", err)
				break
			}

			rejoinRequest := &roomspec.RejoinRoomRequest{
				RoomID:       int32(id),
				RoomSecret:   int32(secret),
				ClientID:     int32(clientID),
				ClientSecret: int32(clientSecret),
			}

			joinReq, _ := proto.Marshal(rejoinRequest)

			payload := &transport.Payload{
				Flag: transport.Payload_REQUEST_RECONNECT,
				Data: joinReq,
			}

			payloadByes, _ := proto.Marshal(payload)

			c.WriteMessage(websocket.BinaryMessage, payloadByes)
		case actionDisconnect:
			c.WriteMessage(websocket.CloseMessage, []byte{})
			<-done
			break
		case actionListClients:
			payload := &transport.Payload{
				Flag: transport.Payload_REQUEST_LIST,
			}

			payloadBytes, _ := proto.Marshal(payload)

			c.WriteMessage(websocket.BinaryMessage, payloadBytes)
		case actionSendMessage:

			fmt.Printf("Message to send: ")
			var msg string
			fmt.Scanln(&msg)

			fmt.Printf("Broadcast (b), target (t), or host (h): ")
			var messageType string
			fmt.Scanln(&messageType)

			payload := &transport.Payload{
				Flag: transport.Payload_REQUEST_RELAY_MESSAGE,
			}

			relayMessage := &relay.Relay{
				Target: nil,
				Data:   []byte(msg),
			}

			switch messageType {
			case "b":
				relayMessage.Type = relay.Relay_BROADCAST
				break
			case "t":
				fmt.Printf("ID of target: ")
				var targetIDStr string
				fmt.Scanln(&targetIDStr)

				targetID, err := strconv.ParseInt(targetIDStr, 10, 32)
				if err != nil {
					fmt.Printf("Invalid target ID, %v", err)
					break
				}

				targetID32 := int32(targetID)

				relayMessage.Target = &targetID32
				relayMessage.Type = relay.Relay_TARGET
			case "h":
				relayMessage.Type = relay.Relay_HOST
			default:
				fmt.Printf("Unknown message type, '%s'", messageType)
				break
			}

			relayBytes, _ := proto.Marshal(relayMessage)

			payload.Data = relayBytes

			payloadBytes, _ := proto.Marshal(payload)

			c.WriteMessage(websocket.BinaryMessage, payloadBytes)
		default:
			fmt.Printf("Unknown message type, '%s'", action)
		}
		// Wait for 1 sec after to receive any messages
		time.Sleep(1 * time.Second)
	}

	// Allow different actions

	// Connect
	// Disconnect
	// Reconnect
	// Send message
}
