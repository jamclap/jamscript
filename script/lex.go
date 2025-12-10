package script

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

func (l *lexer) Lex(source string) []Token {
	l.index = 0
	l.peekedSize = 0
	l.source = source
	l.tokens = l.tokens[:0]
	l.lex()
	return l.tokens
}

type Token struct {
	Kind TokenKind
	Text string
}

func (t Token) String() string {
	return fmt.Sprintf("%s \"%s\"", t.Kind, t.Text)
}

type TokenKind int

const (
	TokenNone TokenKind = iota
	TokenAs
	TokenBreak
	TokenCase
	TokenChange
	TokenClass
	TokenComma
	TokenCommentOpen
	TokenCommentText
	TokenConst
	TokenContinue
	TokenElse
	TokenEnd
	TokenEq
	TokenEqEq
	TokenEnum
	TokenFor
	TokenFrom
	TokenFun
	TokenGe
	TokenGt
	TokenHSpace
	TokenId
	TokenIf
	TokenInt
	TokenIs
	TokenImport
	TokenLe
	TokenLt
	TokenJunk
	TokenNot
	TokenNEq
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
	index      int
	source     string
	peeked     rune
	peekedSize int
	tokens     []Token
}

func (l *lexer) lex() {
	for l.has() {
		r := l.peek()
		switch {
		case unicode.IsLetter(r) || r == '$' || r == '_':
			l.id()
		case r >= '0' && r <= '9':
			l.number()
		case r == ' ' || r == '\t':
			l.hspace()
		default:
			start := l.index
			switch r {
			case '#':
				l.next()
				l.push(TokenCommentOpen, start)
				l.comment()
			case '"':
				l.next()
				l.push(TokenStringOpen, start)
				l.str()
			case '=':
				l.next()
				switch r := l.peek(); r {
				case '=':
					l.next()
					l.push(TokenEqEq, start)
				default:
					l.push(TokenEq, start)
				}
			case '<':
				l.next()
				switch r := l.peek(); r {
				case '=':
					l.next()
					l.push(TokenLe, start)
				default:
					l.push(TokenLt, start)
				}
			case '>':
				l.next()
				switch r := l.peek(); r {
				case '=':
					l.next()
					l.push(TokenGe, start)
				default:
					l.push(TokenGt, start)
				}
			case ',':
				l.next()
				l.push(TokenComma, start)
			case '(':
				l.next()
				l.push(TokenRoundOpen, start)
			case ')':
				l.next()
				l.push(TokenRoundClose, start)
			case '\r':
				l.next()
				if l.peek() == '\n' {
					l.next()
				}
				l.push(TokenVSpace, start)
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
	if l.peekedSize > 0 {
		l.index += l.peekedSize
		l.peekedSize = 0
		return
	}
	if !l.has() {
		return
	}
	_, size := utf8.DecodeRuneInString(l.source[l.index:])
	l.index += size
}

func (l *lexer) peek() rune {
	if l.peekedSize > 0 {
		r := l.peeked
		l.peekedSize = 0
		return r
	}
	if !l.has() {
		return 0
	}
	l.peeked, l.peekedSize = utf8.DecodeRuneInString(l.source[l.index:])
	return l.peeked
}

func (l *lexer) push(kind TokenKind, start int) {
	if start < l.index {
		text := l.source[start:l.index]
		l.tokens = append(l.tokens, Token{Kind: kind, Text: text})
	}
}

func (l *lexer) comment() {
	start := l.index
Comment:
	for l.has() {
		r := l.peek()
		if r == '\n' {
			break Comment
		}
		l.next()
	}
	l.push(TokenCommentText, start)
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

func (l *lexer) number() {
	start := l.index
	// TODO Include negative in int literal?
Int:
	for l.has() {
		r := l.peek()
		switch {
		case r >= '0' && r <= '9':
			l.next()
		default:
			break Int
		}
	}
	// TODO Check if it's a float literal at this point.
	l.push(TokenInt, start)
}

func (l *lexer) str() {
	start := l.index
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
			l.push(kind, start)
			start = l.index
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
	"case":     TokenCase,
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
