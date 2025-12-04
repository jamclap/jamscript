package main_test

import (
	_ "embed"
	"testing"

	"github.com/jamclap/jamscript/script"
)

func BenchmarkProcessHi(b *testing.B) {
	process(hi, b)
}

func BenchmarkResetHi(b *testing.B) {
	processReset(hi, b)
}

func BenchmarkProcessExplore(b *testing.B) {
	process(explore, b)
}

func BenchmarkResetExplore(b *testing.B) {
	processReset(explore, b)
}

func process(source string, b *testing.B) {
	e := script.NewEngine()
	for b.Loop() {
		e.Process(source)
	}
}

func processReset(source string, b *testing.B) {
	for b.Loop() {
		e := script.NewEngine()
		e.Process(source)
	}
}

//go:embed examples/explore.jam
var explore string

//go:embed examples/hi.jam
var hi string
