package goweb

import (
	"net/http"
	"time"
)

type Context struct {
	Engine  *Engine
	Request *http.Request
	Writer  http.ResponseWriter
	CT      time.Time
}

func (c *Context) Success(data interface{}) {
	HandlerResult{Data: data}.Write(c.Writer)
}
func (c *Context) Failed(error string) {
	HandlerResult{Error: error}.Write(c.Writer)
}

type ErrorPageFunc func(c *Context, status int)

func (c *Context) ShowErrorPage(status int) {
	if c.Engine == nil {
		c.Writer.WriteHeader(status)
	} else {
		c.Engine.ErrorPageFunc(c, status)
	}
}
