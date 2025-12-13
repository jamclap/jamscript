package script

import (
	"fmt"
)

type Module struct {
	Core    map[string]Node
	Root    Node // always the last node?
	Sources []Source
	Tops    map[string]Node
}

type Node interface {
	// Idx() int
}

type Idx[T any] int

type NodeFlags uint32

const (
	NodeFlagCapture NodeFlags = 1 << iota
	NodeFlagGlobal
	NodeFlagPlug
	NodeFlagPub
	NodeFlagNone NodeFlags = 0
)

type Block struct {
	NodeInfo
	Kids []Node
}

func (b *Block) Idx() int {
	return b.Index
}

type Call struct {
	NodeInfo
	Callee Node
	Args   []Node
}

type Case struct {
	NodeInfo
	Always   bool
	Patterns []Node
	Gate     Node
	Kids     []Node
}

type Def struct {
	Name  string
	Flags NodeFlags
}

type Fun struct {
	NodeInfo
	Def
	Scope
	Level   int
	Type    FunType
	Params  []Node // always *Var
	RetSpec Node
	Kids    []Node
}

type Get struct {
	NodeInfo
	Subject Node
	Member  Node
}

type Switch struct {
	NodeInfo
	Subject Node
	Kids    []Node
}

type Ref struct {
	NodeInfo
	Node Node
}

type Return struct {
	NodeInfo
	Kind  TokenKind
	Label Node // Required for break.
	Value Node
}

type Scope struct {
	// TODO for []any, which should only store pointers to slices
	Size int
}

type TokenNode struct {
	NodeInfo
	Token
}

type Value struct {
	NodeInfo
	Value any
}

type Var struct {
	NodeInfo
	Def
	Type     Type
	TypeSpec Node
	Value    Node
	Level    int
	Offset   int
}

// Side info for each node that's not expected to be used often.
type NodeInfo struct {
	Index  int
	Source Source
}

type Source struct {
	Path  *string // TODO Module pointer
	Range Range[rune]
}

type NodeKind int

const (
	NodeNone NodeKind = iota
	NodeArgs
	NodeBlock
	NodeCall
	NodeCase
	NodeFun
	NodeGet
	NodeReturn
	NodeSwitch
	NodeToken
	NodeType
	NodeValue
	NodeVar
)

//go:generate stringer -type=NodeKind

type TreePrinter struct {
	Tree *Module
	TreePrinterOptions
}

type TreePrinterOptions struct {
	// TODO Options
}

func (t *Module) Print() {
	p := treePrinting{TreePrinter: TreePrinter{Tree: t}}
	p.printAt(0, t.Root)
}

type treePrinting struct {
	TreePrinter
}

