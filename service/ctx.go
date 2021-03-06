package service

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/disksing/luson/jsonp"
	"github.com/disksing/luson/key"
	"github.com/disksing/luson/util"
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

func (ctx *httpCtx) unmarshalJSON(data []byte, v interface{}) bool {
	err := json.Unmarshal(data, v)
	if err != nil {
		ctx.text(http.StatusBadRequest, err.Error())
		return false
	}
	return true
}

func (ctx *httpCtx) parseJSON(data []byte) (interface{}, bool) {
	var v interface{}
	ok := ctx.unmarshalJSON(data, &v)
	return v, ok
}

func (ctx *httpCtx) readJSON() ([]byte, interface{}, bool) {
	data, ok := ctx.readBody()
	if !ok {
		return nil, nil, false
	}
	v, ok := ctx.parseJSON(data)
	if !ok {
		return data, nil, false
	}
	return data, v, true
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

func (ctx *httpCtx) uriPointer() (string, string, bool) {
	var id string
	var sb strings.Builder
	path := ctx.r.URL.Path
	if ctx.r.URL.RawPath != "" {
		path = ctx.r.URL.RawPath
	}
	if path != "" && path[0] == '/' {
		path = path[1:]
	}
	for _, s := range strings.Split(path, "/") {
		if id == "" {
			if !util.IsUUID(s) {
				ctx.text(http.StatusBadRequest, "expected UUID in request URL")
				return "", "", false
			}
			id = s
			continue
		}
		sb.WriteByte('/')
		s, err := url.PathUnescape(s)
		if err != nil {
			ctx.text(http.StatusBadRequest, "failed to unescape request URI")
			return "", "", false
		}
		sb.WriteString(jsonp.PointerEscaper.Replace(s))
	}
	return id, sb.String(), true
}

func (ctx *httpCtx) text(status int, v string) {
	_ = ctx.render.Text(ctx.w, status, v)
}

func (ctx *httpCtx) statusText(status int) {
	ctx.text(status, http.StatusText(status))
}

func (ctx *httpCtx) json(status int, v interface{}) {
	_ = ctx.render.JSON(ctx.w, status, v)
}

func (ctx *httpCtx) checkAPIKey(apiKey key.APIKey) bool {
	return ctx.r.Header.Get("Authorization") == string(apiKey)
}

func (ctx *httpCtx) probeMergeType(v interface{}) string {
	for _, t := range ctx.r.Header.Values("Content-Type") {
		switch t {
		case "application/json-patch+json":
			return "json-patch"
		case "application/merge-patch+json":
			return "merge-patch"
		}
	}
	if ctx.r.RequestURI == "/" {
		return "json-patch"
	}
	if _, ok := v.([]interface{}); ok {
		return "json-patch"
	}
	return "merge-patch"
}
