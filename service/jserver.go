package service

import (
	"net/http"

	"github.com/disksing/luson/config"
	"github.com/disksing/luson/jsonstore"
	"github.com/disksing/luson/key"
	"github.com/disksing/luson/metastore"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type JServer struct {
	logger *zap.SugaredLogger
	mstore *metastore.Store
	jstore *jsonstore.Store
	conf   *config.Config
	apiKey key.APIKey
}

func NewJServer(mstore *metastore.Store, jstore *jsonstore.Store, apiKey key.APIKey, conf *config.Config, logger *zap.SugaredLogger) *JServer {
	return &JServer{
		logger: logger,
		mstore: mstore,
		jstore: jstore,
		conf:   conf,
		apiKey: apiKey,
	}
}

func (js *JServer) Create(w http.ResponseWriter, r *http.Request) {
	if !checkAPIKey(r, js.apiKey) {
		response(r).Text(w, http.StatusForbidden, "")
		return
	}
	id, err := js.mstore.Create()
	if err != nil {
		response(r).JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	err = js.mstore.Put(&metastore.MetaData{
		ID:     id,
		Access: js.conf.DefaultAccess,
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
	mdata, err := js.mstore.Get(id)
	if err != nil {
		response(r).Text(w, http.StatusInternalServerError, err.Error())
		return
	}
	if mdata == nil {
		response(r).Text(w, http.StatusNotFound, "")
		return
	}
	if mdata.Access == config.Private && !checkAPIKey(r, js.apiKey) {
		response(r).Text(w, http.StatusNotFound, "")
		return
	}
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
	mdata, err := js.mstore.Get(id)
	if err != nil {
		response(r).Text(w, http.StatusInternalServerError, err.Error())
		return
	}
	if mdata == nil {
		response(r).Text(w, http.StatusNotFound, "")
		return
	}
	if mdata.Access == config.Private && !checkAPIKey(r, js.apiKey) {
		response(r).Text(w, http.StatusNotFound, "")
		return
	}
	if mdata.Access == config.Protected && !checkAPIKey(r, js.apiKey) {
		response(r).Text(w, http.StatusForbidden, "")
		return
	}
	err = js.jstore.Put(id, v)
	if err != nil {
		response(r).Text(w, http.StatusInternalServerError, err.Error())
		return
	}
	response(r).Text(w, http.StatusOK, "updated")
}
