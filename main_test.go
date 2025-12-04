package main_test

import (
	_ "embed"
	"testing"

	"github.com/jamclap/jamscript/script"
)

func BenchmarkProcessHi(b *testing.B) {
	process(hi, b)
}

func BenchmarkProcessExplore(b *testing.B) {
	process(explore, b)
}

func process(source string, b *testing.B) {
	e := script.NewEngine()
	for b.Loop() {
		e.Process(source)
	}
}

//go:embed examples/explore.jam
var explore string

//go:embed examples/hi.jam
var hi string
