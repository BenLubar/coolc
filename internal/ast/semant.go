package ast

import (
	"fmt"
	"go/token"
)

var errorIdent = &Ident{
	Pos:  token.NoPos,
	Name: "$error$",
}

var errorClass = &Class{
	Type:    errorIdent,
	Formals: nil,
	Extends: &Extends{
		Type: errorIdent,
	},
	Features: nil,
}

var nativeClass = &Class{}

var nullClass = &Class{
	Type: &Ident{
		Pos:  token.NoPos,
		Name: "Null",
	},
	Formals: nil,
	Extends: &Extends{
		Type: &Ident{
			Pos:  token.NoPos,
			Name: "native",
		},
	},
	Features: nil,
}

var nothingClass = &Class{
	Type: &Ident{
		Pos:  token.NoPos,
		Name: "Nothing",
	},
	Formals: nil,
	Extends: &Extends{
		Type: &Ident{
			Pos:  token.NoPos,
			Name: "native",
		},
	},
	Features: nil,
}

func (p *Program) Semant(fset *token.FileSet) (haveErrors bool) {
	report := func(pos token.Pos, message string) {
		fmt.Printf("%v: %s\n", fset.Position(pos), message)
		haveErrors = true
	}

	p.classMap = map[string]*Class{
		"Nothing": nothingClass,
		"Null":    nullClass,
	}

	for _, c := range p.Classes {
		if o, ok := p.classMap[c.Type.Name]; ok {
			report(c.Type.Pos, "duplicate declaration of class "+c.Type.Name)
			report(o.Type.Pos, "(previous declaration was here)")
		} else {
			p.classMap[c.Type.Name] = c
		}
	}

	// If we continue with duplicate class names, we'll probably give a
	// bunch of useless errors.
	if haveErrors {
		return
	}

	for _, c := range p.Classes {
		c.semantTypes(func(id *Ident) {
			if c, ok := p.classMap[id.Name]; ok {
				id.Class = c
				return
			}

			report(id.Pos, "use of undeclared class "+id.Name)
			id.Class = errorClass
		})
	}

	if haveErrors {
		return
	}

	any := p.classMap["Any"]
	unit := p.classMap["Unit"]
	main := p.classMap["Main"]
	int_ := p.classMap["Int"]
	boolean := p.classMap["Boolean"]

	if any == nil {
		report(token.NoPos, "missing required class: Any")
	}

	if unit == nil {
		report(token.NoPos, "missing required class: Unit")
	}

	if main == nil {
		report(token.NoPos, "missing required class: Main")
	}

	if int_ == nil {
		report(token.NoPos, "missing required class: Int")
	}

	if boolean == nil {
		report(token.NoPos, "missing required class: Boolean")
	}

	less := func(t1, t2 *Class) bool {
		// S-Self
		if t1 == t2 {
			return true
		}
		// S-Nothing
		if t1 == nothingClass {
			return true
		}
		// S-Null
		if t1 == nullClass {
			return t2 != nothingClass && t2 != boolean && t2 != int_ && t2 != unit
		}
		// S-Extends
		for t := t1; t != nativeClass; t = t.Extends.Type.Class {
			if t == t2 {
				return true
			}
		}
		return false
	}

	lub := func(ts ...*Class) *Class {
		t1 := nothingClass

		for _, t2 := range ts {
			// G-Compare
			if less(t1, t2) {
				t1 = t2
				continue
			}
			// G-Commute
			if less(t2, t1) {
				continue
			}
			// G-Extends
			for t1.Depth > t2.Depth {
				t1 = t1.Extends.Type.Class
			}
			for t2.Depth > t1.Depth {
				t2 = t2.Extends.Type.Class
			}
			for t1 != t2 {
				if t1.Depth == 0 {
					t1 = any
					break
				}
				t1 = t1.Extends.Type.Class
				t2 = t2.Extends.Type.Class
			}
		}

		return t1
	}

	children := make(map[*Class][]*Class)
	var free []*Class

	for _, c := range p.Classes {
		if c == any {
			c.Depth = 1
			free = append(free, c)
		} else {
			children[c.Extends.Type.Class] = append(children[c.Extends.Type.Class], c)
		}
	}

	order := 0
	for len(free) != 0 {
		c := free[0]

		for _, cc := range children[c] {
			cc.Depth = c.Depth + 1
		}

		free = append(children[c], free[1:]...)
		delete(children, c)

		order++
		c.Order = order
		for p := c; p != nativeClass; p = p.Extends.Type.Class {
			p.MaxOrder = order
		}
		p.Ordered = append(p.Ordered, c)
	}

	for _, c := range p.Classes {
		if _, ok := children[c]; ok {
			report(c.Type.Pos, "class heirarchy loop: "+c.Type.Name)
		}
	}

	if haveErrors {
		return
	}

	for _, c := range p.Classes {
		c.semantMakeConstructor(unit)
	}

	for _, c := range p.Ordered {
		c.semantMethods(report, less)
	}

	less_report := func(t1 *Class, id *Ident) {
		t2 := id.Class

		if !less(t1, t2) {
			report(id.Pos, "type "+t1.Type.Name+" does not conform to type "+t2.Type.Name)
		}
	}
	for _, c := range p.Classes {
		c.semantIdentifiers(report, less_report, lub)
	}

	p.Main = &StaticCallExpr{
		Recv: &AllocExpr{
			Type: &Ident{
				Pos:   token.NoPos,
				Name:  "Main",
				Class: main,
			},
		},
		Name: &Ident{
			Pos:  token.NoPos,
			Name: "Main",
		},
	}
	p.Main.semantIdentifiers(report, less_report, lub, nil)

	return
}

