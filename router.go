package goweb
import "regexp"
type RouterGroup struct {
	engine   *Engine
}

func (group *RouterGroup) GET(relativePath string, handler HandlerFunc) {
	group.engine.trees=append(group.engine.trees,methodTree{"GET",&node{path:relativePath,handler:handler}})
}
func (group *RouterGroup) RegexMatch(regexp * regexp.Regexp, handler HandlerFunc) {
	group.engine.trees=append(group.engine.trees,methodTree{"GET",&node{regexp:regexp, handler:handler}})
}