package goweb

type WidgetManager struct {
	HandlerWidget HandlerWidget
}

func NewWidgetManager() *WidgetManager {
	var wm = &WidgetManager{}
	wm.HandlerWidget = &DefaultHanderWidget{}
	return wm
}

type HandlerWidget interface {
	Pre_Process(ctx *Context)
	Post_Process(ctx *Context)
}
type DefaultHanderWidget struct {
}

func (w *DefaultHanderWidget) Pre_Process(ctx *Context) {
	outlog.Println("start processing request ->", ctx)
}

func (w *DefaultHanderWidget) Post_Process(ctx *Context) {
	outlog.Println("end processing request ->", ctx)
}
