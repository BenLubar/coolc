package ast

import (
	"fmt"
	"go/token"
	"io"
	"strconv"
)

type genCtx struct {
	w io.Writer

	ints          []int32
	strings       []string
	stringLengths []int

	this int

	label    int
	vars     int
	varsUsed int

	fset *token.FileSet

	opt Options
}

func (ctx *genCtx) Printf(format string, args ...interface{}) {
	_, err := fmt.Fprintf(ctx.w, format, args...)
	if err != nil {
		panic(err)
	}
}

func (ctx *genCtx) AddInt(x int32) int {
	for i, y := range ctx.ints {
		if x == y {
			return i
		}
	}
	ctx.ints = append(ctx.ints, x)
	return len(ctx.ints) - 1
}

func (ctx *genCtx) AddString(x string) int {
	for i, y := range ctx.strings {
		if x == y {
			return i
		}
	}
	ctx.strings = append(ctx.strings, x)
	ctx.stringLengths = append(ctx.stringLengths, ctx.AddInt(int32(len(x))))
	return len(ctx.strings) - 1
}

func (ctx *genCtx) Label() string {
	ctx.label++
	return strconv.Itoa(ctx.label)
}

func (ctx *genCtx) Slot() (int, func()) {
	if ctx.vars == ctx.varsUsed {
		panic("INTERNAL ERROR: too many vars")
	}
	ctx.varsUsed++
	n := ctx.varsUsed
	return -n * 4, func() {
		if ctx.varsUsed != n {
			panic("INTERNAL ERROR: missed var release")
		}
		ctx.varsUsed--
	}
}

func (p *Program) CodeGen(opt Options, fset *token.FileSet, w io.Writer) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	ctx := &genCtx{
		w:    w,
		fset: fset,
		opt:  opt,
	}
	ctx.AddInt(0) // int_lit_0 must be 0
	nullClassID := ctx.AddString("Null")
	p.genCollectLiterals(ctx)

	ctx.Printf(".include \"basic_defs.s\"\n")
	ctx.Printf("\n")
	ctx.Printf(".data\n")
	ctx.Printf("\n")
	for i, x := range ctx.ints {
		ctx.Printf(".align 2\n")
		ctx.Printf("int_lit_%d:\n", i)
		ctx.Printf("\t.long tag_of_Int\n")
		ctx.Printf("\t.long size_of_Int + 4\n")
		ctx.Printf("\t.long gc_tag_root\n")
		ctx.Printf("\t.long %d\n", x)
		ctx.Printf("\n")
	}

	for i, x := range ctx.strings {
		ctx.Printf(".align 2\n")
		ctx.Printf("string_lit_%d:\n", i)
		ctx.Printf("\t.long tag_of_String\n")
		ctx.Printf("\t.long size_of_String + %d\n", len(x))
		ctx.Printf("\t.long gc_tag_root\n")
		ctx.Printf("\t.long int_lit_%d\n", ctx.stringLengths[i])
		ctx.Printf("\t// %q\n", x)
		for j := 0; j < len(x); j++ {
			ctx.Printf("\t.byte %d\n", x[j])
		}
		ctx.Printf("\n")
	}

	ctx.Printf(".globl max_tag\n")
	ctx.Printf(".set max_tag, %d\n", len(p.Ordered))
	ctx.Printf("\n")

	ctx.Printf(".globl gc_sizes\n")
	ctx.Printf(".align 2\n")
	ctx.Printf("gc_sizes:\n")
	ctx.Printf("\t.long 0\n")
	for _, c := range p.Ordered {
		ctx.Printf("\t.long size_of_%s / 4\n", c.Type.Name)
	}
	ctx.Printf("\n")

	for _, c := range p.Ordered {
		ctx.Printf(".align 2\n")
		ctx.Printf("methods_of_%s:\n", c.Type.Name)
		for _, m := range c.Methods {
			ctx.Printf("\t.long %s.%s\n", m.Parent.Type.Name, m.Name.Name)
			if m.Parent == c {
				ctx.Printf("\t.globl method_offset_%s.%s\n", c.Type.Name, m.Name.Name)
				ctx.Printf("\t.set method_offset_%s.%s, %d\n", c.Type.Name, m.Name.Name, m.Order*4)
			}
		}
		ctx.Printf("\n")
	}

	ctx.Printf(".globl class_names\n")
	ctx.Printf(".align 2\n")
	ctx.Printf("class_names:\n")
	ctx.Printf("\t.long string_lit_%d\n", nullClassID)
	for _, c := range p.Ordered {
		ctx.Printf("\t.long string_lit_%d\n", c.NameID)
	}
	ctx.Printf("\n")

	ctx.Printf(".globl method_tables\n")
	ctx.Printf(".align 2\n")
	ctx.Printf("method_tables:\n")
	ctx.Printf("\t.long 0\n")
	for _, c := range p.Ordered {
		ctx.Printf("\t.long methods_of_%s\n", c.Type.Name)
	}
	ctx.Printf("\n")

	for _, c := range p.Ordered {
		c.genConstants(ctx)
	}
	ctx.Printf("\n")
	ctx.Printf(".text\n")

	genMethod(ctx, "main", -1, p.Main)

	for _, c := range p.Ordered {
		c.genCode(ctx)
	}

	return
}

func (p *Program) genCollectLiterals(ctx *genCtx) {
	p.Main.genCollectLiterals(ctx)
	for _, c := range p.Classes {
		c.genCollectLiterals(ctx)
	}
}

func (c *Class) genCollectLiterals(ctx *genCtx) {
	c.NameID = ctx.AddString(c.Type.Name)

	for _, f := range c.Features {
		if m, ok := f.(*Method); ok {
			m.Body.genCollectLiterals(ctx)
		}
	}
}

func (c *Class) genConstants(ctx *genCtx) {
	ctx.Printf(".globl tag_of_%s\n", c.Type.Name)
	ctx.Printf(".set tag_of_%s, %d\n", c.Type.Name, c.Order)
	c.Size = c.Extends.Type.Class.Size
	for _, f := range c.Features {
		if a, ok := f.(*Attribute); ok {
			ctx.Printf(".globl offset_of_%s.%s\n", c.Type.Name, a.Name.Name)
			ctx.Printf(".set offset_of_%s.%s, data_offset + %d\n", c.Type.Name, a.Name.Name, c.Size)
			if _, ok := a.Init.(*NativeExpr); !ok {
				c.Size += 4
			}
		}
	}
	ctx.Printf(".globl size_of_%s\n", c.Type.Name)
	ctx.Printf(".set size_of_%s, %d\n", c.Type.Name, c.Size)
}

