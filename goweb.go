package goweb

import (
	"net/http"
	"time"
)

type Engine struct {
	ErrorPageFunc
	RouterGroup
	trees []methodTree
}

func New() *Engine {
	engine := Engine{
	}
	engine.RouterGroup.engine = &engine
	return &engine
}

type HandlerFunc func(*Context)

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	var handler HandlerFunc
	for _, v := range engine.trees {
		if v.root.path == path || v.root.regexp != nil && v.root.regexp.MatchString(path) {
			if v.method == req.Method {
				handler = v.root.handler
				break
			}
		}
	}
	context := &Context{Engine: engine, Request: req, Writer: w, CT: time.Now()}
	if handler != nil {
		req.ParseForm()
		handler(context)
	} else {
		if engine.ErrorPageFunc == nil {
			http.NotFound(w, req)
		} else {
			context.ShowErrorPage(http.StatusNotFound)
		}
	}
}
