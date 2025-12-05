package script

import (
	"fmt"
	"log"
	"reflect"
)

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
	// TODO Push sys.
	r.runFun(mainFun, len(r.stack))
}

type runner struct {
	module *Module
	stack  []any
}

func (r *runner) runNode(node Node) any {
	fmt.Printf("run node: %+v %T\n", node, node)
	switch n := node.(type) {
	case *Call:
		r.runCall(n)
	case *Ref:
		println("ref")
	}
	return nil
}

func (r *runner) runCall(c *Call) {
	print("call")
	callee := r.runNode(c.Callee)
	f, ok := callee.(*Fun)
	if !ok {
		println("callee not fun")
		return
	}
	r.runFun(f, len(r.stack))
}

func (r *runner) runFun(f *Fun, argsStart int) {
	if len(f.Kids) == 1 {
		if v, ok := f.Kids[0].(reflect.Value); ok {
			fmt.Printf("reflect: %+v\n", v)
		}
	}
	for _, k := range f.Kids {
		r.runNode(k)
	}
}
