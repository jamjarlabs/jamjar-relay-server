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
	"github.com/jamjarlabs/jamjar-relay-server/internal/v1/room"
	"github.com/jamjarlabs/jamjar-relay-server/internal/v1/session"
	"github.com/jamjarlabs/jamjar-relay-server/specs/v1/transport"
)

// Protocol defines the contract that a v1 protocol should fufil, and the actions possible
type Protocol interface {
	// Definitions for client interactions, the boolean response determines if the connection should be stopped,
	// true = break connection, false = keep connection

	// Connect defines a client connecting to a room
	Connect(payload *transport.Payload, connected *session.Session, currentRoom room.Room) (*session.Session, room.Room, bool)
	// Reconnect defines a client reconnecting to a room
	Reconnect(payload *transport.Payload, connected *session.Session, currentRoom room.Room) (*session.Session, room.Room, bool)
	// Disconnect defines a client disconnecting from a room and closing the connection
	Disconnect(connected *session.Session, room room.Room) bool
	// List defines a client requesting a list of all clients connected to a room
	List(payload *transport.Payload, connected *session.Session, room room.Room) bool
	// RelayMessage defines a client sending a message to the room
	RelayMessage(payload *transport.Payload, connected *session.Session, room room.Room) bool
	// GrantHost defines a client transferring the room's host powers to another client
	GrantHost(payload *transport.Payload, connected *session.Session, room room.Room) bool
	// Kick defines a client removing another client from the room
	Kick(payload *transport.Payload, connected *session.Session, room room.Room) bool

	// CloseRoom is a server based control for closing a room and disconnecting all clients
	CloseRoom(roomID int32) error
}
