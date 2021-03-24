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

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/golang/glog"
	relayhttp "github.com/jamjarlabs/jamjar-relay-server/specs/v1/http"
)

// HTTPFail writes a failed API api to the api writer provided.
func HTTPFail(w http.ResponseWriter, failure *relayhttp.Failure) {
	if failure.Code == http.StatusInternalServerError {
		glog.Error(failure.Message)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	// Convert into JSON
	output, err := json.Marshal(failure)
	if err != nil {
		// Should not occur, panic
		panic(err)
	}
	w.WriteHeader(failure.Code)
	_, err = w.Write(output)
	if err != nil {
		glog.Error(err)
	}
}

// HTTPSucceed writes a successful API api to the api writer provided.
func HTTPSucceed(w http.ResponseWriter, success *relayhttp.Success) {
	output, err := json.Marshal(success)
	if err != nil {
		// Should not occur, panic
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(success.Code)
	_, err = w.Write(output)
	if err != nil {
		glog.Error(err)
	}
}

// NotFound provides a handler for HTTP not found events to the API.
func NotFound() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		HTTPFail(w, &relayhttp.Failure{
			Code:    http.StatusNotFound,
			Message: fmt.Sprintf("Endpoint '%s' not found", r.URL.Path),
		})
	}
}

// MethodNotAllowed provides a handler for HTTP method not allowed events to the API.
func MethodNotAllowed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		HTTPFail(w, &relayhttp.Failure{
			Code:    http.StatusMethodNotAllowed,
			Message: fmt.Sprintf("Method '%s' not allowed for endpoint '%s'", r.Method, r.URL.Path),
		})
	}
}
