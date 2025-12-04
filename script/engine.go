package script

type Engine struct {
	// Types map[Type]Type // TODO or use unique.Make(type) instead?
	resolver resolver
}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) Process(source string) *Module {
	tokens := Lex(source)
	parseTree := Parse(tokens)
	// parseTree.Print()
	module := Norm(parseTree)
	module.Core["log"] = &Fun{
		Def: Def{
			Name: "log",
		},
	}
	// tree.Print()
	e.Analyze(module)
	return module
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