func genMethod(ctx *genCtx, name string, args int, body Expr) {
	ctx.Printf("\n")
	ctx.Printf(".globl %s\n", name)
	ctx.Printf(".type %s, @function\n", name)
	ctx.Printf("%s:\n", name)
	ctx.Printf("\t.cfi_startproc\n")

	ctx.this = args*4 + 8
	ctx.vars = body.genCountVars(ctx)
	if ctx.opt.Coroutine {
		ctx.Printf("\tmovl $%d, %%eax\n", (2+ctx.vars+body.genCountStack(ctx))*4)
		ctx.Printf("\tmovl $%d, %%ebx\n", args+1)
		ctx.Printf("\tcall runtime.morestack\n")
	}
	ctx.Printf("\tpush %%ebp\n")
	ctx.Printf("\t.cfi_def_cfa_offset 8\n")
	ctx.Printf("\t.cfi_offset ebp, -8\n")
	ctx.Printf("\tmovl %%esp, %%ebp\n")
	ctx.Printf("\t.cfi_def_cfa_register ebp\n")
	ctx.Printf("\tsubl $%d, %%esp\n", ctx.vars*4)

	//ctx.Printf("\tmovl $0, %%eax\n")
	//ctx.Printf("\tcall gc_check\n")

	ctx.label = 0
	ctx.varsUsed = 0
	body.genCode(ctx)
	for i := 0; i <= args; i++ {
		ctx.Printf("\tmovl %d(%%ebp), %%ebx\n", i*4+8)
		genGC(ctx, "%ebx")
	}

	//ctx.Printf("\tcall gc_check\n")

	ctx.Printf("\tleave\n")
	ctx.Printf("\t.cfi_def_cfa esp, 4\n")
	ctx.Printf("\tret $%d\n", args*4+4)
	ctx.Printf("\t.cfi_endproc\n")
	ctx.Printf("\t.size %s, .-%s\n", name, name)
}

func genRef(ctx *genCtx, reg string) {
	label_done := ctx.Label()

	ctx.Printf("\ttest %s, %s\n", reg, reg)
	ctx.Printf("\tjz %sf\n", label_done)
	ctx.Printf("\tcmpl $0, gc_offset(%s)\n", reg)
	ctx.Printf("\tjl %sf\n", label_done)
	ctx.Printf("\tincl gc_offset(%s)\n", reg)
	ctx.Printf("%s:\n", label_done)
}

func genGC(ctx *genCtx, reg string) {
	label_done := ctx.Label()

	ctx.Printf("\ttest %s, %s\n", reg, reg)
	ctx.Printf("\tjz %sf\n", label_done)
	ctx.Printf("\tcmpl $0, gc_offset(%s)\n", reg)
	ctx.Printf("\tjle %sf\n", label_done)
	ctx.Printf("\tdecl gc_offset(%s)\n", reg)
	ctx.Printf("%s:\n", label_done)
}

func genCodeRawInt(ctx *genCtx, e Expr) {
	if raw, ok := e.(ArithmeticExpr); ok {
		raw.genCodeRawInt(ctx)
	} else {
		e.genCode(ctx)
		genGC(ctx, "%eax")
		ctx.Printf("\tmovl offset_of_Int.value(%%eax), %%eax\n")
	}
}

func genCodeJump(ctx *genCtx, e Expr, l0, l1 string) {
	if raw, ok := e.(JumpExpr); ok {
		raw.genCodeJump(ctx, l0, l1)
	} else {
		e.genCode(ctx)
		ctx.Printf("\tcmpl $boolean_false, %%eax\n")
		ctx.Printf("\tje %s\n", l0)
		ctx.Printf("\tjmp %s\n", l1)
	}
}

func genCodeUnused(ctx *genCtx, e Expr) {
	if raw, ok := e.(UnusedExpr); ok {
		raw.genCodeUnused(ctx)
	} else {
		e.genCode(ctx)
		genGC(ctx, "%eax")
	}
}

func (c *Class) genCode(ctx *genCtx) {
	for _, f := range c.Features {
		if m, ok := f.(*Method); ok {
			if _, ok := m.Body.(*NativeExpr); ok {
				continue
			}

			var name string
			if file := ctx.fset.File(m.Name.Pos); file != nil {
				name = file.Name()
			}

			ctx.Printf("\n")
			ctx.Printf(".file %q\n", name)

			for i, a := range m.Args {
				a.Offset = (len(m.Args)-i)*4 + 4
			}
			genMethod(ctx, c.Type.Name+"."+m.Name.Name, len(m.Args), m.Body)
		}
	}
}

func (e *NotExpr) genCollectLiterals(ctx *genCtx) {
	e.Expr.genCollectLiterals(ctx)
}

func (e *NotExpr) genCountStack(ctx *genCtx) int {
	return e.Expr.genCountStack(ctx)
}

func (e *NotExpr) genCountVars(ctx *genCtx) int {
	return e.Expr.genCountVars(ctx)
}

func (e *NotExpr) genCode(ctx *genCtx) {
	label_true := ctx.Label()
	label_false := ctx.Label()
	label_done := ctx.Label()

	e.genCodeJump(ctx, label_false+"f", label_true+"f")

	ctx.Printf("%s:\n", label_true)
	ctx.Printf("\tlea boolean_true, %%eax\n")
	ctx.Printf("\tjmp %sf\n", label_done)
	ctx.Printf("%s:\n", label_false)
	ctx.Printf("\tlea boolean_false, %%eax\n")
	ctx.Printf("%s:\n", label_done)
}

func (e *NotExpr) genCodeJump(ctx *genCtx, l0, l1 string) {
	genCodeJump(ctx, e.Expr, l1, l0)
}

func (e *NegativeExpr) genCollectLiterals(ctx *genCtx) {
	e.Expr.genCollectLiterals(ctx)
}

func (e *NegativeExpr) genCountStack(ctx *genCtx) int {
	return e.Expr.genCountStack(ctx)
}

func (e *NegativeExpr) genCountVars(ctx *genCtx) int {
	vars := e.Expr.genCountVars(ctx)
	if v := 1; vars < v {
		vars = v
	}
	return vars
}

func (e *NegativeExpr) genCode(ctx *genCtx) {
	offset, unreserve := ctx.Slot()
	ctx.Printf("\tmovl $(size_of_Int + 4), %%eax\n")
	ctx.Printf("\tmovl $tag_of_Int, %%ebx\n")
	ctx.Printf("\tcall gc_alloc\n")
	ctx.Printf("\tmovl %%eax, %d(%%ebp)\n", offset)

	genCodeRawInt(ctx, e.Expr)
	ctx.Printf("\tmovl %%eax, %%ebx\n")

	ctx.Printf("\tmovl %d(%%ebp), %%eax\n", offset)
	unreserve()

	ctx.Printf("\tnegl %%ebx\n")
	ctx.Printf("\tmovl %%ebx, offset_of_Int.value(%%eax)\n")
}

