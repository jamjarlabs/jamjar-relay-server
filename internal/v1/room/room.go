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

type ErrRequestTooManyClients struct {
	Message string
}

func (e ErrRequestTooManyClients) Error() string {
	return "requested too many clients for room"
}

type ErrNoRoomFound struct {
	Message string
}

func (e ErrNoRoomFound) Error() string {
	return "no room found"
}

type ErrNoMatchingClient struct {
	Message string
}

func (e ErrNoMatchingClient) Error() string {
	return "no matching client"
}

type ErrInvalidSecret struct {
	Message string
}

func (e ErrInvalidSecret) Error() string {
	return "invalid secret"
}

type ErrRoomFull struct {
	Message string
}

func (e ErrRoomFull) Error() string {
	return "room full"
}

type ErrMaxClientTooSmall struct {
	Message string
}

func (e ErrMaxClientTooSmall) Error() string {
	return "max clients too small"
}

type RoomFactory func(id int32, secret int32, maxClients int32) (Room, error)

type Room interface {
	RoomMatches(id int32, secret int32) bool

	NewClient(session *session.Session) (*session.Session, error)
	ExistingClient(session *session.Session, clientID int32, clientSecret int32) (*session.Session, error)

	RemoveClient(clientID int32) error

	GetConnected() ([]*session.Session, error)

	IsHost(potentialHost *client.Client) (bool, error)
	SetHost(hostID *int32) (*session.Session, error)
	GetHost() (*session.Session, error)
	GetInfo() (*RoomInfo, error)
}

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
	GetRoomWithID(id int32) (Room, error)
	GetRoomList() ([]Room, error)
	GetRoomsSummary() (*RoomsSummary, error)
	CreateRoom(maxClients int32) (Room, error)
}
