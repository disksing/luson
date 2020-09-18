package tests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPatch(t *testing.T) {
	r := require.New(t)
	env, err := NewEnv()
	r.Nil(err)
	defer env.Close()
	id := mustPostExample(r, env)

	res, err := env.at("/" + id).withAuth().withRawContent(`{"author": "disksing", "loveFrom": null}`).patch()
	r.Nil(err)
	r.Equal(http.StatusOK, res.Status)

	res, err = env.at("/" + id + "/author").get()
	r.Nil(err)
	r.Equal(`"disksing"`, res.RawContent)

	res, err = env.at("/" + id + "/loveFrom").get()
	r.Nil(err)
	r.Equal(http.StatusNotAcceptable, res.Status)

	id = mustPostExample(r, env)

	res, err = env.at("/" + id + "/loveFrom/0").withAuth().withRawContent(`{"language": ["Go", "markdown"]}`).patch()
	r.Nil(err)
	r.Equal(http.StatusOK, res.Status)

	res, err = env.at("/" + id + "/loveFrom/0/language").get()
	r.Nil(err)
	r.Equal(http.StatusOK, res.Status)
	r.Equal([]interface{}{"Go", "markdown"}, res.Value.([]interface{}))
}

func TestJSONPatch(t *testing.T) {
	r := require.New(t)
	env, err := NewEnv()
	r.Nil(err)
	defer env.Close()
	id := mustPostExample(r, env)

	res, err := env.at("/" + id).withAuth().withRawContent(`[{"op": "add", "path": "/loveFrom/-", "value": {"tools": ["thinkpad"]}}]`).patch()
	r.Nil(err)
	r.Equal(http.StatusOK, res.Status, res.RawContent)

	res, err = env.at("/" + id + "/loveFrom/3/tools/0").get()
	r.Nil(err)
	r.Equal(http.StatusOK, res.Status)
	r.Equal("thinkpad", res.Value)
}
