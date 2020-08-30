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
	r.Equal(res.Status, http.StatusCreated)

	id := res.RawContent

	res, err = env.at("/" + id).get()
	r.Nil(err)
	r.Equal(res.Status, http.StatusOK)
	r.Equal(res.Value, v)

	res, err = env.at("/" + id).withPrettyHead().get()
	r.Nil(err)
	r.Equal(res.Status, http.StatusOK)
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
	r.Equal(res2.Status, http.StatusOK)
	r.Equal(res2.RawContent, res.RawContent)
}
