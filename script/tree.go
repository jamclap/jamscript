package script

import (
	"unique"
)

type Node interface{}

type Idx[T any] int

type Source struct {
	Path  unique.Handle[string]
	Start int
	End   int
}

type NodeKind int

const (
	NodeNone NodeKind = iota
	NodeArgs
	NodeBlock
	NodeFun
	NodeCall
	NodeParams
	NodeString
	NodeType
)

//go:generate stringer -type=NodeKind

type NodeFlags uint32

const (
	NodeFlagNone NodeFlags = 0
	NodeFlagPlug NodeFlags = 1 << iota
	NodeFlagPub
)

type treeBuilder struct {
	nodes   []inNode    // TODO convert to array of interface later?
	flags   []NodeFlags // Same length as nodes.
	sources []Source    // same length as nodes.
	source  Source
	funs    []inFun // TODO convert to flat array of Fun?
	vars    []inVar // TODO convert to flat array of Var?
	work    []inWork
}

type inNode struct {
	kind  NodeKind
	index int // array depends on Kind
}

type inFun struct {
	name   unique.Handle[string]
	params []Range[inVar]
	ret    Idx[inNode]
	kids   []Range[inNode]
	// params RangeRef[inVar]
}

type inVar struct {
	name     unique.Handle[string]
	typeInfo Idx[inNode]
}

type inWork struct {
	inNode
	NodeFlags
	Source
}

func newTreeBuilder() treeBuilder {
	// Init some with bogus at index 0 so valid are always nonzero.
	return treeBuilder{
		funs: make([]inFun, 1),
		vars: make([]inVar, 1),
	}
}

func (b *treeBuilder) pushWork(node inNode) {
	// TODO Update source during work.
	// TODO Separate work arrays for each part?
	b.work = append(b.work, inWork{inNode: node, Source: b.source})
}
