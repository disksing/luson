package service

import (
	"fmt"

	"github.com/disksing/luson/util"
	"github.com/gorilla/mux"
)

// NewRouter returns the root HTTP handler.
func NewRouter(js *JServer) *mux.Router {
	r := mux.NewRouter().UseEncodedPath()

	id := fmt.Sprintf("{id:%s}", util.UUIDRegexp)

	r.HandleFunc("/", js.Create).Methods("POST")
	r.PathPrefix("/" + id).HandlerFunc(js.Get).Methods("GET")
	r.PathPrefix("/" + id).HandlerFunc(js.Put).Methods("PUT")
	r.PathPrefix("/" + id).HandlerFunc(js.Patch).Methods("PATCH")

	return r
}