func (p *treePrinting) printAt(indent int, node Node) {
	switch n := node.(type) {
	case nil:
		print("nil")
	case *Block:
		nextIndent := indent
		atRoot := node == p.Tree.Root
		if !atRoot {
			println("then")
			nextIndent++
		}
		for i, kid := range n.Kids {
			if atRoot && i > 0 {
				println()
			}
			PrintIndent(nextIndent)
			p.printAt(nextIndent, kid)
			println()
		}
		if !atRoot {
			PrintIndent(indent)
			print("end")
		}
	case *Call:
		p.printAt(indent, n.Callee)
		print("(")
		for i, a := range n.Args {
			if i > 0 {
				print(", ")
			}
			p.printAt(indent, a)
		}
		print(")")
	case *Case:
		switch {
		case n.Always:
			print("else")
		default:
			print("case")
		}
		for i, m := range n.Patterns {
			switch i {
			case 0:
				print(" ")
			default:
				print(", ")
			}
			p.printAt(indent, m)
		}
		p.printKids(indent, n.Kids, true)
	case *Fun:
		if n.Flags&NodeFlagPub > 0 {
			print("pub ")
		}
		print("fun")
		if n.Name != "" {
			fmt.Printf(" %s", n.Name)
		}
		fmt.Printf("@(%d,%d)", n.Index, n.Level)
		// TODO If wide, print params on separate lines?
		print("(")
		for i, vnode := range n.Params {
			if i > 0 {
				print(", ")
			}
			p.printVar(vnode.(*Var), indent)
		}
		print(")")
		p.printType(n.Type.RetType)
		p.printKids(indent, n.Kids, false)
		PrintIndent(indent)
		print("end")
	case *Get:
		p.printAt(indent, n.Subject)
		print(".")
		p.printAt(indent, n.Member)
	case *Return:
		switch n.Kind {
		case TokenBreak:
			print("break")
		case TokenContinue:
			print("continue")
		case TokenReturn:
			print("return")
		}
		// TODO Label.
		if n.Value != nil {
			print(" ")
			p.printAt(indent, n.Value)
		}
	case *TokenNode:
		switch n.Kind {
		case TokenStringText:
			PrintEscapedString(n.Text)
		default:
			print(n.Text)
		}
	case *Ref:
		switch d := n.Node.(type) {
		case *Fun:
			print(d.Name)
			fmt.Printf("@%d", d.Index)
		case *Var:
			print(d.Name)
			fmt.Printf("@%d", d.Index)
		}
	case *Switch:
		print("switch")
		if n.Subject != nil {
			p.printAt(indent, n.Subject)
		}
		p.printKids(indent, n.Kids, false)
		PrintIndent(indent)
		print("end")
	case *Value:
		switch v := n.Value.(type) {
		case string:
			PrintEscapedString(v)
		default:
			fmt.Printf("%v", n.Value)
		}
	case *Var:
		print("var ")
		p.printVar(n, indent)
	}
}

func (p *treePrinting) printKids(indent int, kids []Node, endless bool) {
	println()
	nextIndent := indent + 1
	for i, kid := range kids {
		PrintIndent(nextIndent)
		p.printAt(nextIndent, kid)
		if !endless || i < len(kids)-1 {
			println()
		}
	}
}

func (p *treePrinting) printType(t Type) {
	switch t {
	case nil, TypeNone:
		print(" Unknown")
	default:
		print(" SomeType")
	}
}

func (p *treePrinting) printVar(n *Var, indent int) {
	print(n.Name)
	fmt.Printf("@(%d,%d/%d)", n.Index, n.Level, n.Offset)
	p.printType(n.Type)
	if n.Value != nil {
		print(" = ")
		p.printAt(indent, n.Value)
	}
}

func PrintEscapedString(s string) {
	print("\"")
	for _, r := range s {
		switch r {
		case '"', '\\':
			print("\\")
			fmt.Printf("%c", r)
		case '\n':
			print("\\n")
		case '\r':
			print("\\r")
		case '\t':
			print("\\t")
		default:
			switch {
			case r < 0x20 || r > 0x7e:
				print("\\u(")
				fmt.Printf("%x", r)
				print(')')
			default:
				fmt.Printf("%c", r)
			}
		}
	}
	print("\"")
}

type treeBuilder struct {
	nodes    []inNode   // TODO convert to array of interface later?
	infos    []NodeInfo // Same length as nodes.
	blocks   []inBlock
	calls    []inCall
	cases    []inCase
	funs     []inFun
	gets     []inGet
	returns  []inReturn
	tokens   []Token
	values   []any
	vars     []inVar // TODO Also workVars for contiguous params?
	work     []inNode
	workInfo []NodeInfo // Same length as work.
	source   Source
	switches []inSwitch
}

type inNode struct {
	kind  NodeKind
	index int // array depends on Kind
}

type inBlock struct {
	kids Range[inNode]
}

type inCall struct {
	callee Idx[inNode]
	args   Range[inNode]
}

type inCase struct {
	always   bool // true for else rather than case
	patterns Range[inNode]
	gate     Idx[inNode]
	kids     Range[inNode]
}

type inFun struct {
	Def
	params Range[inNode]
	// ret Idx[inNode]
	kids Range[inNode]
}

