package script

import (
	"fmt"
	"log"
)

func Norm(p ParseNode) Tree {
	// TODO Convert internal repr to arrays then nodes?
	b := newTreeBuilder()
	b.normNode(p)
	// Fake block to commit the top.
	b.commitBlock(0)
	pop(&b.blocks)
	return b.toTree()
}

func (*treeBuilder) expectNone(part ParseNode) {
	if part.Kind != ParseNone {
		log.Printf("Unexpected: %v\n", part)
	}
}

func (b *treeBuilder) normNode(p ParseNode) {
	switch p.Kind {
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

func (b *treeBuilder) normBlock(p ParseNode) {
	start := len(b.work)
	for _, kid := range p.Kids {
		b.normNode(kid)
	}
	b.commitBlock(start)
}

func (b *treeBuilder) normCall(p ParseNode) {
	// panic("unimplemented")
}

func (b *treeBuilder) normFun(p ParseNode) {
	fun := inFun{}
	next := p.ExpectToken(0, TokenFun)
	next, part := p.Next(next)
	if part.Token.Kind == TokenId {
		fun.name = part.Token.Text
		next, part = p.Next(next)
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
	log.Printf("fun %s %v\n", fun.name.Value(), fun.params)
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
	last(&b.workInfo).Flags |= flags
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
		v.name = part.Token.Text
		next, part = p.Next(next)
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
		log.Printf("Unexpected: %v\n", part)
	}
	_, part = p.Next(next)
	b.expectNone(part)
	b.commitBlock(start)
}

func (b *treeBuilder) normString(p ParseNode) {
	// panic("unimplemented")
}

func (b *treeBuilder) normToken(p ParseNode) {
	// panic("unimplemented")
}
