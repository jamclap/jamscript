package script

import (
	"log"
	"unique"
)

type Tree struct {
	Root    Node
	Flags   []NodeFlags
	Sources []Source
}

type Node interface{}

type Idx[T any] int

type NodeFlags uint32

const (
	NodeFlagPlug NodeFlags = 1 << iota
	NodeFlagPub
	NodeFlagNone NodeFlags = 0
)

type Block struct {
	Kids []Node
}

type Fun struct {
	Index  int
	Name   unique.Handle[string]
	Params []Var
	Ret    Node
	Kids   []Node
}

type Var struct {
	Index    int
	Name     unique.Handle[string]
	TypeInfo Node
}

// Side info for each node that's not expected to be used often.
type NodeInfo struct {
	Flags  NodeFlags
	Source Source
}

type Source struct {
	Path  unique.Handle[string]
	Range Range[rune]
}

type NodeKind int

const (
	NodeNone NodeKind = iota
	NodeArgs
	NodeBlock
	NodeFun
	NodeCall
	NodeString
	NodeVar
	NodeType
)

//go:generate stringer -type=NodeKind

type treeBuilder struct {
	nodes    []inNode   // TODO convert to array of interface later?
	infos    []NodeInfo // Same length as nodes.
	blocks   []inBlock  // TODO convert to flat array of Fun?
	funs     []inFun    // TODO convert to flat array of Fun?
	vars     []inVar    // TODO convert to flat array of Var?
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

type inFun struct {
	name   unique.Handle[string]
	params Range[inNode]
	ret    Idx[inNode]
	kids   Range[inNode]
}

type inVar struct {
	name     unique.Handle[string]
	typeInfo Idx[inNode]
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

func (b *treeBuilder) toTree() (t Tree) {
	log.Printf("norm done")
	log.Printf("nodes: %+v\n", b.nodes)
	log.Printf("infos: %+v\n", b.infos)
	log.Printf("blocks: %+v\n", b.blocks)
	log.Printf("funs: %+v\n", b.funs)
	log.Printf("vars: %+v\n", b.vars)
	funs := make([]Fun, len(b.funs))
	vars := make([]Var, len(b.vars))
	nodes := make([]Node, len(b.nodes))
	for i, node := range b.nodes {
		switch node.kind {
		case NodeFun:
			f := &funs[node.index]
			f.Index = i
			nodes[i] = f
		case NodeVar:
			v := &vars[node.index]
			v.Index = i
			nodes[i] = v
		}
	}
	t.Root = nodes[len(nodes)-1]
	return
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

func (b *treeBuilder) pushNode(node inNode) {
	// TODO Update source during work.
	b.nodes = append(b.nodes, node)
	b.infos = append(b.infos, NodeInfo{Source: b.source})
}

func (b *treeBuilder) pushWork(node inNode) {
	// TODO Update source during work.
	b.work = append(b.work, node)
	b.workInfo = append(b.workInfo, NodeInfo{Source: b.source})
}
