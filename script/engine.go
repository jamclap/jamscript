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
}

func NewEngine() *Engine {
	return &Engine{
		treeBuilder: newTreeBuilder(),
	}
}

func (e *Engine) Process(source string) *Module {
	tokens := e.lexer.Lex(source)
	parseTree := e.parser.Parse(tokens)
	// parseTree.Print()
	module := e.treeBuilder.Norm(parseTree)
	module.Core["log"] = &Fun{
		Def: Def{
			Name: "log",
		},
		Kids: []Node{log.Println},
	}
	// tree.Print()
	e.Analyze(module)
	return module
}

func (e *Engine) Run(m *Module) {
	e.runner.Run(m)
}

func (e *Engine) Analyze(module *Module) {
	// TODO Track changes so we can know if more rounds are needed.
	// TODO What's a good max?
	e.resolver.core = module.Core
	for i := 0; i < 5; i++ {
		// TODO If stable, this shouldn't allocate more on each iteration.
		// TODO But it is, so fix it.
		e.resolver.Resolve(module)
		Typify(module)
	}
}
