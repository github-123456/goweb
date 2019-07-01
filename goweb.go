package goweb

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
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

type HandlerFunc func(ctx *Context)
type HandlersChain []HandlerFunc

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(1 * time.Second)
		timeout <- true
	}()
	context := &Context{Engine: engine, Request: req, Writer: w, CT: time.Now(), Signal: make(chan int), Data: make(map[string]interface{})}
	context.index = -1
	select {
	case engine.ConcurrenceNumSem <- 1:
		path := context.Request.URL.Path
		var handlers HandlersChain
		for _, v := range engine.trees {
			if v.root.path == path || v.root.regexp != nil && v.root.regexp.MatchString(path) {
				if v.method == context.Request.Method {
					handlers = v.root.handlers
					break
				}
			}
		}
		if handlers != nil {
			context.Request.ParseForm()
			context.handlers = handlers
			safelyHandle(context)
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

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (g gzipResponseWriter) Write(b []byte) (int, error) {
	if g.ResponseWriter.Header().Get("Content-Type") == "" {
		g.ResponseWriter.Header().Set("Content-Type", http.DetectContentType(b))
	}
	return g.Writer.Write(b)
}

func safelyHandle(c *Context) {
	if strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") && c.Request.Header.Get("Connection") != "Upgrade" {
		c.Writer.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(c.Writer)
		defer gz.Close()
		w := gzipResponseWriter{Writer: gz, ResponseWriter: c.Writer}
		c.Writer = w
	}
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
	c.Next()
}
