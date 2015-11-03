package ast

import (
	"fmt"
	"go/token"
	"os"
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

type semCtx struct {
	program    *Program
	fset       *token.FileSet
	haveErrors bool

	anyClass     *Class
	unitClass    *Class
	mainClass    *Class
	intClass     *Class
	booleanClass *Class

	opt Options
}

func (ctx *semCtx) Report(pos token.Pos, message string) {
	fmt.Fprintf(os.Stderr, "%v: %s\n", ctx.fset.Position(pos), message)
	ctx.haveErrors = true
}

func (ctx *semCtx) LookupClass(id *Ident) {
	if c, ok := ctx.program.classMap[id.Name]; ok {
		id.Class = c
		return
	}

	ctx.Report(id.Pos, "use of undeclared class "+id.Name)
	id.Class = errorClass
}

func (ctx *semCtx) FindRequiredClasses() {
	ctx.anyClass = ctx.FindRequiredClass("Any")
	ctx.unitClass = ctx.FindRequiredClass("Unit")
	ctx.mainClass = ctx.FindRequiredClass("Main")
	ctx.intClass = ctx.FindRequiredClass("Int")
	ctx.booleanClass = ctx.FindRequiredClass("Boolean")
}

func (ctx *semCtx) FindRequiredClass(name string) *Class {
	if c, ok := ctx.program.classMap[name]; ok {
		return c
	}

	ctx.Report(token.NoPos, "missing required class: "+name)
	return nil
}

func (ctx *semCtx) Less(t1, t2 *Class) bool {
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
		return t2 != nothingClass && t2 != ctx.booleanClass && t2 != ctx.intClass && t2 != ctx.unitClass
	}
	// S-Extends
	for t := t1; t != nativeClass; t = t.Extends.Type.Class {
		if t == t2 {
			return true
		}
	}
	return false
}

func (ctx *semCtx) Lub(ts ...*Class) *Class {
	t1 := nothingClass

	for _, t2 := range ts {
		// G-Compare
		if ctx.Less(t1, t2) {
			t1 = t2
			continue
		}
		// G-Commute
		if ctx.Less(t2, t1) {
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
				t1 = ctx.anyClass
				break
			}
			t1 = t1.Extends.Type.Class
			t2 = t2.Extends.Type.Class
		}
	}

	return t1
}

func (ctx *semCtx) AssertLess(t1 *Class, id *Ident) {
	t2 := id.Class

	if !ctx.Less(t1, t2) {
		ctx.Report(id.Pos, "type "+t1.Type.Name+" does not conform to type "+t2.Type.Name)
	}
}

