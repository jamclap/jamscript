package main

import (
	"fmt"
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
	for _, token := range tokens {
		fmt.Printf("%s\n", token)
	}
}
