/*
Copyright 2020 The Knative Authors

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

package queue

import (
	"net/http"
)

type statsHandler struct {
	proto http.Handler
}

// NewStatsHandler returns a new StatHandler.
func NewStatsHandler(proto http.Handler) http.Handler {
	return &statsHandler{
		proto: proto,
	}
}

// ServeHTTP serves the stats over HTTP. Either protobuf or prometheus stats
// are served, depending on the Accept header.
func (reporter *statsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	reporter.proto.ServeHTTP(w, r)
}
