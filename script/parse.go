package script

import (
	"fmt"
	"log"
)

func Parse(tokens []Token) ParseNode {
	p := parser{
		tokens: tokens,
	}
	p.nodes = make([]inParseNode, 1) // so 0 is none
	p.parse()
	nodes := make([]ParseNode, len(p.nodes))
	for i, node := range p.nodes {
		nodes[i] = ParseNode{
			Kind:  node.kind,
			Kids:  nodes[node.kids.Start:node.kids.End],
			Token: node.token,
		}
	}
	return nodes[len(nodes)-1]
}

type Range[T any] struct {
	Start int
	End   int
}

func (r Range[T]) Slice(items []T) []T {
	return items[r.Start:r.End]
}

func Slice[T any, U any](r Range[T], items []U) []U {
	return items[r.Start:r.End]
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

func (n ParseNode) ExpectToken(start int, kind TokenKind) int {
	next, kid := n.Next(start)
	if kid.Kind == ParseToken && kid.Token.Kind == kind {
		return next
	}
	// TODO Record error.
	log.Printf("Bad kid: %s %s\n", kid.Kind, kid.Token.Kind)
	return len(n.Kids)
}

func (n ParseNode) Next(start int) (int, ParseNode) {
	if n.Kind != ParseToken {
		for i := start; i < len(n.Kids); i++ {
			kid := n.Kids[i]
			switch kid.Kind {
			case ParseToken:
				switch kid.Token.Kind {
				case TokenHSpace:
				default:
					return i + 1, kid
				}
			default:
				return i + 1, kid
			}
		}
	}
	return len(n.Kids), ParseNode{}
}

func (n ParseNode) Print() {
	n.printAt(0)
}

func (n ParseNode) printAt(indent int) {
	PrintIndent(indent)
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

func PrintIndent(indent int) {
	for range indent {
		print("  ")
	}
}

type inParseNode struct {
	kind  ParseKind
	kids  Range[inParseNode]
	token Token
}

type parser struct {
	index  int
	tokens []Token
	nodes  []inParseNode
	work   []inParseNode
}

func (p *parser) parse() {
	p.parseBlockTop()
	// Push the root itself.
	p.commit(ParseBlock, 0)
}

func (p *parser) commit(kind ParseKind, start int) {
	oldLen := len(p.nodes)
	p.nodes = append(p.nodes, p.work[start:]...)
	parent := inParseNode{
		kind: kind,
		kids: Range[inParseNode]{Start: oldLen, End: len(p.nodes)},
	}
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

func (p *parser) push(node inParseNode) {
	p.work = append(p.work, node)
}

func (p *parser) pushToken(t Token) {
	p.push(inParseNode{kind: ParseToken, token: t})
	p.index++
}

func (p *parser) parseArgs() {
	start := len(p.work)
	p.pushToken(p.peek())
Params:
	for p.has() {
		t := p.peek()
		switch t.Kind {
		case TokenComma, TokenVSpace:
			p.pushToken(t)
		case TokenRoundClose:
			p.pushToken(t)
			break Params
		default:
			p.parseExpr()
		}
	}
	p.commit(ParseParams, start)
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
	case TokenPlug, TokenPub:
		p.parseModify()
	case TokenStringOpen:
		p.parseString()
	default:
		start := len(p.work)
		p.pushToken(t)
		p.commit(ParseJunk, start)
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

func (p *parser) parseBlockTop() {
	start := len(p.work)
	for p.has() {
		switch t := p.peek(); t.Kind {
		case TokenVSpace:
			p.pushToken(t)
		default:
			p.parseStatement()
		}
	}
	p.commit(ParseBlock, start)
}

func (p *parser) parseCall() {
	start := len(p.work)
	p.parseAtom()
	if p.peek().Kind == TokenRoundOpen {
		p.parseArgs()
		p.commit(ParseCall, start)
	}
}

func (p *parser) parseExpr() {
	p.parseCall()
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
		p.parseExpr()
	case TokenVSpace:
		p.parseBlock()
	}
	p.commit(ParseFun, start)
}

func (p *parser) parseModify() {
	start := len(p.work)
	found := false
Mods:
	for p.has() {
		t := p.peek()
		switch t.Kind {
		case TokenPlug, TokenPub:
		default:
			break Mods
		}
		found = true
		p.pushToken(t)
	}
	p.parseExpr()
	if found {
		p.commit(ParseModify, start)
	}
}

func (p *parser) parseParam() {
	start := len(p.work)
Param:
	for p.has() {
		t := p.peek()
		switch t.Kind {
		case TokenComma, TokenRoundClose:
			break Param
		case TokenVSpace:
			p.pushToken(t)
		default:
			p.parseExpr()
		}
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
		case TokenComma, TokenVSpace:
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
	p.parseExpr()
}

func (p *parser) parseString() {
	start := len(p.work)
	p.pushToken(p.peek())
Params:
	for p.has() {
		t := p.peek()
		p.pushToken(t)
		if t.Kind == TokenStringClose {
			break Params
		}
	}
	p.commit(ParseString, start)
}
