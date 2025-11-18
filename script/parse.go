package script

func Parse(tokens []Token) ParseNode {
	p := parser{
		tokens: tokens,
	}
	p.parse()
	return ParseNode{}
}

type ParseNode struct {
	Kind  ParseKind
	Kids  []ParseNode
	Token Token
}

type ParseKind int

const (
	ParseNone = iota
	ParseBlock
	ParseCall
	ParseFun
	ParseJunk
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

func (p *parser) parse() {
	p.parseBlock()
}

func (p *parser) has() bool {
	return p.index < len(p.tokens)
}

func (p *parser) peek() Token {
	return p.tokens[p.index]
}

func (p *parser) push(node ParseNode) {
	p.work = append(p.work, node)
}

func (p *parser) parseBlock() {
	for p.has() {
		p.parseStatement()
	}
}

func (p *parser) parseStatement() {
	start := len(p.work)
Statement:
	for p.has() {
		t := p.peek()
		p.push(ParseNode{Kind: ParseToken, Token: t})
		p.index++
		if t.Kind == TokenVSpace {
			break Statement
		}
	}
	_ = start
	// TODO p.commit(ParseJunk, start)
}
