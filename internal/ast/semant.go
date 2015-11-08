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

func semantInline(ctx *semCtx, recv Expr, name *Ident, args []Expr) (Expr, bool) {
	if ctx.opt.OptFold && name.Method.Name.Name == "length" && len(args) == 0 {
		if str, ok := recv.(*StringExpr); ok {
			return &IntExpr{
				Lit: &IntLit{
					Pos:   str.Lit.Pos,
					Int:   int32(len(str.Lit.Str)),
					Class: ctx.intClass,
				},
			}, true
		}
	}

	return nil, false
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

	var mainRun *Method
	if opt.Coroutine {
		mainRun = &Method{
			Override: true,
			Name: &Ident{
				Pos:  token.NoPos,
				Name: "run",
			},
			Args: nil,
			Type: &Ident{
				Pos:  token.NoPos,
				Name: "Unit",
			},
			Body: &ChainExpr{
				Pre: &StaticCallExpr{
					Recv: &AllocExpr{
						Type: &Ident{
							Pos:  token.NoPos,
							Name: "Main",
						},
					},
					Name: &Ident{
						Pos:  token.NoPos,
						Name: "Main",
					},
					Args: nil,
				},
				Expr: &UnitExpr{
					Pos: token.NoPos,
				},
			},
		}
		p.Classes = append(p.Classes, &Class{
			Type: &Ident{
				Pos:  token.NoPos,
				Name: "runtimeMain",
			},
			Formals: nil,
			Extends: &Extends{
				Type: &Ident{
					Pos:  token.NoPos,
					Name: "Runnable",
				},
			},
			Features: []Feature{
				mainRun,
			},
		})
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
	if opt.Coroutine {
		p.Main = &StaticCallExpr{
			Recv: &AllocExpr{
				Type: &Ident{
					Pos:  token.NoPos,
					Name: "Coroutine",
				},
			},
			Name: &Ident{
				Pos:  token.NoPos,
				Name: "Coroutine",
			},
			Args: []Expr{
				&StaticCallExpr{
					Recv: &AllocExpr{
						Type: &Ident{
							Pos:  token.NoPos,
							Name: "runtimeMain",
						},
					},
					Name: &Ident{
						Pos:  token.NoPos,
						Name: "runtimeMain",
					},
				},
			},
		}
	}

	p.Main.semantTypes(ctx, nil)
	p.Main.semantIdentifiers(ctx, nil)

	pmain := &p.Main
	if opt.Coroutine {
		pmain = &mainRun.Body
	}

	if opt.Benchmark != 1 {
		var benchmark VarExpr
		benchmark = VarExpr{
			Name: &Ident{
				Pos:    token.NoPos,
				Name:   "benchmark",
				Object: &benchmark,
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
							Pos:    token.NoPos,
							Name:   "benchmark",
							Object: &benchmark,
						},
					},
					Right: &IntExpr{
						Lit: &IntLit{
							Pos:   token.NoPos,
							Int:   int32(opt.Benchmark),
							Class: ctx.intClass,
						},
					},

					Boolean: &Ident{
						Pos:   token.NoPos,
						Name:  "Boolean",
						Class: ctx.booleanClass,
					},
					Int: &Ident{
						Pos:   token.NoPos,
						Name:  "Int",
						Class: ctx.intClass,
					},
				},
				Body: &ChainExpr{
					Pre: *pmain,
					Expr: &AssignExpr{
						Name: &Ident{
							Pos:    token.NoPos,
							Name:   "benchmark",
							Object: &benchmark,
						},
						Expr: &AddExpr{
							Pos: token.NoPos,
							Left: &NameExpr{
								Name: &Ident{
									Pos:    token.NoPos,
									Name:   "benchmark",
									Object: &benchmark,
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
		*pmain = &benchmark
	}

	if ctx.haveErrors {
		return ctx.haveErrors
	}

	for _, c := range p.Classes {
		for _, f := range c.Features {
			if m, ok := f.(*Method); ok {
				m.Body = m.Body.semantOpt(ctx)
			}
		}
	}
	p.Main = p.Main.semantOpt(ctx)

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
			if c.Type.Name == "String" && f.Name.Name == "str_field" {
				continue
			}

			constructor = &ChainExpr{
				Pre: &AssignExpr{
					Name: f.Name,
					Expr: f.Init,

					Unit: &Ident{
						Pos:  token.NoPos,
						Name: "Unit",
					},
				},
				Expr: constructor,
			}
		}
	}

	var useNative = false

	switch c.Type.Name {
	case "ArrayAny":
		useNative = true

	case "Coroutine", "Channel":
		useNative = ctx.opt.Coroutine
	}

	switch c.Type.Name {
	case "Int", "Boolean", "Unit", "Symbol":
		return

	case "Any":
		// don't call super-constructor because there isn't one.

	default:
		if useNative {
			constructor = &NativeExpr{
				Pos: c.Type.Pos,
			}
		} else {
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
		case ctx.opt.Coroutine && c.Type.Name == "Coroutine" && f.Name.Name == "coroutine_field":
			return
		case ctx.opt.Coroutine && c.Type.Name == "Channel" && f.Name.Name == "channel_field":
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
		case ctx.opt.Coroutine && c.Type.Name == "Coroutine" && f.Name.Name == "Coroutine":
			return
		case ctx.opt.Coroutine && c.Type.Name == "Channel" && f.Name.Name == "recv":
			return
		case ctx.opt.Coroutine && c.Type.Name == "Channel" && f.Name.Name == "send":
			return
		case ctx.opt.Coroutine && c.Type.Name == "Channel" && f.Name.Name == "Channel":
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

func (e *NotExpr) semantOpt(ctx *semCtx) Expr {
	expr := e.Expr.semantOpt(ctx)
	if expr != e.Expr {
		return &NotExpr{
			Expr:    expr,
			Boolean: e.Boolean,
		}
	}
	return e
}

func (e *NotExpr) semantReplaceObject(ctx *semCtx, from, to Object) Expr {
	expr := e.Expr.semantReplaceObject(ctx, from, to)
	if expr != e.Expr {
		return &NotExpr{
			Expr:    expr,
			Boolean: e.Boolean,
		}
	}
	return e
}

func (e *NegativeExpr) semantTypes(ctx *semCtx, c *Class) {
	ctx.LookupClass(e.Int)
	e.Expr.semantTypes(ctx, c)
}

func (e *NegativeExpr) semantIdentifiers(ctx *semCtx, ids semantIdentifiers) *Class {
	ctx.AssertLess(e.Expr.semantIdentifiers(ctx, ids), e.Int)
	return e.Int.Class
}

func (e *NegativeExpr) semantOpt(ctx *semCtx) Expr {
	expr := e.Expr.semantOpt(ctx)
	if i, ok := expr.(*IntExpr); ok && ctx.opt.OptFold {
		return &IntExpr{
			Lit: &IntLit{
				Pos:   i.Lit.Pos,
				Int:   -i.Lit.Int,
				Class: i.Lit.Class,
			},
		}
	}
	if expr != e.Expr {
		return &NegativeExpr{
			Expr: expr,
			Int:  e.Int,
		}
	}
	return e
}

func (e *NegativeExpr) semantReplaceObject(ctx *semCtx, from, to Object) Expr {
	expr := e.Expr.semantReplaceObject(ctx, from, to)
	if expr != e.Expr {
		return &NegativeExpr{
			Expr: expr,
			Int:  e.Int,
		}
	}
	return e
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

func (e *IfExpr) semantOpt(ctx *semCtx) Expr {
	cond := e.Cond.semantOpt(ctx)
	then := e.Then.semantOpt(ctx)
	els := e.Else.semantOpt(ctx)
	if cond != e.Cond || then != e.Then || els != e.Else {
		return &IfExpr{
			Cond:    cond,
			Then:    then,
			Else:    els,
			Boolean: e.Boolean,
		}
	}
	return e
}

func (e *IfExpr) semantReplaceObject(ctx *semCtx, from, to Object) Expr {
	cond := e.Cond.semantReplaceObject(ctx, from, to)
	then := e.Then.semantReplaceObject(ctx, from, to)
	els := e.Else.semantReplaceObject(ctx, from, to)
	if cond != e.Cond || then != e.Then || els != e.Else {
		return &IfExpr{
			Cond:    cond,
			Then:    then,
			Else:    els,
			Boolean: e.Boolean,
		}
	}
	return e
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

func (e *WhileExpr) semantOpt(ctx *semCtx) Expr {
	cond := e.Cond.semantOpt(ctx)
	body := e.Body.semantOpt(ctx)
	if cond != e.Cond || body != e.Body {
		return &WhileExpr{
			Cond:    cond,
			Body:    body,
			Boolean: e.Boolean,
			Unit:    e.Unit,
		}
	}
	return e
}

func (e *WhileExpr) semantReplaceObject(ctx *semCtx, from, to Object) Expr {
	cond := e.Cond.semantReplaceObject(ctx, from, to)
	body := e.Body.semantReplaceObject(ctx, from, to)
	if cond != e.Cond || body != e.Body {
		return &WhileExpr{
			Cond:    cond,
			Body:    body,
			Boolean: e.Boolean,
			Unit:    e.Unit,
		}
	}
	return e
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

func (e *LessOrEqualExpr) semantOpt(ctx *semCtx) Expr {
	left := e.Left.semantOpt(ctx)
	right := e.Right.semantOpt(ctx)
	if left != e.Left || right != e.Right {
		return &LessOrEqualExpr{
			Left:    left,
			Right:   right,
			Pos:     e.Pos,
			Boolean: e.Boolean,
			Int:     e.Int,
		}
	}
	return e
}

func (e *LessOrEqualExpr) semantReplaceObject(ctx *semCtx, from, to Object) Expr {
	left := e.Left.semantReplaceObject(ctx, from, to)
	right := e.Right.semantReplaceObject(ctx, from, to)
	if left != e.Left || right != e.Right {
		return &LessOrEqualExpr{
			Left:    left,
			Right:   right,
			Pos:     e.Pos,
			Boolean: e.Boolean,
			Int:     e.Int,
		}
	}
	return e
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

func (e *LessThanExpr) semantOpt(ctx *semCtx) Expr {
	left := e.Left.semantOpt(ctx)
	right := e.Right.semantOpt(ctx)
	if left != e.Left || right != e.Right {
		return &LessThanExpr{
			Left:    left,
			Right:   right,
			Pos:     e.Pos,
			Boolean: e.Boolean,
			Int:     e.Int,
		}
	}
	return e
}

func (e *LessThanExpr) semantReplaceObject(ctx *semCtx, from, to Object) Expr {
	left := e.Left.semantReplaceObject(ctx, from, to)
	right := e.Right.semantReplaceObject(ctx, from, to)
	if left != e.Left || right != e.Right {
		return &LessThanExpr{
			Left:    left,
			Right:   right,
			Pos:     e.Pos,
			Boolean: e.Boolean,
			Int:     e.Int,
		}
	}
	return e
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

func (e *MultiplyExpr) semantOpt(ctx *semCtx) Expr {
	left := e.Left.semantOpt(ctx)
	right := e.Right.semantOpt(ctx)
	if li, ok := left.(*IntExpr); ok && ctx.opt.OptFold {
		if ri, ok := right.(*IntExpr); ok {
			return &IntExpr{
				Lit: &IntLit{
					Pos:   e.Pos,
					Int:   li.Lit.Int * ri.Lit.Int,
					Class: li.Lit.Class,
				},
			}
		}
	}
	if left != e.Left || right != e.Right {
		return &MultiplyExpr{
			Left:  left,
			Right: right,
			Pos:   e.Pos,
			Int:   e.Int,
		}
	}
	return e
}

func (e *MultiplyExpr) semantReplaceObject(ctx *semCtx, from, to Object) Expr {
	left := e.Left.semantReplaceObject(ctx, from, to)
	right := e.Right.semantReplaceObject(ctx, from, to)
	if left != e.Left || right != e.Right {
		return &MultiplyExpr{
			Left:  left,
			Right: right,
			Pos:   e.Pos,
			Int:   e.Int,
		}
	}
	return e
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

func (e *DivideExpr) semantOpt(ctx *semCtx) Expr {
	left := e.Left.semantOpt(ctx)
	right := e.Right.semantOpt(ctx)
	if li, ok := left.(*IntExpr); ok && ctx.opt.OptFold {
		if ri, ok := right.(*IntExpr); ok && ri.Lit.Int != 0 {
			return &IntExpr{
				Lit: &IntLit{
					Pos:   e.Pos,
					Int:   li.Lit.Int / ri.Lit.Int,
					Class: li.Lit.Class,
				},
			}
		}
	}
	if left != e.Left || right != e.Right {
		return &DivideExpr{
			Left:  left,
			Right: right,
			Pos:   e.Pos,
			Int:   e.Int,
		}
	}
	return e
}

func (e *DivideExpr) semantReplaceObject(ctx *semCtx, from, to Object) Expr {
	left := e.Left.semantReplaceObject(ctx, from, to)
	right := e.Right.semantReplaceObject(ctx, from, to)
	if left != e.Left || right != e.Right {
		return &DivideExpr{
			Left:  left,
			Right: right,
			Pos:   e.Pos,
			Int:   e.Int,
		}
	}
	return e
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

func (e *AddExpr) semantOpt(ctx *semCtx) Expr {
	left := e.Left.semantOpt(ctx)
	right := e.Right.semantOpt(ctx)
	if li, ok := left.(*IntExpr); ok && ctx.opt.OptFold {
		if ri, ok := right.(*IntExpr); ok {
			return &IntExpr{
				Lit: &IntLit{
					Pos:   e.Pos,
					Int:   li.Lit.Int + ri.Lit.Int,
					Class: li.Lit.Class,
				},
			}
		}
	}
	if left != e.Left || right != e.Right {
		return &AddExpr{
			Left:  left,
			Right: right,
			Pos:   e.Pos,
			Int:   e.Int,
		}
	}
	return e
}

func (e *AddExpr) semantReplaceObject(ctx *semCtx, from, to Object) Expr {
	left := e.Left.semantReplaceObject(ctx, from, to)
	right := e.Right.semantReplaceObject(ctx, from, to)
	if left != e.Left || right != e.Right {
		return &AddExpr{
			Left:  left,
			Right: right,
			Pos:   e.Pos,
			Int:   e.Int,
		}
	}
	return e
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

func (e *SubtractExpr) semantOpt(ctx *semCtx) Expr {
	left := e.Left.semantOpt(ctx)
	right := e.Right.semantOpt(ctx)
	if li, ok := left.(*IntExpr); ok && ctx.opt.OptFold {
		if ri, ok := right.(*IntExpr); ok {
			return &IntExpr{
				Lit: &IntLit{
					Pos:   e.Pos,
					Int:   li.Lit.Int - ri.Lit.Int,
					Class: li.Lit.Class,
				},
			}
		}
	}
	if left != e.Left || right != e.Right {
		return &SubtractExpr{
			Left:  left,
			Right: right,
			Pos:   e.Pos,
			Int:   e.Int,
		}
	}
	return e
}

func (e *SubtractExpr) semantReplaceObject(ctx *semCtx, from, to Object) Expr {
	left := e.Left.semantReplaceObject(ctx, from, to)
	right := e.Right.semantReplaceObject(ctx, from, to)
	if left != e.Left || right != e.Right {
		return &SubtractExpr{
			Left:  left,
			Right: right,
			Pos:   e.Pos,
			Int:   e.Int,
		}
	}
	return e
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

func (e *MatchExpr) semantOpt(ctx *semCtx) Expr {
	left := e.Left.semantOpt(ctx)
	cases := make([]*Case, len(e.Cases))
	anyCase := false
	for i, c := range e.Cases {
		body := c.Body.semantOpt(ctx)
		if body != c.Body {
			cases[i] = &Case{
				Name: c.Name,
				Type: c.Type,
				Body: body,
			}
			anyCase = true
		} else {
			cases[i] = c
		}
	}
	if left != e.Left || anyCase {
		var m MatchExpr
		for i, c := range cases {
			name := c.Name.semantReplaceObject(ctx, e, &m)
			cases[i] = &Case{
				Name: name,
				Type: c.Type,
				Body: c.Body.semantReplaceObject(ctx, e, &m),
			}
		}
		m = MatchExpr{
			Left:  left,
			Pos:   e.Pos,
			Cases: cases,
		}
		return &m
	}
	return e
}

func (e *MatchExpr) semantReplaceObject(ctx *semCtx, from, to Object) Expr {
	left := e.Left.semantReplaceObject(ctx, from, to)
	cases := make([]*Case, len(e.Cases))
	anyCase := false
	for i, c := range e.Cases {
		name := c.Name.semantReplaceObject(ctx, from, to)
		body := c.Body.semantReplaceObject(ctx, from, to)
		if name != c.Name || body != c.Body {
			cases[i] = &Case{
				Name: name,
				Type: c.Type,
				Body: body,
			}
			anyCase = true
		} else {
			cases[i] = c
		}
	}
	if left != e.Left || anyCase {
		var m MatchExpr
		for i, c := range cases {
			cases[i] = &Case{
				Name: c.Name.semantReplaceObject(ctx, e, &m),
				Type: c.Type,
				Body: c.Body.semantReplaceObject(ctx, e, &m),
			}
		}
		m = MatchExpr{
			Left:  left,
			Pos:   e.Pos,
			Cases: cases,
		}
		return &m
	}
	return e
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

func (e *DynamicCallExpr) semantOpt(ctx *semCtx) Expr {
	recv := e.Recv.semantOpt(ctx)
	args := make([]Expr, len(e.Args))
	anyArg := false
	for i, a := range e.Args {
		args[i] = a.semantOpt(ctx)
		if args[i] != a {
			anyArg = true
		}
	}
	if recv != e.Recv || anyArg {
		e = &DynamicCallExpr{
			Recv:        recv,
			Name:        e.Name,
			Args:        args,
			HasOverride: e.HasOverride,
		}
	}
	if !e.HasOverride {
		if inl, ok := semantInline(ctx, e.Recv, e.Name, e.Args); ok {
			return inl
		}
	}
	return e
}

func (e *DynamicCallExpr) semantReplaceObject(ctx *semCtx, from, to Object) Expr {
	recv := e.Recv.semantReplaceObject(ctx, from, to)
	args := make([]Expr, len(e.Args))
	anyArg := false
	for i, a := range e.Args {
		args[i] = a.semantReplaceObject(ctx, from, to)
		if args[i] != a {
			anyArg = true
		}
	}
	if recv != e.Recv || anyArg {
		return &DynamicCallExpr{
			Recv:        recv,
			Name:        e.Name,
			Args:        args,
			HasOverride: e.HasOverride,
		}
	}
	return e
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

func (e *SuperCallExpr) semantOpt(ctx *semCtx) Expr {
	args := make([]Expr, len(e.Args))
	anyArg := false
	for i, a := range e.Args {
		args[i] = a.semantOpt(ctx)
		if args[i] != a {
			anyArg = true
		}
	}
	if anyArg {
		e = &SuperCallExpr{
			Pos:   e.Pos,
			Name:  e.Name,
			Args:  args,
			Class: e.Class,
		}
	}
	if inl, ok := semantInline(ctx, &ThisExpr{
		Pos:   e.Pos,
		Class: e.Class,
	}, e.Name, e.Args); ok {
		return inl
	}
	return e
}

func (e *SuperCallExpr) semantReplaceObject(ctx *semCtx, from, to Object) Expr {
	recvPrev := &ThisExpr{
		Pos:   e.Pos,
		Class: e.Class,
	}
	recv := recvPrev.semantReplaceObject(ctx, from, to)
	args := make([]Expr, len(e.Args))
	anyArg := false
	for i, a := range e.Args {
		args[i] = a.semantReplaceObject(ctx, from, to)
		if args[i] != a {
			anyArg = true
		}
	}
	if recv != recvPrev || anyArg {
		return &StaticCallExpr{
			Recv: recv,
			Name: e.Name,
			Args: args,
		}
	}
	return e
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

func (e *StaticCallExpr) semantOpt(ctx *semCtx) Expr {
	recv := e.Recv.semantOpt(ctx)
	args := make([]Expr, len(e.Args))
	anyArg := false
	for i, a := range e.Args {
		args[i] = a.semantOpt(ctx)
		if args[i] != a {
			anyArg = true
		}
	}
	if recv != e.Recv || anyArg {
		e = &StaticCallExpr{
			Recv: recv,
			Name: e.Name,
			Args: args,
		}
	}
	if inl, ok := semantInline(ctx, e.Recv, e.Name, e.Args); ok {
		return inl
	}
	return e
}

func (e *StaticCallExpr) semantReplaceObject(ctx *semCtx, from, to Object) Expr {
	recv := e.Recv.semantReplaceObject(ctx, from, to)
	args := make([]Expr, len(e.Args))
	anyArg := false
	for i, a := range e.Args {
		args[i] = a.semantReplaceObject(ctx, from, to)
		if args[i] != a {
			anyArg = true
		}
	}
	if recv != e.Recv || anyArg {
		return &StaticCallExpr{
			Recv: recv,
			Name: e.Name,
			Args: args,
		}
	}
	return e
}

func (e *AllocExpr) semantTypes(ctx *semCtx, c *Class) {
	ctx.LookupClass(e.Type)
}

func (e *AllocExpr) semantIdentifiers(ctx *semCtx, ids semantIdentifiers) *Class {
	return e.Type.Class
}

func (e *AllocExpr) semantOpt(ctx *semCtx) Expr {
	return e
}

func (e *AllocExpr) semantReplaceObject(ctx *semCtx, from, to Object) Expr {
	return e
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

func (e *AssignExpr) semantOpt(ctx *semCtx) Expr {
	expr := e.Expr.semantOpt(ctx)
	if expr != e.Expr {
		return &AssignExpr{
			Name: e.Name,
			Expr: e.Expr,
			Unit: e.Unit,
		}
	}
	return e
}

func (e *AssignExpr) semantReplaceObject(ctx *semCtx, from, to Object) Expr {
	name := e.Name.semantReplaceObject(ctx, from, to)
	expr := e.Expr.semantReplaceObject(ctx, from, to)
	if name != e.Name || expr != e.Expr {
		return &AssignExpr{
			Name: name,
			Expr: expr,
			Unit: e.Unit,
		}
	}
	return e
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

func (e *VarExpr) semantOpt(ctx *semCtx) Expr {
	init := e.Init.semantOpt(ctx)
	body := e.Body.semantOpt(ctx)
	unused := body == body.semantReplaceObject(ctx, e, nil)
	if unused {
		return &ChainExpr{
			Pre:  init,
			Expr: body,
		}
	}
	if init != e.Init || body != e.Body {
		var v VarExpr
		v = VarExpr{
			Name: e.Name.semantReplaceObject(ctx, e, &v),
			Type: e.Type,
			Init: init,
			Body: body.semantReplaceObject(ctx, e, &v),
		}
		return &v
	}
	return e
}

func (e *VarExpr) semantReplaceObject(ctx *semCtx, from, to Object) Expr {
	name := e.Name.semantReplaceObject(ctx, from, to)
	init := e.Init.semantReplaceObject(ctx, from, to)
	body := e.Body.semantReplaceObject(ctx, from, to)
	if name != e.Name || init != e.Init || body != e.Body {
		var v VarExpr
		v = VarExpr{
			Name: name.semantReplaceObject(ctx, e, &v),
			Type: e.Type,
			Init: init,
			Body: body.semantReplaceObject(ctx, e, &v),
		}
		return &v
	}
	return e
}

func (e *ChainExpr) semantTypes(ctx *semCtx, c *Class) {
	e.Pre.semantTypes(ctx, c)
	e.Expr.semantTypes(ctx, c)
}

func (e *ChainExpr) semantIdentifiers(ctx *semCtx, ids semantIdentifiers) *Class {
	e.Pre.semantIdentifiers(ctx, ids)
	return e.Expr.semantIdentifiers(ctx, ids)
}

func (e *ChainExpr) semantOpt(ctx *semCtx) Expr {
	pre := e.Pre.semantOpt(ctx)
	expr := e.Expr.semantOpt(ctx)
	if pre != e.Pre || expr != e.Expr {
		return &ChainExpr{
			Pre:  pre,
			Expr: expr,
		}
	}
	return e
}

func (e *ChainExpr) semantReplaceObject(ctx *semCtx, from, to Object) Expr {
	pre := e.Pre.semantReplaceObject(ctx, from, to)
	expr := e.Expr.semantReplaceObject(ctx, from, to)
	if pre != e.Pre || expr != e.Expr {
		return &ChainExpr{
			Pre:  pre,
			Expr: expr,
		}
	}
	return e
}

func (e *ThisExpr) semantTypes(ctx *semCtx, c *Class) {
	e.Class = c
}

func (e *ThisExpr) semantIdentifiers(ctx *semCtx, ids semantIdentifiers) *Class {
	return e.Class
}

func (e *ThisExpr) semantOpt(ctx *semCtx) Expr {
	return e
}

func (e *ThisExpr) semantReplaceObject(ctx *semCtx, from, to Object) Expr {
	if from == (*AttributeObject)(nil) {
		return &NameExpr{
			Name: &Ident{
				Pos:    e.Pos,
				Name:   "this",
				Object: to,
			},
		}
	}
	return e
}

func (e *NullExpr) semantTypes(ctx *semCtx, c *Class) {
}

func (e *NullExpr) semantIdentifiers(ctx *semCtx, ids semantIdentifiers) *Class {
	return nullClass
}

func (e *NullExpr) semantOpt(ctx *semCtx) Expr {
	return e
}

func (e *NullExpr) semantReplaceObject(ctx *semCtx, from, to Object) Expr {
	return e
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

func (e *UnitExpr) semantOpt(ctx *semCtx) Expr {
	return e
}

func (e *UnitExpr) semantReplaceObject(ctx *semCtx, from, to Object) Expr {
	return e
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

func (e *NameExpr) semantOpt(ctx *semCtx) Expr {
	return e
}

func (e *NameExpr) semantReplaceObject(ctx *semCtx, from, to Object) Expr {
	name := e.Name.semantReplaceObject(ctx, from, to)
	if name != e.Name {
		return &NameExpr{
			Name: name,
		}
	}
	return e
}

func (e *StringExpr) semantTypes(ctx *semCtx, c *Class) {
	e.Lit.semantTypes(ctx, c)
}

func (e *StringExpr) semantIdentifiers(ctx *semCtx, ids semantIdentifiers) *Class {
	return e.Lit.Class
}

func (e *StringExpr) semantOpt(ctx *semCtx) Expr {
	return e
}

func (e *StringExpr) semantReplaceObject(ctx *semCtx, from, to Object) Expr {
	return e
}

func (e *BoolExpr) semantTypes(ctx *semCtx, c *Class) {
	e.Lit.semantTypes(ctx, c)
}

func (e *BoolExpr) semantIdentifiers(ctx *semCtx, ids semantIdentifiers) *Class {
	return e.Lit.Class
}

func (e *BoolExpr) semantOpt(ctx *semCtx) Expr {
	return e
}

func (e *BoolExpr) semantReplaceObject(ctx *semCtx, from, to Object) Expr {
	return e
}

func (e *IntExpr) semantTypes(ctx *semCtx, c *Class) {
	e.Lit.semantTypes(ctx, c)
}

func (e *IntExpr) semantIdentifiers(ctx *semCtx, ids semantIdentifiers) *Class {
	return e.Lit.Class
}

func (e *IntExpr) semantOpt(ctx *semCtx) Expr {
	return e
}

func (e *IntExpr) semantReplaceObject(ctx *semCtx, from, to Object) Expr {
	return e
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

func (e *NativeExpr) semantOpt(ctx *semCtx) Expr {
	return e
}

func (e *NativeExpr) semantReplaceObject(ctx *semCtx, from, to Object) Expr {
	panic("NativeExpr.semantReplaceObject should never be called")
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

func (i *Ident) semantReplaceObject(ctx *semCtx, from, to Object) *Ident {
	if i.Object == from {
		return &Ident{
			Pos:    i.Pos,
			Name:   i.Name,
			Object: to,
		}
	}

	if from == (*AttributeObject)(nil) {
		if attr, ok := i.Object.(*Attribute); ok {
			return &Ident{
				Pos:  i.Pos,
				Name: i.Name,
				Object: &AttributeObject{
					Object:    to,
					Attribute: attr,
				},
			}
		}

		if _, ok := i.Object.(*AttributeObject); ok {
			panic("INTERNAL COMPILER ERROR: attempt to inline a method twice")
		}
	}

	return i
}
