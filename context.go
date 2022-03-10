package trie

import (
	"net/http"
	"net/url"
	"reflect"
)

type Context struct {
	Writer     http.ResponseWriter
	Request    *http.Request
	Uri        string
	Method     string
	Parameters map[string]string
	StatusCode int
	index      int
	engine     *Engine
}

func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Uri:     r.URL.Path,
		Method:  r.Method,
		Request: r,
		Writer:  w,
		index:   -1,
	}
}

func (c *Context) Param(key string) string {
	return c.Parameters[key]
}

func (c *Context) Post(key string) string {
	return c.Request.FormValue(key)
}

func (c *Context) Get(key string) string {
	return c.Request.URL.Query().Get(key)
}

func (c *Context) Cookie(key string) (string, error) {
	cookie, err := c.Request.Cookie(key)
	if err != nil {
		return "", err
	}
	return url.QueryUnescape(cookie.Value)
}

func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

func (c *Context) Data(code int, contentType string, data []byte) {
	c.Render(code, Data{ContentType: contentType, Data: data})
}

func (c *Context) String(code int, format string, v ...interface{}) {
	c.Render(code, String{Format: format, Data: v})
}

func (c *Context) Json(code int, v interface{}) {
	c.Render(code, Json{Data: v})
}

func (c *Context) Jsonp(code int, v interface{}) {

}

func (c *Context) Xml(code int, v interface{}) {
	c.Render(code, Xml{Data: v})
}

func (c *Context) Yaml(code int, v interface{}) {
	c.Render(code, Yaml{Data: v})
}

func (c *Context) Html(code int, name string, v interface{}) {
	c.Render(code, Html{Template: c.engine.Template, Name: name, Data: v})
}

func (c *Context) contentType(v string) {
	c.Writer.Header().Set("Content-Type", v)
}

func (c *Context) Render(code int, r Render) {
	var contentType string
	reflectorType := reflect.TypeOf(r)
	reflectorValue := reflect.ValueOf(r)
	name := reflectorType.Name()
	switch name {
	case "Data":
		contentType = reflectorValue.FieldByName("ContentType").String()
	case "String":
		contentType = "text/plain; charset=utf-8"
	case "Json":
		contentType = "application/json; charset=utf-8"
	case "Jsonp":
		contentType = "application/javascript; charset=utf-8"
	case "Xml":
		contentType = "application/xml; charset=utf-8"
	case "Yaml":
		contentType = "application/x-yaml; charset=utf-8"
	case "Html":
		contentType = "text/html; charset=utf-8"
	default:
		panic("incorrect rendering format")
	}
	c.contentType(contentType)
	c.Status(code)
	if err := r.Render(c.Writer); err != nil {
		panic(err)
	}
}
