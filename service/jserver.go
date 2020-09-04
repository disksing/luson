package service

import (
	"net/http"

	"github.com/disksing/luson/config"
	"github.com/disksing/luson/jsonp"
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
	ctx := newCtx(w, r)

	if js.conf.DefaultAccess != config.Public && !ctx.checkAPIKey(js.apiKey) {
		ctx.statusText(http.StatusUnauthorized)
		return
	}

	b, err := ctx.readBody()
	if err != nil {
		return
	}

	var v interface{}
	if len(b) > 0 {
		v, err = ctx.parseJSON(b)
		if err != nil {
			return
		}
	}

	id, err := js.mstore.Create()
	if err != nil {
		js.logger.Error("failed to create meta", zap.String("cmd", "create"), zap.String("id", id), zap.Error(err))
		ctx.text(http.StatusInternalServerError, "failed to create meta")
		return
	}
	err = js.mstore.Put(&metastore.MetaData{ID: id, Access: js.conf.DefaultAccess})
	if err != nil {
		js.logger.Error("failed to put meta", zap.String("cmd", "create"), zap.String("id", id), zap.Error(err))
		ctx.text(http.StatusInternalServerError, "failed to write meta")
		return
	}
	err = js.jstore.Put(id, v)
	if err != nil {
		js.logger.Error("failed to put json data", zap.String("cmd", "create"), zap.String("id", id), zap.Error(err))
		ctx.text(http.StatusInternalServerError, "failed to write json data")
		return
	}
	js.logger.Info("create", zap.String("id", id))
	ctx.text(http.StatusCreated, id)
}

func (js *JServer) Get(w http.ResponseWriter, r *http.Request) {
	ctx := newCtx(w, r)

	id := mux.Vars(r)["id"]
	mdata, err := js.mstore.Get(id)
	if err != nil {
		ctx.text(http.StatusInternalServerError, err.Error())
		return
	}
	if mdata == nil {
		ctx.text(http.StatusNotFound, "")
		return
	}
	if mdata.Access == config.Private && !ctx.checkAPIKey(js.apiKey) {
		ctx.text(http.StatusNotFound, "")
		return
	}
	v, hash, err := js.jstore.Get(id)
	if err != nil {
		ctx.text(http.StatusInternalServerError, err.Error())
		return
	}
	p, err := ctx.uriPointer()
	if err != nil {
		return
	}
	v, err = jsonp.Get(v, p)
	if err != nil {
		ctx.text(http.StatusNotAcceptable, err.Error())
		return
	}

	ctx.w.Header().Add("ETag", hash)
	ctx.json(http.StatusOK, v)
}

func (js *JServer) Put(w http.ResponseWriter, r *http.Request) {
	ctx := newCtx(w, r)

	id := mux.Vars(r)["id"]
	v, err := ctx.readJSON()
	if err != nil {
		return
	}
	mdata, err := js.mstore.Get(id)
	if err != nil {
		ctx.text(http.StatusInternalServerError, err.Error())
		return
	}
	if mdata == nil {
		ctx.text(http.StatusNotFound, "")
		return
	}
	if mdata.Access == config.Private && !ctx.checkAPIKey(js.apiKey) {
		ctx.text(http.StatusNotFound, "")
		return
	}
	if mdata.Access == config.Protected && !ctx.checkAPIKey(js.apiKey) {
		ctx.text(http.StatusForbidden, "")
		return
	}

	p, err := ctx.uriPointer()
	if err != nil {
		return
	}
	var hash string
	if p != "" {
		// TODO: TXN
		var old interface{}
		old, hash, err = js.jstore.Get(id)
		if err != nil {
			ctx.text(http.StatusInternalServerError, err.Error())
			return
		}
		v, err = jsonp.Replace(old, p, v)
		if err != nil {
			ctx.text(http.StatusNotAcceptable, err.Error())
			return
		}
	}

	err = js.jstore.Put(id, v)
	if err != nil {
		ctx.text(http.StatusInternalServerError, err.Error())
		return
	}
	if hash != "" {
		ctx.w.Header().Add("ETag", hash)
	}
	ctx.text(http.StatusOK, "")
}

func (js *JServer) Patch(w http.ResponseWriter, r *http.Request) {
	ctx := newCtx(w, r)

	// FIXME: same with PUT
	id := mux.Vars(r)["id"]
	v, err := ctx.readJSON()
	if err != nil {
		return
	}
	mdata, err := js.mstore.Get(id)
	if err != nil {
		ctx.text(http.StatusInternalServerError, err.Error())
		return
	}
	if mdata == nil {
		ctx.text(http.StatusNotFound, "")
		return
	}
	if mdata.Access == config.Private && !ctx.checkAPIKey(js.apiKey) {
		ctx.text(http.StatusNotFound, "")
		return
	}
	if mdata.Access == config.Protected && !ctx.checkAPIKey(js.apiKey) {
		ctx.text(http.StatusForbidden, "")
		return
	}

	// TODO: TXN
	old, hash, err := js.jstore.Get(id)
	if err != nil {
		ctx.text(http.StatusInternalServerError, err.Error())
		return
	}

	p, err := ctx.uriPointer()
	if err != nil {
		return
	}
	sub, err := jsonp.Get(old, p)
	if err != nil {
		ctx.text(http.StatusNotAcceptable, err.Error())
		return
	}
	v, err = jsonp.Replace(old, p, jsonp.Merge(sub, v))
	if err != nil {
		ctx.text(http.StatusNotAcceptable, err.Error())
		return
	}
	err = js.jstore.Put(id, v)
	if err != nil {
		ctx.text(http.StatusInternalServerError, err.Error())
		return
	}
	if hash != "" {
		ctx.w.Header().Add("ETag", hash)
	}
	ctx.text(http.StatusOK, "")
}