func (c *Class) semantTypes(lookup func(*Ident)) {
	lookup(c.Type)
	for _, f := range c.Formals {
		f.semantTypes(lookup, c)
	}
	for _, f := range c.Features {
		f.semantTypes(lookup, c)
	}
	if c.Extends.Type.Name == "native" {
		switch c.Type.Name {
		case "Any", "Null", "Nothing":
			c.Extends.Type.Class = nativeClass
			return
		}
	}
	c.Extends.semantTypes(lookup, c)
}

func (c *Class) semantMakeConstructor(unit *Class) {
	c.Features = append(make([]Feature, len(c.Formals)), c.Features...)
	for i, f := range c.Formals {
		c.Features[i] = &Attribute{
			Name: f.Name,
			Type: f.Type,
			Init: &NameExpr{
				Name: &Ident{
					Pos:  f.Name.Pos,
					Name: "'" + f.Name.Name,
				},
			},

			Parent: c,
		}
		f.Name = &Ident{
			Pos:  f.Name.Pos,
			Name: "'" + f.Name.Name,
		}
	}

	var constructor Expr = &ThisExpr{
		Pos: c.Type.Pos,

		Class: c,
	}

	for i := len(c.Features) - 1; i >= 0; i-- {
		switch f := c.Features[i].(type) {
		case *Init:
			constructor = &ChainExpr{
				Pre:  f.Expr,
				Expr: constructor,
			}

		case *Attribute:
			constructor = &ChainExpr{
				Pre: &AssignExpr{
					Name: f.Name,
					Expr: f.Init,

					Unit: &Ident{
						Pos:   f.Name.Pos,
						Name:  "Unit",
						Class: unit,
					},
				},
				Expr: constructor,
			}
		}
	}

	switch c.Type.Name {
	case "Int", "Boolean", "String", "Unit", "Symbol":
		return

	case "ArrayAny":
		constructor = &NativeExpr{
			Pos: c.Type.Pos,
		}

	case "Any":
		// don't call super-constructor because there isn't one.

	default:
		constructor = &ChainExpr{
			Pre: &StaticCallExpr{
				Recv: &ThisExpr{
					Pos:   c.Extends.Type.Pos,
					Class: c.Extends.Type.Class,
				},
				Name: &Ident{
					Pos:  c.Extends.Type.Pos,
					Name: c.Extends.Type.Name,
				},
				Args: c.Extends.Args,
			},
			Expr: constructor,
		}
	}

	c.Features = append(c.Features, &Method{
		Name: &Ident{
			Pos:  c.Type.Pos,
			Name: c.Type.Name,
		},
		Args: c.Formals,
		Type: c.Type,
		Body: constructor,

		Parent: c,
	})
}

