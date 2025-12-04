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
	NodeFlagPlug NodeFlags = 1 << iota
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
	Index  int
	Callee Node
	Args   []Node
}

type Def struct {
	Name  string
	Flags NodeFlags
}

type Fun struct {
	NodeInfo
	Def
	Scope
	Type    FunType
	Params  []Node // always *Var
	RetSpec Node
	Kids    []Node
}

type Ref struct {
	NodeInfo
	Node Node
}

type Scope struct {
	// TODO for []any, which should only store pointers to slices
	Size int
}

type TokenNode struct {
	NodeInfo
	Token
}

type Var struct {
	NodeInfo
	Def
	Type     Type
	TypeSpec Node
	Value    Node
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
	NodeFun
	NodeString
	NodeToken
	NodeType
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
	case *Fun:
		if n.Flags&NodeFlagPub > 0 {
			print("pub ")
		}
		print("fun")
		if n.Name != "" {
			fmt.Printf(" %s", n.Name)
		}
		fmt.Printf("@%d", n.Index)
		// TODO If wide, print params on separate lines?
		print("(")
		for i, vnode := range n.Params {
			if i > 0 {
				print(", ")
			}
			v := vnode.(*Var)
			print(v.Name)
			fmt.Printf("@%d", v.Index)
		}
		print(")")
		p.printType(n.Type.RetType)
		println()
		nextIndent := indent + 1
		for _, kid := range n.Kids {
			PrintIndent(nextIndent)
			p.printAt(nextIndent, kid)
			println()
		}
		PrintIndent(indent)
		print("end")
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
	case *Var:
		print("var")
		fmt.Printf("@%d", n.Index)
		p.printType(n.Type)
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
	funs     []inFun
	tokens   []Token
	vars     []inVar // TODO Also workVars for contiguous params?
	work     []inNode
	workInfo []NodeInfo // Same length as work.
	source   Source
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

type inFun struct {
	Def
	params Range[inNode]
	// ret Idx[inNode]
	kids Range[inNode]
}

type inVar struct {
	Def
	// typeInfo Idx[inNode]
}

func newTreeBuilder() treeBuilder {
	// Init some with bogus at index 0 so valid are always nonzero.
	return treeBuilder{
		nodes:  make([]inNode, 1),
		infos:  make([]NodeInfo, 1),
		blocks: make([]inBlock, 1),
		funs:   make([]inFun, 1),
		vars:   make([]inVar, 1),
	}
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
	sources := make([]Source, len(b.nodes))
	blocks := make([]Block, len(b.blocks))
	calls := make([]Call, len(b.calls))
	funs := make([]Fun, len(b.funs))
	tokens := make([]TokenNode, len(b.tokens))
	vars := make([]Var, len(b.vars))
	for i, node := range b.nodes {
		switch node.kind {
		case NodeBlock:
			nodes[i] = &blocks[node.index]
		case NodeCall:
			nodes[i] = &calls[node.index]
		case NodeFun:
			nodes[i] = &funs[node.index]
		case NodeToken:
			nodes[i] = &tokens[node.index]
		case NodeVar:
			nodes[i] = &vars[node.index]
		}
		sources[i] = b.infos[i].Source
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
	for i, f := range b.funs {
		funs[i] = Fun{
			Def:    f.Def,
			Params: Slice(f.params, nodes),
			Kids:   Slice(f.kids, nodes),
		}
	}
	for i, tok := range b.tokens {
		tokens[i] = TokenNode{
			Token: tok,
		}
	}
	for i, v := range b.vars {
		vars[i] = Var{
			Def: v.Def,
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
		case NodeFun:
			f := &funs[node.index]
			f.Index = i
		case NodeToken:
			tok := &tokens[node.index]
			tok.Index = i
		case NodeVar:
			v := &vars[node.index]
			v.Index = i
		}
		sources[i] = b.infos[i].Source
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

func (b *treeBuilder) commit(parent inNode, start int) {
	// oldLen := len(b.nodes)
	b.nodes = append(b.nodes, b.work[start:]...)
	b.infos = append(b.infos, b.workInfo[start:]...)
	// parent := inParseNode{
	// 	kind: kind,
	// 	kids: Range[inParseNode]{Start: oldLen, End: len(p.nodes)},
	// }
	// TODO Update source during work.
	b.work = append(b.work[:start], parent)
	b.workInfo = append(b.workInfo[:start], NodeInfo{Source: b.source})
}

func (b *treeBuilder) commitBlock(start int) {
	oldLen := len(b.nodes)
	b.commit(inNode{kind: NodeBlock, index: len(b.blocks)}, start)
	b.blocks = append(
		b.blocks,
		inBlock{kids: Range[inNode]{oldLen, len(b.nodes)}},
	)
}

func (b *treeBuilder) popWork() {
	pop(&b.work)
	pop(&b.workInfo)
}

func (b *treeBuilder) popWorkBlock() inBlock {
	b.popWork()
	return pop(&b.blocks)
}

func (b *treeBuilder) pushWork(node inNode) {
	// TODO Update source during work.
	b.work = append(b.work, node)
	b.workInfo = append(b.workInfo, NodeInfo{Source: b.source})
}
