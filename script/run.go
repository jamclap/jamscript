package script

import "log"

func (r *runner) Run(m *Module) {
	r.module = m
	r.stack = r.stack[:0]
	main, ok := m.Tops["main"]
	if !ok {
		log.Println("no main")
		return
	}
	mainFun, ok := main.(*Fun)
	if !ok {
		// TODO Also check sig.
		log.Println("main not a function")
		return
	}
	r.runFun(mainFun)
}

type runner struct {
	module *Module
	stack  []any
}

func (r *runner) runFun(mainFun *Fun) {
	println("running now")
}