type inGet struct {
	subject Idx[inNode]
	member  Idx[inNode]
}

type inReturn struct {
	kind  TokenKind
	label Idx[inNode] // Required for break.
	value Idx[inNode]
}

type inSwitch struct {
	subject Idx[inNode]
	kids    Range[inNode]
}

type inVar struct {
	Def
	typ   Idx[inNode]
	value Idx[inNode]
}

func newTreeBuilder() treeBuilder {
	// Init some with bogus at index 0 so valid are always nonzero.
	return treeBuilder{
		nodes:    make([]inNode, 1),
		infos:    make([]NodeInfo, 1),
		cases:    make([]inCase, 1),
		blocks:   make([]inBlock, 1),
		funs:     make([]inFun, 1),
		gets:     make([]inGet, 1),
		returns:  make([]inReturn, 1),
		switches: make([]inSwitch, 1),
		vars:     make([]inVar, 1),
	}
}

func (b *treeBuilder) reset() {
	// Start at 1.
	// TODO Any changes needed here?
	b.nodes = b.nodes[:1]
	b.infos = b.infos[:1]
	b.blocks = b.blocks[:1]
	b.cases = b.cases[:1]
	b.funs = b.funs[:1]
	b.gets = b.gets[:1]
	b.returns = b.returns[:1]
	b.vars = b.vars[:1]
	b.switches = b.switches[:1]
	// Start at 0. TODO Should these start at 1 also?
	b.calls = b.calls[:0]
	b.tokens = b.tokens[:0]
	b.work = b.work[:0]
	b.workInfo = b.workInfo[:0]
	b.source = Source{}
	b.values = b.values[:0]
}

