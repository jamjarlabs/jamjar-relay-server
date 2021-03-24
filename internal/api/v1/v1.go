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

package v1

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/jamjarlabs/jamjar-relay-server/internal/api/v1/api"
)

// WebsocketHandler defines the contract for serving websocket requests
type WebsocketHandler interface {
	Websocket(w http.ResponseWriter, r *http.Request)
}

// RoomsHandler defines the contract for serving room requests
type RoomsHandler interface {
	Get(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	Summary(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)
}

// API ties together the API with the router and all of the API handlers.
type API struct {
	Router    chi.Router
	Websocket WebsocketHandler
	Rooms     RoomsHandler
}

// Routes creates the endpoint routes for v1 of the API.
func (a *API) Routes() {
	a.Router.Route("/v1", func(r chi.Router) {
		r.NotFound(api.NotFound())
		r.HandleFunc("/websocket", a.Websocket.Websocket)
		r.Route("/api", func(r chi.Router) {
			r.Get("/summary", a.Rooms.Summary)
			r.Route("/rooms", func(r chi.Router) {
				r.Get("/", a.Rooms.List)
				r.Post("/", a.Rooms.Create)
				r.Route("/{room_id}", func(r chi.Router) {
					r.Get("/", a.Rooms.Get)
					r.Delete("/", a.Rooms.Delete)
				})
			})
		})
	})
}
