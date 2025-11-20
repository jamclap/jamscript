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
	ParseParams
	ParseString
	ParseToken
)

//go:generate stringer -type=ParseKind

func (n ParseNode) Print() {
	n.printAt(0)
}

func (n ParseNode) printAt(indent int) {
	for i := 0; i < indent; i++ {
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
	// Double commit at end.
	// First pushes working nodes. Second pushes the root itself.
	p.commit(ParseBlock, 0)
	p.commit(ParseBlock, 0)
}

func (p *parser) commit(kind ParseKind, start int) {
	oldLen := len(p.nodes)
	p.nodes = append(p.nodes, p.work[start:]...)
	parent := inParseNode{kind: kind, kidsStart: oldLen, kidsEnd: len(p.nodes)}
	p.work = append(p.work[:start], parent)
}

func (p *parser) has() bool {
	return p.index < len(p.tokens)
}

func (p *parser) peek() Token {
	return p.tokens[p.index]
}

func (p *parser) push(node inParseNode) {
	p.work = append(p.work, node)
}

func (p *parser) pushToken(t Token) {
	p.push(inParseNode{kind: ParseToken, token: t})
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
		p.pushToken(t)
		p.index++
		if t.Kind == TokenVSpace {
			break Statement
		}
	}
	_ = start
	p.commit(ParseJunk, start)
}
