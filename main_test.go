package main_test

import (
	_ "embed"
	"io"
	"log"
	"testing"

	"github.com/jamclap/jamscript/script"
)

func BenchmarkProcessHi(b *testing.B) {
	process(hi, b)
}

func BenchmarkRunHi(b *testing.B) {
	run(hi, b)
}

func BenchmarkProcessExplore(b *testing.B) {
	process(explore, b)
}

func BenchmarkRunExplore(b *testing.B) {
	run(explore, b)
}

func process(source string, b *testing.B) {
	e := script.NewEngine()
	for b.Loop() {
		e.Process(source)
	}
}

func run(source string, b *testing.B) {
	e := script.NewEngine()
	log.SetOutput(io.Discard)
	module := e.Process(source)
	for b.Loop() {
		e.Run(module)
	}
}

//go:embed examples/explore.jam
var explore string

//go:embed examples/hi.jam
var hi string
