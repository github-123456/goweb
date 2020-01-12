package goweb

import (
	"fmt"
	"net/http"
	"time"
)

type Context struct {
	Engine   *Engine
	Request  *http.Request
	Writer   http.ResponseWriter
	CT       time.Time
	Signal   chan int
	Data     map[string]interface{}
	index    int
	handlers HandlersChain
	Ok       bool
}

func (c *Context) Next() {
	c.index++
	for c.index < len(c.handlers) {
		c.handlers[c.index](c)
		c.index++
	}
}

func (c *Context) Abort() {
	c.index = 10000000000000
}

func (c *Context) Success(data interface{}) {
	HandlerResult{Data: data}.Write(c.Writer)
}
func (c *Context) Failed(error string) {
	HandlerResult{Error: error}.Write(c.Writer)
}

type ErrorPageFunc func(c *Context, status int, msg string)

func (c *Context) ShowErrorPage(status int, msg string) {
	//must write header before write body
	c.Writer.WriteHeader(status)
	if c.Engine.ErrorPageFunc == nil {
		c.Writer.Write([]byte(msg))
	} else {
		c.Engine.ErrorPageFunc(c, status, msg)
	}
	c.Ok = false
}

func (c *Context) String() string {
	return fmt.Sprintf("method:%s path:%s remote_ip:%s", c.Request.Method, c.Request.URL.Path, c.Request.RemoteAddr)
}