func (c *Class) semantMethods(report func(token.Pos, string), less func(*Class, *Class) bool) {
	parent := c.Extends.Type.Class
	c.Methods = make([]*Method, len(parent.Methods))
	copy(c.Methods, parent.Methods)

	used := make(map[string]token.Pos)

	for _, f := range c.Features {
		if m, ok := f.(*Method); ok {
			if o, ok := used[m.Name.Name]; ok {
				report(m.Name.Pos, "duplicate declaration of "+m.Name.Name)
				report(o, "(previous declaration was here)")
				continue
			}
			used[m.Name.Name] = m.Name.Pos

			if m.Name.Name == c.Type.Name {
				// don't put the constructor in the method table
				continue
			}

			var override *Method

			for _, o := range parent.Methods {
				if o.Name.Name == m.Name.Name {
					override = o
					break
				}
			}

			if !m.Override {
				if override == nil {
					// easiest case, just add the method
					m.Order = len(c.Methods)
					c.Methods = append(c.Methods, m)
				} else {
					report(m.Name.Pos, "missing 'override' on method "+c.Type.Name+"."+m.Name.Name)
					report(override.Name.Pos, "(previous declaration was here)")
				}
			} else {
				if override == nil {
					report(m.Name.Pos, "missing parent for 'override' method "+c.Type.Name+"."+m.Name.Name)

					// add the method anyway. we won't generate
					// code, so this isn't a problem.
					m.Order = len(c.Methods)
					c.Methods = append(c.Methods, m)
				} else {
					m.Order = override.Order
					c.Methods[override.Order] = m

					if len(m.Args) != len(override.Args) {
						report(m.Name.Pos, "invalid override: method "+c.Type.Name+"."+m.Name.Name+" has the wrong number of arguments")
						report(override.Name.Pos, "(parent declaration is here)")
					} else {
						for i, a := range m.Args {
							if t1, t2 := a.Type, override.Args[i].Type; t1.Class != t2.Class {
								report(t1.Pos, "invalid override: method "+c.Type.Name+"."+m.Name.Name+" has an incorrect argument type")
								report(t2.Pos, "(parent declaration is here)")
							}
						}
					}

					if !less(m.Type.Class, override.Type.Class) {
						report(m.Type.Pos, "invalid override: method "+c.Type.Name+"."+m.Name.Name+" has incompatible return type "+m.Type.Name)
						report(override.Type.Pos, "(parent return type is "+override.Type.Name+")")
					}
				}
			}
		}
	}
}

type semantIdentifier struct {
	Name   *Ident
	Type   *Ident
	Object Object
}

type semantIdentifiers []*semantIdentifier

func (ids semantIdentifiers) Lookup(name string) *semantIdentifier {
	for i := len(ids) - 1; i >= 0; i-- {
		if ids[i].Name.Name == name {
			return ids[i]
		}
	}
	return nil
}

func (c *Class) semantInheritedIdentifiers(report func(string)) semantIdentifiers {
	if c == nativeClass {
		return nil
	}

	ids := c.Extends.Type.Class.semantInheritedIdentifiers(report)
	for _, f := range c.Features {
		if a, ok := f.(*Attribute); ok {
			ids = append(ids, &semantIdentifier{
				Name:   a.Name,
				Type:   a.Type,
				Object: a,
			})
			if _, ok := a.Init.(*NativeExpr); ok {
				report("cannot extend " + c.Type.Name)
			}
		}
	}
	return ids
}

func (c *Class) semantIdentifiers(report func(token.Pos, string), less func(*Class, *Ident), lub func(...*Class) *Class) {
	ids := c.Extends.Type.Class.semantInheritedIdentifiers(func(s string) {
		report(c.Extends.Type.Pos, s)
	})
	used := make(map[string]token.Pos)
	for _, id := range ids {
		used[id.Name.Name] = id.Name.Pos
	}
	for _, f := range c.Features {
		if a, ok := f.(*Attribute); ok {
			a.Name.Object = a
			if a.Type.Class == nothingClass {
				report(a.Type.Pos, "cannot declare attribute of type Nothing")
			}
			if pos, ok := used[a.Name.Name]; ok {
				report(a.Name.Pos, "duplicate declaration of "+a.Name.Name)
				report(pos, "(previous declaration was here)")
			} else {
				ids = append(ids, &semantIdentifier{
					Name:   a.Name,
					Type:   a.Type,
					Object: a,
				})
				used[a.Name.Name] = a.Name.Pos
			}
		}
	}

	for _, f := range c.Features {
		if m, ok := f.(*Method); ok {
			m.semantIdentifiers(report, less, lub, ids)
		}
	}
}

