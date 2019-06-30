package goweb

import "regexp"

type methodTree struct {
	method string
	root   *node
}
type node struct {
	path      string
	regexp *regexp.Regexp
	handlers  HandlersChain
}