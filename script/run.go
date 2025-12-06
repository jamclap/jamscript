package script

import (
	"log"
	"reflect"
)

func (r *runner) Run(m *Module) {
	r.module = m
	r.reflectArgs = r.reflectArgs[:0]
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

// TODO Separate runner per coroutine?
type runner struct {
	module      *Module
	reflectArgs []reflect.Value
	stack       []any
}

func (r *runner) runNode(node Node) any {
	// log.Printf("run node: %+v %T\n", node, node)
	switch n := node.(type) {
	case *Call:
		return r.runCall(n)
	case *Ref:
		return r.runRef(n)
	case *TokenNode:
		return r.runToken(n)
	}
	return nil
}

func (r *runner) runCall(c *Call) any {
	// println("call")
	callee := r.runNode(c.Callee)
	f, ok := callee.(*Fun)
	if !ok {
		log.Println("callee not fun")
		return nil
	}
	start := len(r.stack)
	for _, a := range c.Args {
		r.stack = append(r.stack, r.runNode(a))
	}
	r.runFun(f, start)
	r.stack = r.stack[:start]
	return nil
}

func (r *runner) runFun(f *Fun, argsStart int) any {
	if len(f.Kids) == 1 {
		v := f.Kids[0]
		t := reflect.TypeOf(v)
		if t.Kind() == reflect.Func {
			if len(r.stack)-argsStart != reflect.TypeOf(v).NumIn() {
				log.Printf("reflect fun: %+v %d\n", v, reflect.TypeOf(v).NumIn())
				return nil
			}
			for i := argsStart; i < len(r.stack); i++ {
				r.reflectArgs = append(r.reflectArgs, reflect.ValueOf(r.stack[i]))
			}
			// TODO Specialize for certain kinds of funs to reduce allocs?
			results := reflect.ValueOf(v).Call(r.reflectArgs)
			var result any = nil
			if len(results) > 0 {
				result = results[0].Interface()
			}
			r.reflectArgs = r.reflectArgs[:0]
			return result
		}
	}
	for _, k := range f.Kids {
		r.runNode(k)
	}
	return nil
}

func (r *runner) runRef(ref *Ref) any {
	switch d := ref.Node.(type) {
	case *Fun:
		return d
	}
	return nil
}

func (r *runner) runToken(t *TokenNode) any {
	switch t.Kind {
	case TokenStringText:
		return t.Text
	default:
		log.Printf("t: %v\n", t)
	}
	return nil
}