func (e *NegativeExpr) genCodeRawInt(ctx *genCtx) {
	genCodeRawInt(ctx, e.Expr)
	ctx.Printf("\tnegl %%eax\n")
}

func (e *IfExpr) genCollectLiterals(ctx *genCtx) {
	e.Cond.genCollectLiterals(ctx)
	e.Then.genCollectLiterals(ctx)
	e.Else.genCollectLiterals(ctx)
}

func (e *IfExpr) genCountStack(ctx *genCtx) int {
	stack := e.Cond.genCountStack(ctx)
	if s := e.Then.genCountStack(ctx); s > stack {
		stack = s
	}
	if s := e.Else.genCountStack(ctx); s > stack {
		stack = s
	}
	return stack
}

func (e *IfExpr) genCountVars(ctx *genCtx) int {
	vars := e.Cond.genCountVars(ctx)
	if v := e.Then.genCountVars(ctx); v > vars {
		vars = v
	}
	if v := e.Else.genCountVars(ctx); v > vars {
		vars = v
	}
	return vars
}

func (e *IfExpr) genCodeShared(ctx *genCtx, t, f func()) {
	label_true := ctx.Label()
	label_false := ctx.Label()
	label_done := ctx.Label()

	genCodeJump(ctx, e.Cond, label_false+"f", label_true+"f")

	ctx.Printf("%s:\n", label_true)
	t()
	ctx.Printf("\tjmp %sf\n", label_done)
	ctx.Printf("%s:\n", label_false)
	f()
	ctx.Printf("%s:\n", label_done)
}

func (e *IfExpr) genCode(ctx *genCtx) {
	e.genCodeShared(ctx, func() {
		e.Then.genCode(ctx)
	}, func() {
		e.Else.genCode(ctx)
	})
}

func (e *IfExpr) genCodeRawInt(ctx *genCtx) {
	e.genCodeShared(ctx, func() {
		genCodeRawInt(ctx, e.Then)
	}, func() {
		genCodeRawInt(ctx, e.Else)
	})
}

func (e *IfExpr) genCodeJump(ctx *genCtx, l0, l1 string) {
	e.genCodeShared(ctx, func() {
		genCodeJump(ctx, e.Then, l0, l1)
	}, func() {
		genCodeJump(ctx, e.Else, l0, l1)
	})
}

func (e *IfExpr) genCodeUnused(ctx *genCtx) {
	e.genCodeShared(ctx, func() {
		genCodeUnused(ctx, e.Then)
	}, func() {
		genCodeUnused(ctx, e.Else)
	})
}

func (e *WhileExpr) genCollectLiterals(ctx *genCtx) {
	e.Cond.genCollectLiterals(ctx)
	e.Body.genCollectLiterals(ctx)
}

func (e *WhileExpr) genCountStack(ctx *genCtx) int {
	stack := e.Cond.genCountStack(ctx)
	if s := e.Body.genCountStack(ctx); s > stack {
		stack = s
	}
	return stack
}

func (e *WhileExpr) genCountVars(ctx *genCtx) int {
	vars := e.Cond.genCountVars(ctx)
	if v := e.Body.genCountVars(ctx); v > vars {
		vars = v
	}
	return vars
}

func (e *WhileExpr) genCode(ctx *genCtx) {
	e.genCodeUnused(ctx)
	ctx.Printf("\tlea unit_lit, %%eax\n")
}

func (e *WhileExpr) genCodeUnused(ctx *genCtx) {
	label_cond := ctx.Label()
	label_body := ctx.Label()
	label_done := ctx.Label()

	ctx.Printf("%s:\n", label_cond)
	genCodeJump(ctx, e.Cond, label_done+"f", label_body+"f")
	ctx.Printf("%s:\n", label_body)
	genCodeUnused(ctx, e.Body)
	ctx.Printf("\tjmp %sb\n", label_cond)
	ctx.Printf("%s:\n", label_done)
}

func (e *LessOrEqualExpr) genCollectLiterals(ctx *genCtx) {
	e.Left.genCollectLiterals(ctx)
	e.Right.genCollectLiterals(ctx)
}

func (e *LessOrEqualExpr) genCountStack(ctx *genCtx) int {
	stack := e.Left.genCountStack(ctx)
	if s := e.Right.genCountStack(ctx); s > stack {
		stack = s
	}
	return stack
}

func (e *LessOrEqualExpr) genCountVars(ctx *genCtx) int {
	vars := e.Left.genCountVars(ctx)
	if v := 1 + e.Right.genCountVars(ctx); v > vars {
		vars = v
	}
	return vars
}

func (e *LessOrEqualExpr) genCode(ctx *genCtx) {
	label_true := ctx.Label()
	label_false := ctx.Label()
	label_done := ctx.Label()

	e.genCodeJump(ctx, label_false+"f", label_true+"f")
	ctx.Printf("%s:\n", label_false)
	ctx.Printf("\tlea boolean_false, %%eax\n")
	ctx.Printf("\tjmp %sf\n", label_done)
	ctx.Printf("%s:\n", label_true)
	ctx.Printf("\tlea boolean_true, %%eax\n")
	ctx.Printf("%s:\n", label_done)
}

func (e *LessOrEqualExpr) genCodeJump(ctx *genCtx, l0, l1 string) {
	genCodeRawInt(ctx, e.Left)
	offset, unreserve := ctx.Slot()
	ctx.Printf("\tmovl %%eax, %d(%%ebp)\n", offset)
	genCodeRawInt(ctx, e.Right)
	ctx.Printf("\tmovl %d(%%ebp), %%ebx\n", offset)
	unreserve()
	ctx.Printf("\tcmpl %%eax, %%ebx\n")
	ctx.Printf("\tjle %s\n", l1)
	ctx.Printf("\tjmp %s\n", l0)
}

func (e *LessThanExpr) genCollectLiterals(ctx *genCtx) {
	e.Left.genCollectLiterals(ctx)
	e.Right.genCollectLiterals(ctx)
}

func (e *LessThanExpr) genCountStack(ctx *genCtx) int {
	stack := e.Left.genCountStack(ctx)
	if s := e.Right.genCountStack(ctx); s > stack {
		stack = s
	}
	return stack
}

