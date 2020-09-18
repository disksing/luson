package tests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
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

	res, err = env.at("/" + id).get()
	r.Nil(err)
	r.Equal(http.StatusOK, res.Status)
	r.Equal(v, res.Value)

	res, err = env.at("/" + id).withPrettyHead().get()
	r.Nil(err)
	r.Equal(http.StatusOK, res.Status)
	r.Equal(res.RawContent, `{
  "app": "luson",
  "loveFrom": [
    {
      "language": "Go"
    },
    {
      "editor": "vscode"
    },
    "GitHub"
  ]
}
`)

	res2, err := env.at("/" + id).withPrettyParam().get()
	r.Nil(err)
	r.Equal(http.StatusOK, res2.Status)
	r.Equal(res.RawContent, res2.RawContent)

	partial, err := env.at("/" + id + "/loveFrom/1/editor").get()
	r.Nil(err)
	r.Equal(http.StatusOK, partial.Status)
	r.Equal("vscode", partial.Value)

	partial, err = env.at("/" + id + "/loveFrom/0").get()
	r.Nil(err)
	r.Equal(http.StatusOK, partial.Status)
	r.Equal(`{"language":"Go"}`, partial.RawContent)

	partial, err = env.at("/" + id + "/loveFrom/9").get()
	r.Nil(err)
	r.Equal(http.StatusNotAcceptable, partial.Status)
}

func TestUriPath(t *testing.T) {
	r := require.New(t)
	env, err := NewEnv()
	r.Nil(err)
	defer env.Close()

	v := `{"/":{"~~1~0/":{"foo":"bar"}}}`
	res, err := env.at("/").withAuth().withRawContent(v).post()
	r.Nil(err)
	r.Equal(http.StatusCreated, res.Status)

	id := res.RawContent
	res, err = env.at("/" + id + "/%2f/%7e%7e1%7e0%2f/foo").get()
	r.Nil(err)
	r.Equal("bar", res.Value)
}
