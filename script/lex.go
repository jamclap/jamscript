package script

func Lex(source string) {
	l := lexer{
		source: source,
	}
	l.lex()
}

type Token struct {
	kind TokenKind
	text string
}

type TokenKind int

const (
	TokenNone = iota
	TokenId
	TokenString
)

type lexer struct {
	index  int
	source string
	tokens []Token
}

func (l *lexer) lex() {
	print(l.source)
}
