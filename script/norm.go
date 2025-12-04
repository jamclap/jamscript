package script

import (
	"fmt"
	"strings"
)

func (b *treeBuilder) Norm(p ParseNode) *Module {
	b.reset()
	b.normNode(p)
	// Fake block to commit the top.
	b.commitBlock(0)
	pop(&b.blocks)
	return b.toTree()
}

func (*treeBuilder) expectNone(part ParseNode) {
	if part.Kind != ParseNone {
		// log.Printf("Unexpected: %v\n", part)
	}
}

func (b *treeBuilder) normNode(p ParseNode) {
	switch p.Kind {
	case ParseArgs:
		b.normArgs(p)
	case ParseBlock:
		b.normBlock(p)
	case ParseCall:
		b.normCall(p)
	case ParseFun:
		b.normFun(p)
	case ParseJunk:
		b.normJunk(p)
	case ParseModify:
		b.normModify(p)
	case ParseNone:
		b.normNone(p)
	case ParseParam:
		b.normParam(p)
	case ParseParams:
		b.normParams(p)
	case ParseString:
		b.normString(p)
	case ParseToken:
		b.normToken(p)
	default:
		panic(fmt.Sprintf("unexpected script.ParseKind: %#v", p.Kind))
	}
}

func (b *treeBuilder) normArgs(p ParseNode) {
	start := len(b.work)
	next := p.ExpectToken(0, TokenRoundOpen)
	part := ParseNode{}
Args:
	for {
		next, part = p.Next(next)
		switch part.Kind {
		case ParseToken:
			switch part.Token.Kind {
			case TokenComma:
				// TODO Error on repeated.
				continue Args
			case TokenRoundClose:
				break Args
			default:
			}
		case ParseNone:
			break Args
		}
		b.normNode(part)
	}
	if part.Token.Kind != TokenRoundClose {
		// log.Printf("Unexpected: %v\n", part)
	}
	_, part = p.Next(next)
	b.expectNone(part)
	b.commitBlock(start)
}

func (b *treeBuilder) normBlock(p ParseNode) {
	start := len(b.work)
	for _, kid := range p.Kids {
		b.normNode(kid)
	}
	b.commitBlock(start)
}

func (b *treeBuilder) normCall(p ParseNode) {
	start := len(b.work)
	call := inCall{}
	// log.Printf("call\n")
	next, part := p.Next(0)
	b.normNode(part)
	next, part = p.Next(next)
	if part.Kind == ParseArgs {
		b.normArgs(part)
		call.args = b.popWorkBlock().kids
		_, part = p.Next(next)
	}
	b.expectNone(part)
	// We're about to push the callee as the next committed node.
	call.callee = Idx[inNode](len(b.nodes))
	b.commit(inNode{kind: NodeCall, index: len(b.calls)}, start)
	b.calls = append(b.calls, call)
}

func (b *treeBuilder) normFun(p ParseNode) {
	fun := inFun{}
	next := p.ExpectToken(0, TokenFun)
	next, part := p.Next(next)
	if part.Token.Kind == TokenId {
		fun.Name = part.Token.Text
		next, part = p.Next(next)
	} else {
		fun.Name = ""
	}
	if part.Kind == ParseParams {
		b.normParams(part)
		fun.params = b.popWorkBlock().kids
		next, part = p.Next(next)
	}
	// TODO Return type.
	if part.Kind == ParseBlock {
		b.normBlock(part)
		fun.kids = b.popWorkBlock().kids
		_, part = p.Next(next)
	}
	b.expectNone(part)
	b.pushWork(inNode{kind: NodeFun, index: len(b.funs)})
	b.funs = append(b.funs, fun)
	// log.Printf("fun %s %v\n", fun.Name, fun.params)
}

func (b *treeBuilder) normJunk(p ParseNode) {
	// panic("unimplemented")
}

