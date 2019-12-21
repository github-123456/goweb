package goweb

type WidgetManager struct {
	HandlerWidgets []HandlerWidget
}

func NewWidgetManager() *WidgetManager {
	var wm = &WidgetManager{}
	wm.HandlerWidgets = []HandlerWidget{}
	return wm
}

type HandlerWidget interface {
	Pre_Process(ctx *Context)
	Post_Process(ctx *Context)
}

func (wm *WidgetManager) PreProcessHandler(ctx *Context) {
	for _, v := range wm.HandlerWidgets {
		v.Post_Process(ctx)
	}
}

func (wm *WidgetManager) PostProcessHandler(ctx *Context) {
	for _, v := range wm.HandlerWidgets {
		v.Post_Process(ctx)
	}
}
