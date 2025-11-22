package script

import (
	"fmt"
)

func Parse(tokens []Token) ParseNode {
	p := parser{
		tokens: tokens,
	}
	p.parse()
	nodes := make([]ParseNode, len(p.nodes))
	for i, node := range p.nodes {
		nodes[i] = ParseNode{
			Kind:  node.kind,
			Kids:  nodes[node.kidsStart:node.kidsEnd],
			Token: node.token,
		}
	}
	return nodes[len(nodes)-1]
}

type ParseNode struct {
	Kind  ParseKind
	Kids  []ParseNode
	Token Token
}

type ParseKind int

const (
	ParseNone ParseKind = iota
	ParseBlock
	ParseCall
	ParseFun
	ParseJunk
	ParseModify
	ParseParam
	ParseParams
	ParseString
	ParseToken
)

//go:generate stringer -type=ParseKind

func (n ParseNode) Print() {
	n.printAt(0)
}

func (n ParseNode) printAt(indent int) {
	for range indent {
		print("  ")
	}
	switch n.Kind {
	case ParseToken:
		fmt.Printf("%s\n", n.Token)
	default:
		fmt.Printf("%s\n", n.Kind)
		for _, kid := range n.Kids {
			kid.printAt(indent + 1)
		}
	}
}

type inParseNode struct {
	kind      ParseKind
	kidsStart int
	kidsEnd   int
	token     Token
}

type parser struct {
	index  int
	tokens []Token
	nodes  []inParseNode
	work   []inParseNode
}

func (p *parser) parse() {
	p.parseBlock()
	// Push the root itself.
	p.commit(ParseBlock, 0)
}

func (p *parser) commit(kind ParseKind, start int) {
	oldLen := len(p.nodes)
	p.nodes = append(p.nodes, p.work[start:]...)
	parent := inParseNode{kind: kind, kidsStart: oldLen, kidsEnd: len(p.nodes)}
	p.work = append(p.work[:start], parent)
}

func (p *parser) has() bool {
	for p.index < len(p.tokens) {
		t := p.tokens[p.index]
		if t.Kind != TokenHSpace {
			return true
		}
		p.pushToken(t)
	}
	return false
}

func (p *parser) peek() (t Token) {
	if p.has() {
		t = p.tokens[p.index]
	}
	return
}

func (p *parser) pending(start int) int {
	return len(p.work) - start
}

func (p *parser) push(node inParseNode) {
	p.work = append(p.work, node)
}

func (p *parser) pushToken(t Token) {
	p.push(inParseNode{kind: ParseToken, token: t})
	p.index++
}

func (p *parser) parseAtom() {
	if !p.has() {
		return
	}
	switch t := p.peek(); t.Kind {
	case TokenFun:
		p.parseFun()
	case TokenId:
		p.pushToken(t)
	case TokenPlug:
	case TokenPub:
		p.parseModify()
	default:
		// TODO Fix.
		p.pushToken(t)
	}
}

func (p *parser) parseBlock() {
	start := len(p.work)
Block:
	for p.has() {
		switch t := p.peek(); t.Kind {
		case TokenVSpace:
			p.pushToken(t)
		case TokenEnd:
			p.pushToken(t)
			break Block
		default:
			p.parseStatement()
		}
	}
	p.commit(ParseBlock, start)
}

func (p *parser) parseFun() {
	start := len(p.work)
	p.pushToken(p.peek())
	if t := p.peek(); t.Kind == TokenId {
		p.pushToken(t)
	}
	if p.peek().Kind == TokenRoundOpen {
		p.parseParams()
	}
	switch t := p.peek(); t.Kind {
	case TokenThen:
		p.pushToken(t)
		// TODO Parse expr.
		p.parseAtom()
	case TokenVSpace:
		p.parseBlock()
	}
	p.commit(ParseFun, start)
}

func (p *parser) parseJunk() {
	start := len(p.work)
Junk:
	for p.has() {
		t := p.peek()
		if t.Kind == TokenVSpace {
			break Junk
		}
		p.parseAtom()
	}
	if p.pending(start) > 1 {
		p.commit(ParseJunk, start)
	}
}

func (p *parser) parseModify() {
	start := len(p.work)
	found := false
Mods:
	for p.has() {
		t := p.peek()
		switch t.Kind {
		case TokenPlug:
		case TokenPub:
		default:
			break Mods
		}
		found = true
		p.pushToken(t)
	}
	// TODO Parse assignment?
	p.parseAtom()
	if found {
		p.commit(ParseModify, start)
	}
}

func (p *parser) parseParam() {
	start := len(p.work)
	if t := p.peek(); t.Kind == TokenId {
		p.pushToken(t)
	}
Param:
	for p.has() {
		t := p.peek()
		switch t.Kind {
		case TokenComma:
		case TokenRoundClose:
			break Param
		}
		// TODO Parse type and such.
		p.parseAtom()
	}
	p.commit(ParseParam, start)
}

func (p *parser) parseParams() {
	start := len(p.work)
	p.pushToken(p.peek())
Params:
	for p.has() {
		t := p.peek()
		switch t.Kind {
		case TokenComma:
			p.pushToken(t)
		case TokenRoundClose:
			p.pushToken(t)
			break Params
		default:
			p.parseParam()
		}
	}
	p.commit(ParseParams, start)
}

func (p *parser) parseStatement() {
	p.parseJunk()
}