func (b *treeBuilder) normModify(p ParseNode) {
	next := 0
	part := ParseNode{}
	var flags NodeFlags
Modify:
	for {
		next, part = p.Next(next)
		switch part.Token.Kind {
		case TokenPlug:
			flags |= NodeFlagPlug
		case TokenPub:
			flags |= NodeFlagPub
		default:
			break Modify
		}
	}
	b.normNode(part)
	_, part = p.Next(next)
	b.expectNone(part)
	w := b.work[len(b.work)-1]
	switch w.kind {
	case NodeFun:
		b.funs[w.index].Flags |= flags
	case NodeVar:
		b.vars[w.index].Flags |= flags
	}
	// log.Printf(
	// 	"flags: %+v %+v\n",
	// 	b.workInfo[len(b.workInfo)-1],
	// 	b.work[len(b.work)-1],
	// )
}

func (b *treeBuilder) normNone(p ParseNode) {
	// panic("unimplemented")
}

func (b *treeBuilder) normParam(p ParseNode) {
	v := inVar{}
	next, part := p.Next(0)
	if part.Token.Kind == TokenId {
		v.Name = part.Token.Text
		next, part = p.Next(next)
	} else {
		v.Name = ""
	}
	if part.Kind != ParseNone {
		// TODO Fix logic, and make it easy to do things like this.
		// TODO Presumably need to commit the top of work and get that.
		// TODO Use a wrapper with anonymous function for the helper?
		// n := len(b.nodes)
		b.normNode(part)
		// if len(b.nodes) > n {
		// 	v.typeInfo = Idx[inNode](n)
		// }
		_, part = p.Next(next)
	}
	b.expectNone(part)
	// Param instances are only referenced directly through params, so nodes can
	// go straight on without prior work storage.
	b.pushWork(inNode{kind: NodeVar, index: len(b.vars)})
	b.vars = append(b.vars, v)
}

func (b *treeBuilder) normParams(p ParseNode) {
	start := len(b.work)
	next := p.ExpectToken(0, TokenRoundOpen)
	part := ParseNode{}
Params:
	for {
		next, part = p.Next(next)
		switch part.Kind {
		case ParseParam:
			b.normParam(part)
		default:
			switch part.Token.Kind {
			case TokenComma:
				// TODO Error on repeated.
			default:
				break Params
			}
		}
	}
	if part.Token.Kind != TokenRoundClose {
		// log.Printf("Unexpected: %v\n", part)
	}
	_, part = p.Next(next)
	b.expectNone(part)
	b.commitBlock(start)
}

func (b *treeBuilder) normString(p ParseNode) {
	builder := strings.Builder{}
	next := p.ExpectToken(0, TokenStringOpen)
	part := ParseNode{}
Parts:
	for {
		next, part = p.Next(next)
		switch part.Token.Kind {
		case TokenStringText:
			builder.WriteString(part.Token.Text)
		case TokenStringEscape:
			// TODO Multi-rune escapes.
			switch r := RuneAt(part.Token.Text, 1); r {
			case '"', '\\':
				builder.WriteRune(r)
			case 'n':
				builder.WriteRune('\n')
			case 'r':
				builder.WriteRune('\r')
			case 't':
				builder.WriteRune('\t')
			}
		case TokenStringClose, TokenNone:
			break Parts
		}
	}
	text := Token{Kind: TokenStringText, Text: builder.String()}
	b.pushWork(inNode{kind: NodeToken, index: len(b.tokens)})
	b.tokens = append(b.tokens, text)
}

func RuneAt(s string, i int) rune {
	for _, r := range s {
		if i == 0 {
			return r
		}
		i--
	}
	return -1
}

func (b *treeBuilder) normToken(p ParseNode) {
	switch p.Token.Kind {
	case TokenId: // TODO numeric tokens or such like
	default:
		return
	}
	b.pushWork(inNode{kind: NodeToken, index: len(b.tokens)})
	b.tokens = append(b.tokens, p.Token)
}