func (e *Extends) semantTypes(lookup func(*Ident), c *Class) {
	lookup(e.Type)
	for _, a := range e.Args {
		a.semantTypes(lookup, c)
	}
}

func (f *Init) semantTypes(lookup func(*Ident), c *Class) {
	f.Expr.semantTypes(lookup, c)
}

func (f *Attribute) semantTypes(lookup func(*Ident), c *Class) {
	f.Parent = c

	if _, ok := f.Init.(*NativeExpr); ok {
		switch {
		case c.Type.Name == "Int" && f.Name.Name == "value":
			return
		case c.Type.Name == "Boolean" && f.Name.Name == "value":
			return
		case c.Type.Name == "String" && f.Name.Name == "str_field":
			return
		case c.Type.Name == "ArrayAny" && f.Name.Name == "array_field":
			return
		}
	}
	lookup(f.Type)
	f.Init.semantTypes(lookup, c)
}

func (f *Method) semantTypes(lookup func(*Ident), c *Class) {
	f.Parent = c

	for _, a := range f.Args {
		a.semantTypes(lookup, c)
	}
	lookup(f.Type)
	if _, ok := f.Body.(*NativeExpr); ok {
		switch {
		case c.Type.Name == "Any" && f.Name.Name == "toString":
			return
		case c.Type.Name == "Any" && f.Name.Name == "equals":
			return
		case c.Type.Name == "IO" && f.Name.Name == "abort":
			return
		case c.Type.Name == "IO" && f.Name.Name == "out":
			return
		case c.Type.Name == "IO" && f.Name.Name == "in":
			return
		case c.Type.Name == "IO" && f.Name.Name == "symbol":
			return
		case c.Type.Name == "IO" && f.Name.Name == "symbol_name":
			return
		case c.Type.Name == "Int" && f.Name.Name == "equals":
			return
		case c.Type.Name == "String" && f.Name.Name == "equals":
			return
		case c.Type.Name == "String" && f.Name.Name == "concat":
			return
		case c.Type.Name == "String" && f.Name.Name == "substring":
			return
		case c.Type.Name == "String" && f.Name.Name == "charAt":
			return
		case c.Type.Name == "ArrayAny" && f.Name.Name == "get":
			return
		case c.Type.Name == "ArrayAny" && f.Name.Name == "set":
			return
		case c.Type.Name == "ArrayAny" && f.Name.Name == "ArrayAny":
			return
		}
	}
	f.Body.semantTypes(lookup, c)
}

func (f *Method) semantIdentifiers(report func(token.Pos, string), less func(*Class, *Ident), lub func(...*Class) *Class, ids semantIdentifiers) {
	used := make(map[string]token.Pos)
	for _, a := range f.Args {
		if o, ok := used[a.Name.Name]; ok {
			report(a.Name.Pos, "duplicate declaration of "+a.Name.Name)
			report(o, "(previous declaration was here)")
		} else {
			ids = append(ids, &semantIdentifier{
				Name:   a.Name,
				Type:   a.Type,
				Object: a,
			})
			used[a.Name.Name] = a.Name.Pos
		}
	}

	less(f.Body.semantIdentifiers(report, less, lub, ids), f.Type)
}

func (a *Formal) semantTypes(lookup func(*Ident), c *Class) {
	lookup(a.Type)
}

func (e *NotExpr) semantTypes(lookup func(*Ident), c *Class) {
	lookup(e.Boolean)
	e.Expr.semantTypes(lookup, c)
}

func (e *NotExpr) semantIdentifiers(report func(token.Pos, string), less func(*Class, *Ident), lub func(...*Class) *Class, ids semantIdentifiers) *Class {
	less(e.Expr.semantIdentifiers(report, less, lub, ids), e.Boolean)
	return e.Boolean.Class
}

func (e *NegativeExpr) semantTypes(lookup func(*Ident), c *Class) {
	lookup(e.Int)
	e.Expr.semantTypes(lookup, c)
}