func (e *LessThanExpr) genCountVars(ctx *genCtx) int {
	vars := e.Left.genCountVars(ctx)
	if v := 1 + e.Right.genCountVars(ctx); v > vars {
		vars = v
	}
	return vars
}

func (e *LessThanExpr) genCode(ctx *genCtx) {
	label_true := ctx.Label()
	label_false := ctx.Label()
	label_done := ctx.Label()

	e.genCodeJump(ctx, label_false+"f", label_true+"f")
	ctx.Printf("%s:\n", label_false)
	ctx.Printf("\tlea boolean_false, %%eax\n")
	ctx.Printf("\tjmp %sf\n", label_done)
	ctx.Printf("%s:\n", label_true)
	ctx.Printf("\tlea boolean_true, %%eax\n")
	ctx.Printf("%s:\n", label_done)
}

func (e *LessThanExpr) genCodeJump(ctx *genCtx, l0, l1 string) {
	genCodeRawInt(ctx, e.Left)
	offset, unreserve := ctx.Slot()
	ctx.Printf("\tmovl %%eax, %d(%%ebp)\n", offset)
	genCodeRawInt(ctx, e.Right)
	ctx.Printf("\tmovl %d(%%ebp), %%ebx\n", offset)
	unreserve()
	ctx.Printf("\tcmpl %%eax, %%ebx\n")
	ctx.Printf("\tjl %s\n", l1)
	ctx.Printf("\tjmp %s\n", l0)
}

func genArithmetic(ctx *genCtx, left, right Expr, compute func(), box bool) {
	genCodeRawInt(ctx, left)
	offset, unreserve := ctx.Slot()
	ctx.Printf("\tmovl %%eax, %d(%%ebp)\n", offset)
	genCodeRawInt(ctx, right)
	ctx.Printf("\tmovl %%eax, %%ecx\n")
	ctx.Printf("\tmovl %d(%%ebp), %%ebx\n", offset)
	compute()
	if box {
		ctx.Printf("\tmovl %%eax, %d(%%ebp)\n", offset)
		ctx.Printf("\tmovl $(size_of_Int + 4), %%eax\n")
		ctx.Printf("\tmovl $tag_of_Int, %%ebx\n")
		ctx.Printf("\tcall gc_alloc\n")
		ctx.Printf("\tmovl %d(%%ebp), %%ecx\n", offset)
		ctx.Printf("\tmovl %%ecx, offset_of_Int.value(%%eax)\n")
	}
	unreserve()
}

func (e *MultiplyExpr) genCollectLiterals(ctx *genCtx) {
	e.Left.genCollectLiterals(ctx)
	e.Right.genCollectLiterals(ctx)
}

func (e *MultiplyExpr) genCountStack(ctx *genCtx) int {
	stack := e.Left.genCountStack(ctx)
	if s := e.Right.genCountStack(ctx); s > stack {
		stack = s
	}
	return stack
}

func (e *MultiplyExpr) genCountVars(ctx *genCtx) int {
	vars := e.Left.genCountVars(ctx)
	if v := 1 + e.Right.genCountVars(ctx); v > vars {
		vars = v
	}
	return vars
}

func (e *MultiplyExpr) genCode(ctx *genCtx) {
	genArithmetic(ctx, e.Left, e.Right, func() {
		ctx.Printf("\timul %%ebx, %%ecx\n")
		ctx.Printf("\tmovl %%ecx, %%eax\n")
	}, true)
}

func (e *MultiplyExpr) genCodeRawInt(ctx *genCtx) {
	genArithmetic(ctx, e.Left, e.Right, func() {
		ctx.Printf("\timul %%ebx, %%ecx\n")
		ctx.Printf("\tmovl %%ecx, %%eax\n")
	}, false)
}

func (e *DivideExpr) genCollectLiterals(ctx *genCtx) {
	e.Left.genCollectLiterals(ctx)
	e.Right.genCollectLiterals(ctx)
}

func (e *DivideExpr) genCountStack(ctx *genCtx) int {
	stack := e.Left.genCountStack(ctx)
	if s := e.Right.genCountStack(ctx); s > stack {
		stack = s
	}
	return stack
}

func (e *DivideExpr) genCountVars(ctx *genCtx) int {
	vars := e.Left.genCountVars(ctx)
	if v := 1 + e.Right.genCountVars(ctx); v > vars {
		vars = v
	}
	return vars
}

func (e *DivideExpr) genCode(ctx *genCtx) {
	genArithmetic(ctx, e.Left, e.Right, func() {
		ctx.Printf("\tmovl %%ebx, %%eax\n")
		ctx.Printf("\tcdq\n")
		ctx.Printf("\tidiv %%ecx\n")
	}, true)
}

func (e *DivideExpr) genCodeRawInt(ctx *genCtx) {
	genArithmetic(ctx, e.Left, e.Right, func() {
		ctx.Printf("\tmovl %%ebx, %%eax\n")
		ctx.Printf("\tcdq\n")
		ctx.Printf("\tidiv %%ecx\n")
	}, false)
}

func (e *AddExpr) genCollectLiterals(ctx *genCtx) {
	e.Left.genCollectLiterals(ctx)
	e.Right.genCollectLiterals(ctx)
}

func (e *AddExpr) genCountStack(ctx *genCtx) int {
	stack := e.Left.genCountStack(ctx)
	if s := e.Right.genCountStack(ctx); s > stack {
		stack = s
	}
	return stack
}

func (e *AddExpr) genCountVars(ctx *genCtx) int {
	vars := e.Left.genCountVars(ctx)
	if v := 1 + e.Right.genCountVars(ctx); v > vars {
		vars = v
	}
	return vars
}

func (e *AddExpr) genCode(ctx *genCtx) {
	genArithmetic(ctx, e.Left, e.Right, func() {
		ctx.Printf("\taddl %%ebx, %%ecx\n")
		ctx.Printf("\tmovl %%ecx, %%eax\n")
	}, true)
}

func (e *AddExpr) genCodeRawInt(ctx *genCtx) {
	genArithmetic(ctx, e.Left, e.Right, func() {
		ctx.Printf("\taddl %%ebx, %%ecx\n")
		ctx.Printf("\tmovl %%ecx, %%eax\n")
	}, false)
}

func (e *SubtractExpr) genCollectLiterals(ctx *genCtx) {
	e.Left.genCollectLiterals(ctx)
	e.Right.genCollectLiterals(ctx)
}

func (e *SubtractExpr) genCountStack(ctx *genCtx) int {
	stack := e.Left.genCountStack(ctx)
	if s := e.Right.genCountStack(ctx); s > stack {
		stack = s
	}
	return stack
}

