package script

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

func Lex(source string) []Token {
	l := lexer{
		source: source,
	}
	l.lex()
	return l.tokens
}

type Token struct {
	Kind TokenKind
	Text string
}

func (t Token) String() string {
	return fmt.Sprintf("Token{%s \"%s\"}", t.Kind, t.Text)
}

type TokenKind int

const (
	TokenNone TokenKind = iota
	TokenEnd
	TokenFun
	TokenId
	TokenPub
	TokenString
)

//go:generate stringer -trimprefix=Token -type=TokenKind

type lexer struct {
	index  int
	source string
	tokens []Token
}

func (l *lexer) lex() {
	for l.has() {
		r := l.peek()
		switch {
		case unicode.IsLetter(r) || r == '_':
			l.id()
		default:
			l.next()
		}
	}
}

func (l *lexer) has() bool {
	return l.index < len(l.source)
}

func (l *lexer) next() {
	if !l.has() {
		return
	}
	_, size := utf8.DecodeRuneInString(l.source[l.index:])
	l.index += size
}

func (l *lexer) peek() rune {
	if !l.has() {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.source[l.index:])
	return r
}

func (l *lexer) push(kind TokenKind, start int) {
	l.tokens = append(l.tokens, Token{Kind: kind, Text: l.source[start:l.index]})
}

func (l *lexer) id() {
	start := l.index
Id:
	for l.has() {
		r := l.peek()
		switch {
		case unicode.IsDigit(r):
		case unicode.IsLetter(r):
		case r == '_':
		default:
			break Id
		}
		l.next()
	}
	text := l.source[start:l.index]
	kind, isKey := keys[text]
	if !isKey {
		kind = TokenId
	}
	l.push(kind, start)
}

var keys = map[string]TokenKind{
	"end": TokenEnd,
	"fun": TokenFun,
	"pub": TokenPub,
}
