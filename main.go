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
	module := script.Process(string(b))
	module.Print()
}
