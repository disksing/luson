package service

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/disksing/luson/key"
	"github.com/unrolled/render"
)

type httpCtx struct {
	w      http.ResponseWriter
	r      *http.Request
	render *render.Render
}

func newCtx(w http.ResponseWriter, r *http.Request) *httpCtx {
	var pretty bool
	if _, ok := r.Header["X-Pretty-Json"]; ok {
		pretty = true
	}
	if _, ok := r.URL.Query()["pretty"]; ok {
		pretty = true
	}
	render := render.New(render.Options{IndentJSON: pretty})
	return &httpCtx{
		w:      w,
		r:      r,
		render: render,
	}
}

func (ctx *httpCtx) readBody() ([]byte, bool) {
	defer ctx.r.Body.Close()
	data, err := ioutil.ReadAll(ctx.r.Body)
	if err != nil {
		ctx.text(http.StatusInternalServerError, err.Error())
		return nil, false
	}
	return data, true
}

func (ctx *httpCtx) parseJSON(data []byte) (interface{}, bool) {
	var v interface{}
	err := json.Unmarshal(data, &v)
	if err != nil {
		ctx.text(http.StatusBadRequest, err.Error())
		return nil, false
	}
	return v, true
}

func (ctx *httpCtx) readJSON() (interface{}, bool) {
	data, ok := ctx.readBody()
	if !ok {
		return nil, false
	}
	v, ok := ctx.parseJSON(data)
	if !ok {
		return nil, false
	}
	return v, true
}

// returns v, empty, ok
func (ctx *httpCtx) readJSONEx() (interface{}, bool, bool) {
	data, ok := ctx.readBody()
	if !ok {
		return nil, false, false
	}
	if len(data) == 0 {
		return nil, true, true
	}
	v, ok := ctx.parseJSON(data)
	if !ok {
		return nil, false, false
	}
	return v, false, true
}

func (ctx *httpCtx) uriPointer() (string, bool) {
	splits := strings.Split(ctx.r.RequestURI, "/")[2:] // trim /id prefix
	for i := range splits {
		s, err := url.PathUnescape(splits[i])
		if err != nil {
			ctx.text(http.StatusBadRequest, err.Error())
			return "", false
		}
		splits[i] = "/" + strings.NewReplacer("~", "~0", "/", "~1").Replace(s)
	}
	return strings.Join(splits, ""), true
}

func (ctx *httpCtx) text(status int, v string) {
	_ = ctx.render.Text(ctx.w, status, v)
}

func (ctx *httpCtx) json(status int, v interface{}) {
	_ = ctx.render.JSON(ctx.w, status, v)
}

func (ctx *httpCtx) checkAPIKey(apiKey key.APIKey) bool {
	return ctx.r.Header.Get("Authorization") == string(apiKey)
}