func (e *SubtractExpr) genCountVars(ctx *genCtx) int {
	vars := e.Left.genCountVars(ctx)
	if v := 1 + e.Right.genCountVars(ctx); v > vars {
		vars = v
	}
	return vars
}

func (e *SubtractExpr) genCode(ctx *genCtx) {
	genArithmetic(ctx, e.Left, e.Right, func() {
		ctx.Printf("\tsubl %%ecx, %%ebx\n")
		ctx.Printf("\tmovl %%ebx, %%eax\n")
	}, true)
}

func (e *SubtractExpr) genCodeRawInt(ctx *genCtx) {
	genArithmetic(ctx, e.Left, e.Right, func() {
		ctx.Printf("\tsubl %%ecx, %%ebx\n")
		ctx.Printf("\tmovl %%ebx, %%eax\n")
	}, false)
}

func (e *MatchExpr) genCollectLiterals(ctx *genCtx) {
	e.Left.genCollectLiterals(ctx)
	for _, c := range e.Cases {
		c.Body.genCollectLiterals(ctx)
	}
}

func (e *MatchExpr) genCountStack(ctx *genCtx) int {
	stack := e.Left.genCountStack(ctx)
	for _, c := range e.Cases {
		if s := c.Body.genCountStack(ctx); s > stack {
			stack = s
		}
	}
	return stack
}

func (e *MatchExpr) genCountVars(ctx *genCtx) int {
	vars := e.Left.genCountVars(ctx)
	for _, c := range e.Cases {
		if v := c.Body.genCountVars(ctx); v > vars {
			vars = v
		}
	}
	return vars + 1
}

func (e *MatchExpr) genCode(ctx *genCtx) {
	label_null := ctx.Label()
	label_done := ctx.Label()

	e.Left.genCode(ctx)
	offset, unreserve := ctx.Slot()
	e.Offset = offset
	ctx.Printf("\tmovl %%eax, %d(%%ebp)\n", offset)
	ctx.Printf("\ttest %%eax, %%eax\n")
	ctx.Printf("\tjz %sf\n", label_null)
	ctx.Printf("\tmovl tag_offset(%%eax), %%eax\n")
	ctx.Printf("%s:\n", label_null)

	labels := make([]string, len(e.Cases))
	for i := range labels {
		labels[i] = ctx.Label()
	}

	for i, c := range e.Cases {
		if c.Type.Class.Order == c.Type.Class.MaxOrder {
			ctx.Printf("\tcmpl $%d, %%eax\n", c.Type.Class.Order)
			ctx.Printf("\tje %sf\n", labels[i])
		} else {
			label_skip := ctx.Label()
			ctx.Printf("\tcmpl $%d, %%eax\n", c.Type.Class.Order)
			ctx.Printf("\tjl %sf\n", label_skip)
			ctx.Printf("\tcmpl $%d, %%eax\n", c.Type.Class.MaxOrder)
			ctx.Printf("\tjle %sf\n", labels[i])
			ctx.Printf("%s:\n", label_skip)
		}
	}
	ctx.Printf("\tjmp runtime.case_panic\n")

	for i, c := range e.Cases {
		ctx.Printf("%s:\n", labels[i])
		c.genCode(ctx)
		ctx.Printf("\tjmp %sf\n", label_done)
	}

	ctx.Printf("%s:\n", label_done)
	ctx.Printf("\tmovl %d(%%ebp), %%ebx\n", offset)
	genGC(ctx, "%ebx")
	unreserve()
}

func (e *MatchExpr) genCodeRawInt(ctx *genCtx) {
	label_null := ctx.Label()
	label_done := ctx.Label()

	e.Left.genCode(ctx)
	offset, unreserve := ctx.Slot()
	e.Offset = offset
	ctx.Printf("\tmovl %%eax, %d(%%ebp)\n", offset)
	ctx.Printf("\ttest %%eax, %%eax\n")
	ctx.Printf("\tjz %sf\n", label_null)
	ctx.Printf("\tmovl tag_offset(%%eax), %%eax\n")
	ctx.Printf("%s:\n", label_null)

	labels := make([]string, len(e.Cases))
	for i := range labels {
		labels[i] = ctx.Label()
	}

	for i, c := range e.Cases {
		if c.Type.Class.Order == c.Type.Class.MaxOrder {
			ctx.Printf("\tcmpl $%d, %%eax\n", c.Type.Class.Order)
			ctx.Printf("\tje %sf\n", labels[i])
		} else {
			label_skip := ctx.Label()
			ctx.Printf("\tcmpl $%d, %%eax\n", c.Type.Class.Order)
			ctx.Printf("\tjl %sf\n", label_skip)
			ctx.Printf("\tcmpl $%d, %%eax\n", c.Type.Class.MaxOrder)
			ctx.Printf("\tjle %sf\n", labels[i])
			ctx.Printf("%s:\n", label_skip)
		}
	}
	ctx.Printf("\tjmp runtime.case_panic\n")

	for i, c := range e.Cases {
		ctx.Printf("%s:\n", labels[i])
		c.genCodeRawInt(ctx)
		ctx.Printf("\tjmp %sf\n", label_done)
	}

	ctx.Printf("%s:\n", label_done)
	ctx.Printf("\tmovl %d(%%ebp), %%ebx\n", offset)
	genGC(ctx, "%ebx")
	unreserve()
}

func (e *MatchExpr) genCodeJump(ctx *genCtx, l0, l1 string) {
	label_null := ctx.Label()
	label_true := ctx.Label()
	label_false := ctx.Label()

	e.Left.genCode(ctx)
	offset, unreserve := ctx.Slot()
	e.Offset = offset
	ctx.Printf("\tmovl %%eax, %d(%%ebp)\n", offset)
	ctx.Printf("\ttest %%eax, %%eax\n")
	ctx.Printf("\tjz %sf\n", label_null)
	ctx.Printf("\tmovl tag_offset(%%eax), %%eax\n")
	ctx.Printf("%s:\n", label_null)

	labels := make([]string, len(e.Cases))
	for i := range labels {
		labels[i] = ctx.Label()
	}

	for i, c := range e.Cases {
		if c.Type.Class.Order == c.Type.Class.MaxOrder {
			ctx.Printf("\tcmpl $%d, %%eax\n", c.Type.Class.Order)
			ctx.Printf("\tje %sf\n", labels[i])
		} else {
			label_skip := ctx.Label()
			ctx.Printf("\tcmpl $%d, %%eax\n", c.Type.Class.Order)
			ctx.Printf("\tjl %sf\n", label_skip)
			ctx.Printf("\tcmpl $%d, %%eax\n", c.Type.Class.MaxOrder)
			ctx.Printf("\tjle %sf\n", labels[i])
			ctx.Printf("%s:\n", label_skip)
		}
	}
	ctx.Printf("\tjmp runtime.case_panic\n")

	for i, c := range e.Cases {
		ctx.Printf("%s:\n", labels[i])
		c.genCodeJump(ctx, label_false+"f", label_true+"f")
	}

	ctx.Printf("%s:\n", label_false)
	ctx.Printf("\tmovl %d(%%ebp), %%ebx\n", offset)
	genGC(ctx, "%ebx")
	ctx.Printf("\tjmp %s\n", l0)

	ctx.Printf("%s:\n", label_true)
	ctx.Printf("\tmovl %d(%%ebp), %%ebx\n", offset)
	genGC(ctx, "%ebx")
	ctx.Printf("\tjmp %s\n", l1)

	unreserve()
}

