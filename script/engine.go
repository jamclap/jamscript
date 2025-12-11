package script

import "log"

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
	}
}

func doLog(s string) {
	// TODO Option to select where `log` goes?
	// fmt.Fprintln(os.Stderr, s)
	log.Println(s)
}
