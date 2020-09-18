package service

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/disksing/luson/config"
	"github.com/disksing/luson/jsonp"
	"github.com/disksing/luson/jsonstore"
	"github.com/disksing/luson/key"
	"github.com/disksing/luson/metastore"
	"github.com/disksing/luson/util"
	"go.uber.org/zap"
)

// JServer services JSON data.
type JServer struct {
	logger *util.Logger
	mstore *metastore.Store
	jstore *jsonstore.Store
	conf   *config.Config
	apiKey key.APIKey
}

// NewJServer creates the JSON service handler.
func NewJServer(mstore *metastore.Store, jstore *jsonstore.Store, apiKey key.APIKey, conf *config.Config, logger *util.Logger) *JServer {
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
		ctx.statusText(http.StatusUnauthorized)
		return
	}

	v, _, ok := ctx.readJSONEx()
	if !ok {
		return
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

// Get handles JSON GET requests.
func (js *JServer) Get(w http.ResponseWriter, r *http.Request) {
	ctx := newCtx(w, r)
	id, p, ok := ctx.uriPointer()
	if !ok {
		return
	}

	if !js.checkMetaForRead(ctx, id) {
		return
	}

	v, hash, err := js.jstore.Get(id)
	if err != nil {
		ctx.text(http.StatusInternalServerError, err.Error())
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

	id, p, ok := ctx.uriPointer()
	if !ok {
		return
	}
	_, v, ok := ctx.readJSON()
	if !ok {
		return
	}
	if !js.checkMetaForWrite(ctx, id) {
		return
	}

	var hash string
	if p != "" {
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
	id, p, ok := ctx.uriPointer()
	if !ok {
		return
	}
	raw, v, ok := ctx.readJSON()
	if !ok {
		return
	}
	if ctx.probeMergeType(v) == "merge-patch" {
		js.mergePatch(ctx, id, p, v)
	} else {
		js.jsonPatch(ctx, id, p, raw)
	}
}

func (js *JServer) mergePatch(ctx *httpCtx, id, p string, v interface{}) {
	if id == "" {
		ctx.text(http.StatusBadRequest, "expect resource id")
		return
	}

	if !js.checkMetaForWrite(ctx, id) {
		return
	}

	old, hash, err := js.jstore.Get(id)
	if err != nil {
		ctx.text(http.StatusInternalServerError, err.Error())
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
	ctx.statusText(http.StatusOK)
}

func (js *JServer) jsonPatch(ctx *httpCtx, id, basePath string, data []byte) {
	ps, ok := js.readJSONPatch(ctx, id, basePath, data)
	if !ok {
		return
	}
	txn := js.jstore.NewTxn()
	for _, p := range ps {
		switch p.Op {
		case "test":
			if !js.checkMetaForWrite(ctx, p.pathID) {
				return
			}
			v, err := txn.Get(p.pathID)
			if err != nil {
				ctx.text(http.StatusInternalServerError, "id:"+p.pathID)
				return
			}
			v, err = jsonp.Get(v, p.Path)
			if err != nil {
				ctx.text(http.StatusBadRequest, err.Error())
				return
			}
			if !reflect.DeepEqual(v, p.Value) {
				ctx.text(http.StatusPreconditionFailed, "value does not match")
				return
			}

		}
	}
}

type jsonPatch struct {
	Op     string      `json:"op"`
	Path   string      `json:"path"`
	Value  interface{} `json:"value"`
	From   string      `json:"from"`
	pathID string
	fromID string
}

func (js *JServer) readJSONPatch(ctx *httpCtx, id, basePath string, data []byte) (ps []jsonPatch, ok bool) {
	if !ctx.unmarshalJSON(data, &ps) {
		return nil, false
	}
	for _, p := range ps {
		switch p.Op {
		case "move", "copy":
			// check from
			p.fromID, p.From, ok = js.adjustPath(ctx, id, basePath, p.From, "from")
			if !ok {
				return
			}
			fallthrough
		case "test", "remove", "add", "replace":
			// check path
			p.pathID, p.Path, ok = js.adjustPath(ctx, id, basePath, p.Path, "path")
			if !ok {
				return
			}
		default:
			ctx.text(http.StatusBadRequest, "invalid optype "+p.Op)
			return nil, false
		}
	}
	return ps, true
}

func (js *JServer) adjustPath(ctx *httpCtx, id, basePath, path, typ string) (string, string, bool) {
	if path != "" && path[0] != '/' {
		// start with UUID
		if len(path) < util.UUIDLen || !util.IsUUID(path[:util.UUIDLen]) {
			ctx.text(http.StatusBadRequest, fmt.Sprintf("expect uuid in %s `%s`", typ, path))
			return "", "", false
		}
		return path[:util.UUIDLen], path[util.UUIDLen:], true
	}
	if id == "" {
		ctx.text(http.StatusBadRequest, fmt.Sprintf("expect uuid in %s `%s`", typ, path))
		return "", "", false
	}
	return id, basePath + path, true
}

func (js *JServer) checkMeta(ctx *httpCtx, id string, mut bool) (ok bool) {
	mdata, err := js.mstore.Get(id)
	if err != nil {
		ctx.text(http.StatusInternalServerError, err.Error())
		return
	}
	if mdata == nil {
		ctx.text(http.StatusNotFound, id)
		return
	}
	if mdata.Access == config.Private && !ctx.checkAPIKey(js.apiKey) {
		ctx.text(http.StatusNotFound, id)
		return
	}
	if mut && mdata.Access == config.Protected && !ctx.checkAPIKey(js.apiKey) {
		ctx.text(http.StatusUnauthorized, id)
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
