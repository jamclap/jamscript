package script

import "unique"

func (t *typer) Type(m *Module) {
	t.funTypes = t.funTypes[:0]
	t.typeTypes = t.typeTypes[:0]
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
	// Stack of wanted types by labeled blocks/functions.
	// TODO Also stack of found types for the same.
	funTypes  []FunType
	typeTypes []TypeType
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
	case *Case:
		return t.typeCase(n, wanted, nil)
	case *Fun:
		return t.typeFun(n, wanted)
	case *Get:
		// TODO Resolve on the spot for members?
		// TODO Or retain subject type for later resolve?
	case *Ref:
		return t.typeRef(n, wanted)
	case *Switch:
		return t.typeSwitch(n, wanted)
	case *TokenNode:
		return t.typeToken(n, wanted)
	case *Value:
		return t.typeValue(n, wanted)
	case *Var:
		return t.typeVar(n, wanted)
	}
	return nil
}

func (t *typer) typeBlock(b *Block, wanted Type) Type {
	return t.typeBlockKids(b.Kids, wanted)
}

func (t *typer) typeBlockKids(kids []Node, wanted Type) Type {
	var typ Type = TypeVoid
	var kidWanted Type
	for i, n := range kids {
		// TODO Break with value, including labeled, complicates this.
		// TODO Resolve break/return targets so we can directly assign there.
		if i == len(kids)-1 {
			kidWanted = wanted
		}
		typ = t.typeNode(n, kidWanted)
	}
	return typ
}

func (t *typer) typeCall(c *Call, wanted Type) Type {
	wantedFunType := push(&t.funTypes, FunType{RetType: wanted})
	defer pop(&t.funTypes)
	calleeType := t.typeNode(c.Callee, wantedFunType)
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

func (t *typer) typeCase(c *Case, wanted Type, subjectWanted Type) Type {
	for _, pattern := range c.Patterns {
		t.typeNode(pattern, subjectWanted)
	}
	if c.Gate != nil {
		t.typeNode(c.Gate, TypeBool)
	}
	return t.typeBlockKids(c.Kids, wanted)
}

func (t *typer) typeFun(f *Fun, wanted Type) Type {
	// TODO If already typed, just fill in blanks.
	// TODO Could we have blanks only in type parameters?
	var retType Type
	var wantedRetType Type
	wantedFunType, wantedOk := wanted.(*FunType)
	if wantedOk {
		wantedRetType = wantedFunType.RetType
	}
	wantedTypeType := push(&t.typeTypes, TypeType{Type: wantedRetType})
	defer pop(&t.typeTypes)
	specType := t.typeNode(f.RetSpec, wantedTypeType)
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
		return &n.Type
	case *Var:
		return n.Type
	}
	return nil
}

func (t *typer) typeSwitch(s *Switch, wanted Type) Type {
	var typ Type
	var subjectType Type = TypeBool
	if s.Subject != nil {
		subjectType = t.typeNode(s.Subject, nil)
	}
	for i, k := range s.Kids {
		switch c := k.(type) {
		case *Case:
			caseType := t.typeCase(c, wanted, subjectType)
			if i == 0 {
				typ = caseType
			}
		default:
			t.typeNode(c, nil)
		}
	}
	return typ
}

func (t *typer) typeToken(tok *TokenNode, wanted Type) Type {
	// TODO Use this for integer literals to find when float?
	_ = tok
	_ = wanted
	return nil
}

func (t *typer) typeValue(value *Value, wanted Type) Type {
	_ = wanted
	switch value.Value.(type) {
	case int32:
		return TypeInt
	case string:
		return TypeString
	}
	return nil
}

func (t *typer) typeVar(v *Var, wanted Type) Type {
	typ := v.Type
	valueTyped := false
	switch v.TypeSpec {
	case nil:
		switch v.Value {
		case nil:
			switch wanted {
			case nil:
				// TODO Prioritize vartype specs over defaults.
				switch v.Name {
				case "i", "j", "k", "l", "m", "n":
					typ = TypeInt
				case "w", "x", "y", "Z":
					typ = TypeFloat
				}
			default:
				typ = wanted
			}
		default:
			typ = t.typeNode(v.Value, wanted)
			valueTyped = true
		}
	default:
		wantedTypeType := push(&t.typeTypes, TypeType{Type: wanted})
		defer pop(&t.typeTypes)
		typeType := t.typeNode(v.TypeSpec, wantedTypeType)
		if typeType, ok := typeType.(*TypeType); ok {
			typ = typeType.Type
		}
	}
	if v.Type == nil {
		// TODO Could we have blanks only in type parameters?
		v.Type = typ
	}
	if v.Value != nil && !valueTyped {
		t.typeNode(v.Value, v.Type)
	}
	// The var declaration itself is type nil.
	return nil
}
