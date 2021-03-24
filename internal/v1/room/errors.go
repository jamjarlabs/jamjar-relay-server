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

// ErrRequestTooManyClients occurs if a room is requested with more clients than are available
type ErrRequestTooManyClients struct {
	Message string
}

func (e ErrRequestTooManyClients) Error() string {
	return "requested too many clients for room"
}

// ErrNoRoomFound occurs if a room cannot be found
type ErrNoRoomFound struct {
	Message string
}

func (e ErrNoRoomFound) Error() string {
	return "no room found"
}

// ErrNoMatchingClient occurs when a client cannot be found in a room
type ErrNoMatchingClient struct {
	Message string
}

func (e ErrNoMatchingClient) Error() string {
	return "no matching client"
}

// ErrInvalidSecret occurs when a client's secret is not valid
type ErrInvalidSecret struct {
	Message string
}

func (e ErrInvalidSecret) Error() string {
	return "invalid secret"
}

// ErrRoomFull occurs when a room is full
type ErrRoomFull struct {
	Message string
}

func (e ErrRoomFull) Error() string {
	return "room full"
}

// ErrMaxClientTooSmall occurs when trying to create a room with a max client value that is too small
type ErrMaxClientTooSmall struct {
	Message string
}

func (e ErrMaxClientTooSmall) Error() string {
	return "max clients too small"
}
