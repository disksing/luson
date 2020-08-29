package service

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/disksing/luson/key"
	"github.com/unrolled/render"
)

func response(r *http.Request) *render.Render {
	var pretty bool
	if _, ok := r.Header["X-Pretty-Json"]; ok {
		pretty = true
	}
	if _, ok := r.URL.Query()["pretty"]; ok {
		pretty = true
	}
	return render.New(render.Options{
		IndentJSON: pretty,
		IndentXML:  pretty,
	})
}

func readBody(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		response(r).Text(w, http.StatusInternalServerError, err.Error())
		return nil, err
	}
	return b, nil
}

func readJSON(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	b, err := readBody(w, r)
	if err != nil {
		return nil, err
	}
	var v interface{}
	err = json.Unmarshal(b, &v)
	if err != nil {
		response(r).Text(w, http.StatusBadRequest, err.Error())
		return nil, err
	}
	return v, nil
}

func checkAPIKey(r *http.Request, apiKey key.APIKey) bool {
	return r.Header.Get("Authorization") == "Token "+string(apiKey)
}
