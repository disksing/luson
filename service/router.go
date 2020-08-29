package service

import (
	"fmt"

	"github.com/disksing/luson/util"
	"github.com/gorilla/mux"
)

func NewRouter(js *JServer) *mux.Router {
	r := mux.NewRouter()

	id := fmt.Sprintf("{id:%s}", util.UUIDRegexp)

	r.HandleFunc("/", js.Create).Methods("POST")
	r.HandleFunc("/"+id, js.Get).Methods("GET")
	r.HandleFunc("/"+id, js.Put).Methods("PUT")

	return r
}
