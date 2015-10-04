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
	p.classMap = map[string]*Class{
		"Nothing": nothingClass,
		"Null":    nullClass,
	}

	for _, c := range p.Classes {
		if o, ok := p.classMap[c.Type.Name]; ok {
			fmt.Printf("%v: duplicate declaration of class %s\n", fset.Position(c.Type.Pos), c.Type.Name)
			fmt.Printf("%v: (previous declaration was here)\n", fset.Position(o.Type.Pos))
			haveErrors = true
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
				id.Object = c
				return
			}

			fmt.Printf("%v: use of undeclared class %s\n", fset.Position(id.Pos), id.Name)
			haveErrors = true
			id.Object = errorClass
		})
	}

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
			c.Extends.Type.Object = nativeClass
			return
		}
	}
	c.Extends.semantTypes(lookup, c)
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
		case c.Type.Name == "Int" && f.Name.Name == "toString":
			return
		case c.Type.Name == "Int" && f.Name.Name == "equals":
			return
		case c.Type.Name == "Boolean" && f.Name.Name == "equals":
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
		}
	}
	f.Body.semantTypes(lookup, c)
}

func (a *Formal) semantTypes(lookup func(*Ident), c *Class) {
	lookup(a.Type)
}

func (e *NotExpr) semantTypes(lookup func(*Ident), c *Class) {
	e.Expr.semantTypes(lookup, c)
}

func (e *NegativeExpr) semantTypes(lookup func(*Ident), c *Class) {
	e.Expr.semantTypes(lookup, c)
}

func (e *IfExpr) semantTypes(lookup func(*Ident), c *Class) {
	e.Cond.semantTypes(lookup, c)
	e.Then.semantTypes(lookup, c)
	e.Else.semantTypes(lookup, c)
}

func (e *WhileExpr) semantTypes(lookup func(*Ident), c *Class) {
	e.Cond.semantTypes(lookup, c)
	e.Body.semantTypes(lookup, c)
}

func (e *LessOrEqualExpr) semantTypes(lookup func(*Ident), c *Class) {
	e.Left.semantTypes(lookup, c)
	e.Right.semantTypes(lookup, c)
}

func (e *LessThanExpr) semantTypes(lookup func(*Ident), c *Class) {
	e.Left.semantTypes(lookup, c)
	e.Right.semantTypes(lookup, c)
}

func (e *MultiplyExpr) semantTypes(lookup func(*Ident), c *Class) {
	e.Left.semantTypes(lookup, c)
	e.Right.semantTypes(lookup, c)
}

func (e *DivideExpr) semantTypes(lookup func(*Ident), c *Class) {
	e.Left.semantTypes(lookup, c)
	e.Right.semantTypes(lookup, c)
}

func (e *AddExpr) semantTypes(lookup func(*Ident), c *Class) {
	e.Left.semantTypes(lookup, c)
	e.Right.semantTypes(lookup, c)
}

func (e *SubtractExpr) semantTypes(lookup func(*Ident), c *Class) {
	e.Left.semantTypes(lookup, c)
	e.Right.semantTypes(lookup, c)
}

func (e *MatchExpr) semantTypes(lookup func(*Ident), c *Class) {
	e.Left.semantTypes(lookup, c)
	for _, a := range e.Cases {
		a.semantTypes(lookup, c)
	}
}

func (e *DynamicCallExpr) semantTypes(lookup func(*Ident), c *Class) {
	e.Recv.semantTypes(lookup, c)
	for _, a := range e.Args {
		a.semantTypes(lookup, c)
	}
}

func (e *SuperCallExpr) semantTypes(lookup func(*Ident), c *Class) {
	i := Ident{
		Pos:  e.Pos,
		Name: c.Extends.Type.Name,
	}
	lookup(&i)
	e.Class = i.Object.(*Class)
	for _, a := range e.Args {
		a.semantTypes(lookup, c)
	}
}

func (e *StaticCallExpr) semantTypes(lookup func(*Ident), c *Class) {
	e.Recv.semantTypes(lookup, c)
	for _, a := range e.Args {
		a.semantTypes(lookup, c)
	}
}

func (e *AllocExpr) semantTypes(lookup func(*Ident), c *Class) {
	lookup(e.Type)
}

func (e *AssignExpr) semantTypes(lookup func(*Ident), c *Class) {
	i := Ident{
		Pos:  e.Name.Pos,
		Name: "Unit",
	}
	lookup(&i)
	e.Class = i.Object.(*Class)
	e.Expr.semantTypes(lookup, c)
}

func (e *VarExpr) semantTypes(lookup func(*Ident), c *Class) {
	lookup(e.Type)
	e.Init.semantTypes(lookup, c)
	e.Body.semantTypes(lookup, c)
}

func (e *ChainExpr) semantTypes(lookup func(*Ident), c *Class) {
	e.Pre.semantTypes(lookup, c)
	e.Expr.semantTypes(lookup, c)
}

func (e *ThisExpr) semantTypes(lookup func(*Ident), c *Class) {
	e.Class = c
}

func (e *NullExpr) semantTypes(lookup func(*Ident), c *Class) {
}

func (e *UnitExpr) semantTypes(lookup func(*Ident), c *Class) {
	i := Ident{
		Pos:  e.Pos,
		Name: "Unit",
	}
	lookup(&i)
	e.Class = i.Object.(*Class)
}

func (e *NameExpr) semantTypes(lookup func(*Ident), c *Class) {
}

func (e *StringExpr) semantTypes(lookup func(*Ident), c *Class) {
	e.Lit.semantTypes(lookup, c)
}

func (e *BoolExpr) semantTypes(lookup func(*Ident), c *Class) {
	e.Lit.semantTypes(lookup, c)
}

func (e *IntExpr) semantTypes(lookup func(*Ident), c *Class) {
	e.Lit.semantTypes(lookup, c)
}

func (e *NativeExpr) semantTypes(lookup func(*Ident), c *Class) {
	lookup(&Ident{
		Pos:  e.Pos,
		Name: "native",
	})
}

func (l *IntLit) semantTypes(lookup func(*Ident), c *Class) {
	i := Ident{
		Pos:  l.Pos,
		Name: "Int",
	}
	lookup(&i)
	l.Class = i.Object.(*Class)
}

func (l *StringLit) semantTypes(lookup func(*Ident), c *Class) {
	i := Ident{
		Pos:  l.Pos,
		Name: "String",
	}
	lookup(&i)
	l.Class = i.Object.(*Class)
}

func (l *BoolLit) semantTypes(lookup func(*Ident), c *Class) {
	i := Ident{
		Pos:  l.Pos,
		Name: "Boolean",
	}
	lookup(&i)
	l.Class = i.Object.(*Class)
}

func (a *Case) semantTypes(lookup func(*Ident), c *Class) {
	lookup(a.Type)
	a.Body.semantTypes(lookup, c)
}
