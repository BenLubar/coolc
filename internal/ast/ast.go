package ast

import (
	"fmt"
	"go/token"
)

// Program is a set of classes with a generated main method that is called by
// the runtime.
type Program struct {
	// Classes is the list of Cool classes in the order they were read by
	// the compiler.
	Classes []*Class

	// Ordered is a topological ordering of the classes.
	Ordered []*Class

	// Main is a generated expression that is equivalent to the Cool code
	// `new Main()`. It is used as an entry point for the runtime.
	Main Expr

	classMap map[string]*Class
}

// Class is a Cool class as defined in CoolAid section 3.
type Class struct {
	// Type is the name of the class.
	Type *Ident
	// Formals is the set of arguments to the class's constructor as they
	// are defined in the source code.
	Formals []*Formal
	// Extends is the extends declaration, as defined in section 3.2 of
	// CoolAid, or a generated one for classes implicitly extending Any.
	Extends *Extends
	// Features is the set of features, as defined in section 3.1 of
	// CoolAid, in the order they appear in source code. Features also
	// contains generated features for the formals of the class and a
	// generated constructor for user-creatable classes.
	Features []Feature

	// Order is this class's position in the program's topological ordering
	// of classes. A class with Order x and MaxOrder y is an ancestor of
	// every class with x < Order <= y.
	Order int
	// MaxOrder is the highest Order of any class with this class as
	// an ancestor, or Order if this class has no children.
	MaxOrder int
	// Depth is the depth of this class. 1 is Any, 2 is a class with no
	// "extends" declaration, 3 is a class that extends 2, and so on.
	Depth int

	// Methods is this class's method table, with the first N methods
	// (where N is the number of methods in the parent's method table)
	// overriding the parent's methods. The methods in Methods may be
	// from an ancestor class.
	Methods []*Method
	// HasOverride is true for any method that was overridden in a child
	// class. Used for optimization.
	HasOverride []bool

	// Size is the size in bytes of this class, not counting metadata or
	// native fields.
	Size int
	// NameID is the ID of the string literal with the name of this class,
	// used for the Any.toString method.
	NameID int
}

// Extends is the "extends" declaration of a class, as defined in section 3.2
// of CoolAid.
type Extends struct {
	// Type is the name of the parent class.
	Type *Ident
	// Args is the list of arguments given to the parent constructor.
	Args []Expr
}

// Formal is a method argument.
type Formal struct {
	// Name is the name of the argument.
	Name *Ident
	// Type is the declared type of the argument.
	Type *Ident

	// Offset is the offset from the base pointer of the temporary used for
	// this variable.
	Offset int
}

// Base implements Object.
func (a *Formal) Base(this int) string {
	return "%ebp"
}

// Offs implements Object.
func (a *Formal) Offs() string {
	return fmt.Sprintf("%d", a.Offset)
}

// Stack implements Object.
func (a *Formal) Stack() bool {
	return true
}

// RawInt implements Object.
func (a *Formal) RawInt() bool {
	return false
}

// Feature is a feature as defined by section 3.1 of CoolAid.
type Feature interface {
	semantTypes(func(*Ident), *Class)
}

// Init is a block feature. It is inlined into the constructor in the order it
// was defined.
type Init struct {
	// Expr is the contents of the block feature.
	Expr Expr
}

// Attribute is a var feature, as defined by section 5 of CoolAid. Its
// initialization is inlined into the constructor in the order it was defined.
type Attribute struct {
	// Name is the name of this attribute.
	Name *Ident
	// Type is the declared type of this attribute.
	Type *Ident
	// Init is the initializer for this attribute.
	Init Expr

	// Parent is the class that this attribute is declared within.
	Parent *Class
}

// Base implements Object.
func (a *Attribute) Base(this int) string {
	return fmt.Sprintf("%d(%%ebp)", this)
}

// Offs implements Object.
func (a *Attribute) Offs() string {
	return fmt.Sprintf("offset_of_%s.%s", a.Parent.Type.Name, a.Name.Name)
}

// Stack implements Object.
func (a *Attribute) Stack() bool {
	return false
}

// RawInt implements Object.
func (a *Attribute) RawInt() bool {
	return false
}

// Method is a def feature, as defined by section 6 of CoolAid.
type Method struct {
	// Override is true if and only if the `override` keyword was used in
	// this method declaration.
	Override bool
	// Name is the name of this method.
	Name *Ident
	// Args are the formal arguments of this method.
	Args []*Formal
	// Type is the declared return type of this method.
	Type *Ident
	// Body is the expression given for this method.
	Body Expr

	// Parent is the class that this method is declared within.
	Parent *Class

	// Order is the method table offset of this method.
	Order int
}

// Expr is an expression.
type Expr interface {
	semantTypes(func(*Ident), *Class)
	semantIdentifiers(func(token.Pos, string), func(*Class, *Ident), func(...*Class) *Class, semantIdentifiers) *Class

	genCollectLiterals(*genCtx)
	genCountVars(*genCtx) int
	genCode(*genCtx)
}