func (e *NegativeExpr) semantIdentifiers(report func(token.Pos, string), less func(*Class, *Ident), lub func(...*Class) *Class, ids semantIdentifiers) *Class {
	less(e.Expr.semantIdentifiers(report, less, lub, ids), e.Int)
	return e.Int.Class
}

func (e *IfExpr) semantTypes(lookup func(*Ident), c *Class) {
	lookup(e.Boolean)
	e.Cond.semantTypes(lookup, c)
	e.Then.semantTypes(lookup, c)
	e.Else.semantTypes(lookup, c)
}

func (e *IfExpr) semantIdentifiers(report func(token.Pos, string), less func(*Class, *Ident), lub func(...*Class) *Class, ids semantIdentifiers) *Class {
	less(e.Cond.semantIdentifiers(report, less, lub, ids), e.Boolean)
	return lub(e.Then.semantIdentifiers(report, less, lub, ids), e.Else.semantIdentifiers(report, less, lub, ids))
}

func (e *WhileExpr) semantTypes(lookup func(*Ident), c *Class) {
	lookup(e.Boolean)
	lookup(e.Unit)
	e.Cond.semantTypes(lookup, c)
	e.Body.semantTypes(lookup, c)
}

func (e *WhileExpr) semantIdentifiers(report func(token.Pos, string), less func(*Class, *Ident), lub func(...*Class) *Class, ids semantIdentifiers) *Class {
	less(e.Cond.semantIdentifiers(report, less, lub, ids), e.Boolean)
	e.Body.semantIdentifiers(report, less, lub, ids)
	return e.Unit.Class
}

func (e *LessOrEqualExpr) semantTypes(lookup func(*Ident), c *Class) {
	lookup(e.Boolean)
	lookup(e.Int)
	e.Left.semantTypes(lookup, c)
	e.Right.semantTypes(lookup, c)
}

func (e *LessOrEqualExpr) semantIdentifiers(report func(token.Pos, string), less func(*Class, *Ident), lub func(...*Class) *Class, ids semantIdentifiers) *Class {
	less(e.Left.semantIdentifiers(report, less, lub, ids), e.Int)
	less(e.Right.semantIdentifiers(report, less, lub, ids), e.Int)
	return e.Boolean.Class
}

func (e *LessThanExpr) semantTypes(lookup func(*Ident), c *Class) {
	lookup(e.Boolean)
	lookup(e.Int)
	e.Left.semantTypes(lookup, c)
	e.Right.semantTypes(lookup, c)
}

func (e *LessThanExpr) semantIdentifiers(report func(token.Pos, string), less func(*Class, *Ident), lub func(...*Class) *Class, ids semantIdentifiers) *Class {
	less(e.Left.semantIdentifiers(report, less, lub, ids), e.Int)
	less(e.Right.semantIdentifiers(report, less, lub, ids), e.Int)
	return e.Boolean.Class
}

func (e *MultiplyExpr) semantTypes(lookup func(*Ident), c *Class) {
	lookup(e.Int)
	e.Left.semantTypes(lookup, c)
	e.Right.semantTypes(lookup, c)
}

func (e *MultiplyExpr) semantIdentifiers(report func(token.Pos, string), less func(*Class, *Ident), lub func(...*Class) *Class, ids semantIdentifiers) *Class {
	less(e.Left.semantIdentifiers(report, less, lub, ids), e.Int)
	less(e.Right.semantIdentifiers(report, less, lub, ids), e.Int)
	return e.Int.Class
}

func (e *DivideExpr) semantTypes(lookup func(*Ident), c *Class) {
	lookup(e.Int)
	e.Left.semantTypes(lookup, c)
	e.Right.semantTypes(lookup, c)
}

func (e *DivideExpr) semantIdentifiers(report func(token.Pos, string), less func(*Class, *Ident), lub func(...*Class) *Class, ids semantIdentifiers) *Class {
	less(e.Left.semantIdentifiers(report, less, lub, ids), e.Int)
	less(e.Right.semantIdentifiers(report, less, lub, ids), e.Int)
	return e.Int.Class
}

func (e *AddExpr) semantTypes(lookup func(*Ident), c *Class) {
	lookup(e.Int)
	e.Left.semantTypes(lookup, c)
	e.Right.semantTypes(lookup, c)
}