func (b *treeBuilder) toTree() *Module {
	// log.Printf("norm done")
	// log.Printf("nodes: %+v\n", b.nodes)
	// log.Printf("infos: %+v\n", b.infos)
	// log.Printf("blocks: %+v\n", b.blocks)
	// log.Printf("calls: %+v\n", b.calls)
	// log.Printf("funs: %+v\n", b.funs)
	// log.Printf("tokens: %+v\n", b.tokens)
	// log.Printf("vars: %+v\n", b.vars)
	nodes := make([]Node, len(b.nodes))
	blocks := make([]Block, len(b.blocks))
	calls := make([]Call, len(b.calls))
	cases := make([]Case, len(b.cases))
	funs := make([]Fun, len(b.funs))
	gets := make([]Get, len(b.gets))
	returns := make([]Return, len(b.returns))
	switches := make([]Switch, len(b.switches))
	tokens := make([]TokenNode, len(b.tokens))
	values := make([]Value, len(b.values))
	vars := make([]Var, len(b.vars))
	for i, node := range b.nodes {
		switch node.kind {
		case NodeBlock:
			nodes[i] = &blocks[node.index]
		case NodeCall:
			nodes[i] = &calls[node.index]
		case NodeCase:
			nodes[i] = &cases[node.index]
		case NodeFun:
			nodes[i] = &funs[node.index]
		case NodeGet:
			nodes[i] = &gets[node.index]
		case NodeReturn:
			nodes[i] = &returns[node.index]
		case NodeSwitch:
			nodes[i] = &switches[node.index]
		case NodeToken:
			nodes[i] = &tokens[node.index]
		case NodeValue:
			nodes[i] = &values[node.index]
		case NodeVar:
			nodes[i] = &vars[node.index]
		}
	}
	for i, b := range b.blocks {
		blocks[i] = Block{
			Kids: Slice(b.kids, nodes),
		}
	}
	for i, c := range b.calls {
		calls[i] = Call{
			Callee: nodes[c.callee],
			Args:   Slice(c.args, nodes),
		}
	}
	for i, c := range b.cases {
		cases[i] = Case{
			Always:   c.always,
			Patterns: Slice(c.patterns, nodes),
			Gate:     nodes[c.gate],
			Kids:     Slice(c.kids, nodes),
		}
	}
	for i, f := range b.funs {
		funs[i] = Fun{
			Def:    f.Def,
			Params: Slice(f.params, nodes),
			Kids:   Slice(f.kids, nodes),
		}
	}
	for i, g := range b.gets {
		gets[i] = Get{
			Subject: nodes[g.subject],
			Member:  nodes[g.member],
		}
	}
	for i, r := range b.returns {
		returns[i] = Return{
			Kind:  r.kind,
			Label: nodes[r.label],
			Value: nodes[r.value],
		}
	}
	for i, s := range b.switches {
		switches[i] = Switch{
			Subject: nodes[s.subject],
			Kids:    Slice(s.kids, nodes),
		}
	}
	for i, tok := range b.tokens {
		tokens[i] = TokenNode{
			Token: tok,
		}
	}
	for i, v := range b.values {
		values[i] = Value{
			Value: v,
		}
	}
	for i, v := range b.vars {
		vars[i] = Var{
			Def:   v.Def,
			Value: nodes[v.value],
		}
	}
	for i, node := range b.nodes {
		switch node.kind {
		case NodeBlock:
			b := &blocks[node.index]
			b.Index = i
		case NodeCall:
			c := &calls[node.index]
			c.Index = i
		case NodeCase:
			c := &cases[node.index]
			c.Index = i
		case NodeFun:
			f := &funs[node.index]
			f.Index = i
		case NodeGet:
			g := &gets[node.index]
			g.Index = i
		case NodeReturn:
			r := &returns[node.index]
			r.Index = i
		case NodeSwitch:
			s := &switches[node.index]
			s.Index = i
		case NodeToken:
			tok := &tokens[node.index]
			tok.Index = i
		case NodeValue:
			v := &values[node.index]
			v.Index = i
		case NodeVar:
			v := &vars[node.index]
			v.Index = i
		}
	}
	// log.Printf("copy done\n")
	// log.Printf("nodes: %+v\n", nodes)
	// log.Printf("blocks: %+v\n", blocks)
	// log.Printf("calls: %+v\n", calls)
	// log.Printf("funs: %+v\n", funs)
	// log.Printf("tokens: %+v\n", tokens)
	// log.Printf("vars: %+v\n", vars)
	return &Module{
		Core: map[string]Node{},
		Root: nodes[len(nodes)-1],
	}
}

func (b *treeBuilder) commitHeadless(start int) {
	b.nodes = append(b.nodes, b.work[start:]...)
	b.infos = append(b.infos, b.workInfo[start:]...)
	b.work = b.work[:start]
	b.workInfo = b.workInfo[:start]
}

func (b *treeBuilder) commit(parent inNode, start int) {
	b.commitHeadless(start)
	// TODO Update source during work.
	b.work = append(b.work, parent)
	b.workInfo = append(b.workInfo, NodeInfo{Source: b.source})
}

func (b *treeBuilder) commitBlock(start int) {
	oldLen := len(b.nodes)
	b.commit(inNode{kind: NodeBlock, index: len(b.blocks)}, start)
	b.blocks = append(
		b.blocks,
		inBlock{kids: Range[inNode]{oldLen, len(b.nodes)}},
	)
}

func (b *treeBuilder) latest(start int, latestStart int) Idx[inNode] {
	latestEnd := len(b.work)
	switch {
	case latestEnd > latestStart:
		return Idx[inNode](len(b.nodes) + latestEnd - start)
	default:
		return 0
	}
}

func (b *treeBuilder) popWork() {
	pop(&b.work)
	pop(&b.workInfo)
}

func (b *treeBuilder) popWorkBlock() Range[inNode] {
	b.popWork()
	return pop(&b.blocks).kids
}

func (b *treeBuilder) pushWork(node inNode) {
	// TODO Update source during work.
	b.work = append(b.work, node)
	b.workInfo = append(b.workInfo, NodeInfo{Source: b.source})
}
