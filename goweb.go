package goweb

import "net/http"

type Engine struct {
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
	if handler != nil {
		handler(&Context{Request: req, Writer: w})
	} else {
		http.NotFound(w, req)
	}
}
