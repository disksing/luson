package tests

import (
	"net/http"
	"testing"

	"github.com/disksing/luson/config"
	"github.com/disksing/luson/util"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	r := require.New(t)
	env, err := NewEnv()
	r.Nil(err)
	defer env.Close()

	res, err := env.at("/").post()
	r.Nil(err)
	r.Equal(res.Status, http.StatusUnauthorized)

	res, err = env.at("/").withAuth().post()
	r.Nil(err)
	r.Equal(res.Status, http.StatusCreated)
	r.True(util.IsUUID(res.RawContent))

	res, err = env.at("/" + res.RawContent).get()
	r.Nil(err)
	r.Equal(res.Status, http.StatusOK, res.RawContent)
	r.True(res.IsJSON)
	r.Nil(res.Value)

	res, err = env.at("/").withAuth().
		withContent([]string{"hello", "world"}).
		post()
	r.Nil(err)
	r.True(util.IsUUID(res.RawContent))
	res, err = env.at("/" + res.RawContent).get()
	r.Nil(err)
	r.Equal(res.RawContent, `["hello","world"]`)
}

func TestCreatePublic(t *testing.T) {
	r := require.New(t)
	env, err := NewEnv()
	r.Nil(err)
	defer env.Close()
	env.Conf.DefaultAccess = config.Public

	res, err := env.at("/").post()
	r.Nil(err)
	r.Equal(res.Status, http.StatusCreated)
	r.True(util.IsUUID(res.RawContent))
}
