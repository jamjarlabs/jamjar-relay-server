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
	"fmt"
	"math"
	"math/rand"

	"github.com/jamjarlabs/jamjar-relay-server/internal/v1/session"
	sessionv1 "github.com/jamjarlabs/jamjar-relay-server/internal/v1/session"
	clientv1 "github.com/jamjarlabs/jamjar-relay-server/specs/v1/client"
)

// NewMemoryManager creates a new memory room manager with some default options
func NewMemoryManager(maxClients int32, roomFactory Factory, ceilCommittedToNearest int32) *MemoryManager {
	return &MemoryManager{
		MaxClients:             maxClients,
		Rooms:                  make(map[int32]Room),
		RoomFactory:            roomFactory,
		CeilCommittedToNearest: ceilCommittedToNearest,
	}
}

// MemoryManager manages rooms in memory
type MemoryManager struct {
	RoomFactory            Factory
	MaxClients             int32
	Rooms                  map[int32]Room
	CeilCommittedToNearest int32
}

// GetRoom retrieves a room specified by an ID
func (m *MemoryManager) GetRoom(id int32) (Room, error) {
	room := m.Rooms[id]
	if room == nil {
		return nil, ErrNoRoomFound{
			Message: fmt.Sprintf("No room found with the ID %d", id),
		}
	}
	return m.Rooms[id], nil
}

// DeleteRoom deletes a room from memory specified by an ID
func (m *MemoryManager) DeleteRoom(id int32) error {
	delete(m.Rooms, id)
	return nil
}

// ListRooms returns a list of all the room manager's rooms
func (m *MemoryManager) ListRooms() ([]Room, error) {
	list := make([]Room, 0, len(m.Rooms))
	for _, room := range m.Rooms {
		list = append(list, room)
	}
	return list, nil
}

// Summary generates a rooms summary from all the rooms in the room manager
func (m *MemoryManager) Summary() (*Summary, error) {
	currentClients := int32(0)
	committedClients := int32(0)
	for _, room := range m.Rooms {
		info, err := room.GetInfo()
		if err != nil {
			return nil, err
		}
		currentClients += info.CurrentClients
		committedClients += int32(math.Ceil(float64(info.MaxClients)/float64(m.CeilCommittedToNearest)) * float64(m.CeilCommittedToNearest))
	}
	return &Summary{
		NumberOfRooms:    int32(len(m.Rooms)),
		MaxClients:       m.MaxClients,
		CurrentClients:   currentClients,
		CommittedClients: committedClients,
	}, nil
}

// CreateRoom creates a new room in the room manager
func (m *MemoryManager) CreateRoom(maxClients int32) (Room, error) {

	summary, err := m.Summary()

	if err != nil {
		return nil, err
	}

	newCommittedClients := maxClients + summary.CommittedClients

	if summary.MaxClients-(newCommittedClients) < 0 {
		return nil, ErrRequestTooManyClients{
			Message: fmt.Sprintf(
				"Cannot create this room, this would result in more committed clients than the max (%d/%d)",
				newCommittedClients, summary.MaxClients),
		}
	}

	hasUniqueID := false
	roomID := int32(0)
	for !hasUniqueID {
		roomID = rand.Int31()
		hasUniqueID = true
		for _, room := range m.Rooms {
			info, err := room.GetInfo()
			if err != nil {
				return nil, err
			}
			if info.ID == roomID {
				hasUniqueID = false
				break
			}
		}
	}

	room, err := m.RoomFactory(roomID, rand.Int31(), maxClients)
	if err != nil {
		return nil, err
	}
	m.Rooms[roomID] = room

	return room, nil
}

// NewMemoryRoom creates a new memory room with some default options, it can return an error if the maxClients value
// is invalid (less than 1)
func NewMemoryRoom(id int32, secret int32, maxClients int32) (*MemoryRoom, error) {
	if maxClients <= 0 {
		return nil, ErrMaxClientTooSmall{
			Message: fmt.Sprintf("The room must have a maximum clients value of 1 or more, %d is invalid", maxClients),
		}
	}

	return &MemoryRoom{
		ID:                  id,
		Secret:              secret,
		MaxClients:          maxClients,
		ConnectedClients:    []*sessionv1.Session{},
		DisconnectedClients: []*clientv1.Client{},
		RoomStatus:          StatusRunning,
	}, nil
}

// MemoryRoom represents a room in memory, with the connected clients and options stored in memory
type MemoryRoom struct {
	ID                  int32
	Secret              int32
	MaxClients          int32
	HostID              *int32
	ConnectedClients    []*sessionv1.Session
	DisconnectedClients []*clientv1.Client
	RoomStatus          Status
}

// GetStatus returns the room's status
func (r *MemoryRoom) GetStatus() Status {
	return r.RoomStatus
}

// SetStatus sets the room's status
func (r *MemoryRoom) SetStatus(status Status) {
	r.RoomStatus = status
}