func (p *Program) Semant(opt Options, fset *token.FileSet) bool {
	ctx := &semCtx{
		program: p,
		fset:    fset,
		opt:     opt,
	}
	p.classMap = map[string]*Class{
		"Nothing": nothingClass,
		"Null":    nullClass,
	}

	for _, c := range p.Classes {
		if o, ok := p.classMap[c.Type.Name]; ok {
			ctx.Report(c.Type.Pos, "duplicate declaration of class "+c.Type.Name)
			ctx.Report(o.Type.Pos, "(previous declaration was here)")
		} else {
			p.classMap[c.Type.Name] = c
		}
	}

	// If we continue with duplicate class names, we'll probably give a
	// bunch of useless errors.
	if ctx.haveErrors {
		return true
	}

	for _, c := range p.Classes {
		c.semantTypes(ctx)
	}

	if ctx.haveErrors {
		return true
	}

	ctx.FindRequiredClasses()

	children := make(map[*Class][]*Class)
	var free []*Class

	for _, c := range p.Classes {
		if c == ctx.anyClass {
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
			ctx.Report(c.Type.Pos, "class heirarchy loop: "+c.Type.Name)
		}
	}

	if ctx.haveErrors {
		return true
	}

	for _, c := range p.Classes {
		c.semantMakeConstructor(ctx)
	}

	for _, c := range p.Ordered {
		c.semantMethods(ctx)
	}

	for _, c := range p.Classes {
		c.semantIdentifiers(ctx)
	}

	p.Main = &StaticCallExpr{
		Recv: &AllocExpr{
			Type: &Ident{
				Pos:   token.NoPos,
				Name:  "Main",
				Class: ctx.mainClass,
			},
		},
		Name: &Ident{
			Pos:  token.NoPos,
			Name: "Main",
		},
	}
	if opt.Benchmark != 1 {
		p.Main = &VarExpr{
			Name: &Ident{
				Pos:  token.NoPos,
				Name: "benchmark",
			},
			Type: &Ident{
				Pos:   token.NoPos,
				Name:  "Int",
				Class: ctx.intClass,
			},
			Init: &IntExpr{
				Lit: &IntLit{
					Pos:   token.NoPos,
					Int:   0,
					Class: ctx.intClass,
				},
			},
			Body: &WhileExpr{
				Cond: &LessThanExpr{
					Pos: token.NoPos,
					Left: &NameExpr{
						Name: &Ident{
							Pos:  token.NoPos,
							Name: "benchmark",
						},
					},
					Right: &IntExpr{
						Lit: &IntLit{
							Pos:   token.NoPos,
							Int:   int32(opt.Benchmark),
							Class: ctx.intClass,
						},
					},
					Int: &Ident{
						Pos:   token.NoPos,
						Name:  "Int",
						Class: ctx.intClass,
					},
					Boolean: &Ident{
						Pos:   token.NoPos,
						Name:  "Boolean",
						Class: ctx.booleanClass,
					},
				},
				Body: &ChainExpr{
					Pre: p.Main,
					Expr: &AssignExpr{
						Name: &Ident{
							Pos:  token.NoPos,
							Name: "benchmark",
						},
						Expr: &AddExpr{
							Pos: token.NoPos,
							Left: &NameExpr{
								Name: &Ident{
									Pos:  token.NoPos,
									Name: "benchmark",
								},
							},
							Right: &IntExpr{
								Lit: &IntLit{
									Pos:   token.NoPos,
									Int:   1,
									Class: ctx.intClass,
								},
							},
							Int: &Ident{
								Pos:   token.NoPos,
								Name:  "Int",
								Class: ctx.intClass,
							},
						},
						Unit: &Ident{
							Pos:   token.NoPos,
							Name:  "Unit",
							Class: ctx.unitClass,
						},
					},
				},
				Boolean: &Ident{
					Pos:   token.NoPos,
					Name:  "Boolean",
					Class: ctx.booleanClass,
				},
				Unit: &Ident{
					Pos:   token.NoPos,
					Name:  "Unit",
					Class: ctx.unitClass,
				},
			},
		}
	}
	p.Main.semantIdentifiers(ctx, nil)

	return ctx.haveErrors
}

func (c *Class) semantTypes(ctx *semCtx) {
	ctx.LookupClass(c.Type)
	for _, f := range c.Formals {
		f.semantTypes(ctx, c)
	}
	for _, f := range c.Features {
		f.semantTypes(ctx, c)
	}
	if c.Extends.Type.Name == "native" {
		switch c.Type.Name {
		case "Any", "Null", "Nothing":
			c.Extends.Type.Class = nativeClass
			return
		}
	}
	c.Extends.semantTypes(ctx, c)
}

