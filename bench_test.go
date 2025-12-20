package main_test

import (
	_ "embed"
	"io"
	"log"
	"testing"

	"github.com/jamclap/jamscript/rio"
)

func BenchmarkProcessBranch(b *testing.B) {
	process(branch, b)
}

func BenchmarkRunBranch(b *testing.B) {
	run(branch, b)
}

func BenchmarkProcessExplore(b *testing.B) {
	process(explore, b)
}

func BenchmarkRunExplore(b *testing.B) {
	run(explore, b)
}

func BenchmarkProcessFib(b *testing.B) {
	process(fib, b)
}

func BenchmarkRunFib(b *testing.B) {
	run(fib, b)
}

func BenchmarkProcessHi(b *testing.B) {
	process(hi, b)
}

func BenchmarkRunHi(b *testing.B) {
	run(hi, b)
}

func process(source string, b *testing.B) {
	e := rio.NewEngine()
	for b.Loop() {
		e.Process(source)
	}
}

func run(source string, b *testing.B) {
	engine := rio.NewEngine()
	log.SetOutput(io.Discard)
	module := engine.Process(source)
	for b.Loop() {
		engine.Run(module)
	}
}

//go:embed testdata/branch.rio
var branch string

//go:embed testdata/explore.rio
var explore string

//go:embed testdata/fib.rio
var fib string

//go:embed testdata/hi.rio
var hi string
