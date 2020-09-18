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

	v := map[string]interface{}{
		"app": "luson",
		"loveFrom": []interface{}{
			map[string]interface{}{"language": "Go"},
			map[string]interface{}{"editor": "vscode"},
			"GitHub",
		},
	}
	res, err := env.at("/").withAuth().withContent(v).post()
	r.Nil(err)
	r.Equal(http.StatusCreated, res.Status)

	id := res.RawContent

	res, err = env.at("/" + id).withAuth().withRawContent(`{"author": "disksing", "loveFrom": null}`).patch()
	r.Nil(err)
	r.Equal(http.StatusOK, res.Status)

	res, err = env.at("/" + id + "/author").get()
	r.Nil(err)
	r.Equal(`"disksing"`, res.RawContent)

	res, err = env.at("/" + id + "/loveFrom").get()
	r.Nil(err)
	r.Equal(http.StatusNotAcceptable, res.Status)

	res, err = env.at("/" + id).withAuth().withContent(v).put()
	r.Nil(err)
	r.Equal(http.StatusOK, res.Status)

	res, err = env.at("/" + id + "/loveFrom/0").withAuth().withRawContent(`{"language": ["Go", "markdown"]}`).patch()
	r.Nil(err)
	r.Equal(http.StatusOK, res.Status)

	res, err = env.at("/" + id + "/loveFrom/0/language").get()
	r.Nil(err)
	r.Equal(http.StatusOK, res.Status)
	r.Equal([]interface{}{"Go", "markdown"}, res.Value.([]interface{}))
}
