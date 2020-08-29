package service

import (
	"net/http"

	"github.com/disksing/luson/jsonstore"
	"github.com/disksing/luson/metastore"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type JServer struct {
	logger *zap.SugaredLogger
	mstore *metastore.Store
	jstore *jsonstore.Store
}

func NewJServer(mstore *metastore.Store, jstore *jsonstore.Store, logger *zap.SugaredLogger) *JServer {
	return &JServer{
		logger: logger,
		mstore: mstore,
		jstore: jstore,
	}
}

func (js *JServer) Create(w http.ResponseWriter, r *http.Request) {
	id, err := js.mstore.Create()
	if err != nil {
		response(r).JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	err = js.mstore.Put(&metastore.MetaData{
		ID:     id,
		Access: metastore.Protected,
	})
	if err != nil {
		response(r).JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	err = js.jstore.Put(id, nil)
	if err != nil {
		response(r).JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	response(r).Text(w, http.StatusCreated, id)
}

func (js *JServer) Get(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	v, hash, err := js.jstore.Get(id)
	if err != nil {
		response(r).Text(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Add("ETag", hash)
	response(r).JSON(w, http.StatusOK, v)
}

func (js *JServer) Put(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	v, err := readJSON(w, r)
	if err != nil {
		return
	}
	err = js.jstore.Put(id, v)
	if err != nil {
		response(r).Text(w, http.StatusInternalServerError, err.Error())
		return
	}
	response(r).Text(w, http.StatusOK, "updated")
}
