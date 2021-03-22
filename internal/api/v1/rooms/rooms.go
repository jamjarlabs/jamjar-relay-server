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

package rooms

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/jamjarlabs/jamjar-relay-server/internal/api/v1/api"
	"github.com/jamjarlabs/jamjar-relay-server/internal/v1/room"
	relayhttp "github.com/jamjarlabs/jamjar-relay-server/specs/v1/http"
)

type Handle struct {
	RoomManager room.RoomManager
}

type RoomCreationRequest struct {
	MaxClients int32 `json:"max_clients"`
}

func (h *Handle) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "room_id")
	id64, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		api.HTTPFail(w, &relayhttp.Failure{
			Code:    http.StatusBadRequest,
			Message: "Invalid room ID provided, must be a 32-bit integer",
		})
		return
	}

	id := int32(id64)

	retrievedRoom, err := h.RoomManager.GetRoomWithID(id)
	if err != nil {
		switch v := err.(type) {
		case room.ErrNoRoomFound:
			api.HTTPFail(w, &relayhttp.Failure{
				Code:    http.StatusBadRequest,
				Message: v.Message,
			})
			return
		default:
			api.HTTPFail(w, &relayhttp.Failure{
				Code:    http.StatusInternalServerError,
				Message: fmt.Sprintf("Internal Server Error: %s", err.Error()),
			})
			return
		}
	}

	info, err := retrievedRoom.GetInfo()
	if err != nil {
		api.HTTPFail(w, &relayhttp.Failure{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("Internal Server Error: %s", err.Error()),
		})
		return
	}

	api.HTTPSucceed(w, &relayhttp.Success{
		Code: http.StatusOK,
		Data: info,
	})
}

func (h *Handle) Summary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.RoomManager.GetRoomsSummary()
	if err != nil {
		api.HTTPFail(w, &relayhttp.Failure{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("Internal Server Error: %s", err.Error()),
		})
		return
	}

	api.HTTPSucceed(w, &relayhttp.Success{
		Code: http.StatusOK,
		Data: summary,
	})
}

func (h *Handle) Create(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		api.HTTPFail(w, &relayhttp.Failure{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprint("Missing body in request"),
		})
		return
	}

	var createRoom RoomCreationRequest
	err := json.NewDecoder(r.Body).Decode(&createRoom)
	if err != nil {
		api.HTTPFail(w, &relayhttp.Failure{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Invalid room creation request provided; %s", err.Error()),
		})
		return
	}

	newRoom, err := h.RoomManager.CreateRoom(createRoom.MaxClients)
	if err != nil {
		switch v := err.(type) {
		case room.ErrRequestTooManyClients:
			api.HTTPFail(w, &relayhttp.Failure{
				Code:    http.StatusBadRequest,
				Message: v.Message,
			})
			return
		case room.ErrMaxClientTooSmall:
			api.HTTPFail(w, &relayhttp.Failure{
				Code:    http.StatusBadRequest,
				Message: v.Message,
			})
			return
		default:
			api.HTTPFail(w, &relayhttp.Failure{
				Code:    http.StatusInternalServerError,
				Message: fmt.Sprintf("Internal Server Error: %s", err.Error()),
			})
			return
		}
	}

	info, err := newRoom.GetInfo()
	if err != nil {
		api.HTTPFail(w, &relayhttp.Failure{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("Internal Server Error: %s", err.Error()),
		})
		return
	}

	api.HTTPSucceed(w, &relayhttp.Success{
		Code: http.StatusOK,
		Data: info,
	})
}

func (h *Handle) List(w http.ResponseWriter, r *http.Request) {
	rooms, err := h.RoomManager.GetRoomList()
	if err != nil {
		api.HTTPFail(w, &relayhttp.Failure{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("Internal Server Error: %s", err.Error()),
		})
		return
	}

	infos := []*room.RoomInfo{}

	for _, room := range rooms {
		info, err := room.GetInfo()
		if err != nil {
			api.HTTPFail(w, &relayhttp.Failure{
				Code:    http.StatusInternalServerError,
				Message: fmt.Sprintf("Internal Server Error: %s", err.Error()),
			})
			return
		}
		infos = append(infos, info)
	}

	api.HTTPSucceed(w, &relayhttp.Success{
		Code: http.StatusOK,
		Data: infos,
	})
}
