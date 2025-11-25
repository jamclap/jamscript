package script

import (
	"fmt"
	"log"
)

func Norm(p ParseNode) {
	// TODO Convert internal repr to arrays then nodes?
	b := treeBuilder{}
	b.normNode(p)
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
	for _, kid := range p.Kids {
		b.normNode(kid)
	}
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
		next, part = p.Next(next)
	}
	// TODO Return type.
	if part.Kind == ParseBlock {
		b.normBlock(part)
		_, part = p.Next(next)
	}
	b.expectNone(part)
	b.pushFun(fun)
}

func (b *treeBuilder) normJunk(p ParseNode) {
	// panic("unimplemented")
}

func (b *treeBuilder) normModify(p ParseNode) {
	next := 0
	part := ParseNode{}
	plug := false
	pub := false
Modify:
	for {
		next, part = p.Next(next)
		switch part.Token.Kind {
		case TokenPlug:
			log.Printf("plug\n")
			plug = true
		case TokenPub:
			log.Printf("pub\n")
			pub = true
		default:
			break Modify
		}
	}
	// TODO Set modifiers for the part.
	b.normNode(part)
	_, part = p.Next(next)
	b.expectNone(part)
	_ = plug
	_ = pub
}

func (b *treeBuilder) normNone(p ParseNode) {
	// panic("unimplemented")
}

func (b *treeBuilder) normParam(p ParseNode) {
	// panic("unimplemented")
}

func (b *treeBuilder) normParams(p ParseNode) {
	// panic("unimplemented")
}

func (b *treeBuilder) normString(p ParseNode) {
	// panic("unimplemented")
}

func (b *treeBuilder) normToken(p ParseNode) {
	// panic("unimplemented")
}