func (e *AddExpr) semantIdentifiers(report func(token.Pos, string), less func(*Class, *Ident), lub func(...*Class) *Class, ids semantIdentifiers) *Class {
	less(e.Left.semantIdentifiers(report, less, lub, ids), e.Int)
	less(e.Right.semantIdentifiers(report, less, lub, ids), e.Int)
	return e.Int.Class
}

func (e *SubtractExpr) semantTypes(lookup func(*Ident), c *Class) {
	lookup(e.Int)
	e.Left.semantTypes(lookup, c)
	e.Right.semantTypes(lookup, c)
}

func (e *SubtractExpr) semantIdentifiers(report func(token.Pos, string), less func(*Class, *Ident), lub func(...*Class) *Class, ids semantIdentifiers) *Class {
	less(e.Left.semantIdentifiers(report, less, lub, ids), e.Int)
	less(e.Right.semantIdentifiers(report, less, lub, ids), e.Int)
	return e.Int.Class
}

func (e *MatchExpr) semantTypes(lookup func(*Ident), c *Class) {
	e.Left.semantTypes(lookup, c)
	for _, a := range e.Cases {
		a.semantTypes(lookup, c)
	}
}

func (e *MatchExpr) semantIdentifiers(report func(token.Pos, string), less func(*Class, *Ident), lub func(...*Class) *Class, ids semantIdentifiers) *Class {
	left := e.Left.semantIdentifiers(report, less, lub, ids)

	possible := make(map[int]bool)
	if left == nothingClass {
		// no possible types
	} else if left == nullClass {
		// only Null is possible
		possible[0] = true
	} else {
		if lub(nullClass, left) == left {
			// left is a nullable type
			possible[0] = true
		}
		// left and all of left's children
		for i := left.Order; i <= left.MaxOrder; i++ {
			possible[i] = true
		}
		// and all of left's parents
		for p := left.Extends.Type.Class; p != nativeClass; p = p.Extends.Type.Class {
			possible[p.Order] = true
		}
	}

	var ts []*Class
	for _, c := range e.Cases {
		ts = append(ts, c.semantIdentifiers(report, less, lub, ids, e, possible))
	}

	return lub(ts...)
}

func (e *DynamicCallExpr) semantTypes(lookup func(*Ident), c *Class) {
	e.Recv.semantTypes(lookup, c)
	for _, a := range e.Args {
		a.semantTypes(lookup, c)
	}
}

func (e *DynamicCallExpr) semantIdentifiers(report func(token.Pos, string), less func(*Class, *Ident), lub func(...*Class) *Class, ids semantIdentifiers) *Class {
	left := e.Recv.semantIdentifiers(report, less, lub, ids)

	for _, m := range left.Methods {
		if m.Name.Name == e.Name.Name {
			e.Name.Method = m

			if len(m.Args) != len(e.Args) {
				report(e.Name.Pos, "wrong number of method arguments")
				report(m.Name.Pos, "(method is declared here)")
			} else {
				for i, a := range e.Args {
					less(a.semantIdentifiers(report, less, lub, ids), m.Args[i].Type)
				}
			}

			return m.Type.Class
		}
	}

	report(e.Name.Pos, "undeclared method "+left.Type.Name+"."+e.Name.Name)
	return nothingClass
}

func (e *SuperCallExpr) semantTypes(lookup func(*Ident), c *Class) {
	i := Ident{
		Pos:  e.Pos,
		Name: c.Extends.Type.Name,
	}
	lookup(&i)
	e.Class = i.Class
	for _, a := range e.Args {
		a.semantTypes(lookup, c)
	}
}

func (e *SuperCallExpr) semantIdentifiers(report func(token.Pos, string), less func(*Class, *Ident), lub func(...*Class) *Class, ids semantIdentifiers) *Class {
	for _, m := range e.Class.Methods {
		if m.Name.Name == e.Name.Name {
			e.Name.Method = m

			if len(m.Args) != len(e.Args) {
				report(e.Name.Pos, "wrong number of method arguments")
				report(m.Name.Pos, "(method is declared here)")
			} else {
				for i, a := range e.Args {
					less(a.semantIdentifiers(report, less, lub, ids), m.Args[i].Type)
				}
			}

			return m.Type.Class
		}
	}

	report(e.Name.Pos, "undeclared method "+e.Class.Type.Name+"."+e.Name.Name)
	return nothingClass
}

