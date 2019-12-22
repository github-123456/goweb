package goweb

import (
	"compress/gzip"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/swishcloud/gostudy/logger"
)

var outlog = logger.NewLogger(os.Stdout, "GOWEB INFO")
var errlog = logger.NewLogger(os.Stderr, "GOWEB ERROR")

type Engine struct {
	ErrorPageFunc
	RouterGroup
	trees             []methodTree
	ConcurrenceNumSem chan int
	WM                *WidgetManager
}

func Default() *Engine {
	engine := Engine{}
	engine.RouterGroup.engine = &engine
	engine.ConcurrenceNumSem = make(chan int, 5)
	engine.WM = NewWidgetManager()
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
			safelyHandle(engine, context)
		} else {
			if context.Request.Method == "GET" {
				if engine.ErrorPageFunc == nil {
					http.NotFound(context.Writer, context.Request)
				} else {
					context.ShowErrorPage(http.StatusNotFound, "page not found")
				}
			} else {
				context.Failed(fmt.Sprintf("%s", string(http.StatusNotFound)))
			}
		}
		<-engine.ConcurrenceNumSem
	case <-timeout:
		context.ShowErrorPage(http.StatusBadRequest, "server overload")
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

func safelyHandle(engine *Engine, c *Context) {
	engine.WM.HandlerWidget.Pre_Process(c)
	defer func() {
		if err := recover(); err != nil {
			err_desc := fmt.Sprintf("%s", err)
			errlog.Println(err)
			if c.Request.Method == "GET" {
				c.ShowErrorPage(http.StatusInternalServerError, err_desc)
			} else {
				c.Failed(fmt.Sprintf("%s", err))
			}

		}
		engine.WM.HandlerWidget.Post_Process(c)
	}()
	if strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") && c.Request.Header.Get("Connection") != "Upgrade" {
		c.Writer.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(c.Writer)
		defer gz.Close()
		w := gzipResponseWriter{Writer: gz, ResponseWriter: c.Writer}
		c.Writer = w
	}
	c.Next()
}

func (ctx *Context) RenderPage(data interface{}, filenames ...string) {
	tmpl, err := template.ParseFiles(filenames...)
	if err != nil {
		errlog.Println(err)
		ctx.Writer.Write([]byte(fmt.Sprintf("%s", err)))
		return
	}
	err = tmpl.Execute(ctx.Writer, data)
	if err != nil {
		errlog.Println(err)
		return
	}
}