func (e *MatchExpr) genCodeUnused(ctx *genCtx) {
	label_null := ctx.Label()
	label_done := ctx.Label()

	e.Left.genCode(ctx)
	offset, unreserve := ctx.Slot()
	e.Offset = offset
	ctx.Printf("\tmovl %%eax, %d(%%ebp)\n", offset)
	ctx.Printf("\ttest %%eax, %%eax\n")
	ctx.Printf("\tjz %sf\n", label_null)
	ctx.Printf("\tmovl tag_offset(%%eax), %%eax\n")
	ctx.Printf("%s:\n", label_null)

	labels := make([]string, len(e.Cases))
	for i := range labels {
		labels[i] = ctx.Label()
	}

	for i, c := range e.Cases {
		if c.Type.Class.Order == c.Type.Class.MaxOrder {
			ctx.Printf("\tcmpl $%d, %%eax\n", c.Type.Class.Order)
			ctx.Printf("\tje %sf\n", labels[i])
		} else {
			label_skip := ctx.Label()
			ctx.Printf("\tcmpl $%d, %%eax\n", c.Type.Class.Order)
			ctx.Printf("\tjl %sf\n", label_skip)
			ctx.Printf("\tcmpl $%d, %%eax\n", c.Type.Class.MaxOrder)
			ctx.Printf("\tjle %sf\n", labels[i])
			ctx.Printf("%s:\n", label_skip)
		}
	}
	ctx.Printf("\tjmp runtime.case_panic\n")

	for i, c := range e.Cases {
		ctx.Printf("%s:\n", labels[i])
		c.genCodeUnused(ctx)
		ctx.Printf("\tjmp %sf\n", label_done)
	}

	ctx.Printf("%s:\n", label_done)
	ctx.Printf("\tmovl %d(%%ebp), %%ebx\n", offset)
	genGC(ctx, "%ebx")
	unreserve()
}

func (e *DynamicCallExpr) genCollectLiterals(ctx *genCtx) {
	e.Recv.genCollectLiterals(ctx)
	for _, a := range e.Args {
		a.genCollectLiterals(ctx)
	}
}

func (e *DynamicCallExpr) genCountStack(ctx *genCtx) int {
	stack := e.Recv.genCountStack(ctx)
	for i, a := range e.Args {
		if s := 1 + i + a.genCountStack(ctx); s > stack {
			stack = s
		}
	}
	if s := 1 + len(e.Args) + 1; s > stack {
		stack = s
	}
	return stack
}

func (e *DynamicCallExpr) genCountVars(ctx *genCtx) int {
	vars := e.Recv.genCountVars(ctx)
	for _, a := range e.Args {
		if v := a.genCountVars(ctx); v > vars {
			vars = v
		}
	}
	return vars
}

func (e *DynamicCallExpr) genCode(ctx *genCtx) {
	e.Recv.genCode(ctx)
	ctx.Printf("\ttest %%eax, %%eax\n")
	ctx.Printf("\tjz runtime.null_panic\n")
	ctx.Printf("\tpush %%eax\n")
	for _, a := range e.Args {
		a.genCode(ctx)
		ctx.Printf("\tpush %%eax\n")
	}
	if e.HasOverride {
		ctx.Printf("\tmovl %d(%%esp), %%eax\n", len(e.Args)*4)
		ctx.Printf("\tmovl tag_offset(%%eax), %%eax\n")
		ctx.Printf("\tshll $2, %%eax\n")
		ctx.Printf("\tmovl method_tables(%%eax), %%eax\n")
		ctx.Printf("\tmovl method_offset_%s.%s(%%eax), %%eax\n", e.Name.Method.Parent.Type.Name, e.Name.Method.Name.Name)
		ctx.Printf("\tcall *%%eax\n")
	} else {
		ctx.Printf("\tcall %s.%s\n", e.Name.Method.Parent.Type.Name, e.Name.Method.Name.Name)
	}
}

func (e *SuperCallExpr) genCollectLiterals(ctx *genCtx) {
	for _, a := range e.Args {
		a.genCollectLiterals(ctx)
	}
}

func (e *SuperCallExpr) genCountStack(ctx *genCtx) int {
	stack := 1
	for i, a := range e.Args {
		if s := 1 + i + a.genCountStack(ctx); s > stack {
			stack = s
		}
	}
	if s := 1 + len(e.Args) + 1; s > stack {
		stack = s
	}
	return stack
}

func (e *SuperCallExpr) genCountVars(ctx *genCtx) int {
	vars := 0
	for _, a := range e.Args {
		if v := a.genCountVars(ctx); v > vars {
			vars = v
		}
	}
	return vars
}

func (e *SuperCallExpr) genCode(ctx *genCtx) {
	ctx.Printf("\tmovl %d(%%ebp), %%eax\n", ctx.this)
	genRef(ctx, "%eax")
	ctx.Printf("\tpush %%eax\n")
	for _, a := range e.Args {
		a.genCode(ctx)
		ctx.Printf("\tpush %%eax\n")
	}
	ctx.Printf("\tcall %s.%s\n", e.Name.Method.Parent.Type.Name, e.Name.Method.Name.Name)
}

func (e *StaticCallExpr) genCollectLiterals(ctx *genCtx) {
	e.Recv.genCollectLiterals(ctx)
	for _, a := range e.Args {
		a.genCollectLiterals(ctx)
	}
}

func (e *StaticCallExpr) genCountStack(ctx *genCtx) int {
	stack := e.Recv.genCountStack(ctx)
	for i, a := range e.Args {
		if s := 1 + i + a.genCountStack(ctx); s > stack {
			stack = s
		}
	}
	if s := 1 + len(e.Args) + 1; s > stack {
		stack = s
	}
	return stack
}

