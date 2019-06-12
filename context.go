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
	Signal chan int
	Data map[string]interface{}
}

func (c *Context) Success(data interface{}) {
	HandlerResult{Data: data}.Write(c.Writer)
}
func (c *Context) Failed(error string) {
	HandlerResult{Error: error}.Write(c.Writer)
}

type ErrorPageFunc func(c *Context, status int,msg string)

func (c *Context) ShowErrorPage(status int,msg string) {
	if c.Engine.ErrorPageFunc == nil {
		c.Writer.WriteHeader(status)
	} else {
		c.Engine.ErrorPageFunc(c, status,msg)
	}
}
