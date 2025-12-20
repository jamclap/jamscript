package rio

import (
	"errors"
	"fmt"
	"log"
	"reflect"
)

func (r *runner) Run(m *Module) (err error) {
	r.module = m
	r.reflectArgs = r.reflectArgs[:0]
	r.returnKind = TokenNone
	r.stack = r.stack[:0]
	r.levels = append(r.levels[:0], runLevel{})
	main, ok := m.Tops["main"]
	if !ok {
		return errors.New("no main")
	}
	mainFun, ok := main.(*Fun)
	if !ok {
		// TODO Also check sig.
		return errors.New("main not a function")
	}
	// TODO Push sys.
	defer func() {
		if rec := recover(); rec != nil {
			// log.Println(rec)
			err = fmt.Errorf("%v", rec)
		}
	}()
	r.runFun(mainFun)
	return
}

// TODO Separate runner per coroutine?
type runner struct {
	levels      []runLevel
	module      *Module
	reflectArgs []reflect.Value
	returnKind  TokenKind
	stack       []any
}

type runLevel struct {
	stackStart int
}

func (r *runner) levelStart() int {
	return last(&r.levels).stackStart
}

func (r *runner) popLevel() {
	start := r.levelStart()
	r.stack = r.stack[:start]
	pop(&r.levels)
	// fmt.Printf("popLevel r.levels: %+v\n", r.levels)
}

func (r *runner) pushLevel(stackStart int) {
	// TODO Track the statically outer function start for each nested fun?
	// TODO Part of the state of the runtime closure object?
	level := runLevel{stackStart: stackStart}
	r.levels = append(r.levels, level)
	// fmt.Printf("pushLevel r.levels: %+v %+v\n", r.levels, r.stack)
}

func (r *runner) runNode(node Node) any {
	// log.Printf("run node: %+v %T\n", node, node)
	switch n := node.(type) {
	case *Call:
		return r.runCall(n)
	case *Get:
		return r.runGet(n)
	case *Ref:
		return r.runRef(n)
	case *Return:
		return r.runReturn(n)
	case *Switch:
		return r.runSwitch(n)
	case *Value:
		return r.runValue(n)
	case *Var:
		return r.runVar(n)
	}
	return nil
}

func (r *runner) runBlockKids(kids []Node) any {
	var value any
	for _, k := range kids {
		value = r.runNode(k)
		if r.returnKind != TokenNone {
			return value
		}
	}
	return value
}

func (r *runner) runCall(c *Call) any {
	// fmt.Printf("args for f.Name: %v\n", f.Name)
	stackStart := len(r.stack)
	// println("call")
	var callee any
	switch calleeNode := c.Callee.(type) {
	case *Get:
		var subject any
		// Split these out to prevent binding allocation.
		subject = r.runNode(calleeNode.Subject)
		callee = r.runNode(calleeNode.Member)
		// log.Printf("subject: %v\n", subject)
		// log.Printf("callee: %v\n", callee)
		r.stack = append(r.stack, subject)
	default:
		callee = r.runNode(c.Callee)
	}
	f, ok := callee.(*Fun)
	if !ok {
		panic("callee not fun")
	}
	// TODO How to handle nested funs and captures right?
	for _, a := range c.Args {
		arg := r.runNode(a)
		// log.Printf("arg: %v\n", arg)
		r.stack = append(r.stack, arg)
	}
	r.pushLevel(stackStart)
	// fmt.Printf("call f.Name: %v %v %+v\n", f.Name, stackStart, r.stack)
	value := r.runFun(f)
	// log.Printf("return value: %v\n", value)
	r.popLevel()
	return value
}

