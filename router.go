package trie

import (
	"fmt"
	"net/http"
	"path"
	"strings"
)

type Router struct {
	prefix     string
	middleware []Middleware
	trees      map[string]*Tree
}

func (r *Router) Get(pattern string, handler Handler) {
	r.Handle(http.MethodGet, pattern, handler)
}

func (r *Router) Post(pattern string, handler Handler) {
	r.Handle(http.MethodPost, pattern, handler)
}

func (r *Router) Delete(pattern string, handler Handler) {
	r.Handle(http.MethodDelete, pattern, handler)
}

func (r *Router) Put(pattern string, handler Handler) {
	r.Handle(http.MethodPut, pattern, handler)
}

func (r *Router) Patch(pattern string, handler Handler) {
	r.Handle(http.MethodPatch, pattern, handler)
}

func (r *Router) Head(pattern string, handler Handler) {
	r.Handle(http.MethodHead, pattern, handler)
}

func (r *Router) Options(pattern string, handler Handler) {
	r.Handle(http.MethodOptions, pattern, handler)
}

func (r *Router) Any(pattern string, handler Handler) {
	r.Handle(http.MethodGet, pattern, handler)
	r.Handle(http.MethodPost, pattern, handler)
	r.Handle(http.MethodPut, pattern, handler)
	r.Handle(http.MethodPatch, pattern, handler)
	r.Handle(http.MethodHead, pattern, handler)
	r.Handle(http.MethodOptions, pattern, handler)
	r.Handle(http.MethodDelete, pattern, handler)
	r.Handle(http.MethodConnect, pattern, handler)
	r.Handle(http.MethodTrace, pattern, handler)
}

func (r *Router) Static(pattern, root string) {
	pattern = path.Join(r.prefix, pattern)
	patternArray := strings.Split(pattern, "/")
	patternSlice := make([]string, 0)
	for _, item := range patternArray {
		if item != "" && item[0] == '{' {
			break
		}
		patternSlice = append(patternSlice, item)
	}
	relativePath := strings.Join(patternSlice, "/")
	fileServer := http.StripPrefix(relativePath, http.FileServer(http.Dir(root)))
	handler := func(c *Context) {
		fileServer.ServeHTTP(c.Writer, c.Request)
	}
	r.Get(pattern, handler)
	r.Head(pattern, handler)
}

func (r *Router) Use(middleware ...Middleware) {
	if len(middleware) > 0 {
		r.middleware = append(r.middleware, middleware...)
	}
}

func (r *Router) Group(prefix string) *Router {
	return &Router{
		prefix:     prefix,
		trees:      r.trees,
		middleware: r.middleware,
	}
}

func (r *Router) Handle(method string, pattern string, handler Handler) {
	if _, ok := methods[method]; !ok {
		panic(fmt.Errorf("invalid method"))
	}
	tree, ok := r.trees[method]
	if !ok {
		tree = NewTree()
		r.trees[method] = tree
	}
	if r.prefix != "" {
		pattern = path.Join(r.prefix, pattern)
	}
	tree.Add(pattern, handler, r.middleware...)
}