func (e *StaticCallExpr) genCountVars(ctx *genCtx) int {
	vars := e.Recv.genCountVars(ctx)
	for _, a := range e.Args {
		if v := a.genCountVars(ctx); v > vars {
			vars = v
		}
	}
	return vars
}

func (e *StaticCallExpr) genCode(ctx *genCtx) {
	e.Recv.genCode(ctx)
	ctx.Printf("\tpush %%eax\n")
	for _, a := range e.Args {
		a.genCode(ctx)
		ctx.Printf("\tpush %%eax\n")
	}
	ctx.Printf("\tcall %s.%s\n", e.Name.Method.Parent.Type.Name, e.Name.Method.Name.Name)
}

func (e *AllocExpr) genCollectLiterals(ctx *genCtx) {
}

func (e *AllocExpr) genCountStack(ctx *genCtx) int {
	return 0
}

func (e *AllocExpr) genCountVars(ctx *genCtx) int {
	return 0
}

func (e *AllocExpr) genCode(ctx *genCtx) {
	ctx.Printf("\tmovl $size_of_%s, %%eax\n", e.Type.Name)
	ctx.Printf("\tmovl $tag_of_%s, %%ebx\n", e.Type.Name)
	ctx.Printf("\tcall gc_alloc\n")
	var gen func(c *Class)
	gen = func(c *Class) {
		if c == nativeClass {
			return
		}
		gen(c.Extends.Type.Class)
		for _, f := range c.Features {
			if a, ok := f.(*Attribute); ok {
				var sym string
				switch a.Type.Name {
				case "Int":
					sym = "int_lit_0"
				case "Boolean":
					sym = "boolean_false"
				case "Unit":
					sym = "unit_lit"
				default:
					continue
				}
				ctx.Printf("\tmovl $%s, offset_of_%s.%s(%%eax)\n", sym, c.Type.Name, a.Name.Name)
			}
		}
	}
	gen(e.Type.Class)
}

func (e *AssignExpr) genCollectLiterals(ctx *genCtx) {
	e.Expr.genCollectLiterals(ctx)
}

func (e *AssignExpr) genCountStack(ctx *genCtx) int {
	return e.Expr.genCountVars(ctx)
}

func (e *AssignExpr) genCountVars(ctx *genCtx) int {
	return e.Expr.genCountVars(ctx)
}

func (e *AssignExpr) genCode(ctx *genCtx) {
	e.genCodeUnused(ctx)
	ctx.Printf("\tleal unit_lit, %%eax\n")
}

func (e *AssignExpr) genCodeUnused(ctx *genCtx) {
	var rawInt bool
	if e.Name.Object.RawInt() {
		if raw, ok := e.Expr.(ArithmeticExpr); ok {
			rawInt = true
			raw.genCodeRawInt(ctx)
		}
	}
	if !rawInt {
		e.Expr.genCode(ctx)
	}
	ctx.Printf("\tmovl %s, %%edx\n", e.Name.Object.Base(ctx.this))
	if e.Name.Object.Stack() {
		if !e.Name.Object.RawInt() {
			ctx.Printf("\tmovl %s(%%edx), %%ebx\n", e.Name.Object.Offs())
			genGC(ctx, "%ebx")
		}
	} else if !rawInt {
		genGC(ctx, "%eax")
	}
	if e.Name.Object.RawInt() && !rawInt {
		if e.Name.Object.Stack() {
			genGC(ctx, "%eax")
		}
		ctx.Printf("\tmovl offset_of_Int.value(%%eax), %%eax\n")
	}
	ctx.Printf("\tmovl %%eax, %s(%%edx)\n", e.Name.Object.Offs())
}

func (e *VarExpr) genCollectLiterals(ctx *genCtx) {
	e.Init.genCollectLiterals(ctx)
	e.Body.genCollectLiterals(ctx)
}

func (e *VarExpr) genCountStack(ctx *genCtx) int {
	stack := e.Init.genCountStack(ctx)
	if s := e.Body.genCountStack(ctx); s > stack {
		stack = s
	}
	return stack
}

func (e *VarExpr) genCountVars(ctx *genCtx) int {
	vars := e.Init.genCountVars(ctx)
	if v := e.Body.genCountVars(ctx); v > vars {
		vars = v
	}
	return vars + 1
}

func (e *VarExpr) genCodeShared(ctx *genCtx, body func()) {
	var rawInt bool
	if e.RawInt() {
		if raw, ok := e.Init.(ArithmeticExpr); ok {
			rawInt = true
			raw.genCodeRawInt(ctx)
		}
	}
	if !rawInt {
		e.Init.genCode(ctx)
		if e.RawInt() {
			genGC(ctx, "%eax")
			ctx.Printf("\tmovl offset_of_Int.value(%%eax), %%eax\n")
		}
	}
	offset, unreserve := ctx.Slot()
	e.Offset = offset
	ctx.Printf("\tmovl %%eax, %d(%%ebp)\n", offset)
	body()
	ctx.Printf("\tmovl %d(%%ebp), %%ebx\n", offset)
	if !e.RawInt() {
		genGC(ctx, "%ebx")
	}
	unreserve()
}

func (e *VarExpr) genCode(ctx *genCtx) {
	e.genCodeShared(ctx, func() {
		e.Body.genCode(ctx)
	})
}

func (e *VarExpr) genCodeRawInt(ctx *genCtx) {
	e.genCodeShared(ctx, func() {
		genCodeRawInt(ctx, e.Body)
	})
}

func (e *VarExpr) genCodeUnused(ctx *genCtx) {
	e.genCodeShared(ctx, func() {
		genCodeUnused(ctx, e.Body)
	})
}

func (e *ChainExpr) genCollectLiterals(ctx *genCtx) {
	e.Pre.genCollectLiterals(ctx)
	e.Expr.genCollectLiterals(ctx)
}

func (e *ChainExpr) genCountStack(ctx *genCtx) int {
	stack := e.Pre.genCountStack(ctx)
	if s := e.Expr.genCountStack(ctx); s > stack {
		stack = s
	}
	return stack
}

func (e *ChainExpr) genCountVars(ctx *genCtx) int {
	vars := e.Pre.genCountVars(ctx)
	if v := e.Expr.genCountVars(ctx); v > vars {
		vars = v
	}
	return vars
}

func (e *ChainExpr) genCode(ctx *genCtx) {
	genCodeUnused(ctx, e.Pre)
	e.Expr.genCode(ctx)
}

func (e *ChainExpr) genCodeRawInt(ctx *genCtx) {
	genCodeUnused(ctx, e.Pre)
	genCodeRawInt(ctx, e.Expr)
}

