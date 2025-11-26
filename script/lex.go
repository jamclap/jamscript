package script

import (
	"fmt"
	"unicode"
	"unicode/utf8"
	"unique"
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
	Text unique.Handle[string]
}

func (t Token) String() string {
	return fmt.Sprintf("%s \"%s\"", t.Kind, t.Text.Value())
}

type TokenKind int

const (
	TokenNone TokenKind = iota
	TokenAs
	TokenBreak
	TokenChange
	TokenClass
	TokenComma
	TokenConst
	TokenContinue
	TokenElse
	TokenEnd
	TokenEnum
	TokenFor
	TokenFrom
	TokenFun
	TokenHSpace
	TokenId
	TokenIf
	TokenIs
	TokenImport
	TokenJunk
	TokenPlug
	TokenPub
	TokenReturn
	TokenRoundClose
	TokenRoundOpen
	TokenStringEscape
	TokenStringText
	TokenStringClose
	TokenStringOpen
	TokenStruct
	TokenSwitch
	TokenThen
	TokenVSpace
	TokenUnion
	TokenUse
	TokenVar
	TokenVartype
)

//go:generate stringer -type=TokenKind

type lexer struct {
	index  int
	source string
	tokens []Token
}

func (l *lexer) lex() {
	for l.has() {
		r := l.peek()
		switch {
		case unicode.IsLetter(r) || r == '$' || r == '_':
			l.id()
		case r == ' ' || r == '\t':
			l.hspace()
		default:
			start := l.index
			switch r {
			case '"':
				l.next()
				l.push(TokenStringOpen, start)
				l.str()
			case ',':
				l.next()
				l.push(TokenComma, start)
			case '(':
				l.next()
				l.push(TokenRoundOpen, start)
			case ')':
				l.next()
				l.push(TokenRoundClose, start)
			case '\n':
				l.next()
				l.push(TokenVSpace, start)
			default:
				l.next()
				l.push(TokenJunk, start)
			}
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
	if start < l.index {
		text := unique.Make(l.source[start:l.index])
		l.tokens = append(l.tokens, Token{Kind: kind, Text: text})
	}
}

func (l *lexer) hspace() {
	start := l.index
HSpace:
	for l.has() {
		l.next()
		r := l.peek()
		switch r {
		case ' ':
		case '\t':
		default:
			break HSpace
		}
	}
	l.push(TokenHSpace, start)
}

func (l *lexer) id() {
	start := l.index
	l.next()
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

func (l *lexer) str() {
	start := l.index
	l.next()
	kind := TokenStringText
Str:
	for l.has() {
		r := l.peek()
		if r == '\n' {
			break Str
		}
		if kind == TokenStringEscape {
			l.next()
			l.push(kind, start)
			kind = TokenStringText
			start = l.index
			continue Str
		}
		switch r {
		case '"':
			l.next()
			kind = TokenStringClose
			break Str
		case '\n':
			break Str
		case '\\':
			l.push(kind, start)
			kind = TokenStringEscape
			start = l.index
		}
		l.next()
	}
	l.push(kind, start)
}

// We have keys only for things that affect parsing?
var keys = map[string]TokenKind{
	"as":       TokenAs,
	"break":    TokenBreak,
	"class":    TokenClass,
	"change":   TokenChange,
	"const":    TokenConst,
	"continue": TokenContinue,
	"else":     TokenElse,
	"end":      TokenEnd,
	"if":       TokenIf,
	"is":       TokenIs,
	"import":   TokenImport,
	"enum":     TokenEnum,
	"for":      TokenFor,
	"from":     TokenFrom,
	"fun":      TokenFun,
	"plug":     TokenPlug,
	"pub":      TokenPub,
	"return":   TokenReturn,
	"struct":   TokenStruct,
	"switch":   TokenSwitch,
	"then":     TokenThen,
	"union":    TokenUnion,
	"use":      TokenUse,
	"var":      TokenVar,
	"vartype":  TokenVartype,
}
