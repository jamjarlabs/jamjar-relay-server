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
	"bufio"
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
	actionKick        = "k"
	actionGrantHost   = "g"
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
			mt, inputBytes, err := c.ReadMessage()
			if err != nil {
				log.Fatalf("read: %v", err)
				return
			}

			switch mt {
			case websocket.BinaryMessage:
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
				case transport.Payload_RESPONSE_KICK:
					kickResponse := &roomspec.KickResponse{}

					proto.Unmarshal(payload.Data, kickResponse)
					fmt.Printf("\nKicked client with ID %d\n", kickResponse.ClientID)
				case transport.Payload_RESPONSE_BEGIN_HOST_MIGRATE:
					fmt.Println("\nHost migration begun")
				case transport.Payload_RESPONSE_FINISH_HOST_MIGRATE:
					fmt.Println("\nHost migration finished")
				}
			case websocket.CloseMessage:
				fmt.Println("Connection closed")
				return
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
		fmt.Printf("%s - Kick client\n", actionKick)
		fmt.Printf("%s - Grant client host\n", actionGrantHost)
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
			in := bufio.NewReader(os.Stdin)
			msg, _ := in.ReadString('\n')

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
		case actionKick:
			fmt.Printf("ID of client to kick: ")
			var clientIDStr string
			fmt.Scanln(&clientIDStr)

			clientID, err := strconv.ParseInt(clientIDStr, 10, 32)
			if err != nil {
				fmt.Printf("Invalid client ID, %v", err)
				break
			}

			clientID32 := int32(clientID)

			kickRequest := &roomspec.KickRequest{
				ClientID: clientID32,
			}

			kickReq, _ := proto.Marshal(kickRequest)

			payload := &transport.Payload{
				Flag: transport.Payload_REQUEST_KICK,
				Data: kickReq,
			}

			payloadByes, _ := proto.Marshal(payload)

			c.WriteMessage(websocket.BinaryMessage, payloadByes)
		case actionGrantHost:
			fmt.Printf("ID of client to grant host to: ")
			var clientIDStr string
			fmt.Scanln(&clientIDStr)

			clientID, err := strconv.ParseInt(clientIDStr, 10, 32)
			if err != nil {
				fmt.Printf("Invalid client ID, %v", err)
				break
			}

			clientID32 := int32(clientID)

			grantHostRequest := &roomspec.GrantHostRequest{
				HostID: clientID32,
			}

			grantHostReq, _ := proto.Marshal(grantHostRequest)

			payload := &transport.Payload{
				Flag: transport.Payload_REQUEST_GRANT_HOST,
				Data: grantHostReq,
			}

			payloadByes, _ := proto.Marshal(payload)

			c.WriteMessage(websocket.BinaryMessage, payloadByes)
		default:
			fmt.Printf("Unknown message type, '%s'", action)
		}
		// Wait for 1 sec after to receive any messages
		time.Sleep(1 * time.Second)
	}
}
