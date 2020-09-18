package tests

import (
	"net/http"

	"github.com/stretchr/testify/require"
)

func mustPostExample(r *require.Assertions, env *Env) string {
	v := `{"app":"luson", "loveFrom":[{"language":"Go"}, {"editor":"vscode"}, "GitHub"]}`
	res, err := env.at("/").withAuth().withRawContent(v).post()
	r.Nil(err)
	r.Equal(http.StatusCreated, res.Status)
	return res.RawContent
}
