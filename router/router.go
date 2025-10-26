package router

import (
	"fmt"
	"net/http"
	"regexp"
)

//
// Path Parameters: '/v1/user/{id}' accessed with r.PathValue("id")
//
// rg := NewRouterGroup("/v1")
// rg.Group("/user", AuthMiddleware)
// 	rg.GET("/{id}", GetUser)

var duplicateSlashes = regexp.MustCompile("/{2,}")

type RouterGroup struct {
	mux        *http.ServeMux
	root       *RouterGroup
	basePath   string
	middleware []Middleware
	groups     []*RouterGroup
}

func cleanPath(path string) string {
	if path == "" {
		path = "/"
	}

	if path[0] != '/' {
		path = "/" + path
	}

	if len(path) > 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	return duplicateSlashes.ReplaceAllString(path, "/")
}

func cleanMiddleware(middleware []Middleware) []Middleware {
	if middleware == nil {
		return []Middleware{}
	}

	for i := 0; i < len(middleware); i++ {
		if middleware[i] == nil {
			if i == len(middleware)-1 {
				middleware = middleware[:i]
				break
			}

			middleware = append(middleware[:i], middleware[i+1:]...)
		}
	}

	return middleware
}

func NewRouterGroup(mux *http.ServeMux, path string, middleware ...Middleware) (*RouterGroup, error) {
	if mux == nil {
		return nil, fmt.Errorf("no mux provided")
	}

	path = cleanPath(path)
	middleware = cleanMiddleware(middleware)

	return &RouterGroup{
		mux:        mux,
		basePath:   path,
		middleware: middleware,
		groups:     []*RouterGroup{},
	}, nil
}

func (group *RouterGroup) Group(path string, middleware ...Middleware) *RouterGroup {
	r := group.root
	if r == nil {
		r = group
	}

	path = cleanPath(path)
	rg := RouterGroup{
		root:     r,
		basePath: cleanPath(group.basePath + "/" + path),
	}

	if len(middleware) > 0 && middleware[0] != nil {
		rg.middleware = append(rg.middleware, middleware...)
	}

	return &rg
}

func (group *RouterGroup) ANY(path string, handler http.Handler, middleware ...Middleware) {
	h := group.genHandler(handler, middleware)
	path = cleanPath(path)

	mux := group.mux
	if mux == nil {
		mux = group.root.mux
	}
	mux.Handle(cleanPath(group.basePath+"/"+path), h)
}

func (group *RouterGroup) GET(path string, handler http.Handler, middleware ...Middleware) {
	h := group.genHandler(handler, middleware)
	path = cleanPath(path)

	mux := group.mux
	if mux == nil {
		mux = group.root.mux
	}
	mux.Handle("GET "+cleanPath(group.basePath+"/"+path), h)
}

func (group RouterGroup) genHandler(h http.Handler, middleware []Middleware) http.Handler {
	if len(middleware) != 0 {
		h = compile(h, middleware)
	}

	if len(group.middleware) != 0 {
		h = compile(h, group.middleware)
	}

	return h
}

func compile(h http.Handler, middleware []Middleware) http.Handler {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}

	return h
}