func (e *StaticCallExpr) semantTypes(lookup func(*Ident), c *Class) {
	e.Recv.semantTypes(lookup, c)
	for _, a := range e.Args {
		a.semantTypes(lookup, c)
	}
}

func (e *StaticCallExpr) semantIdentifiers(report func(token.Pos, string), less func(*Class, *Ident), lub func(...*Class) *Class, ids semantIdentifiers) *Class {
	left := e.Recv.semantIdentifiers(report, less, lub, ids)

	for _, f := range left.Features {
		if m, ok := f.(*Method); ok && m.Name.Name == e.Name.Name {
			e.Name.Method = m

			if len(m.Args) != len(e.Args) {
				report(e.Name.Pos, "wrong number of method arguments")
				report(m.Name.Pos, "(method is declared here)")
			} else {
				for i, a := range e.Args {
					less(a.semantIdentifiers(report, less, lub, ids), m.Args[i].Type)
				}
			}

			return m.Type.Class
		}
	}

	report(e.Name.Pos, "undeclared method "+left.Type.Name+"."+e.Name.Name)
	return nothingClass
}

func (e *AllocExpr) semantTypes(lookup func(*Ident), c *Class) {
	lookup(e.Type)
}

func (e *AllocExpr) semantIdentifiers(report func(token.Pos, string), less func(*Class, *Ident), lub func(...*Class) *Class, ids semantIdentifiers) *Class {
	return e.Type.Class
}

func (e *AssignExpr) semantTypes(lookup func(*Ident), c *Class) {
	lookup(e.Unit)
	e.Expr.semantTypes(lookup, c)
}

func (e *AssignExpr) semantIdentifiers(report func(token.Pos, string), less func(*Class, *Ident), lub func(...*Class) *Class, ids semantIdentifiers) *Class {
	if o := ids.Lookup(e.Name.Name); o == nil {
		report(e.Name.Pos, "undeclared identifier "+e.Name.Name)
	} else {
		e.Name.Object = o.Object
		less(e.Expr.semantIdentifiers(report, less, lub, ids), o.Type)
	}
	return e.Unit.Class
}

func (e *VarExpr) semantTypes(lookup func(*Ident), c *Class) {
	lookup(e.Type)
	e.Init.semantTypes(lookup, c)
	e.Body.semantTypes(lookup, c)
}

func (e *VarExpr) semantIdentifiers(report func(token.Pos, string), less func(*Class, *Ident), lub func(...*Class) *Class, ids semantIdentifiers) *Class {
	if o := ids.Lookup(e.Name.Name); o != nil {
		report(e.Name.Pos, "duplicate declaration of "+e.Name.Name)
		report(o.Name.Pos, "(previous declaration was here)")
	} else {
		e.Name.Object = e
		less(e.Init.semantIdentifiers(report, less, lub, ids), e.Type)
		ids = append(ids, &semantIdentifier{
			Name:   e.Name,
			Type:   e.Type,
			Object: e,
		})
	}
	return e.Body.semantIdentifiers(report, less, lub, ids)
}

func (e *ChainExpr) semantTypes(lookup func(*Ident), c *Class) {
	e.Pre.semantTypes(lookup, c)
	e.Expr.semantTypes(lookup, c)
}

func (e *ChainExpr) semantIdentifiers(report func(token.Pos, string), less func(*Class, *Ident), lub func(...*Class) *Class, ids semantIdentifiers) *Class {
	e.Pre.semantIdentifiers(report, less, lub, ids)
	return e.Expr.semantIdentifiers(report, less, lub, ids)
}

func (e *ThisExpr) semantTypes(lookup func(*Ident), c *Class) {
	e.Class = c
}

func (e *ThisExpr) semantIdentifiers(report func(token.Pos, string), less func(*Class, *Ident), lub func(...*Class) *Class, ids semantIdentifiers) *Class {
	return e.Class
}

func (e *NullExpr) semantTypes(lookup func(*Ident), c *Class) {
}

