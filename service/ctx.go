package service

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

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

func (ctx *httpCtx) readBody() ([]byte, error) {
	defer ctx.r.Body.Close()
	data, err := ioutil.ReadAll(ctx.r.Body)
	if err != nil {
		ctx.text(http.StatusInternalServerError, err.Error())
		return nil, err
	}
	return data, nil
}

func (ctx *httpCtx) parseJSON(data []byte) (interface{}, error) {
	var v interface{}
	err := json.Unmarshal(data, &v)
	if err != nil {
		ctx.text(http.StatusBadRequest, err.Error())
		return nil, err
	}
	return v, nil
}

func (ctx *httpCtx) readJSON() (interface{}, error) {
	data, err := ctx.readBody()
	if err != nil {
		return nil, err
	}
	v, err := ctx.parseJSON(data)
	if err != nil {
		return nil, err
	}
	return v, nil
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
