package main_test

import (
	_ "embed"
	"testing"

	"github.com/jamclap/jamscript/script"
)

func BenchmarkLexHi(b *testing.B) {
	lex(hi, b)
}

func BenchmarkParseHi(b *testing.B) {
	parse(hi, b)
}

func BenchmarkNormHi(b *testing.B) {
	norm(hi, b)
}

func BenchmarkResolveHi(b *testing.B) {
	resolve(hi, b)
}

func BenchmarkLexExplore(b *testing.B) {
	lex(explore, b)
}

func BenchmarkParseExplore(b *testing.B) {
	parse(explore, b)
}

func BenchmarkNormExplore(b *testing.B) {
	norm(explore, b)
}

func BenchmarkResolveExplore(b *testing.B) {
	resolve(explore, b)
}

func lex(source string, b *testing.B) {
	for i := 0; i < b.N; i++ {
		script.Lex(source)
	}
}

func parse(source string, b *testing.B) {
	for b.Loop() {
		tokens := script.Lex(source)
		script.Parse(tokens)
	}
}

func norm(source string, b *testing.B) {
	for b.Loop() {
		tokens := script.Lex(source)
		parseTree := script.Parse(tokens)
		script.Norm(parseTree)
	}
}

func resolve(source string, b *testing.B) {
	for b.Loop() {
		tokens := script.Lex(source)
		parseTree := script.Parse(tokens)
		tree := script.Norm(parseTree)
		script.Resolve(tree)
	}
}

//go:embed examples/explore.jam
var explore string

//go:embed examples/hi.jam
var hi string
