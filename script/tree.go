package script

import "unique"

type Node interface{}

type Idx[T any] int

type Ref[T any] struct {
	Slice *[]T
	Index int
}

func (r Ref[T]) Get() T {
	return (*r.Slice)[r.Index]
}

func (r Ref[T]) Set(value T) {
	(*r.Slice)[r.Index] = value
}

type RangeRef[T any] struct {
	Slice *[]T
	Range Range[T]
}

func (r RangeRef[T]) AssertIndex(index int) {
	if index >= r.Len() {
		panic("out of range")
	}
}

func (r RangeRef[T]) Len() int {
	return r.Range.End - r.Range.Start
}

func (r RangeRef[T]) Get(index int) T {
	r.AssertIndex(index)
	return (*r.Slice)[r.Range.Start+index]
}

func (r RangeRef[T]) Set(index int, value T) {
	r.AssertIndex(index)
	(*r.Slice)[r.Range.Start+index] = value
}

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
	// params RangeRef[inVar]
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
