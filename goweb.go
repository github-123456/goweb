package goweb

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

var outlog = log.New(os.Stdout, fmt.Sprintf("INFO [%s] ", "GOWEB"), log.Ldate|log.Ltime|log.Lshortfile)
var errlog = log.New(os.Stderr, fmt.Sprintf("ERROR [%s] ", "GOWEB"), log.Ldate|log.Ltime|log.Lshortfile)

type Engine struct {
	ErrorPageFunc
	RouterGroup
	trees             []methodTree
	ConcurrenceNumSem chan int
}

func Default() *Engine {
	engine := Engine{
	}
	engine.RouterGroup.engine = &engine
	engine.ConcurrenceNumSem = make(chan int, 5)
	return &engine
}

type HandlerFunc func(*Context)

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(1 * time.Second)
		timeout <- true
	}()
	context := &Context{Engine: engine, Request: req, Writer: w, CT: time.Now(), Signal: make(chan int),Data:make(map[string]interface{})}
	select {
	case engine.ConcurrenceNumSem <- 1:
		path := context.Request.URL.Path
		var handler HandlerFunc
		for _, v := range engine.trees {
			if v.root.path == path || v.root.regexp != nil && v.root.regexp.MatchString(path) {
				if v.method == context.Request.Method {
					handler = v.root.handler
					break
				}
			}
		}
		if handler != nil {
			context.Request.ParseForm()
			safelyHandle(handler, context)
		} else {
			if context.Request.Method == "GET" {
				if engine.ErrorPageFunc == nil {
					http.NotFound(context.Writer, context.Request)
				} else {
					context.ShowErrorPage(http.StatusNotFound, "page not found")
				}
			} else {
				context.Failed(fmt.Sprintf("%s", "404"))
			}
		}
		<-engine.ConcurrenceNumSem
	case <-timeout:
		context.ShowErrorPage(http.StatusBadGateway, "server overload")
	}
}

func safelyHandle(hf HandlerFunc, c *Context) {
	outlog.Println(fmt.Sprintf("start processing request->ip:%s path：%s", c.Request.RemoteAddr, c.Request.RequestURI))
	defer func() {
		outlog.Println(fmt.Sprintf("end processing request->ip:%s path：%s", c.Request.RemoteAddr, c.Request.URL.Path))
		if err := recover(); err != nil {
			errlog.Println(err)
			if c.Request.Method == "GET" {
				if c.Engine.ErrorPageFunc == nil {
					http.NotFound(c.Writer, c.Request)
				} else {
					c.ShowErrorPage(http.StatusBadGateway, "server error")
				}
			} else {
				c.Failed(fmt.Sprintf("%s", err))
			}

		}
	}()
	hf(c)
}
