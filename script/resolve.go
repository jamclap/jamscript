package script

func (r *resolver) Resolve(m *Module) {
	if m.Tops == nil {
		m.Tops = make(map[string]Node)
	}
	r.levels = append(r.levels[:0], 0)
	r.scope = r.scope[:0]
	r.tops = m.Tops
	r.resolveRoot(m.Root.(*Block))
}

type Pair[A, B any] struct {
	First  A
	Second B
}

type resolver struct {
	core   map[string]Node
	levels []int // Indices into scope. TODO Useless? Or needed for closures?
	scope  []Pair[string, Node]
	tops   map[string]Node
}

func (r *resolver) popLevel() int {
	start := *last(&r.levels)
	pop(&r.levels)
	size := len(r.scope) - start
	r.scope = r.scope[:start]
	return size
}

func (r *resolver) pushLevel() {
	r.levels = append(r.levels, len(r.scope))
}

func (r *resolver) resolveRoot(root *Block) {
	// Clear current tops while retaining capacity.
	// TODO Presume they don't change?
	for k := range r.tops {
		delete(r.tops, k)
	}
	// Extract tops.
	// TODO Also struct fields.
Tops:
	for _, kid := range root.Kids {
		name := ""
		switch k := kid.(type) {
		case *Fun:
			name = k.Name
		case *Var:
			name = k.Name
		default:
			continue Tops
		}
		switch _, found := r.tops[name]; found {
		case true:
			// TODO Report error
		default:
			r.tops[name] = kid
		}
	}
	// Now recurse.
	for i, kid := range root.Kids {
		switch k := kid.(type) {
		case *Fun:
			r.resolveFun(k)
		case *Var:
			r.resolveVar(k)
		default:
			r.resolveNode(&root.Kids[i])
		}
	}
}

func (r *resolver) resolveBlock(b *Block) {
	r.pushLevel()
	for i := range b.Kids {
		r.resolveNode(&b.Kids[i])
	}
	r.popLevel()
}

func (r *resolver) resolveCall(c *Call) {
	r.resolveNode(&c.Callee)
	for i := range c.Args {
		r.resolveNode(&c.Args[i])
	}
}

func (r *resolver) resolveCase(c *Case) {
	r.pushLevel()
	// TODO Vars introduced in pattern matching.
	for _, pattern := range c.Patterns {
		r.resolveNode(&pattern)
	}
	r.resolveNode(&c.Gate)
	for _, kid := range c.Kids {
		r.resolveNode(&kid)
	}
	r.popLevel()
}

func (r *resolver) resolveFun(f *Fun) {
	if len(r.levels) > 1 {
		r.scope = append(r.scope, Pair[string, Node]{f.Name, f})
	}
	r.pushLevel()
	for _, p := range f.Params {
		r.resolveNode(&p)
	}
	for i := range f.Kids {
		r.resolveNode(&f.Kids[i])
	}
	f.Size = r.popLevel()
}

func (r *resolver) resolveGet(g *Get) {
	r.resolveNode(&g.Subject)
	// TODO Resolve member based on the type of subject.
}

func (r *resolver) resolveNode(node *Node) {
	switch n := (*node).(type) {
	case *Block:
		r.resolveBlock(n)
	case *Call:
		r.resolveCall(n)
	case *Case:
		r.resolveCase(n)
	case *Fun:
		r.resolveFun(n)
	case *Get:
		r.resolveGet(n)
	case *Return:
		r.resolveReturn(n)
	case *Switch:
		r.resolveSwitch(n)
	case *TokenNode:
		r.resolveToken(node, n)
	case *Var:
		r.resolveVar(n)
	}
}

func (r *resolver) resolveReturn(ret *Return) {
	// TODO Resolve label?
	r.resolveNode(&ret.Value)
}

func (r *resolver) resolveSwitch(s *Switch) {
	r.resolveNode(&s.Subject)
	for _, kid := range s.Kids {
		r.resolveNode(&kid)
	}
}

func (r *resolver) resolveToken(node *Node, t *TokenNode) {
	if t.Kind != TokenId {
		return
	}
	info := t.NodeInfo
	for i := len(r.scope) - 1; i >= 0; i-- {
		pair := r.scope[i]
		if pair.First == t.Text {
			// TODO Store side table of resolutions for later bulkier allocation?
			*node = &Ref{NodeInfo: info, Node: pair.Second}
			return
		}
	}
	if top, ok := r.tops[t.Text]; ok {
		// TODO Store side table of resolutions for later bulkier allocation?
		*node = &Ref{NodeInfo: info, Node: top}
		_ = top
	}
	if top, ok := r.core[t.Text]; ok {
		// TODO Store side table of resolutions for later bulkier allocation?
		// TODO Force top-level defs for imports? Focus on qualified access too?
		*node = &Ref{NodeInfo: info, Node: top}
		_ = top
	}
}

func (r *resolver) resolveVar(v *Var) {
	if len(r.levels) > 1 {
		v.Offset = len(r.scope)
		r.scope = append(r.scope, Pair[string, Node]{v.Name, v})
	}
	r.resolveNode(&v.TypeSpec)
	r.resolveNode(&v.Value)
}
