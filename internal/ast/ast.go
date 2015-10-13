package ast

import (
	"fmt"
	"go/token"
	"io"
)

type Program struct {
	Classes []*Class

	Ordered []*Class

	Main Expr

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

	Size   int
	NameID int
}

type Extends struct {
	Type *Ident
	Args []Expr
}

type Formal struct {
	Name *Ident
	Type *Ident

	Offset int
}

func (a *Formal) Base(this int) string {
	return "%ebp"
}

func (a *Formal) Offs() string {
	return fmt.Sprintf("%d", a.Offset)
}

func (a *Formal) Stack() bool {
	return true
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

	Parent *Class
}

func (a *Attribute) Base(this int) string {
	return fmt.Sprintf("%d(%%ebp)", this)
}

func (a *Attribute) Offs() string {
	return fmt.Sprintf("offset_of_%s.%s", a.Parent.Type.Name, a.Name.Name)
}

func (a *Attribute) Stack() bool {
	return false
}

type Method struct {
	Override bool
	Name     *Ident
	Args     []*Formal
	Type     *Ident
	Body     Expr

	Parent *Class

	Order int
}

type Expr interface {
	semantTypes(func(*Ident), *Class)
	semantIdentifiers(func(token.Pos, string), func(*Class, *Ident), func(...*Class) *Class, semantIdentifiers) *Class

	genCollectLiterals(func(int32) int, func(string) int)
	genCountVars(int) int
	genCode(io.Writer, func() string, func() (int, func()))
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

	Offset int
}

func (e *MatchExpr) Base(this int) string {
	return "%ebp"
}

func (e *MatchExpr) Offs() string {
	return fmt.Sprintf("%d", e.Offset)
}

func (e *MatchExpr) Stack() bool {
	return true
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

	This int
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

	This int
}

type VarExpr struct {
	Name *Ident
	Type *Ident
	Init Expr
	Body Expr

	Offset int
}

func (e *VarExpr) Base(this int) string {
	return "%ebp"
}

func (e *VarExpr) Offs() string {
	return fmt.Sprintf("%d", e.Offset)
}

func (e *VarExpr) Stack() bool {
	return true
}

type ChainExpr struct {
	Pre  Expr
	Expr Expr
}

type ThisExpr struct {
	Pos token.Pos

	Class *Class

	Offset int
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

	This int
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

type Object interface {
	Base(int) string
	Offs() string
	Stack() bool
}

type Ident struct {
	Name string
	Pos  token.Pos

	Class  *Class
	Method *Method
	Object Object
}

type IntLit struct {
	Int int32
	Pos token.Pos

	Class *Class

	LitID int
}

type StringLit struct {
	Str string
	Pos token.Pos

	Class *Class

	LitID int
}

type BoolLit struct {
	Bool bool
	Pos  token.Pos

	Class *Class
}
