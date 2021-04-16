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

package api

// RoomCreationRequest defines the data needed to create a new room
type RoomCreationRequest struct {
	MaxClients int32 `json:"max_clients"`
}

// RoomInfo defines useful information about a room that can be easily serialised
type RoomInfo struct {
	ID             int32  `json:"id"`
	Secret         int32  `json:"secret"`
	MaxClients     int32  `json:"max_clients"`
	CurrentClients int32  `json:"current_clients"`
	RoomStatus     string `json:"room_status"`
}

// RoomsSummary defines a grouped summary of multiple rooms, useful for seeing the overall state of the relay server
type RoomsSummary struct {
	NumberOfRooms    int32 `json:"number_of_rooms"`
	MaxClients       int32 `json:"max_clients"`
	CurrentClients   int32 `json:"current_clients"`
	CommittedClients int32 `json:"committed_clients"`
}
