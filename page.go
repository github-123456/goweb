package goweb

import (
	"fmt"
	"html/template"
)

type PageModel interface {
	Prepare(c *Context) interface{}
}

func RenderPage(c *Context,pageModel PageModel, filenames ...string) {
	tmpl,err:= template.ParseFiles(filenames...)
	if err != nil {
		fmt.Fprintf(c.Writer, err.Error())
		return
	}
	pageModel.Prepare(c)
	err= tmpl.Execute(c.Writer, pageModel)
	if err != nil {
		errlog.Println(err)
		return
	}
}