// ArithmeticExpr is an expression that can return an unboxed integer.
type ArithmeticExpr interface {
	Expr

	genCodeRawInt(*genCtx)
}

// JumpExpr is an expression that can jump instead of returning a boolean.
type JumpExpr interface {
	Expr

	genCodeJump(*genCtx, string, string)
}

// NotExpr is an expression of the form `!x`.
type NotExpr struct {
	// Expr is `x` in the expression `!x`.
	Expr Expr

	// Boolean is a generated identifier for the Boolean class in
	// basic.cool.
	Boolean *Ident
}

// NegativeExpr is an expression of the form `-x`.
type NegativeExpr struct {
	// Expr is `x` in the expression `-x`.
	Expr Expr

	// Int is a generated identifier for the Int class in basic.cool.
	Int *Ident
}

// IfExpr is an expression of the form `if (x) y else z`.
type IfExpr struct {
	// Cond is `x` in the expression `if (x) y else z`.
	Cond Expr
	// Then is `y` in the expression `if (x) y else z`.
	Then Expr
	// Else is `z` in the expression `if (x) y else z`.
	Else Expr

	// Boolean is a generated identifier for the Boolean class in
	// basic.cool.
	Boolean *Ident
}

// WhileExpr is an expression of the form `while (x) y`.
type WhileExpr struct {
	// Cond is `x` in the expression `while (x) y`.
	Cond Expr
	// Body is `y` in the expression `while (x) y`.
	Body Expr

	// Boolean is a generated identifier for the Boolean class in
	// basic.cool.
	Boolean *Ident
	// Unit is a generated identifier for the Unit class in basic.cool.
	Unit *Ident
}

// BinaryOperator is the shared data structure for expressions of the form
// `x op y`.
type BinaryOperator struct {
	// Pos is the position of the first byte of the operator.
	Pos token.Pos
	// Left is `x` in the expression `x op y`.
	Left Expr
	// Right is `y` in the expression `x op y`.
	Right Expr

	// Boolean is a generated identifier for the Boolean class in
	// basic.cool. It may be nil if it is not needed.
	Boolean *Ident
	// Int is a generated identifier for the Int class in basic.cool. It
	// may be nil if it is not needed.
	Int *Ident
}

type (
	// LessOrEqualExpr is an expression of the form `x <= y`.
	LessOrEqualExpr BinaryOperator
	// LessThanExpr is an expression of the form `x < y`.
	LessThanExpr BinaryOperator
	// MultiplyExpr is an expression of the form `x * y`.
	MultiplyExpr BinaryOperator
	// DivideExpr is an expression of the form `x / y`.
	DivideExpr BinaryOperator
	// AddExpr is an expression of the form `x + y`.
	AddExpr BinaryOperator
	// SubtractExpr is an expression of the form `x - y`.
	SubtractExpr BinaryOperator
)

// MatchExpr is an expression of the form `x match { ... }`
type MatchExpr struct {
	// Pos is the position of the `match` keyword.
	Pos token.Pos
	// Left is `x` in the expression `x match { ... }`
	Left Expr
	// Cases are the cases given, in source code order.
	Cases []*Case

	// Offset is the stack offset of the temporary for the value of Left.
	Offset int
}

// Base implements Object.
func (e *MatchExpr) Base(this int) string {
	return "%ebp"
}

// Offs implements Object.
func (e *MatchExpr) Offs() string {
	return fmt.Sprintf("%d", e.Offset)
}

// Stack implements Object.
func (e *MatchExpr) Stack() bool {
	return true
}

// RawInt implements Object.
func (e *MatchExpr) RawInt() bool {
	return false
}

// DynamicCallExpr is an expression of the form `x(...)` or `y.x(...)`.
type DynamicCallExpr struct {
	// Recv is `this` in `x(...)` or `y` in `y.x(...)`.
	Recv Expr
	// Name is `x` in `x(...)` or `y.x(...)`.
	Name *Ident
	// Args are the arguments given to the method call.
	Args []Expr

	// HasOverride is true if the method called is unknown at compile time.
	HasOverride bool
}

// SuperCallExpr is an expression of the form `super.x(...)`.
type SuperCallExpr struct {
	// Pos is the position of the `super` keyword.
	Pos token.Pos
	// Name is `x` in the expression `super.x(...)`.
	Name *Ident
	// Args are the arguments given to the method call.
	Args []Expr

	// Class is the parent of the class this expression is lexically
	// within.
	Class *Class
}

// StaticCallExpr is used for expressions of the form `new X(...)`.
type StaticCallExpr struct {
	// Recv is the receiver for this method call. In the case of
	// `new X(...)`, it is an AllocExpr.
	Recv Expr
	// Name is the method name for this method call.
	Name *Ident
	// Args are the arguments given to the method call.
	Args []Expr
}

// AllocExpr is used for expressions of the form `new X(...)`.
type AllocExpr struct {
	// Type is X in the expression `new X(...)`.
	Type *Ident
}

