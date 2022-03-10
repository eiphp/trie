package trie

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type H map[string]interface{}

type Handler func(*Context)

type Middleware func(next Handler) Handler

type Parameter map[string]string

var (
	defaultRegex = `[\w]+`
	idRegex      = `[\d]+`
	idKey        = `id`
	methods      = map[string]struct{}{
		http.MethodGet:    {},
		http.MethodPost:   {},
		http.MethodPut:    {},
		http.MethodDelete: {},
		http.MethodPatch:  {},
		http.MethodHead:   {},
	}
)

func Debug(format string, v ...interface{}) {
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	fmt.Fprintf(os.Stdout, format, v...)
}

type Engine struct {
	Router
	Template *template.Template
	FuncMap  template.FuncMap
	notFound Handler
}

func NewEngine() *Engine {
	return &Engine{
		Router: Router{
			trees: make(map[string]*Tree),
		},
	}
}

func (e *Engine) parse(uri string, pattern string) (parameters Parameter, b bool) {
	var (
		parameterName []string
		regex         string
	)
	b = true
	parameters = make(Parameter)
	res := strings.Split(pattern, "/")
	for _, str := range res {
		if str == "" {
			continue
		}
		length := len(str)
		firstChar := str[0]
		lastChar := str[length-1]
		if string(firstChar) == "{" && string(lastChar) == "}" {
			match := string(str[1 : length-1])
			result := strings.Split(match, ":")
			parameterName = append(parameterName, result[0])
			regex = regex + "/" + "(" + result[1] + ")"
		} else if string(firstChar) == ":" {
			match := str
			result := strings.Split(match, ":")
			parameterName = append(parameterName, result[1])
			if result[1] == idKey {
				regex = regex + "/" + "(" + idRegex + ")"
			} else {
				regex = regex + "/" + "(" + defaultRegex + ")"
			}
		} else {
			regex = regex + "/" + str
		}
	}
	if strings.HasSuffix(uri, "/") {
		regex = regex + "/"
	}
	reg := regexp.MustCompile(regex)
	if subMatch := reg.FindSubmatch([]byte(uri)); subMatch != nil {
		if string(subMatch[0]) == uri {
			subMatch = subMatch[1:]
			for k, v := range subMatch {
				parameters[parameterName[k]] = string(v)
			}
			return
		}
	}
	return nil, false
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	uri := r.URL.Path
	method := r.Method
	c := NewContext(w, r)
	c.engine = e
	if _, ok := e.trees[method]; !ok {
		e.HandleNotFound(c, e.middleware)
		return
	}
	nodes := e.trees[method].Find(uri, false)
	if len(nodes) > 0 {
		node := nodes[0]
		if node.handle != nil {
			if node.pattern == uri {
				handle(c, node.handle, node.middleware)
				return
			}
			if node.pattern == uri[1:] {
				handle(c, node.handle, node.middleware)
				return
			}
		}
	}
	if len(nodes) == 0 {
		res := strings.Split(uri, "/")
		prefix := res[1]
		nodes := e.trees[method].Find(prefix, true)
		for _, node := range nodes {
			if handler := node.handle; handler != nil && node.pattern != uri {
				if parameters, ok := e.parse(uri, node.pattern); ok {
					r = r.WithContext(context.WithValue(r.Context(), struct{}{}, parameters))
					c.Parameters = parameters
					c.Request = r
					handle(c, handler, node.middleware)
					return
				}
			}
		}

	}
	e.HandleNotFound(c, e.middleware)
}

func (e *Engine) HandleNotFound(c *Context, middleware []Middleware) {
	if e.notFound != nil {
		handle(c, e.notFound, middleware)
		return
	}
	http.NotFound(c.Writer, c.Request)
}

func handle(c *Context, handler Handler, middlewares []Middleware) {
	var runHandler = handler
	for _, middleware := range middlewares {
		runHandler = middleware(runHandler)
	}
	runHandler(c)
}

func (e *Engine) SetFuncMap(funcMap template.FuncMap) {
	e.FuncMap = funcMap
}

func (e *Engine) LoadHtmlGlob(pattern string) {
	e.Template = template.Must(template.New("").Funcs(e.FuncMap).ParseGlob(pattern))
}

func (e *Engine) Run(address string) (err error) {
	defer func() { Debug("[Trie-debug] %v", err) }()
	Debug("Listening and serving HTTP on %v\n", address)
	return http.ListenAndServe(address, e)
}

func (e *Engine) RunTLS(address, cert, key string) (err error) {
	defer func() { Debug("[Trie-debug] %v", err) }()
	Debug("Listening and serving HTTPS on %v\n", address)
	Debug("Cert file: %v\n", cert)
	Debug("Key file: %v\n", key)
	err = http.ListenAndServeTLS(address, cert, key, e)
	return err
}
