package script

import "unique"

func Typify(m *Module) {
	t := typer{}
	t.typeRoot(m.Root.(*Block))
}

func NormType(t Type) Type {
	switch t.(type) {
	case nil:
		return TypeNone
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
	TypeString
	TypeVoid
)

type EitherType struct {
	YesType Type
	NoType  Type
}

type FunType struct {
	ParamTypes []Type
	RetType    Type
}

type ListType struct {
	ItemType Type
}

type TypeType struct {
	Type Type
}

type typer struct {
	// TODO Stack of wanted types by labeled blocks/functions.
	// TODO Also stack of found types for the same.
}

func (t *typer) typeRoot(b *Block) {
	for _, n := range b.Kids {
		t.typeNode(n, nil)
	}
}

func (t *typer) typeNode(node Node, wanted Type) Type {
	switch n := node.(type) {
	case *Block:
		return t.typeBlock(n, wanted)
	case *Call:
		return t.typeCall(n, wanted)
	case *Fun:
		return t.typeFun(n, wanted)
	case *Ref:
		return t.typeRef(n, wanted)
	case *TokenNode:
		return t.typeToken(n, wanted)
	case *Var:
		return t.typeVar(n, wanted)
	}
	return nil
}

func (t *typer) typeBlock(b *Block, wanted Type) Type {
	var typ Type = TypeVoid
	var kidWanted Type
	for i, n := range b.Kids {
		// TODO Break with value, including labeled, complicates this.
		if i == len(b.Kids)-1 {
			kidWanted = wanted
		}
		typ = t.typeNode(n, kidWanted)
	}
	return typ
}

func (t *typer) typeCall(c *Call, wanted Type) Type {
	calleeType := t.typeNode(c.Callee, &FunType{RetType: wanted})
	var retType Type
	funType, ok := calleeType.(*FunType)
	if ok {
		retType = funType.RetType
	}
	for i, a := range c.Args {
		var paramType Type
		if ok && i < len(funType.ParamTypes) {
			paramType = funType.ParamTypes[i]
		}
		t.typeNode(a, paramType)
	}
	return retType
}

func (t *typer) typeFun(f *Fun, wanted Type) Type {
	// TODO If already typed, just fill in blanks?
	// TODO Could we have blanks only in type parameters?
	var retType Type
	var wantedRetType Type
	wantedFunType, wantedOk := wanted.(*FunType)
	if wantedOk {
		wantedRetType = wantedFunType.RetType
	}
	specType := t.typeNode(f.RetSpec, &TypeType{Type: wantedRetType})
	if specTypeType, ok := specType.(*TypeType); ok {
		retType = specTypeType.Type
	}
	for i, p := range f.Params {
		var paramWanted Type
		if wantedOk && i < len(wantedFunType.ParamTypes) {
			paramWanted = wantedFunType.ParamTypes[i]
		}
		paramType := t.typeNode(p, paramWanted)
		if i < len(f.Type.ParamTypes) {
			f.Type.ParamTypes[i] = paramType
		} else {
			// TODO Independently allocated param types slices ok?
			f.Type.ParamTypes = append(f.Type.ParamTypes, paramType)
		}
	}
	for _, n := range f.Kids {
		t.typeNode(n, nil)
	}
	f.Type.RetType = retType
	return &f.Type
}

func (t *typer) typeRef(r *Ref, wanted Type) Type {
	_ = wanted
	switch n := r.Node.(type) {
	case *Fun:
		return n.Type
	case *Var:
		return n.Type
	}
	return nil
}

func (t *typer) typeToken(tok *TokenNode, wanted Type) Type {
	// TODO Use this for integer literals to find when float?
	_ = wanted
	switch tok.Kind {
	case TokenStringText:
		return TypeString
	}
	return nil
}

func (t *typer) typeVar(v *Var, wanted Type) Type {
	// TODO Resolve type spec.
	// TODO Type initializer.
	v.Type = wanted
	return nil
}
