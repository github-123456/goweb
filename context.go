package goweb

import "net/http"

type Context struct{
	Request   *http.Request
	Writer    http.ResponseWriter
}