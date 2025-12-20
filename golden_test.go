package main_test

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/jamclap/jamscript/rio"
)

func TestGolden(t *testing.T) {
	engine := rio.NewEngine()
	names := []string{"branch", "fib", "hi"}
	for _, name := range names {
		updateGolden(engine, name)
	}
}

func updateGolden(engine *rio.Engine, name string) {
	dir := "testdata"
	// Prep output.
	outDir := filepath.Join(dir, "out")
	err := os.MkdirAll(outDir, 0o755)
	if err != nil {
		log.Panic(err)
	}
	out, err := os.Create(filepath.Join(outDir, fmt.Sprintf("%v.txt", name)))
	if err != nil {
		log.Panic(err)
	}
	defer out.Close()
	// Process input, and write tree.
	source, err := os.ReadFile(filepath.Join(dir, fmt.Sprintf("%v.rio", name)))
	if err != nil {
		log.Panic(err)
	}
	module := engine.Process(string(source))
	module.Print(out)
	fmt.Fprint(out, "\n--- run log ---\n\n")
	// Run program, capturing log.
	oldOut := log.Writer()
	log.SetOutput(out)
	defer log.SetOutput(oldOut)
	oldFlags := log.Flags()
	log.SetFlags(0)
	defer log.SetFlags(oldFlags)
	err = engine.Run(module)
	if err != nil {
		log.Println(err)
	}
}
