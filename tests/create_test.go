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
	r.Equal(http.StatusUnauthorized, res.Status)

	res, err = env.at("/").withAuth().post()
	r.Nil(err)
	r.Equal(http.StatusCreated, res.Status)
	r.True(util.IsUUID(res.RawContent))

	res, err = env.at("/" + res.RawContent).get()
	r.Nil(err)
	r.Equal(http.StatusOK, res.Status, res.RawContent)
	r.True(res.IsJSON)
	r.Nil(res.Value)

	res, err = env.at("/").withAuth().
		withContent([]string{"hello", "world"}).
		post()
	r.Nil(err)
	r.True(util.IsUUID(res.RawContent))
	res, err = env.at("/" + res.RawContent).get()
	r.Nil(err)
	r.Equal(`["hello","world"]`, res.RawContent)
}

func TestCreatePublic(t *testing.T) {
	r := require.New(t)
	env, err := NewEnv()
	r.Nil(err)
	defer env.Close()
	env.Conf.DefaultAccess = config.Public

	res, err := env.at("/").post()
	r.Nil(err)
	r.Equal(http.StatusCreated, res.Status)
	r.True(util.IsUUID(res.RawContent))
}

func TestCreatePrivate(t *testing.T) {
	r := require.New(t)
	env, err := NewEnv()
	r.Nil(err)
	defer env.Close()
	env.Conf.DefaultAccess = config.Private

	res, err := env.at("/").post()
	r.Nil(err)
	r.Equal(http.StatusUnauthorized, res.Status)

	res, err = env.at("/").withAuth().post()
	r.Nil(err)
	r.Equal(http.StatusCreated, res.Status)
	r.True(util.IsUUID(res.RawContent))

	id := res.RawContent
	res, err = env.at("/" + id).get()
	r.Nil(err)
	r.Equal(http.StatusNotFound, res.Status)

	res, err = env.at("/" + id).withAuth().get()
	r.Nil(err)
	r.Equal(http.StatusOK, res.Status)
	r.Nil(res.Value)
}