func (e *ChainExpr) genCodeJump(ctx *genCtx, l0, l1 string) {
	genCodeUnused(ctx, e.Pre)
	genCodeJump(ctx, e.Expr, l0, l1)
}

func (e *ChainExpr) genCodeUnused(ctx *genCtx) {
	genCodeUnused(ctx, e.Pre)
	genCodeUnused(ctx, e.Expr)
}

func (e *ThisExpr) genCollectLiterals(ctx *genCtx) {
}

func (e *ThisExpr) genCountStack(ctx *genCtx) int {
	return 0
}

func (e *ThisExpr) genCountVars(ctx *genCtx) int {
	return 0
}

func (e *ThisExpr) genCode(ctx *genCtx) {
	ctx.Printf("\tmovl %d(%%ebp), %%eax\n", ctx.this)
	genRef(ctx, "%eax")
}

func (e *ThisExpr) genCodeUnused(ctx *genCtx) {
}

func (e *NullExpr) genCollectLiterals(ctx *genCtx) {
}

func (e *NullExpr) genCountStack(ctx *genCtx) int {
	return 0
}

func (e *NullExpr) genCountVars(ctx *genCtx) int {
	return 0
}

func (e *NullExpr) genCode(ctx *genCtx) {
	ctx.Printf("\tmovl $0, %%eax\n")
}

func (e *NullExpr) genCodeUnused(ctx *genCtx) {
}

func (e *UnitExpr) genCollectLiterals(ctx *genCtx) {
}

func (e *UnitExpr) genCountStack(ctx *genCtx) int {
	return 0
}

func (e *UnitExpr) genCountVars(ctx *genCtx) int {
	return 0
}

func (e *UnitExpr) genCode(ctx *genCtx) {
	ctx.Printf("\tleal unit_lit, %%eax\n")
}

func (e *UnitExpr) genCodeUnused(ctx *genCtx) {
}

func (e *NameExpr) genCollectLiterals(ctx *genCtx) {
}

func (e *NameExpr) genCountStack(ctx *genCtx) int {
	return 0
}

func (e *NameExpr) genCountVars(ctx *genCtx) int {
	return 0
}

func (e *NameExpr) genCode(ctx *genCtx) {
	if e.Name.Object.RawInt() {
		ctx.Printf("\tmovl $(size_of_Int + 4), %%eax\n")
		ctx.Printf("\tmovl $tag_of_Int, %%ebx\n")
		ctx.Printf("\tcall gc_alloc\n")
		ctx.Printf("\tmovl %s, %%edx\n", e.Name.Object.Base(ctx.this))
		ctx.Printf("\tmovl %s(%%edx), %%ebx\n", e.Name.Object.Offs())
		ctx.Printf("\tmovl %%ebx, offset_of_Int.value(%%eax)\n")
	} else {
		ctx.Printf("\tmovl %s, %%edx\n", e.Name.Object.Base(ctx.this))
		ctx.Printf("\tmovl %s(%%edx), %%eax\n", e.Name.Object.Offs())
		genRef(ctx, "%eax")
	}
}

func (e *NameExpr) genCodeRawInt(ctx *genCtx) {
	ctx.Printf("\tmovl %s, %%edx\n", e.Name.Object.Base(ctx.this))
	ctx.Printf("\tmovl %s(%%edx), %%eax\n", e.Name.Object.Offs())
	if !e.Name.Object.RawInt() {
		ctx.Printf("\tmovl offset_of_Int.value(%%eax), %%eax\n")
	}
}

func (e *NameExpr) genCodeUnused(ctx *genCtx) {
}

func (e *StringExpr) genCollectLiterals(ctx *genCtx) {
	e.Lit.LitID = ctx.AddString(e.Lit.Str)
}

func (e *StringExpr) genCountStack(ctx *genCtx) int {
	return 0
}

func (e *StringExpr) genCountVars(ctx *genCtx) int {
	return 0
}

func (e *StringExpr) genCode(ctx *genCtx) {
	ctx.Printf("\tleal string_lit_%d, %%eax\n", e.Lit.LitID)
}

func (e *StringExpr) genCodeUnused(ctx *genCtx) {
}

func (e *BoolExpr) genCollectLiterals(ctx *genCtx) {
}

func (e *BoolExpr) genCountStack(ctx *genCtx) int {
	return 0
}

func (e *BoolExpr) genCountVars(ctx *genCtx) int {
	return 0
}

func (e *BoolExpr) genCode(ctx *genCtx) {
	if e.Lit.Bool {
		ctx.Printf("\tleal boolean_true, %%eax\n")
	} else {
		ctx.Printf("\tleal boolean_false, %%eax\n")
	}
}

func (e *BoolExpr) genCodeJump(ctx *genCtx, l0, l1 string) {
	if e.Lit.Bool {
		ctx.Printf("\tjmp %s\n", l1)
	} else {
		ctx.Printf("\tjmp %s\n", l0)
	}
}

func (e *BoolExpr) genCodeUnused(ctx *genCtx) {
}

func (e *IntExpr) genCollectLiterals(ctx *genCtx) {
	e.Lit.LitID = ctx.AddInt(e.Lit.Int)
}

func (e *IntExpr) genCountStack(ctx *genCtx) int {
	return 0
}

func (e *IntExpr) genCountVars(ctx *genCtx) int {
	return 0
}

func (e *IntExpr) genCode(ctx *genCtx) {
	ctx.Printf("\tleal int_lit_%d, %%eax\n", e.Lit.LitID)
}

func (e *IntExpr) genCodeRawInt(ctx *genCtx) {
	ctx.Printf("\tmovl $%d, %%eax\n", e.Lit.Int)
}

func (e *IntExpr) genCodeUnused(ctx *genCtx) {
}

func (e *NativeExpr) genCollectLiterals(ctx *genCtx) {
}

func (e *NativeExpr) genCountStack(ctx *genCtx) int {
	panic("NativeExpr.genCountStack should never be called")
}

func (e *NativeExpr) genCountVars(ctx *genCtx) int {
	panic("NativeExpr.genCountVars should never be called")
}

func (e *NativeExpr) genCode(ctx *genCtx) {
	panic("NativeExpr.genCode should never be called")
}

func (c *Case) genCode(ctx *genCtx) {
	c.Body.genCode(ctx)
}

func (c *Case) genCodeRawInt(ctx *genCtx) {
	genCodeRawInt(ctx, c.Body)
}

func (c *Case) genCodeJump(ctx *genCtx, l0, l1 string) {
	genCodeJump(ctx, c.Body, l0, l1)
}

func (c *Case) genCodeUnused(ctx *genCtx) {
	genCodeUnused(ctx, c.Body)
}
