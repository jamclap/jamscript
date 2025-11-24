package script

import "unique"

func Norm(parse []ParseNode) Node {
	// TODO Convert internal repr to arrays then nodes
	return nil
}

type Node interface{}

type inTree struct {
	nodes []inNode // TODO convert to array of interface later?
	funs  []inFun  // TODO convert to flat array of Fun?
	vars  []inVar  // TODO convert to flat array of Var?
}

type inFun struct {
	id     unique.Handle[string]
	params []Range // Var
	kids   []Range // Node
}

type inNode struct {
	kind  NodeKind
	index int // array depends on Kind
}

type inVar struct {
	id unique.Handle[string]
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
