package ast

import "go/token"

type Program struct {
	Classes  []*Class
	classMap map[string]*Class
}

type Class struct {
	Type     *Ident
	Formals  []*Formal
	Extends  *Extends
	Features []Feature

	Order    int
	MaxOrder int
	Depth    int

	Methods []*Method
}

type Extends struct {
	Type *Ident
	Args []Expr
}

type Formal struct {
	Name *Ident
	Type *Ident
}

type Feature interface {
	semantTypes(func(*Ident), *Class)
}

type Init struct {
	Expr Expr
}

type Attribute struct {
	Name *Ident
	Type *Ident
	Init Expr
}

type Method struct {
	Override bool
	Name     *Ident
	Args     []*Formal
	Type     *Ident
	Body     Expr

	Order int
}

type Expr interface {
	semantTypes(func(*Ident), *Class)
	semantIdentifiers(func(token.Pos, string), func(*Class, *Ident), func(...*Class) *Class, semantIdentifiers) *Class
}

type NotExpr struct {
	Expr Expr

	Boolean *Ident
}

type NegativeExpr struct {
	Expr Expr

	Int *Ident
}

type IfExpr struct {
	Cond Expr
	Then Expr
	Else Expr

	Boolean *Ident
}

type WhileExpr struct {
	Cond Expr
	Body Expr

	Boolean *Ident
	Unit    *Ident
}

type BinaryOperator struct {
	Pos   token.Pos
	Left  Expr
	Right Expr

	Boolean *Ident
	Int     *Ident
}

type (
	LessOrEqualExpr BinaryOperator
	LessThanExpr    BinaryOperator
	MultiplyExpr    BinaryOperator
	DivideExpr      BinaryOperator
	AddExpr         BinaryOperator
	SubtractExpr    BinaryOperator
)

type MatchExpr struct {
	Pos   token.Pos
	Left  Expr
	Cases []*Case
}

type DynamicCallExpr struct {
	Recv Expr
	Name *Ident
	Args []Expr
}

type SuperCallExpr struct {
	Pos  token.Pos
	Name *Ident
	Args []Expr

	Class *Class
}

type StaticCallExpr struct {
	Recv Expr
	Name *Ident
	Args []Expr
}

type AllocExpr struct {
	Type *Ident
}

type AssignExpr struct {
	Name *Ident
	Expr Expr

	Unit *Ident
}

type VarExpr struct {
	Name *Ident
	Type *Ident
	Init Expr
	Body Expr
}

type ChainExpr struct {
	Pre  Expr
	Expr Expr
}

type ThisExpr struct {
	Pos token.Pos

	Class *Class
}

type NullExpr struct {
	Pos token.Pos
}

type UnitExpr struct {
	Pos token.Pos

	Class *Class
}

type NameExpr struct {
	Name *Ident
}

type StringExpr struct {
	Lit *StringLit
}

type BoolExpr struct {
	Lit *BoolLit
}

type IntExpr struct {
	Lit *IntLit
}

type NativeExpr struct {
	Pos token.Pos
}

type Case struct {
	Name *Ident
	Type *Ident
	Body Expr

	Tags []int
}

type Ident struct {
	Name string
	Pos  token.Pos

	Object interface{}
}

type IntLit struct {
	Int int32
	Pos token.Pos

	Class *Class
}

type StringLit struct {
	Str string
	Pos token.Pos

	Class *Class
}

type BoolLit struct {
	Bool bool
	Pos  token.Pos

	Class *Class
}
