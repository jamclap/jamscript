package script

import (
	"log"
)

type Engine struct {
	// Types map[Type]Type // TODO or use unique.Make(type) instead?
	lexer       lexer
	parser      parser
	resolver    resolver
	runner      runner
	treeBuilder treeBuilder
	typer       typer
}

func NewEngine() *Engine {
	return &Engine{
		treeBuilder: newTreeBuilder(),
	}
}

func (e *Engine) Process(source string) *Module {
	tokens := e.lexer.Lex(source)
	// for _, token := range tokens {
	// 	fmt.Printf("token: %v\n", token)
	// }
	parseTree := e.parser.Parse(tokens)
	// parseTree.Print(os.Stdout)
	module := e.treeBuilder.Norm(parseTree)
	module.Core["log"] = doLog
	// module.Print(os.Stdout)
	e.analyze(module)
	// module.Print(os.Stdout)
	return module
}

func (e *Engine) Run(m *Module) error {
	return e.runner.Run(m)
}

func (e *Engine) analyze(module *Module) {
	// TODO Track changes so we can know if more rounds are needed.
	// TODO What's a good max?
	e.resolver.core = module.Core
	for i := 0; i < 5; i++ {
		// If stable, this shouldn't allocate more on each iteration.
		e.resolver.Resolve(module)
		e.typer.Type(module)
		// TODO Include inlining/macro run phase in the loop?
	}
}

var doLog = &Fun{
	Def: Def{
		Name: "log",
	},
	Type: FunType{
		ParamTypes: []Type{TypeAny},
		RetType:    TypeVoid,
	},
	Kids: []Node{func(a any) {
		// TODO Option to select where `log` goes?
		// TODO Specialized formatting for records and more.
		// TODO Call toString() with some fallback for classes.
		log.Println(a)
	}},
}

var intAdd = &Fun{
	Def: Def{
		Name: "add",
	},
	Type: FunType{
		ParamTypes: []Type{TypeInt, TypeInt},
		RetType:    TypeInt,
	},
	Kids: []Node{func(i, j int32) int32 { return i + j }},
}

var intEq = &Fun{
	Def: Def{
		Name: "eq",
	},
	Type: FunType{
		ParamTypes: []Type{TypeInt, TypeInt},
		RetType:    TypeBool,
	},
	Kids: []Node{func(i, j int32) bool { return i == j }},
}

var intGt = &Fun{
	Def: Def{
		Name: "gt",
	},
	Type: FunType{
		ParamTypes: []Type{TypeInt, TypeInt},
		RetType:    TypeBool,
	},
	Kids: []Node{func(i, j int32) bool { return i > j }},
}

var intLt = &Fun{
	Def: Def{
		Name: "lt",
	},
	Type: FunType{
		ParamTypes: []Type{TypeInt, TypeInt},
		RetType:    TypeBool,
	},
	Kids: []Node{func(i, j int32) bool { return i < j }},
}

var intSub = &Fun{
	Def: Def{
		Name: "sub",
	},
	Type: FunType{
		ParamTypes: []Type{TypeInt, TypeInt},
		RetType:    TypeBool,
	},
	Kids: []Node{func(i, j int32) int32 { return i - j }},
}

var intType = func() *Record {
	members := []Node{
		intAdd,
		intEq,
		intGt,
		intLt,
		intSub,
	}
	memberMap := make(map[string]Node, len(members))
	for _, m := range members {
		memberMap[m.(*Fun).Name] = m
	}
	return &Record{
		Members:   members,
		MemberMap: memberMap,
	}
}()
