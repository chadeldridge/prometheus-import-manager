package router

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

var testMiddleware = func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

var testMiddleware2 = func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func TestRouterNewRouterGroup(t *testing.T) {
	require := require.New(t)
	mux := &http.ServeMux{}

	t.Run("nil mux", func(t *testing.T) {
		rg, err := NewRouterGroup(nil, "/v1")
		require.Error(err, "NewRouterGroup() did not return an error")
		require.Equal("no mux provided", err.Error(), "NewRouterGroup() error did not match")
		require.Nil(rg, "NewRouterGroup() did not return nil")
	})

	t.Run("empty path", func(t *testing.T) {
		rg, err := NewRouterGroup(mux, "")
		require.NoError(err, "NewRouterGroup() returned an error: %s", err)
		require.NotNil(rg, "NewRouterGroup() returned nil")
		require.Equal(mux, rg.mux, "NewRouterGroup() mux did not match")
		require.Equal("/", rg.basePath, "NewRouterGroup() basePath did not match")
		require.NotNil(rg.middleware, "NewRouterGroup() middleware was nil")
		require.NotNil(rg.groups, "NewRouterGroup() groups was nil")
	})

	t.Run("no leading slash", func(t *testing.T) {
		rg, err := NewRouterGroup(mux, "v1")
		require.NoError(err, "NewRouterGroup() returned an error: %s", err)
		require.NotNil(rg, "NewRouterGroup() returned nil")
		require.Equal(mux, rg.mux, "NewRouterGroup() mux did not match")
		require.Equal("/v1", rg.basePath, "NewRouterGroup() basePath did not match")
		require.NotNil(rg.middleware, "NewRouterGroup() middleware was nil")
		require.NotNil(rg.groups, "NewRouterGroup() groups was nil")
	})

	t.Run("no middleware", func(t *testing.T) {
		rg, err := NewRouterGroup(mux, "/v1")
		require.NoError(err, "NewRouterGroup() returned an error: %s", err)
		require.NotNil(rg, "NewRouterGroup() returned nil")
		require.Equal(mux, rg.mux, "NewRouterGroup() mux did not match")
		require.Equal("/v1", rg.basePath, "NewRouterGroup() basePath did not match")
		require.NotNil(rg.middleware, "NewRouterGroup() middleware was nil")
		require.Len(rg.middleware, 0, "NewRouterGroup() middleware was not empty")
		require.NotNil(rg.groups, "NewRouterGroup() groups was nil")
	})

	t.Run("with middleware", func(t *testing.T) {
		rg, err := NewRouterGroup(mux, "/v1", testMiddleware)
		require.NoError(err, "NewRouterGroup() returned an error: %s", err)
		require.NotNil(rg, "NewRouterGroup() returned nil")
		require.Equal(mux, rg.mux, "NewRouterGroup() mux did not match")
		require.Equal("/v1", rg.basePath, "NewRouterGroup() basePath did not match")
		require.NotNil(rg.middleware, "NewRouterGroup() middleware was nil")
		require.Len(rg.middleware, 1, "NewRouterGroup() middleware was not empty")
		require.NotNil(rg.groups, "NewRouterGroup() groups was nil")
	})
}

func TestRouterGroup(t *testing.T) {
	require := require.New(t)
	mux := &http.ServeMux{}

	root, err := NewRouterGroup(mux, "/v1")
	require.NoError(err, "NewRouterGroup() returned an error: %s", err)
	require.NotNil(root, "NewRouterGroup() returned nil")

	t.Run("no root", func(t *testing.T) {
		group := root.Group("/user")
		require.NotNil(group, "Group() returned nil")
		require.Equal(root, group.root, "Group() root did not match")
		require.Equal("/v1/user", group.basePath, "Group() basePath did not match")
		require.Len(group.middleware, 0, "Group() middleware was not empty")
	})

	t.Run("with root", func(t *testing.T) {
		group := root.Group("/user")
		require.NotNil(group, "Group() returned nil")
		require.Equal(root, group.root, "Group() root did not match")
		require.Equal("/v1/user", group.basePath, "Group() basePath did not match")
		require.Len(group.middleware, 0, "Group() middleware was not empty")

		group2 := group.Group("/profile")
		require.NotNil(group2, "Group() returned nil")
		require.Equal(root, group2.root, "Group() root did not match")
		require.Equal("/v1/user/profile", group2.basePath, "Group() basePath did not match")
		require.Len(group2.middleware, 0, "Group() middleware was not empty")
	})

	t.Run("with middleware", func(t *testing.T) {
		group := root.Group("/user", testMiddleware)
		require.NotNil(group, "Group() returned nil")
		require.Equal(root, group.root, "Group() root did not match")
		require.Equal("/v1/user", group.basePath, "Group() basePath did not match")
		require.Len(group.middleware, 1, "Group() middleware was not empty")
	})

	t.Run("with middleware appended", func(t *testing.T) {
		group := root.Group("/user", testMiddleware)
		require.NotNil(group, "Group() returned nil")
		require.Equal(root, group.root, "Group() root did not match")
		require.Equal("/v1/user", group.basePath, "Group() basePath did not match")
		require.Len(group.middleware, 1, "Group() middleware was not empty")

		group2 := group.Group("/profile", testMiddleware, testMiddleware2)
		require.NotNil(group2, "Group() returned nil")
		require.Equal(root, group2.root, "Group() root did not match")
		require.Equal("/v1/user/profile", group2.basePath, "Group() basePath did not match")
		require.Len(group2.middleware, 2, "Group() middleware was not empty")
	})
}

func TestRouterGET(t *testing.T) {
	require := require.New(t)
	mux := &http.ServeMux{}

	root, err := NewRouterGroup(mux, "/v1")
	require.NoError(err, "NewRouterGroup() returned an error: %s", err)
	require.NotNil(root, "NewRouterGroup() returned nil")

	user := root.Group("/user")
	require.NotNil(user, "Group() returned nil")
	require.Equal("/v1/user", user.basePath, "Group() basePath did not match")

	t.Run("no middleware", func(t *testing.T) {
		root.GET("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	})

	t.Run("with middleware", func(t *testing.T) {
		root.GET("/test2", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}), testMiddleware)
	})
}
