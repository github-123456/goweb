package goweb

import "regexp"

type RouterGroup struct {
	engine   *Engine
	Handlers HandlersChain
}

func (group *RouterGroup) Group() *RouterGroup {
	return &RouterGroup{
		Handlers: group.Handlers,
		engine:   group.engine,
	}
}

func (group *RouterGroup) GET(path string, handler HandlerFunc) {
	group.engine.trees = append(group.engine.trees, methodTree{"GET", &node{path: path, handlers: append(group.Handlers, handler)}})
}
func (group *RouterGroup) POST(path string, handler HandlerFunc) {
	group.engine.trees = append(group.engine.trees, methodTree{"POST", &node{path: path, handlers: append(group.Handlers, handler)}})
}
func (group *RouterGroup) RegexMatch(regexp *regexp.Regexp, handler HandlerFunc) {
	group.engine.trees = append(group.engine.trees, methodTree{"GET", &node{regexp: regexp, handlers: append(group.Handlers, handler)}})
}
func (group *RouterGroup) Use(middleware ...HandlerFunc) {
	group.Handlers = append(group.Handlers, middleware...)
}