func (r *runner) runFun(f *Fun) any {
	// fmt.Printf("runFun f.Name: %v\n", f.Name)
	levelStart := r.levelStart()
	argCount := len(r.stack) - levelStart
	if len(f.Kids) == 1 {
		v := f.Kids[0]
		t := reflect.TypeOf(v)
		if t.Kind() == reflect.Func {
			switch f2 := v.(type) {
			case func(int32, int32) bool:
				if argCount != 2 {
					panic("bad arg count")
				}
				i, ok := r.stack[len(r.stack)-2].(int32)
				if !ok {
					panic("bad arg type")
				}
				j, ok := r.stack[len(r.stack)-1].(int32)
				if !ok {
					panic("bad arg type")
				}
				return f2(i, j)
			case func(int32, int32) int32:
				if argCount != 2 {
					panic("bad arg count")
				}
				i, ok := r.stack[len(r.stack)-2].(int32)
				if !ok {
					panic("bad arg type")
				}
				j, ok := r.stack[len(r.stack)-1].(int32)
				if !ok {
					panic("bad arg type")
				}
				return f2(i, j)
			case func(any):
				if argCount != 1 {
					panic("bad arg count")
				}
				f2(r.stack[len(r.stack)-1])
				return nil
			}
			if argCount != reflect.TypeOf(v).NumIn() {
				panic(fmt.Sprintf("reflect fun: %+v %d\n", v, reflect.TypeOf(v).NumIn()))
			}
			for i := levelStart; i < len(r.stack); i++ {
				r.reflectArgs = append(r.reflectArgs, reflect.ValueOf(r.stack[i]))
			}
			// log.Printf("r.stack: %v\n", r.stack)
			// log.Printf("r.reflectArgs: %v\n", r.reflectArgs)
			// TODO Specialize for certain kinds of funs to reduce allocs?
			results := reflect.ValueOf(v).Call(r.reflectArgs)
			var result any = nil
			if len(results) > 0 {
				result = results[0].Interface()
			}
			// log.Printf("result: %v\n", result)
			r.reflectArgs = r.reflectArgs[:0]
			return result
		}
	}
	// TODO Check arg count.
	for _, k := range f.Kids {
		value := r.runNode(k)
		// TODO Break returns should have been handled before here.
		if r.returnKind != TokenNone {
			// log.Printf("returning value: %v\n", value)
			r.returnKind = TokenNone
			return value
		}
	}
	return nil
}

func (r *runner) runGet(g *Get) any {
	subject := r.runNode(g.Subject)
	member := r.runNode(g.Member)
	// TODO If member is a method, bind subject here?
	_ = subject
	return member
}

func (r *runner) runRef(ref *Ref) any {
	switch d := ref.Target.(type) {
	case *Fun:
		return d
	case *Var:
		// fmt.Printf("d.Name: %v at %v+%v\n", d.Name, start, d.Offset)
		// fmt.Printf("ref r.levels: %+v %+v\n", r.levels, r.stack)
		value := r.stack[r.levelStart()+d.Offset]
		// fmt.Printf("var value: %v\n", value)
		return value
	}
	return nil
}

func (r *runner) runReturn(ret *Return) any {
	value := r.runNode(ret.Value)
	r.returnKind = TokenReturn
	// log.Printf("runReturn value: %+v\n", value)
	return value
}

func (r *runner) runSwitch(n *Switch) any {
	subject := n.Subject
	switch subject {
	case nil:
		subject = true
	default:
		log.Printf("support switch subject: %+v\n", subject)
		return nil
	}
Cases:
	for _, k := range n.Kids {
		c, ok := k.(*Case)
		if !ok {
			log.Printf("not case: %+v\n", k)
			continue Cases
		}
		// fmt.Printf("c: %+v\n", c)
		matched := false
		switch {
		case c.Always:
			matched = true
		default:
		Patterns:
			for _, p := range c.Patterns {
				value := r.runNode(p)
				if value == subject {
					matched = true
					break Patterns
				}
			}
			// TODO Require guard also, if not nil.
		}
		if matched {
			// println("matches")
			var value = r.runBlockKids(c.Kids)
			// log.Printf("switch value: %v\n", value)
			return value
		}
	}
	// log.Println("No switch value")
	return nil
}

func (r *runner) runValue(value *Value) any {
	return value.Value
}

func (r *runner) runVar(v *Var) any {
	value := r.runNode(v.Value)
	// fmt.Printf("v: %v %+v\n", v.Name, value)
	// log.Printf("runVar value: %v\n", value)
	r.stack = append(r.stack, value)
	// The var statement itself has value nil.
	return nil
}
