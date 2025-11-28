package main_test

import (
	_ "embed"
	"testing"

	"github.com/jamclap/jamscript/script"
)

func lex(source string) {
	script.Lex(source)
}

func BenchmarkLex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		lex(hi)
	}
}

func parse(source string) {
	tokens := script.Lex(source)
	script.Parse(tokens)
}

func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		parse(hi)
	}
}

func norm(source string) {
	tokens := script.Lex(source)
	parseTree := script.Parse(tokens)
	script.Norm(parseTree)
}

func BenchmarkNorm(b *testing.B) {
	for i := 0; i < b.N; i++ {
		norm(hi)
	}
}

//go:embed examples/hi.jam
var hi string
