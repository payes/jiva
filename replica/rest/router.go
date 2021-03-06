/*
 Copyright © 2020 The OpenEBS Authors

 This file was originally authored by Rancher Labs
 under Apache License 2018.

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

package rest

import (
	"net/http"
	_ "net/http/pprof" /* for profiling */

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rancher/go-rancher/api"
	"github.com/rancher/go-rancher/client"
)

func HandleError(s *client.Schemas, t func(http.ResponseWriter, *http.Request) error) http.Handler {
	return api.ApiHandler(s, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if err := t(rw, req); err != nil {
			apiContext := api.GetApiContext(req)
			apiContext.WriteErr(err)
		}
	}))
}

func checkAction(s *Server, t func(http.ResponseWriter, *http.Request) error) func(http.ResponseWriter, *http.Request) error {
	return func(rw http.ResponseWriter, req *http.Request) error {
		replica := s.Replica(api.GetApiContext(req))
		if replica.Actions[req.URL.Query().Get("action")] == "" {
			rw.WriteHeader(http.StatusNotFound)
			return nil
		}
		return t(rw, req)
	}
}

func NewRouter(s *Server) *mux.Router {
	schemas := NewSchema()
	router := mux.NewRouter().StrictSlash(true)
	f := HandleError

	router.Methods("GET").Path("/ping").Handler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte("pong"))
	}))

	// API framework routes
	router.Methods("GET").Path("/").Handler(api.VersionsHandler(schemas, "v1"))
	router.Methods("GET").Path("/v1/schemas").Handler(api.SchemasHandler(schemas))
	router.Methods("GET").Path("/v1/schemas/{id}").Handler(api.SchemaHandler(schemas))
	router.Methods("GET").Path("/v1").Handler(api.VersionHandler(schemas, "v1"))

	// Stats
	router.Methods("GET").Path("/v1/stats").Handler(f(schemas, s.GetStats))

	// rebuildinfo
	router.Methods("GET").Path("/v1/rebuildinfo").Handler(f(schemas, s.GetRebuildInfo))

	// Replicas
	router.Methods("GET").Path("/v1/replicas").Handler(f(schemas, s.ListReplicas))
	router.Methods("GET").Path("/v1/replicas/{id}").Handler(f(schemas, s.GetReplica))
	router.Methods("GET").Path("/v1/replicas/{id}/volusage").Handler(f(schemas, s.GetVolUsage))
	router.Methods("DELETE").Path("/v1/replicas/{id}").Handler(f(schemas, s.DeleteReplica))

	router.Methods("DELETE").Path("/v1/delete").Handler(f(schemas, s.DeleteVolume))
	router.Handle("/metrics", promhttp.Handler())

	// Actions
	actions := map[string]func(http.ResponseWriter, *http.Request) error{
		"start":              s.StartReplica,
		"reload":             s.ReloadReplica,
		"updatecloneinfo":    s.UpdateCloneInfo,
		"snapshot":           s.SnapshotReplica,
		"open":               s.OpenReplica,
		"close":              s.CloseReplica,
		"resize":             s.Resize,
		"removedisk":         s.RemoveDisk,
		"replacedisk":        s.ReplaceDisk,
		"setrebuilding":      s.SetRebuilding,
		"setlogging":         s.SetLogging,
		"create":             s.Create,
		"revert":             s.RevertReplica,
		"prepareremovedisk":  s.PrepareRemoveDisk,
		"setrevisioncounter": s.SetRevisionCounter,
		"setreplicamode":     s.SetReplicaMode,
		"setcheckpoint":      s.SetCheckpoint,
	}

	for name, action := range actions {
		router.Methods("POST").Path("/v1/replicas/{id}").Queries("action", name).Handler(f(schemas, checkAction(s, action)))
	}
	router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)

	return router
}
