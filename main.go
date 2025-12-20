package main

import (
	"log"
	"os"

	"github.com/jamclap/jamscript/rio"
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
	e := rio.NewEngine()
	module := e.Process(string(b))
	// module.Print()
	e.Run(module)
}