// IsHost determines if a client is the room's host
func (r *MemoryRoom) IsHost(potentialHost *clientv1.Client) (bool, error) {
	// Not host if no host assigned, or ID doesn't match host ID
	return &potentialHost.ID == r.HostID, nil
}

// RoomMatches determines if a room matches the ID and secret provided
func (r *MemoryRoom) RoomMatches(id int32, secret int32) bool {
	return r.ID == id && r.Secret == secret
}

// GetInfo generates the room's info
func (r *MemoryRoom) GetInfo() (*Info, error) {
	return &Info{
		ID:             r.ID,
		Secret:         r.Secret,
		MaxClients:     r.MaxClients,
		CurrentClients: int32(len(r.ConnectedClients)),
		RoomStatus:     r.RoomStatus.String(),
	}, nil
}

// NewClient handles creating a new client for the room for the connection provided
func (r *MemoryRoom) NewClient(connected *sessionv1.Session) (*sessionv1.Session, error) {
	if int32(len(r.ConnectedClients)) >= r.MaxClients {
		return connected, ErrRoomFull{
			Message: fmt.Sprintf("Room with ID %d is full", r.ID),
		}
	}

	newID := int32(0)
	for i := 0; i < len(r.ConnectedClients); i++ {
		if newID <= r.ConnectedClients[i].Client.ID {
			newID = r.ConnectedClients[i].Client.ID + 1
		}
	}
	for i := 0; i < len(r.DisconnectedClients); i++ {
		if newID <= r.DisconnectedClients[i].ID {
			newID = r.DisconnectedClients[i].ID + 1
		}
	}

	connected.Client = &clientv1.Client{
		ID:     newID,
		Secret: rand.Int31(),
	}

	r.ConnectedClients = append(r.ConnectedClients, connected)

	return connected, nil
}

// ExistingClient handles regenerating a client based on a previously disconnected client for the connection provided
func (r *MemoryRoom) ExistingClient(connected *sessionv1.Session, clientID int32, clientSecret int32) (*sessionv1.Session, error) {
	if int32(len(r.ConnectedClients)) >= r.MaxClients {
		return connected, ErrRoomFull{
			Message: fmt.Sprintf("Room with ID %d is full", r.ID),
		}
	}

	for i := 0; i < len(r.DisconnectedClients); i++ {
		matchClient := r.DisconnectedClients[i]
		if clientID == matchClient.ID {
			if clientSecret == matchClient.Secret {
				connected.Client = &clientv1.Client{
					ID:     clientID,
					Secret: clientSecret,
				}
				r.DisconnectedClients = append(r.DisconnectedClients[:i], r.DisconnectedClients[i+1:]...)
				r.ConnectedClients = append(r.ConnectedClients, connected)

				return connected, nil
			}
			return connected, ErrInvalidSecret{
				Message: fmt.Sprintf("Invalid secret provided for client with ID %d", clientID),
			}
		}
	}

	return connected, ErrNoMatchingClient{
		Message: fmt.Sprintf("No client found with ID %d", clientID),
	}
}

// GetClient returns a client with the ID provided, if none found an error is returned
func (r *MemoryRoom) GetClient(clientID int32) (*session.Session, error) {
	for _, connectedClient := range r.ConnectedClients {
		if connectedClient.Client.ID == clientID {
			return connectedClient, nil
		}
	}

	return nil, ErrNoMatchingClient{
		Message: fmt.Sprintf("No connected client found with ID %d", clientID),
	}
}

// RemoveClient handles removing a client from the room
func (r *MemoryRoom) RemoveClient(clientID int32) error {
	for i, connectedClient := range r.ConnectedClients {
		if clientID == connectedClient.Client.ID {
			r.ConnectedClients = append(r.ConnectedClients[:i], r.ConnectedClients[i+1:]...)
			r.DisconnectedClients = append(r.DisconnectedClients, connectedClient.Client)
			return nil
		}
	}
	return ErrNoMatchingClient{
		Message: fmt.Sprintf("No connected client found with ID %d", clientID),
	}
}

// GetConnected returns a list of all currently connected sessions
func (r *MemoryRoom) GetConnected() ([]*sessionv1.Session, error) {
	return r.ConnectedClients, nil
}

// SetHost sets a room's host, can be set to nil for no host
func (r *MemoryRoom) SetHost(hostID *int32) (*sessionv1.Session, error) {
	if hostID == nil {
		r.HostID = nil
		return nil, nil
	}

	host, err := r.GetClient(*hostID)
	if err != nil {
		return nil, err
	}

	r.HostID = &host.Client.ID

	return host, nil
}

// GetHost gets a room's host
func (r *MemoryRoom) GetHost() (*sessionv1.Session, error) {
	if r.HostID == nil {
		r.HostID = nil
		return nil, nil
	}

	host, err := r.GetClient(*r.HostID)
	if err != nil {
		switch err.(type) {
		case ErrNoMatchingClient:
			return r.SetHost(nil)
		default:
			return nil, err
		}
	}

	return host, nil
}
