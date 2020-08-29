package tests

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
)

type Req struct {
	addr    string
	url     string
	heads   map[string]string
	params  map[string]string
	content []byte
}

func (env *Env) at(url string) *Req {
	return &Req{
		addr:   env.addr,
		url:    url,
		heads:  make(map[string]string),
		params: make(map[string]string),
	}
}

func (r *Req) withHead(k, v string) *Req {
	r.heads[k] = v
	return r
}

func (r *Req) withParam(k, v string) *Req {
	r.params[k] = v
	return r
}

func (r *Req) withAuth() *Req {
	return r.withHead("Authorization", MockAPIKey)
}

func (r *Req) withPrettyParam() *Req {
	return r.withParam("pretty", "")
}

func (r *Req) withPrettyHead() *Req {
	return r.withHead("X-Pretty-Json", "")
}

func (r *Req) withContent(v interface{}) *Req {
	data, _ := json.Marshal(v)
	r.content = data
	return r
}

func (r *Req) withRawContent(data string) *Req {
	r.content = []byte(data)
	return r
}

func (r *Req) get() (*Res, error) {
	return r.exec("GET")
}

func (r *Req) post() (*Res, error) {
	return r.exec("POST")
}

func (r *Req) put() (*Res, error) {
	return r.exec("PUT")
}

func (r *Req) delete() (*Res, error) {
	return r.exec("DELETE")
}

func (r *Req) patch() (*Res, error) {
	return r.exec("PATCH")
}

func (r *Req) exec(method string) (*Res, error) {
	u := "http://" + path.Join(r.addr, r.url)
	var p string
	for k, v := range r.params {
		if p == "" {
			p += "?"
		} else {
			p += "&"
		}
		p += url.PathEscape(k)
		if v != "" {
			p += "=" + url.PathEscape(v)
		}
	}
	req, err := http.NewRequest(method,
		u+p,
		bytes.NewBuffer(r.content))
	if err != nil {
		return nil, err
	}
	for k, v := range r.heads {
		req.Header.Add(k, v)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var v interface{}
	err = json.Unmarshal(b, &v)
	return &Res{
		Status:     res.StatusCode,
		RawContent: string(b),
		IsJSON:     err == nil,
		Value:      v,
	}, nil
}

type Res struct {
	Status     int
	RawContent string
	IsJSON     bool
	Value      interface{}
}
