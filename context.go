package goweb

import (
	"compress/gzip"
	"fmt"
	"net/http"
	"time"
)

type Context struct {
	Engine     *Engine
	Request    *http.Request
	Writer     *ResponseWriter
	CT         time.Time
	Signal     chan int
	Data       map[string]interface{}
	index      int
	handlers   HandlersChain
	Ok         bool
	StatusCode int
	FuncMap    map[string]interface{}
	Err        error
}
type ResponseWriter struct {
	http.ResponseWriter
	gz          *gzip.Writer
	ctx         *Context
	Compress    bool
	initialized bool
}

func (w *ResponseWriter) EnsureInitialzed(compress bool) {
	if !w.initialized {
		w.Compress = compress
		if compress {
			w.ResponseWriter.Header().Set("Content-Encoding", "gzip")
			w.gz = gzip.NewWriter(w.ResponseWriter)

		}
		w.initialized = true
	}
}
func (w *ResponseWriter) Close() {
	if w.gz != nil {
		w.gz.Close()
	}
}
func (w *ResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}
func (w *ResponseWriter) Write(b []byte) (int, error) {
	w.EnsureInitialzed(false)
	if w.ResponseWriter.Header().Get("Content-Type") == "" {
		w.ResponseWriter.Header().Set("Content-Type", http.DetectContentType(b))
	}
	if w.gz != nil {
		return w.gz.Write(b)
	}
	return w.ResponseWriter.Write(b)
}
func (w *ResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.ctx.StatusCode = statusCode
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
	HandlerResult{Error: &error}.Write(c.Writer)
}

type ErrorPageFunc func(c *Context, status int, msg string)

func (c *Context) ShowErrorPage(status int, msg string) {
}

func (c *Context) String() string {
	return fmt.Sprintf("method:%s path:%s remote_ip:%s", c.Request.Method, c.Request.URL.Path, c.Request.RemoteAddr)
}
