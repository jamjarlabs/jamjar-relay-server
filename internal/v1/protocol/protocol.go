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

type Protocol interface {
	Connect(payload *transport.Payload, connected *session.Session, currentRoom room.Room) (*session.Session, room.Room, bool)
	Reconnect(payload *transport.Payload, connected *session.Session, currentRoom room.Room) (*session.Session, room.Room, bool)
	Disconnect(connected *session.Session, room room.Room) bool
	List(payload *transport.Payload, connected *session.Session, room room.Room) bool
	RelayMessage(payload *transport.Payload, connected *session.Session, room room.Room) bool
}
