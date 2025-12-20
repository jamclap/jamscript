package rio

import (
	"fmt"
	"io"
)

func (p *parser) Parse(tokens []Token) ParseNode {
	p.index = 0
	p.nodes = p.nodes[:0]
	p.tokens = tokens
	p.work = p.work[:0]
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
	ParseArgs
	ParseBlock
	ParseCall
	ParseCase
	ParseComment
	ParseElse
	ParseFun
	ParseInfix
	ParseJunk
	ParseModify
	ParseParam
	ParseParams
	ParsePrefix
	ParseReturn
	ParseString
	ParseSwitch
	ParseSwitchEmpty
	ParseToken
	ParseVar
)

//go:generate stringer -type=ParseKind

func (n ParseNode) ExpectToken(start int, kind TokenKind) int {
	next, kid := n.Next(start)
	if kid.Kind == ParseToken && kid.Token.Kind == kind {
		return next
	}
	// TODO Record error.
	// log.Printf("Bad kid: %s %s\n", kid.Kind, kid.Token.Kind)
	return len(n.Kids)
}

func (n ParseNode) Next(start int) (int, ParseNode) {
	return n.NextEx(start, false)
}

func (n ParseNode) NextEx(start int, keepVSpace bool) (int, ParseNode) {
	if n.Kind != ParseToken {
		for i := start; i < len(n.Kids); i++ {
			kid := n.Kids[i]
			switch kid.Kind {
			case ParseComment:
			case ParseToken:
				switch kid.Token.Kind {
				case TokenCommentText, TokenCommentOpen, TokenHSpace:
				case TokenVSpace:
					if keepVSpace {
						return i + 1, kid
					}
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

func (n ParseNode) Print(w io.Writer) {
	n.printAt(w, 0)
}

func (n ParseNode) printAt(w io.Writer, indent int) {
	PrintIndent(w, indent)
	switch n.Kind {
	case ParseToken:
		fmt.Fprintf(w, "%s\n", n.Token)
	default:
		fmt.Fprintf(w, "%s\n", n.Kind)
		for _, kid := range n.Kids {
			kid.printAt(w, indent+1)
		}
	}
}

func PrintIndent(w io.Writer, indent int) {
	for range indent {
		fmt.Fprint(w, "    ")
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
Has:
	for p.index < len(p.tokens) {
		t := p.tokens[p.index]
		switch t.Kind {
		case TokenCommentOpen:
			start := len(p.work)
			p.pushToken(t)
			if t = p.tokens[p.index]; t.Kind == TokenCommentText {
				p.pushToken(t)
				p.commit(ParseComment, start)
			}
			continue Has
		case TokenCommentText, TokenHSpace:
		default:
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
	p.commit(ParseArgs, start)
}

func (p *parser) parseAdd() {
	start := len(p.work)
	p.parseCall()
	for {
		switch t := p.peek(); t.Kind {
		case TokenAdd, TokenSub:
			p.pushToken(t)
			p.parseCall()
			p.commit(ParseInfix, start)
		default:
			return
		}
	}
}

func (p *parser) parseAtom() {
	if !p.has() {
		return
	}
	switch t := p.peek(); t.Kind {
	case TokenCase:
		p.parseCase(t)
	case TokenElse:
		p.parseElse(t)
	case TokenFun:
		p.parseFun(t)
	case TokenId, TokenInt:
		p.pushToken(t)
	case TokenPlug, TokenPub:
		p.parseModify(t)
	case TokenReturn:
		p.parseReturn(t)
	case TokenStringOpen:
		p.parseString(t)
	case TokenSub:
		p.parsePrefix(t)
	case TokenSwitch:
		p.parseSwitch(t)
	case TokenVar:
		p.parseVar(t)
	case TokenVSpace:
	default:
		start := len(p.work)
		p.pushToken(t)
		p.commit(ParseJunk, start)
	}
}

func (p *parser) parseBlock() {
	start := len(p.work)
	if t := p.peek(); t.Kind == TokenThen {
		p.pushToken(t)
	}
	switch t := p.peek(); t.Kind {
	case TokenVSpace:
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
	default:
		p.parseExpr()
	}
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

func (p *parser) parseCase(t Token) {
	start := len(p.work)
	p.pushToken(t)
	// Patterns.
	patternsStart := len(p.work)
	p.parseExpr()
	// TODO Multiple patterns separated by comma.
	p.commit(ParseArgs, patternsStart)
	// Finish.
	p.parseCaseFinish()
	p.commit(ParseCase, start)
}

func (p *parser) parseCaseFinish() {
	start := len(p.work)
	if t := p.peek(); t.Kind == TokenThen {
		p.pushToken(t)
	}
	switch t := p.peek(); t.Kind {
	case TokenVSpace:
	Block:
		for p.has() {
			switch t := p.peek(); t.Kind {
			case TokenVSpace:
				p.pushToken(t)
			case TokenCase, TokenElse, TokenEnd:
				break Block
			default:
				p.parseStatement()
			}
		}
	default:
		p.parseExpr()
	}
	p.commit(ParseBlock, start)
}

func (p *parser) parseCompare() {
	start := len(p.work)
	p.parseAdd()
	for {
		switch t := p.peek(); t.Kind {
		case TokenEqEq, TokenGe, TokenGt, TokenLe, TokenLt, TokenNEq:
			p.pushToken(t)
			p.parseAdd()
			p.commit(ParseInfix, start)
		default:
			return
		}
	}
}

func (p *parser) parseElse(t Token) {
	start := len(p.work)
	p.pushToken(t)
	p.parseCaseFinish()
	p.commit(ParseElse, start)
}

func (p *parser) parseExpr() {
	p.parseCompare()
}

func (p *parser) parseFun(t Token) {
	start := len(p.work)
	p.pushToken(t)
	if t := p.peek(); t.Kind == TokenId {
		p.pushToken(t)
	}
	if p.peek().Kind == TokenRoundOpen {
		p.parseParams()
	}
	p.parseBlock()
	p.commit(ParseFun, start)
}

func (p *parser) parseModify(t Token) {
	start := len(p.work)
	p.pushToken(t)
Mods:
	for p.has() {
		t := p.peek()
		switch t.Kind {
		case TokenPlug, TokenPub:
		default:
			break Mods
		}
		p.pushToken(t)
	}
	p.parseExpr()
	p.commit(ParseModify, start)
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

func (p *parser) parsePrefix(t Token) {
	start := len(p.work)
	p.pushToken(t)
	p.parseExpr()
	p.commit(ParsePrefix, start)
}

func (p *parser) parseReturn(t Token) {
	start := len(p.work)
	p.pushToken(t)
	switch t := p.peek(); t.Kind {
	case TokenVSpace:
	default:
		p.parseExpr()
	}
	p.commit(ParseReturn, start)
}

func (p *parser) parseStatement() {
	p.parseExpr()
}

func (p *parser) parseString(t Token) {
	start := len(p.work)
	p.pushToken(t)
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

func (p *parser) parseSwitch(t Token) {
	start := len(p.work)
	kind := ParseSwitch
	p.pushToken(t)
	switch t := p.peek(); t.Kind {
	case TokenThen, TokenVSpace:
		kind = ParseSwitchEmpty
	default:
		p.parseExpr()
	}
	p.parseBlock()
	p.commit(kind, start)
}

func (p *parser) parseVar(t Token) {
	start := len(p.work)
	p.pushToken(t)
	// Check for name.
	if t := p.peek(); t.Kind == TokenId {
		p.pushToken(t)
	}
	// Check for type.
	switch t := p.peek(); t.Kind {
	case TokenEq:
	case TokenVSpace:
	default:
		// Type.
		p.parseExpr()
	}
	// Check for init.
	if t := p.peek(); t.Kind == TokenEq {
		p.pushToken(t)
		if p.peek().Kind == TokenVSpace {
			p.pushToken(t)
		}
		p.parseExpr()
	}
	p.commit(ParseVar, start)
}