func (e *NullExpr) semantIdentifiers(report func(token.Pos, string), less func(*Class, *Ident), lub func(...*Class) *Class, ids semantIdentifiers) *Class {
	return nullClass
}

func (e *UnitExpr) semantTypes(lookup func(*Ident), c *Class) {
	i := Ident{
		Pos:  e.Pos,
		Name: "Unit",
	}
	lookup(&i)
	e.Class = i.Class
}

func (e *UnitExpr) semantIdentifiers(report func(token.Pos, string), less func(*Class, *Ident), lub func(...*Class) *Class, ids semantIdentifiers) *Class {
	return e.Class
}

func (e *NameExpr) semantTypes(lookup func(*Ident), c *Class) {
}

func (e *NameExpr) semantIdentifiers(report func(token.Pos, string), less func(*Class, *Ident), lub func(...*Class) *Class, ids semantIdentifiers) *Class {
	if o := ids.Lookup(e.Name.Name); o != nil {
		e.Name.Object = o.Object
		return o.Type.Class
	}
	report(e.Name.Pos, "undeclared identifier "+e.Name.Name)
	return nothingClass
}

func (e *StringExpr) semantTypes(lookup func(*Ident), c *Class) {
	e.Lit.semantTypes(lookup, c)
}

func (e *StringExpr) semantIdentifiers(report func(token.Pos, string), less func(*Class, *Ident), lub func(...*Class) *Class, ids semantIdentifiers) *Class {
	return e.Lit.Class
}

func (e *BoolExpr) semantTypes(lookup func(*Ident), c *Class) {
	e.Lit.semantTypes(lookup, c)
}

func (e *BoolExpr) semantIdentifiers(report func(token.Pos, string), less func(*Class, *Ident), lub func(...*Class) *Class, ids semantIdentifiers) *Class {
	return e.Lit.Class
}

func (e *IntExpr) semantTypes(lookup func(*Ident), c *Class) {
	e.Lit.semantTypes(lookup, c)
}

func (e *IntExpr) semantIdentifiers(report func(token.Pos, string), less func(*Class, *Ident), lub func(...*Class) *Class, ids semantIdentifiers) *Class {
	return e.Lit.Class
}

func (e *NativeExpr) semantTypes(lookup func(*Ident), c *Class) {
	lookup(&Ident{
		Pos:  e.Pos,
		Name: "native",
	})
}

func (e *NativeExpr) semantIdentifiers(report func(token.Pos, string), less func(*Class, *Ident), lub func(...*Class) *Class, ids semantIdentifiers) *Class {
	return nothingClass
}

func (l *IntLit) semantTypes(lookup func(*Ident), c *Class) {
	i := Ident{
		Pos:  l.Pos,
		Name: "Int",
	}
	lookup(&i)
	l.Class = i.Class
}

func (l *StringLit) semantTypes(lookup func(*Ident), c *Class) {
	i := Ident{
		Pos:  l.Pos,
		Name: "String",
	}
	lookup(&i)
	l.Class = i.Class
}

func (l *BoolLit) semantTypes(lookup func(*Ident), c *Class) {
	i := Ident{
		Pos:  l.Pos,
		Name: "Boolean",
	}
	lookup(&i)
	l.Class = i.Class
}

func (a *Case) semantTypes(lookup func(*Ident), c *Class) {
	lookup(a.Type)
	a.Body.semantTypes(lookup, c)
}

func (a *Case) semantIdentifiers(report func(token.Pos, string), less func(*Class, *Ident), lub func(...*Class) *Class, ids semantIdentifiers, m *MatchExpr, possible map[int]bool) *Class {
	left := a.Type.Class
	any := false

	check := func(min, max int) {
		for i := min; i <= max; i++ {
			if possible[i] {
				possible[i] = false
				any = true
			}
		}
	}

	if left == nothingClass {
		// there are no values of type Nothing
	} else if left == nullClass {
		// check if Null is possible
		check(0, 0)
	} else {
		// left and all of left's children
		check(left.Order, left.MaxOrder)
	}

	if !any {
		report(a.Type.Pos, "unreachable case for type "+a.Type.Name)
	}

	ids = append(ids, &semantIdentifier{
		Name:   a.Name,
		Type:   a.Type,
		Object: m,
	})

	return a.Body.semantIdentifiers(report, less, lub, ids)
}
