package script

func Parse(tokens []Token) ParseNode {
	return ParseNode{}
}

type ParseNode struct {
	kind  ParseNodeKind
	token Token
	kids  []ParseNode
}

type ParseNodeKind int

const (
	ParseNone = iota
	ParseToken
)
