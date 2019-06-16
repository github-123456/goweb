package goweb

import (
	"encoding/json"
	"net/http"
)

type HandlerResult struct{
	Error string `json:"error"`
	Data interface{} `json:"data"`
}

func (hr HandlerResult)Write(w http.ResponseWriter)  {
	json, err := json.Marshal(hr)
	if err!=nil{
		panic(err)
	}
	w.Header().Add("Content-Type","application/json")
	w.Write(json)
}

func SanitizeHtml(html string)string {
	return bluemondayPolicy.Sanitize(html)
}
