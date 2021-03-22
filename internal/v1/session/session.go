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

package session

import (
	"github.com/jamjarlabs/jamjar-relay-server/specs/v1/client"
)

type Session struct {
	CloseSignal chan struct{}
	Closed      bool
	Write       chan []byte
	Client      *client.Client
	RoomID      *int32
}

func (s *Session) Close() {
	s.Closed = true
	close(s.CloseSignal)
}
