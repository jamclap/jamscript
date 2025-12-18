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
	// parseTree.Print()
	module := e.treeBuilder.Norm(parseTree)
	module.Core["log"] = &Fun{
		Def: Def{
			Name: "log",
		},
		Kids: []Node{doLog},
	}
	// module.Print()
	e.analyze(module)
	// module.Print()
	return module
}

func (e *Engine) Run(m *Module) {
	e.runner.Run(m)
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

func doLog(s string) {
	// TODO Option to select where `log` goes?
	// fmt.Fprintln(os.Stderr, s)
	log.Println(s)
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

var intType = func() *Record {
	members := []Node{
		intGt,
		intLt,
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
