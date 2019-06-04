package goweb

import (
	"fmt"
	"net/http"
	"html/template"
)

func RenderPage(w http.ResponseWriter,data interface{}, filenames ...string) {
	tmpl,err:= template.ParseFiles(filenames...)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}
	err= tmpl.Execute(w, data)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}
}