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
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/golang/glog"
	v1 "github.com/jamjarlabs/jamjar-relay-server/internal/api/v1"
	"github.com/jamjarlabs/jamjar-relay-server/internal/api/v1/rooms"
	"github.com/jamjarlabs/jamjar-relay-server/internal/api/v1/websockets"
	"github.com/jamjarlabs/jamjar-relay-server/internal/v1/protocol"
	roomv1 "github.com/jamjarlabs/jamjar-relay-server/internal/v1/room"
)

const (
	portEnv    = "PORT"
	addressEnv = "ADDRESS"
)

const (
	maxClients    = 100
	ceilToNearest = 5
)

func main() {
	flag.Parse()

	portStr, exists := os.LookupEnv(portEnv)
	if !exists {
		glog.Fatalf("Missing %s environment variable", portEnv)
	}

	address, exists := os.LookupEnv(addressEnv)
	if !exists {
		glog.Fatalf("Missing %s environment variable", addressEnv)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		glog.Fatalf("Invalid %s variable provided, must be integer, %v", portEnv, err)
	}

	roomFactory := func(id, secret, maxClients int32) (roomv1.Room, error) {
		return roomv1.NewMemoryRoom(id, secret, maxClients)
	}

	rand.Seed(time.Now().UTC().UnixNano())

	roomManager := roomv1.NewRoomMemoryManager(maxClients, roomFactory, ceilToNearest)

	protocol := &protocol.StandardProtocol{
		RoomManager: roomManager,
	}

	router := chi.NewRouter()

	// Set up API
	api := &v1.API{
		Router: router,
		Websocket: &websockets.Handle{
			Protocol: protocol,
		},
		Rooms: &rooms.Handle{
			RoomManager: roomManager,
			Protocol:    protocol,
		},
	}
	api.Routes()

	srv := http.Server{
		Addr:    fmt.Sprintf("%s:%d", address, port),
		Handler: router,
	}

	glog.V(0).Infof("Starting API over HTTP on %s:%d", address, port)
	err = srv.ListenAndServe()
	if err != http.ErrServerClosed {
		glog.Fatalf("HTTP API Error: %s", err)
	}
}
