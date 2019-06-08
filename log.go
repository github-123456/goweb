package goweb

import "fmt"

type Log interface {
	Info(a interface{})
	Error(a interface{})
}

type Logger struct {
	Name string
}

func (l Logger)Info(a interface{})  {
	fmt.Println(fmt.Sprintf("Info[%s]:%s",l.Name,a))
}
func (l Logger)Error(a interface{}) {
	fmt.Println(fmt.Sprintf("Error[%s]:%s", l.Name, a))
}