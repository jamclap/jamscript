package main_test

import (
	_ "embed"
	"io"
	"log"
	"testing"

	"github.com/jamclap/jamscript/script"
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

func BenchmarkProcessHi(b *testing.B) {
	process(hi, b)
}

func BenchmarkRunHi(b *testing.B) {
	run(hi, b)
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

//go:embed testdata/branch.jam
var branch string

//go:embed testdata/explore.jam
var explore string

//go:embed testdata/hi.jam
var hi string
