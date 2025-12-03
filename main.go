package main

import (
	"log"
	"os"

	"github.com/jamclap/jamscript/script"
)

func main() {
	if len(os.Args) < 2 {
		return
	}
	path := os.Args[1]
	b, err := os.ReadFile(path)
	if err != nil {
		log.Panic(err)
	}
	source := string(b)
	tokens := script.Lex(source)
	parseTree := script.Parse(tokens)
	// parseTree.Print()
	module := script.Norm(parseTree)
	module.Core["log"] = &script.Fun{
		Def: script.Def{
			Name: "log",
		},
	}
	// tree.Print()
	script.Resolve(module)
	script.Typify(module)
	module.Print()
}
