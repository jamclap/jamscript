package script

import "unique"

func FindTypes(m *Module) {
	//
}

func NormType(t Type) Type {
	switch t.(type) {
	case BaseType:
		return t
	case unique.Handle[Type]:
		return t
	default:
		return unique.Make(t)
	}
}

type Type interface{}

type BaseType int

const (
	TypeNone BaseType = iota
	TypeBool
	TypeFloat
	TypeInt
	TypeVoid
)

type EitherType struct {
	YesType Type
	NoType  Type
}

type ListType struct {
	ItemType Type
}
