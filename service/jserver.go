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

// JServer services JSON data.
type JServer struct {
	logger *zap.SugaredLogger
	mstore *metastore.Store
	jstore *jsonstore.Store
	conf   *config.Config
	apiKey key.APIKey
}

// NewJServer creates the JSON service handler.
func NewJServer(mstore *metastore.Store, jstore *jsonstore.Store, apiKey key.APIKey, conf *config.Config, logger *zap.SugaredLogger) *JServer {
	return &JServer{
		logger: logger,
		mstore: mstore,
		jstore: jstore,
		conf:   conf,
		apiKey: apiKey,
	}
}

// Create handles JSON POST requests.
func (js *JServer) Create(w http.ResponseWriter, r *http.Request) {
	ctx := newCtx(w, r)

	if js.conf.DefaultAccess != config.Public && !ctx.checkAPIKey(js.apiKey) {
		ctx.text(http.StatusUnauthorized, "")
		return
	}

	v, _, ok := ctx.readJSONEx()
	if !ok {
		return
	}

	id, err := js.mstore.Create()
	if err != nil {
		ctx.json(http.StatusInternalServerError, err.Error())
		return
	}
	err = js.mstore.Put(&metastore.MetaData{ID: id, Access: js.conf.DefaultAccess})
	if err != nil {
		ctx.json(http.StatusInternalServerError, err.Error())
		return
	}
	err = js.jstore.Put(id, v)
	if err != nil {
		ctx.json(http.StatusInternalServerError, err.Error())
		return
	}
	ctx.text(http.StatusCreated, id)
}

func (js *JServer) checkMeta(ctx *httpCtx, id string, mut bool) (ok bool) {
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
	if mut && mdata.Access == config.Protected && !ctx.checkAPIKey(js.apiKey) {
		ctx.text(http.StatusUnauthorized, "")
		return
	}
	return true
}

func (js *JServer) checkMetaForRead(ctx *httpCtx, id string) bool {
	return js.checkMeta(ctx, id, false)
}

func (js *JServer) checkMetaForWrite(ctx *httpCtx, id string) bool {
	return js.checkMeta(ctx, id, true)
}

// Get handles JSON GET requests.
func (js *JServer) Get(w http.ResponseWriter, r *http.Request) {
	ctx := newCtx(w, r)
	id := mux.Vars(r)["id"]

	if !js.checkMetaForRead(ctx, id) {
		return
	}

	v, hash, err := js.jstore.Get(id)
	if err != nil {
		ctx.text(http.StatusInternalServerError, err.Error())
		return
	}
	p, ok := ctx.uriPointer()
	if !ok {
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

// Put handles JSON PUT requests.
func (js *JServer) Put(w http.ResponseWriter, r *http.Request) {
	ctx := newCtx(w, r)
	id := mux.Vars(r)["id"]

	v, ok := ctx.readJSON()
	if !ok {
		return
	}
	if !js.checkMetaForWrite(ctx, id) {
		return
	}

	p, ok := ctx.uriPointer()
	if !ok {
		return
	}
	var hash string
	if p != "" {
		// TODO: TXN
		var old interface{}
		var err error
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

	err := js.jstore.Put(id, v)
	if err != nil {
		ctx.text(http.StatusInternalServerError, err.Error())
		return
	}
	if hash != "" {
		ctx.w.Header().Add("ETag", hash)
	}
	ctx.text(http.StatusOK, "")
}

// Patch handles JSON PATCH requests.
func (js *JServer) Patch(w http.ResponseWriter, r *http.Request) {
	ctx := newCtx(w, r)

	id := mux.Vars(r)["id"]
	v, ok := ctx.readJSON()
	if !ok {
		return
	}
	if !js.checkMetaForWrite(ctx, id) {
		return
	}

	// TODO: TXN
	old, hash, err := js.jstore.Get(id)
	if err != nil {
		ctx.text(http.StatusInternalServerError, err.Error())
		return
	}

	p, ok := ctx.uriPointer()
	if !ok {
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
