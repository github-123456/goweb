package goweb

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/swishcloud/gostudy/logger"
)

var outlog = logger.NewLogger(os.Stdout, "GOWEB INFO")
var errlog = logger.NewLogger(os.Stderr, "GOWEB ERROR")

type Engine struct {
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
	context := &Context{Engine: engine, Request: req, CT: time.Now(), Signal: make(chan int), Ok: true, Data: make(map[string]interface{}), FuncMap: map[string]interface{}{}}
	context.Writer = &ResponseWriter{ResponseWriter: w, ctx: context}
	context.index = -1
	context.FuncMap["formatTime"] = func(t time.Time, layout string) (string, error) {
		if layout == "" {
			layout = "01/02/2006 15:04"
		}
		tom := 0
		c, err := context.Request.Cookie("tom")
		if err == nil {
			tom, err = strconv.Atoi(c.Value)
			if err != nil {
				panic(err)
			}
		}
		t = t.Add(-time.Duration(int64(time.Minute) * int64(tom)))
		return t.Format(layout), nil
	}
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
		context.handlers = handlers
		safelyHandle(engine, context)
		<-engine.ConcurrenceNumSem
	case <-timeout:
		context.ShowErrorPage(http.StatusBadRequest, "server overload")
	}
}
func safelyHandle(engine *Engine, c *Context) {
	engine.WM.HandlerWidget.Pre_Process(c)
	defer func() {
		if err := recover(); err != nil {
			c.Ok = false
			err_desc := fmt.Sprintf("%s", err)
			c.Err = errors.New(err_desc)
			errlog.Println(err)
		}
		engine.WM.HandlerWidget.Post_Process(c)
		c.Writer.Close()
	}()
	if c.handlers == nil {
		c.Err = errors.New("page not found")
	} else {
		err := c.Request.ParseForm()
		if err != nil {
			panic(err)
		}
		c.Next()
	}
}

func (ctx *Context) RenderPage(data interface{}, filenames ...string) {
	tmpl := template.New(path.Base(filenames[0])).Funcs(ctx.FuncMap)
	tmpl, err := tmpl.ParseFiles(filenames...)
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
