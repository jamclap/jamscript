package script

import "unique"

type Node interface{}

type Idx[T any] int

type Source struct {
	Path  unique.Handle[string]
	Start int
	End   int
}

type inTreeBuilder struct {
	tree    inTree
	working inTree
}

type inTree struct {
	nodes   []inNode // TODO convert to array of interface later?
	sources []Source // same length as `nodes` TODO Or need a ref idx?
	funs    []inFun  // TODO convert to flat array of Fun?
	vars    []inVar  // TODO convert to flat array of Var?
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
}

type inVar struct {
	name     unique.Handle[string]
	typeInfo Idx[inNode]
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
