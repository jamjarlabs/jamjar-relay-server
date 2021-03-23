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

package room

import (
	"github.com/jamjarlabs/jamjar-relay-server/internal/v1/session"
	"github.com/jamjarlabs/jamjar-relay-server/specs/v1/client"
)

type RoomFactory func(id int32, secret int32, maxClients int32) (Room, error)

type Room interface {
	RoomMatches(id int32, secret int32) bool

	NewClient(session *session.Session) (*session.Session, error)
	ExistingClient(session *session.Session, clientID int32, clientSecret int32) (*session.Session, error)

	GetClient(clientID int32) (*session.Session, error)
	RemoveClient(clientID int32) error

	GetConnected() ([]*session.Session, error)

	IsHost(potentialHost *client.Client) (bool, error)
	SetHost(hostID *int32) (*session.Session, error)
	GetHost() (*session.Session, error)
	GetInfo() (*RoomInfo, error)

	SetStatus(RoomStatus)
	GetStatus() RoomStatus
}

type RoomStatus int32

const (
	RoomStatus_RUNNING RoomStatus = iota
	RoomStatus_CLOSING RoomStatus = iota
)

type RoomInfo struct {
	ID             int32 `json:"id"`
	Secret         int32 `json:"secret"`
	MaxClients     int32 `json:"max_clients"`
	CurrentClients int32 `json:"current_clients"`
}

type RoomsSummary struct {
	NumberOfRooms    int32 `json:"number_of_rooms"`
	MaxClients       int32 `json:"max_clients"`
	CurrentClients   int32 `json:"current_clients"`
	CommittedClients int32 `json:"committed_clients"`
}

type RoomManager interface {
	GetRoom(id int32) (Room, error)
	DeleteRoom(id int32) error
	CreateRoom(maxClients int32) (Room, error)

	ListRooms() ([]Room, error)

	Summary() (*RoomsSummary, error)
}
