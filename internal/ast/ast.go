package ast

import "go/token"

type Program struct {
	Classes []*Class
}

type Class struct {
	Type     *Ident
	Formals  []*Formal
	Extends  *Extends
	Features []Feature
}

type Extends struct {
	Type *Ident
	Args []Expr
}

type Formal struct {
	Name *Ident
	Type *Ident
}

type Feature interface{}

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
}

type Expr interface{}

type NotExpr struct {
	Expr Expr
}

type NegativeExpr struct {
	Expr Expr
}

type IfExpr struct {
	Cond Expr
	Then Expr
	Else Expr
}

type WhileExpr struct {
	Cond Expr
	Body Expr
}

type BinaryOperator struct {
	Pos   token.Pos
	Left  Expr
	Right Expr
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
}

type NullExpr struct {
	Pos token.Pos
}

type UnitExpr struct {
	Pos token.Pos
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
}

type Ident struct {
	Name string
	Pos  token.Pos
}

type IntLit struct {
	Int int32
	Pos token.Pos
}

type StringLit struct {
	Str string
	Pos token.Pos
}

type BoolLit struct {
	Bool bool
	Pos  token.Pos
}