func (c *Class) semantMakeConstructor(ctx *semCtx) {
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
						Class: ctx.unitClass,
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

func (c *Class) semantMethods(ctx *semCtx) {
	parent := c.Extends.Type.Class
	c.Methods = make([]*Method, len(parent.Methods))
	copy(c.Methods, parent.Methods)

	used := make(map[string]token.Pos)

	for _, f := range c.Features {
		if m, ok := f.(*Method); ok {
			if o, ok := used[m.Name.Name]; ok {
				ctx.Report(m.Name.Pos, "duplicate declaration of "+m.Name.Name)
				ctx.Report(o, "(previous declaration was here)")
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
					ctx.Report(m.Name.Pos, "missing 'override' on method "+c.Type.Name+"."+m.Name.Name)
					ctx.Report(override.Name.Pos, "(previous declaration was here)")
				}
			} else {
				if override == nil {
					ctx.Report(m.Name.Pos, "missing parent for 'override' method "+c.Type.Name+"."+m.Name.Name)

					// add the method anyway. we won't generate
					// code, so this isn't a problem.
					m.Order = len(c.Methods)
					c.Methods = append(c.Methods, m)
				} else {
					m.Order = override.Order
					c.Methods[override.Order] = m

					if len(m.Args) != len(override.Args) {
						ctx.Report(m.Name.Pos, "invalid override: method "+c.Type.Name+"."+m.Name.Name+" has the wrong number of arguments")
						ctx.Report(override.Name.Pos, "(parent declaration is here)")
					} else {
						for i, a := range m.Args {
							if t1, t2 := a.Type, override.Args[i].Type; t1.Class != t2.Class {
								ctx.Report(t1.Pos, "invalid override: method "+c.Type.Name+"."+m.Name.Name+" has an incorrect argument type")
								ctx.Report(t2.Pos, "(parent declaration is here)")
							}
						}
					}

					if !ctx.Less(m.Type.Class, override.Type.Class) {
						ctx.Report(m.Type.Pos, "invalid override: method "+c.Type.Name+"."+m.Name.Name+" has incompatible return type "+m.Type.Name)
						ctx.Report(override.Type.Pos, "(parent return type is "+override.Type.Name+")")
					}

					for p := c.Extends.Type.Class; p != nativeClass; p = p.Extends.Type.Class {
						if m.Order < len(p.HasOverride) {
							p.HasOverride[m.Order] = true
						} else {
							break
						}
					}
				}
			}
		}
	}

	c.HasOverride = make([]bool, len(c.Methods))
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

func (c *Class) semantIdentifiers(ctx *semCtx) {
	ids := c.Extends.Type.Class.semantInheritedIdentifiers(func(s string) {
		ctx.Report(c.Extends.Type.Pos, s)
	})
	used := make(map[string]token.Pos)
	for _, id := range ids {
		used[id.Name.Name] = id.Name.Pos
	}
	for _, f := range c.Features {
		if a, ok := f.(*Attribute); ok {
			a.Name.Object = a
			if a.Type.Class == nothingClass {
				ctx.Report(a.Type.Pos, "cannot declare attribute of type Nothing")
			}
			if pos, ok := used[a.Name.Name]; ok {
				ctx.Report(a.Name.Pos, "duplicate declaration of "+a.Name.Name)
				ctx.Report(pos, "(previous declaration was here)")
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
			m.semantIdentifiers(ctx, ids)
		}
	}
}

func (e *Extends) semantTypes(ctx *semCtx, c *Class) {
	ctx.LookupClass(e.Type)
	for _, a := range e.Args {
		a.semantTypes(ctx, c)
	}
}

func (f *Init) semantTypes(ctx *semCtx, c *Class) {
	f.Expr.semantTypes(ctx, c)
}

func (f *Attribute) semantTypes(ctx *semCtx, c *Class) {
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
	ctx.LookupClass(f.Type)
	f.Init.semantTypes(ctx, c)
}

func (f *Method) semantTypes(ctx *semCtx, c *Class) {
	f.Parent = c

	for _, a := range f.Args {
		a.semantTypes(ctx, c)
	}
	ctx.LookupClass(f.Type)
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
	f.Body.semantTypes(ctx, c)
}

func (f *Method) semantIdentifiers(ctx *semCtx, ids semantIdentifiers) {
	used := make(map[string]token.Pos)
	for _, a := range f.Args {
		if o, ok := used[a.Name.Name]; ok {
			ctx.Report(a.Name.Pos, "duplicate declaration of "+a.Name.Name)
			ctx.Report(o, "(previous declaration was here)")
		} else {
			ids = append(ids, &semantIdentifier{
				Name:   a.Name,
				Type:   a.Type,
				Object: a,
			})
			used[a.Name.Name] = a.Name.Pos
		}
	}

	ctx.AssertLess(f.Body.semantIdentifiers(ctx, ids), f.Type)
}

func (a *Formal) semantTypes(ctx *semCtx, c *Class) {
	ctx.LookupClass(a.Type)
}

func (e *NotExpr) semantTypes(ctx *semCtx, c *Class) {
	ctx.LookupClass(e.Boolean)
	e.Expr.semantTypes(ctx, c)
}

func (e *NotExpr) semantIdentifiers(ctx *semCtx, ids semantIdentifiers) *Class {
	ctx.AssertLess(e.Expr.semantIdentifiers(ctx, ids), e.Boolean)
	return e.Boolean.Class
}

func (e *NegativeExpr) semantTypes(ctx *semCtx, c *Class) {
	ctx.LookupClass(e.Int)
	e.Expr.semantTypes(ctx, c)
}

func (e *NegativeExpr) semantIdentifiers(ctx *semCtx, ids semantIdentifiers) *Class {
	ctx.AssertLess(e.Expr.semantIdentifiers(ctx, ids), e.Int)
	return e.Int.Class
}

func (e *IfExpr) semantTypes(ctx *semCtx, c *Class) {
	ctx.LookupClass(e.Boolean)
	e.Cond.semantTypes(ctx, c)
	e.Then.semantTypes(ctx, c)
	e.Else.semantTypes(ctx, c)
}

func (e *IfExpr) semantIdentifiers(ctx *semCtx, ids semantIdentifiers) *Class {
	ctx.AssertLess(e.Cond.semantIdentifiers(ctx, ids), e.Boolean)
	return ctx.Lub(e.Then.semantIdentifiers(ctx, ids), e.Else.semantIdentifiers(ctx, ids))
}

func (e *WhileExpr) semantTypes(ctx *semCtx, c *Class) {
	ctx.LookupClass(e.Boolean)
	ctx.LookupClass(e.Unit)
	e.Cond.semantTypes(ctx, c)
	e.Body.semantTypes(ctx, c)
}

func (e *WhileExpr) semantIdentifiers(ctx *semCtx, ids semantIdentifiers) *Class {
	ctx.AssertLess(e.Cond.semantIdentifiers(ctx, ids), e.Boolean)
	e.Body.semantIdentifiers(ctx, ids)
	return e.Unit.Class
}

func (e *LessOrEqualExpr) semantTypes(ctx *semCtx, c *Class) {
	ctx.LookupClass(e.Boolean)
	ctx.LookupClass(e.Int)
	e.Left.semantTypes(ctx, c)
	e.Right.semantTypes(ctx, c)
}

func (e *LessOrEqualExpr) semantIdentifiers(ctx *semCtx, ids semantIdentifiers) *Class {
	ctx.AssertLess(e.Left.semantIdentifiers(ctx, ids), e.Int)
	ctx.AssertLess(e.Right.semantIdentifiers(ctx, ids), e.Int)
	return e.Boolean.Class
}

func (e *LessThanExpr) semantTypes(ctx *semCtx, c *Class) {
	ctx.LookupClass(e.Boolean)
	ctx.LookupClass(e.Int)
	e.Left.semantTypes(ctx, c)
	e.Right.semantTypes(ctx, c)
}

func (e *LessThanExpr) semantIdentifiers(ctx *semCtx, ids semantIdentifiers) *Class {
	ctx.AssertLess(e.Left.semantIdentifiers(ctx, ids), e.Int)
	ctx.AssertLess(e.Right.semantIdentifiers(ctx, ids), e.Int)
	return e.Boolean.Class
}

func (e *MultiplyExpr) semantTypes(ctx *semCtx, c *Class) {
	ctx.LookupClass(e.Int)
	e.Left.semantTypes(ctx, c)
	e.Right.semantTypes(ctx, c)
}

func (e *MultiplyExpr) semantIdentifiers(ctx *semCtx, ids semantIdentifiers) *Class {
	ctx.AssertLess(e.Left.semantIdentifiers(ctx, ids), e.Int)
	ctx.AssertLess(e.Right.semantIdentifiers(ctx, ids), e.Int)
	return e.Int.Class
}

func (e *DivideExpr) semantTypes(ctx *semCtx, c *Class) {
	ctx.LookupClass(e.Int)
	e.Left.semantTypes(ctx, c)
	e.Right.semantTypes(ctx, c)
}

func (e *DivideExpr) semantIdentifiers(ctx *semCtx, ids semantIdentifiers) *Class {
	ctx.AssertLess(e.Left.semantIdentifiers(ctx, ids), e.Int)
	ctx.AssertLess(e.Right.semantIdentifiers(ctx, ids), e.Int)
	return e.Int.Class
}

func (e *AddExpr) semantTypes(ctx *semCtx, c *Class) {
	ctx.LookupClass(e.Int)
	e.Left.semantTypes(ctx, c)
	e.Right.semantTypes(ctx, c)
}

func (e *AddExpr) semantIdentifiers(ctx *semCtx, ids semantIdentifiers) *Class {
	ctx.AssertLess(e.Left.semantIdentifiers(ctx, ids), e.Int)
	ctx.AssertLess(e.Right.semantIdentifiers(ctx, ids), e.Int)
	return e.Int.Class
}

func (e *SubtractExpr) semantTypes(ctx *semCtx, c *Class) {
	ctx.LookupClass(e.Int)
	e.Left.semantTypes(ctx, c)
	e.Right.semantTypes(ctx, c)
}

func (e *SubtractExpr) semantIdentifiers(ctx *semCtx, ids semantIdentifiers) *Class {
	ctx.AssertLess(e.Left.semantIdentifiers(ctx, ids), e.Int)
	ctx.AssertLess(e.Right.semantIdentifiers(ctx, ids), e.Int)
	return e.Int.Class
}

func (e *MatchExpr) semantTypes(ctx *semCtx, c *Class) {
	e.Left.semantTypes(ctx, c)
	for _, a := range e.Cases {
		a.semantTypes(ctx, c)
	}
}

func (e *MatchExpr) semantIdentifiers(ctx *semCtx, ids semantIdentifiers) *Class {
	left := e.Left.semantIdentifiers(ctx, ids)

	possible := make(map[int]bool)
	if left == nothingClass {
		// no possible types
	} else if left == nullClass {
		// only Null is possible
		possible[0] = true
	} else {
		if ctx.Lub(nullClass, left) == left {
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
		ts = append(ts, c.semantIdentifiers(ctx, ids, e, possible))
	}

	return ctx.Lub(ts...)
}

func (e *DynamicCallExpr) semantTypes(ctx *semCtx, c *Class) {
	e.Recv.semantTypes(ctx, c)
	for _, a := range e.Args {
		a.semantTypes(ctx, c)
	}
}

func (e *DynamicCallExpr) semantIdentifiers(ctx *semCtx, ids semantIdentifiers) *Class {
	left := e.Recv.semantIdentifiers(ctx, ids)

	for i, m := range left.Methods {
		if m.Name.Name == e.Name.Name {
			e.Name.Method = m

			if len(m.Args) != len(e.Args) {
				ctx.Report(e.Name.Pos, "wrong number of method arguments")
				ctx.Report(m.Name.Pos, "(method is declared here)")
			} else {
				for i, a := range e.Args {
					ctx.AssertLess(a.semantIdentifiers(ctx, ids), m.Args[i].Type)
				}
			}

			e.HasOverride = left.HasOverride[i]

			return m.Type.Class
		}
	}

	ctx.Report(e.Name.Pos, "undeclared method "+left.Type.Name+"."+e.Name.Name)
	return nothingClass
}

func (e *SuperCallExpr) semantTypes(ctx *semCtx, c *Class) {
	i := Ident{
		Pos:  e.Pos,
		Name: c.Extends.Type.Name,
	}
	ctx.LookupClass(&i)
	e.Class = i.Class
	for _, a := range e.Args {
		a.semantTypes(ctx, c)
	}
}

func (e *SuperCallExpr) semantIdentifiers(ctx *semCtx, ids semantIdentifiers) *Class {
	for _, m := range e.Class.Methods {
		if m.Name.Name == e.Name.Name {
			e.Name.Method = m

			if len(m.Args) != len(e.Args) {
				ctx.Report(e.Name.Pos, "wrong number of method arguments")
				ctx.Report(m.Name.Pos, "(method is declared here)")
			} else {
				for i, a := range e.Args {
					ctx.AssertLess(a.semantIdentifiers(ctx, ids), m.Args[i].Type)
				}
			}

			return m.Type.Class
		}
	}

	ctx.Report(e.Name.Pos, "undeclared method "+e.Class.Type.Name+"."+e.Name.Name)
	return nothingClass
}

func (e *StaticCallExpr) semantTypes(ctx *semCtx, c *Class) {
	e.Recv.semantTypes(ctx, c)
	for _, a := range e.Args {
		a.semantTypes(ctx, c)
	}
}

func (e *StaticCallExpr) semantIdentifiers(ctx *semCtx, ids semantIdentifiers) *Class {
	left := e.Recv.semantIdentifiers(ctx, ids)

	for _, f := range left.Features {
		if m, ok := f.(*Method); ok && m.Name.Name == e.Name.Name {
			e.Name.Method = m

			if len(m.Args) != len(e.Args) {
				ctx.Report(e.Name.Pos, "wrong number of method arguments")
				ctx.Report(m.Name.Pos, "(method is declared here)")
			} else {
				for i, a := range e.Args {
					ctx.AssertLess(a.semantIdentifiers(ctx, ids), m.Args[i].Type)
				}
			}

			return m.Type.Class
		}
	}

	ctx.Report(e.Name.Pos, "undeclared method "+left.Type.Name+"."+e.Name.Name)
	return nothingClass
}

func (e *AllocExpr) semantTypes(ctx *semCtx, c *Class) {
	ctx.LookupClass(e.Type)
}

func (e *AllocExpr) semantIdentifiers(ctx *semCtx, ids semantIdentifiers) *Class {
	return e.Type.Class
}

func (e *AssignExpr) semantTypes(ctx *semCtx, c *Class) {
	ctx.LookupClass(e.Unit)
	e.Expr.semantTypes(ctx, c)
}

func (e *AssignExpr) semantIdentifiers(ctx *semCtx, ids semantIdentifiers) *Class {
	if o := ids.Lookup(e.Name.Name); o == nil {
		ctx.Report(e.Name.Pos, "undeclared identifier "+e.Name.Name)
	} else {
		e.Name.Object = o.Object
		ctx.AssertLess(e.Expr.semantIdentifiers(ctx, ids), o.Type)
	}
	return e.Unit.Class
}

func (e *VarExpr) semantTypes(ctx *semCtx, c *Class) {
	ctx.LookupClass(e.Type)
	e.Init.semantTypes(ctx, c)
	e.Body.semantTypes(ctx, c)
}

func (e *VarExpr) semantIdentifiers(ctx *semCtx, ids semantIdentifiers) *Class {
	if o := ids.Lookup(e.Name.Name); o != nil {
		ctx.Report(e.Name.Pos, "duplicate declaration of "+e.Name.Name)
		ctx.Report(o.Name.Pos, "(previous declaration was here)")
	} else {
		e.Name.Object = e
		ctx.AssertLess(e.Init.semantIdentifiers(ctx, ids), e.Type)
		ids = append(ids, &semantIdentifier{
			Name:   e.Name,
			Type:   e.Type,
			Object: e,
		})
	}
	return e.Body.semantIdentifiers(ctx, ids)
}

func (e *ChainExpr) semantTypes(ctx *semCtx, c *Class) {
	e.Pre.semantTypes(ctx, c)
	e.Expr.semantTypes(ctx, c)
}

func (e *ChainExpr) semantIdentifiers(ctx *semCtx, ids semantIdentifiers) *Class {
	e.Pre.semantIdentifiers(ctx, ids)
	return e.Expr.semantIdentifiers(ctx, ids)
}

func (e *ThisExpr) semantTypes(ctx *semCtx, c *Class) {
	e.Class = c
}

func (e *ThisExpr) semantIdentifiers(ctx *semCtx, ids semantIdentifiers) *Class {
	return e.Class
}

func (e *NullExpr) semantTypes(ctx *semCtx, c *Class) {
}

func (e *NullExpr) semantIdentifiers(ctx *semCtx, ids semantIdentifiers) *Class {
	return nullClass
}

func (e *UnitExpr) semantTypes(ctx *semCtx, c *Class) {
	i := Ident{
		Pos:  e.Pos,
		Name: "Unit",
	}
	ctx.LookupClass(&i)
	e.Class = i.Class
}

func (e *UnitExpr) semantIdentifiers(ctx *semCtx, ids semantIdentifiers) *Class {
	return e.Class
}

func (e *NameExpr) semantTypes(ctx *semCtx, c *Class) {
}

func (e *NameExpr) semantIdentifiers(ctx *semCtx, ids semantIdentifiers) *Class {
	if o := ids.Lookup(e.Name.Name); o != nil {
		e.Name.Object = o.Object
		return o.Type.Class
	}
	ctx.Report(e.Name.Pos, "undeclared identifier "+e.Name.Name)
	return nothingClass
}

func (e *StringExpr) semantTypes(ctx *semCtx, c *Class) {
	e.Lit.semantTypes(ctx, c)
}

func (e *StringExpr) semantIdentifiers(ctx *semCtx, ids semantIdentifiers) *Class {
	return e.Lit.Class
}

func (e *BoolExpr) semantTypes(ctx *semCtx, c *Class) {
	e.Lit.semantTypes(ctx, c)
}

func (e *BoolExpr) semantIdentifiers(ctx *semCtx, ids semantIdentifiers) *Class {
	return e.Lit.Class
}

func (e *IntExpr) semantTypes(ctx *semCtx, c *Class) {
	e.Lit.semantTypes(ctx, c)
}

func (e *IntExpr) semantIdentifiers(ctx *semCtx, ids semantIdentifiers) *Class {
	return e.Lit.Class
}

func (e *NativeExpr) semantTypes(ctx *semCtx, c *Class) {
	ctx.LookupClass(&Ident{
		Pos:  e.Pos,
		Name: "native",
	})
}

func (e *NativeExpr) semantIdentifiers(ctx *semCtx, ids semantIdentifiers) *Class {
	return nothingClass
}

func (l *IntLit) semantTypes(ctx *semCtx, c *Class) {
	i := Ident{
		Pos:  l.Pos,
		Name: "Int",
	}
	ctx.LookupClass(&i)
	l.Class = i.Class
}

func (l *StringLit) semantTypes(ctx *semCtx, c *Class) {
	i := Ident{
		Pos:  l.Pos,
		Name: "String",
	}
	ctx.LookupClass(&i)
	l.Class = i.Class
}

func (l *BoolLit) semantTypes(ctx *semCtx, c *Class) {
	i := Ident{
		Pos:  l.Pos,
		Name: "Boolean",
	}
	ctx.LookupClass(&i)
	l.Class = i.Class
}

func (a *Case) semantTypes(ctx *semCtx, c *Class) {
	ctx.LookupClass(a.Type)
	a.Body.semantTypes(ctx, c)
}

func (a *Case) semantIdentifiers(ctx *semCtx, ids semantIdentifiers, m *MatchExpr, possible map[int]bool) *Class {
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
		ctx.Report(a.Type.Pos, "unreachable case for type "+a.Type.Name)
	}

	ids = append(ids, &semantIdentifier{
		Name:   a.Name,
		Type:   a.Type,
		Object: m,
	})

	return a.Body.semantIdentifiers(ctx, ids)
}
