package trie

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"net/http"

	"gopkg.in/yaml.v2"
)

type Render interface {
	Render(http.ResponseWriter) error
}

type Data struct {
	ContentType string
	Data        []byte
}

type String struct {
	Format string
	Data   interface{}
}

type Json struct {
	Data interface{}
}

type Jsonp struct {
	Callback string
	Data     interface{}
}

type Xml struct {
	Data interface{}
}

type Yaml struct {
	Data interface{}
}

type Html struct {
	Template *template.Template
	Name     string
	Data     interface{}
}

func (d Data) Render(writer http.ResponseWriter) error {
	_, err := writer.Write(d.Data)
	return err
}

func (s String) Render(writer http.ResponseWriter) error {
	_, err := writer.Write([]byte(fmt.Sprintf(s.Format, s.Data)))
	return err
}

func (j Json) Render(writer http.ResponseWriter) (err error) {
	fmt.Println(writer.Header())
	return json.NewEncoder(writer).Encode(j.Data)
}

func (jp Jsonp) Render(writer http.ResponseWriter) error {
	ret, err := json.Marshal(jp.Data)
	if err != nil {
		return err
	}
	if jp.Callback == "" {
		_, err = writer.Write(ret)
		return err
	}
	callback := template.JSEscapeString(jp.Callback)
	_, err = writer.Write([]byte(callback))
	if err != nil {
		return err
	}
	_, err = writer.Write([]byte("("))
	if err != nil {
		return err
	}
	_, err = writer.Write(ret)
	if err != nil {
		return err
	}
	_, err = writer.Write([]byte(");"))
	if err != nil {
		return err
	}
	return nil
}

func (x Xml) Render(writer http.ResponseWriter) error {
	return xml.NewEncoder(writer).Encode(x.Data)
}

func (y Yaml) Render(writer http.ResponseWriter) error {
	data, err := yaml.Marshal(y.Data)
	if err != nil {
		return nil
	}
	_, err = writer.Write(data)
	return err
}

func (h Html) Render(writer http.ResponseWriter) error {
	fmt.Println()
	if h.Name == "" {
		return h.Template.Execute(writer, h.Data)
	}
	return h.Template.ExecuteTemplate(writer, h.Name, h.Data)
}
