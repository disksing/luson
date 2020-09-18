package tests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPut(t *testing.T) {
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
	v["version"] = "v0.1"
	res, err = env.at("/" + id).withAuth().withContent(v).put()
	r.Nil(err)
	r.Equal(http.StatusOK, res.Status)

	res, err = env.at("/" + id + "/version").get()
	r.Nil(err)
	r.Equal(http.StatusOK, res.Status)
	r.Equal("v0.1", res.Value.(string))

	res, err = env.at("/" + id + "/loveFrom/0/language").withAuth().withRawContent(`["Go","markdown"]`).put()
	r.Nil(err)
	r.Equal(http.StatusOK, res.Status)

	res, err = env.at("/" + id + "/loveFrom/0/language").get()
	r.Nil(err)
	r.Equal(http.StatusOK, res.Status)
	r.Equal(`["Go","markdown"]`, res.RawContent)
}