// AssignExpr is an expression of the form `x = y`.
type AssignExpr struct {
	// Name is `x` in the expression `x = y`.
	Name *Ident
	// Expr is `y` in the expression `x = y`.
	Expr Expr

	// Unit is a generated identifier for the Unit class in basic.cool.
	Unit *Ident
}

// VarExpr is an expression of the form `var x : X = y; z`.
type VarExpr struct {
	// Name is `x` in the expression `var x : X = y; z`.
	Name *Ident
	// Type is `X` in the expression `var x : X = y; z`.
	Type *Ident
	// Init is `y` in the expression `var x : X = y; z`.
	Init Expr
	// Body is `z` in the expression `var x : X = y; z`.
	Body Expr

	// Offset is the stack offset of this variable.
	Offset int
}

// Base implements Object.
func (e *VarExpr) Base(this int) string {
	return "%ebp"
}

// Offs implements Object.
func (e *VarExpr) Offs() string {
	return fmt.Sprintf("%d", e.Offset)
}

// Stack implements Object.
func (e *VarExpr) Stack() bool {
	return true
}

// RawInt implements Object.
func (e *VarExpr) RawInt() bool {
	return e.Type.Name == "Int"
}

// ChainExpr is an expression of the form `x; y`.
type ChainExpr struct {
	Pre  Expr
	Expr Expr
}

// ThisExpr is an expression of the form `this`. It is also used for method
// calls with no explicit reciever.
type ThisExpr struct {
	// Pos is the position of the `this` keyword.
	Pos token.Pos

	// Class is the current class.
	Class *Class
}

// NullExpr is an expression of the form `null`.
type NullExpr struct {
	// Pos is the position of the `null` keyword.
	Pos token.Pos
}

// UnitExpr is an expression of the form `()`.
type UnitExpr struct {
	// Pos is the position of (.
	Pos token.Pos

	// Class is the Unit class from basic.cool.
	Class *Class
}

// NameExpr is an expression of the form `identifier`.
type NameExpr struct {
	// Name is the identifier.
	Name *Ident
}

// StringExpr is an expression of the form `"string"` or `"""raw string"""`.
type StringExpr struct {
	// Lit is the string literal.
	Lit *StringLit
}

// BoolExpr is an expression of the form `true` or `false`.
type BoolExpr struct {
	// Lit is the boolean literal.
	Lit *BoolLit
}

// IntExpr is an expression of the form `1234`.
type IntExpr struct {
	// Lit is the integer literal.
	Lit *IntLit
}

// NativeExpr is an expression of the form `native`. It can only be used by
// code in basic.cool and cannot be used as part of another expression.
type NativeExpr struct {
	// Pos is the position of the `native` keyword.
	Pos token.Pos
}

// Case is a case of the form `case x : Y => z` or `case null => z`.
type Case struct {
	// Name is `x` in `case x : Y => z` or `null` in `case null => z`.
	Name *Ident
	// Type is `Y` in `case x : Y => z` or `Null` in `case null => z`.
	Type *Ident
	// Body is `z` in `case x : Y => z` or `case null => z`.
	Body Expr
}

// Object is a stored value.
type Object interface {
	// Base is the base location for this object. The argument is the
	// offset of `this` from %ebp in the current method.
	Base(int) string
	// Offs is the offset of this object from Base.
	Offs() string
	// Stack returns true if this is stored on the stack. In this
	// implementation, the stack is reference-counted while the heap is
	// garbage-collected.
	Stack() bool
	// RawInt returns true if this is an unboxed integer. RawInt requres
	// Stack.
	RawInt() bool
}

// Ident is an object or type identifier. Exactly one of Class, Method, Object
// will be set.
type Ident struct {
	// Name is the name given as an identifier.
	Name string
	// Pos is the position of the first byte of the name.
	Pos token.Pos

	// Class is the class this type identifier refers to.
	Class *Class
	// Method is the method this object identifier refers to.
	Method *Method
	// Object is the variable this object identifier refers to.
	Object Object
}

// IntLit is an integer literal.
type IntLit struct {
	// Int is the value of this integer literal. Negative integer literals
	// do not exist. Instead, a NegativeExpr will wrap an IntExpr.
	Int int32
	// Pos is the position of the first digit of this literal.
	Pos token.Pos

	// Class is the Int class from basic.cool.
	Class *Class

	// LitID is the ID of this integer literal. It is used to generate a
	// symbol name during code generation.
	LitID int
}

// StringLit is a string literal.
type StringLit struct {
	// Str is the contents of this string literal. Backslash escapes are
	// already processed by this point.
	Str string
	// Pos is the position of the first `"` of this literal.
	Pos token.Pos

	// Class is the String class from basic.cool.
	Class *Class

	// LitID is the ID of this string literal. It is used to generate a
	// symbol name during code generation.
	LitID int
}

// BoolLit is `true` or `false`.
type BoolLit struct {
	// Bool is the value of this boolean literal.
	Bool bool
	// Pos is the position of the `true` or `false` keyword.
	Pos token.Pos

	// Class is the Boolean class from basic.cool.
	Class *Class
}
