package script

func Parse(tokens []Token) ParseNode {
	p := parser{
		tokens: tokens,
	}
	p.parse()
	return ParseNode{}
}

type ParseNode struct {
	kind  ParseKind
	kids  []ParseNode
	token Token
}

type ParseKind int

const (
	ParseNone = iota
	ParseBlock
	ParseCall
	ParseFun
	ParseModify
	ParseParams
	ParseString
	ParseToken
)

type parser struct {
	index  int
	tokens []Token
	nodes  []ParseNode
	work   []ParseNode
}
