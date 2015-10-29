package ast

import (
	"fmt"
	"io"
)

func (p *Program) CodeGen(w io.Writer) {
	var ints []int32
	var strings []string
	var stringLengths []int
	addInt := func(x int32) int {
		for i, y := range ints {
			if x == y {
				return i
			}
		}
		ints = append(ints, x)
		return len(ints) - 1
	}
	addInt(0) // int_lit_0 must be 0
	addString := func(x string) int {
		for i, y := range strings {
			if x == y {
				return i
			}
		}
		strings = append(strings, x)
		stringLengths = append(stringLengths, addInt(int32(len(x))))
		return len(strings) - 1
	}
	var byteIntIDs [256]int
	for i := range byteIntIDs {
		byteIntIDs[i] = addInt(int32(i))
	}
	nullClassID := addString("Null")
	p.genCollectLiterals(addInt, addString)

	fmt.Fprintf(w, ".include \"basic_defs.s\"\n")
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, ".data\n")
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, ".globl boolean_false\n")
	fmt.Fprintf(w, ".align 2\n")
	fmt.Fprintf(w, "boolean_false:\n")
	fmt.Fprintf(w, "\t.long tag_of_Boolean\n")
	fmt.Fprintf(w, "\t.long size_of_Boolean\n")
	fmt.Fprintf(w, "\t.long gc_tag_root\n")
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, ".globl boolean_true\n")
	fmt.Fprintf(w, ".align 2\n")
	fmt.Fprintf(w, "boolean_true:\n")
	fmt.Fprintf(w, "\t.long tag_of_Boolean\n")
	fmt.Fprintf(w, "\t.long size_of_Boolean\n")
	fmt.Fprintf(w, "\t.long gc_tag_root\n")
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, ".globl unit_lit\n")
	fmt.Fprintf(w, ".align 2\n")
	fmt.Fprintf(w, "unit_lit:\n")
	fmt.Fprintf(w, "\t.long tag_of_Unit\n")
	fmt.Fprintf(w, "\t.long size_of_Unit\n")
	fmt.Fprintf(w, "\t.long gc_tag_root\n")
	fmt.Fprintf(w, "\n")
	for i, x := range ints {
		fmt.Fprintf(w, ".align 2\n")
		fmt.Fprintf(w, "int_lit_%d:\n", i)
		fmt.Fprintf(w, "\t.long tag_of_Int\n")
		fmt.Fprintf(w, "\t.long size_of_Int + 4\n")
		fmt.Fprintf(w, "\t.long gc_tag_root\n")
		fmt.Fprintf(w, "\t.long %d\n", x)
		fmt.Fprintf(w, "\n")
	}

	for i, x := range strings {
		fmt.Fprintf(w, ".align 2\n")
		fmt.Fprintf(w, "string_lit_%d:\n", i)
		fmt.Fprintf(w, "\t.long tag_of_String\n")
		fmt.Fprintf(w, "\t.long size_of_String + %d\n", len(x))
		fmt.Fprintf(w, "\t.long gc_tag_root\n")
		fmt.Fprintf(w, "\t.long int_lit_%d\n", stringLengths[i])
		for j := 0; j < len(x); j++ {
			fmt.Fprintf(w, "\t.byte %d\n", x[j])
		}
		fmt.Fprintf(w, "\n")
	}

	fmt.Fprintf(w, ".globl byte_ints\n")
	fmt.Fprintf(w, ".align 2\n")
	fmt.Fprintf(w, "byte_ints:\n")
	for _, x := range byteIntIDs {
		fmt.Fprintf(w, "\t.long int_lit_%d\n", x)
	}
	fmt.Fprintf(w, "\n")

	fmt.Fprintf(w, ".globl max_tag\n")
	fmt.Fprintf(w, ".set max_tag, %d\n", len(p.Ordered))
	fmt.Fprintf(w, "\t")

	fmt.Fprintf(w, ".globl gc_sizes\n")
	fmt.Fprintf(w, ".align 2\n")
	fmt.Fprintf(w, "gc_sizes:\n")
	fmt.Fprintf(w, "\t.long 0\n")
	for _, c := range p.Ordered {
		fmt.Fprintf(w, "\t.long size_of_%s / 4\n", c.Type.Name)
	}
	fmt.Fprintf(w, "\n")

	for _, c := range p.Ordered {
		fmt.Fprintf(w, ".align 2\n")
		fmt.Fprintf(w, "methods_of_%s:\n", c.Type.Name)
		for _, m := range c.Methods {
			fmt.Fprintf(w, "\t.long %s.%s\n", m.Parent.Type.Name, m.Name.Name)
		}
		fmt.Fprintf(w, "\n")
	}

	fmt.Fprintf(w, ".globl class_names\n")
	fmt.Fprintf(w, ".align 2\n")
	fmt.Fprintf(w, "class_names:\n")
	fmt.Fprintf(w, "\t.long string_lit_%d\n", nullClassID)
	for _, c := range p.Ordered {
		fmt.Fprintf(w, "\t.long string_lit_%d\n", c.NameID)
	}
	fmt.Fprintf(w, "\n")

	fmt.Fprintf(w, ".globl method_tables\n")
	fmt.Fprintf(w, ".align 2\n")
	fmt.Fprintf(w, "method_tables:\n")
	fmt.Fprintf(w, "\t.long 0\n")
	for _, c := range p.Ordered {
		fmt.Fprintf(w, "\t.long methods_of_%s\n", c.Type.Name)
	}
	fmt.Fprintf(w, "\n")

	for _, c := range p.Ordered {
		c.genConstants(w)
	}
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, ".text\n")

	genMethod(w, "main", 0, p.Main, false)

	for _, c := range p.Ordered {
		c.genCode(w)
	}
}

func (p *Program) genCollectLiterals(ints func(int32) int, strings func(string) int) {
	for _, c := range p.Classes {
		c.genCollectLiterals(ints, strings)
	}
}

func (c *Class) genCollectLiterals(ints func(int32) int, strings func(string) int) {
	c.NameID = strings(c.Type.Name)

	for _, f := range c.Features {
		if m, ok := f.(*Method); ok {
			m.Body.genCollectLiterals(ints, strings)
		}
	}
}

func (c *Class) genConstants(w io.Writer) {
	fmt.Fprintf(w, ".globl tag_of_%s\n", c.Type.Name)
	fmt.Fprintf(w, ".set tag_of_%s, %d\n", c.Type.Name, c.Order)
	c.Size = c.Extends.Type.Class.Size
	for _, f := range c.Features {
		if a, ok := f.(*Attribute); ok {
			fmt.Fprintf(w, ".globl offset_of_%s.%s\n", c.Type.Name, a.Name.Name)
			fmt.Fprintf(w, ".set offset_of_%s.%s, data_offset + %d\n", c.Type.Name, a.Name.Name, c.Size)
			if _, ok := a.Init.(*NativeExpr); !ok {
				c.Size += 4
			}
		}
	}
	fmt.Fprintf(w, ".globl size_of_%s\n", c.Type.Name)
	fmt.Fprintf(w, ".set size_of_%s, %d\n", c.Type.Name, c.Size)
}

func genMethod(w io.Writer, name string, args int, body Expr, hasArgs bool) {
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, ".globl %s\n", name)
	fmt.Fprintf(w, "%s:\n", name)
	vars := body.genCountVars(args*4 + 8)
	fmt.Fprintf(w, "\tenter $%d, $0\n", vars*4)
	varsUsed := 0
	label := 0
	mkLabel := func() string {
		label++
		return fmt.Sprintf("%d", label)
	}
	body.genCode(w, mkLabel, func() (int, func()) {
		if vars == varsUsed {
			panic("INTERNAL ERROR: too many vars")
		}
		varsUsed++
		n := varsUsed
		return -n * 4, func() {
			if varsUsed != n {
				panic("INTERNAL ERROR: missed var release")
			}
			varsUsed--
		}
	})
	if hasArgs {
		for i := 0; i <= args; i++ {
			fmt.Fprintf(w, "\tmovl %d(%%ebp), %%ebx\n", i*4+8)
			genGC(w, "%ebx", mkLabel)
		}
	}
	fmt.Fprintf(w, "\tleave\n")
	fmt.Fprintf(w, "\tret $%d\n", args*4+4)
}

func genRef(w io.Writer, reg string, label func() string) {
	label_done := label()

	fmt.Fprintf(w, "\ttest %s, %s\n", reg, reg)
	fmt.Fprintf(w, "\tjz %sf\n", label_done)
	fmt.Fprintf(w, "\tcmpl $0, gc_offset(%s)\n", reg)
	fmt.Fprintf(w, "\tjl %sf\n", label_done)
	fmt.Fprintf(w, "\tincl gc_offset(%s)\n", reg)
	fmt.Fprintf(w, "%s:\n", label_done)
}

func genGC(w io.Writer, reg string, label func() string) {
	label_done := label()

	fmt.Fprintf(w, "\ttest %s, %s\n", reg, reg)
	fmt.Fprintf(w, "\tjz %sf\n", label_done)
	fmt.Fprintf(w, "\tcmpl $0, gc_offset(%s)\n", reg)
	fmt.Fprintf(w, "\tjle %sf\n", label_done)
	fmt.Fprintf(w, "\tdecl gc_offset(%s)\n", reg)
	fmt.Fprintf(w, "%s:\n", label_done)
}

func (c *Class) genCode(w io.Writer) {
	for _, f := range c.Features {
		if m, ok := f.(*Method); ok {
			if _, ok := m.Body.(*NativeExpr); ok {
				continue
			}

			for i, a := range m.Args {
				a.Offset = (len(m.Args)-i)*4 + 4
			}
			genMethod(w, c.Type.Name+"."+m.Name.Name, len(m.Args), m.Body, true)
		}
	}
}

func (e *NotExpr) genCollectLiterals(ints func(int32) int, strings func(string) int) {
	e.Expr.genCollectLiterals(ints, strings)
}

func (e *NotExpr) genCountVars(this int) int {
	return e.Expr.genCountVars(this)
}

func (e *NotExpr) genCode(w io.Writer, label func() string, slot func() (int, func())) {
	label_true := label()
	label_false := label()
	label_done := label()

	e.genCodeJump(w, label_false+"f", label_true+"f", label, slot)

	fmt.Fprintf(w, "%s:\n", label_true)
	fmt.Fprintf(w, "\tlea boolean_true, %%eax\n")
	fmt.Fprintf(w, "\tjmp %sf\n", label_done)
	fmt.Fprintf(w, "%s:\n", label_false)
	fmt.Fprintf(w, "\tlea boolean_false, %%eax\n")
	fmt.Fprintf(w, "%s:\n", label_done)
}

func (e *NotExpr) genCodeJump(w io.Writer, l0, l1 string, label func() string, slot func() (int, func())) {
	if raw, ok := e.Expr.(JumpExpr); ok {
		raw.genCodeJump(w, l1, l0, label, slot)
	} else {
		e.Expr.genCode(w, label, slot)
		fmt.Fprintf(w, "\tcmpl $boolean_false, %%eax\n")
		fmt.Fprintf(w, "\tje %s\n", l1)
		fmt.Fprintf(w, "\tjmp %s\n", l0)
	}
}

func (e *NegativeExpr) genCollectLiterals(ints func(int32) int, strings func(string) int) {
	e.Expr.genCollectLiterals(ints, strings)
}

func (e *NegativeExpr) genCountVars(this int) int {
	vars := e.Expr.genCountVars(this)
	if v := 1; vars < v {
		vars = v
	}
	return vars
}

func (e *NegativeExpr) genCode(w io.Writer, label func() string, slot func() (int, func())) {
	offset, unreserve := slot()
	fmt.Fprintf(w, "\tmovl $(size_of_Int + 4), %%eax\n")
	fmt.Fprintf(w, "\tmovl $tag_of_Int, %%ebx\n")
	fmt.Fprintf(w, "\tcall gc_alloc\n")
	fmt.Fprintf(w, "\tmovl %%eax, %d(%%ebp)\n", offset)

	if raw, ok := e.Expr.(ArithmeticExpr); ok {
		raw.genCodeRawInt(w, label, slot)
		fmt.Fprintf(w, "\tmovl %%eax, %%ebx\n")
	} else {
		e.Expr.genCode(w, label, slot)
		genGC(w, "%eax", label)
		fmt.Fprintf(w, "\tmovl offset_of_Int.value(%%eax), %%ebx\n")
	}

	fmt.Fprintf(w, "\tmovl %d(%%ebp), %%eax\n", offset)
	unreserve()

	fmt.Fprintf(w, "\tnegl %%ebx\n")
	fmt.Fprintf(w, "\tmovl %%ebx, offset_of_Int.value(%%eax)\n")
}

func (e *NegativeExpr) genCodeRawInt(w io.Writer, label func() string, slot func() (int, func())) {
	if raw, ok := e.Expr.(ArithmeticExpr); ok {
		raw.genCodeRawInt(w, label, slot)
	} else {
		e.Expr.genCode(w, label, slot)
		genGC(w, "%eax", label)
		fmt.Fprintf(w, "\tmovl offset_of_Int.value(%%eax), %%eax\n")
	}

	fmt.Fprintf(w, "\tnegl %%eax\n")
}

func (e *IfExpr) genCollectLiterals(ints func(int32) int, strings func(string) int) {
	e.Cond.genCollectLiterals(ints, strings)
	e.Then.genCollectLiterals(ints, strings)
	e.Else.genCollectLiterals(ints, strings)
}

func (e *IfExpr) genCountVars(this int) int {
	vars := e.Cond.genCountVars(this)
	if v := e.Then.genCountVars(this); v > vars {
		vars = v
	}
	if v := e.Else.genCountVars(this); v > vars {
		vars = v
	}
	return vars
}

func (e *IfExpr) genCode(w io.Writer, label func() string, slot func() (int, func())) {
	label_true := label()
	label_false := label()
	label_done := label()

	if raw, ok := e.Cond.(JumpExpr); ok {
		raw.genCodeJump(w, label_false+"f", label_true+"f", label, slot)
	} else {
		e.Cond.genCode(w, label, slot)
		fmt.Fprintf(w, "\tcmpl $boolean_false, %%eax\n")
		fmt.Fprintf(w, "\tje %sf\n", label_false)
	}

	fmt.Fprintf(w, "%s:\n", label_true)
	e.Then.genCode(w, label, slot)
	fmt.Fprintf(w, "\tjmp %sf\n", label_done)
	fmt.Fprintf(w, "%s:\n", label_false)
	e.Else.genCode(w, label, slot)
	fmt.Fprintf(w, "%s:\n", label_done)
}

func (e *IfExpr) genCodeRawInt(w io.Writer, label func() string, slot func() (int, func())) {
	label_true := label()
	label_false := label()
	label_done := label()

	if raw, ok := e.Cond.(JumpExpr); ok {
		raw.genCodeJump(w, label_false+"f", label_true+"f", label, slot)
	} else {
		e.Cond.genCode(w, label, slot)
		fmt.Fprintf(w, "\tcmpl $boolean_false, %%eax\n")
		fmt.Fprintf(w, "\tje %sf\n", label_false)
	}

	fmt.Fprintf(w, "%s:\n", label_true)
	if raw, ok := e.Then.(ArithmeticExpr); ok {
		raw.genCodeRawInt(w, label, slot)
	} else {
		e.Then.genCode(w, label, slot)
		genGC(w, "%eax", label)
		fmt.Fprintf(w, "\tmovl offset_of_Int.value(%%eax), %%eax\n")
	}
	fmt.Fprintf(w, "\tjmp %sf\n", label_done)
	fmt.Fprintf(w, "%s:\n", label_false)
	if raw, ok := e.Else.(ArithmeticExpr); ok {
		raw.genCodeRawInt(w, label, slot)
	} else {
		e.Else.genCode(w, label, slot)
		genGC(w, "%eax", label)
		fmt.Fprintf(w, "\tmovl offset_of_Int.value(%%eax), %%eax\n")
	}
	fmt.Fprintf(w, "%s:\n", label_done)
}

func (e *IfExpr) genCodeJump(w io.Writer, l0, l1 string, label func() string, slot func() (int, func())) {
	label_true := label()
	label_false := label()

	if raw, ok := e.Cond.(JumpExpr); ok {
		raw.genCodeJump(w, label_false+"f", label_true+"f", label, slot)
	} else {
		e.Cond.genCode(w, label, slot)
		fmt.Fprintf(w, "\tcmpl $boolean_false, %%eax\n")
		fmt.Fprintf(w, "\tje %sf\n", label_false)
	}

	fmt.Fprintf(w, "%s:\n", label_true)
	if raw, ok := e.Then.(JumpExpr); ok {
		raw.genCodeJump(w, l0, l1, label, slot)
	} else {
		e.Then.genCode(w, label, slot)
		fmt.Fprintf(w, "\tcmpl $boolean_false, %%eax\n")
		fmt.Fprintf(w, "\tje %s\n", l0)
		fmt.Fprintf(w, "\tjmp %s\n", l1)
	}
	fmt.Fprintf(w, "%s:\n", label_false)
	if raw, ok := e.Else.(JumpExpr); ok {
		raw.genCodeJump(w, l0, l1, label, slot)
	} else {
		e.Else.genCode(w, label, slot)
		fmt.Fprintf(w, "\tcmpl $boolean_false, %%eax\n")
		fmt.Fprintf(w, "\tje %s\n", l0)
		fmt.Fprintf(w, "\tjmp %s\n", l1)
	}
}

func (e *WhileExpr) genCollectLiterals(ints func(int32) int, strings func(string) int) {
	e.Cond.genCollectLiterals(ints, strings)
	e.Body.genCollectLiterals(ints, strings)
}

func (e *WhileExpr) genCountVars(this int) int {
	vars := e.Cond.genCountVars(this)
	if v := e.Body.genCountVars(this); v > vars {
		vars = v
	}
	return vars
}

func (e *WhileExpr) genCode(w io.Writer, label func() string, slot func() (int, func())) {
	label_cond := label()
	label_body := label()
	label_done := label()

	fmt.Fprintf(w, "%s:\n", label_cond)
	if raw, ok := e.Cond.(JumpExpr); ok {
		raw.genCodeJump(w, label_done+"f", label_body+"f", label, slot)
	} else {
		e.Cond.genCode(w, label, slot)
		fmt.Fprintf(w, "\tcmpl $boolean_false, %%eax\n")
		fmt.Fprintf(w, "\tje %sf\n", label_done)
	}
	fmt.Fprintf(w, "%s:\n", label_body)
	e.Body.genCode(w, label, slot)
	genGC(w, "%eax", label)
	fmt.Fprintf(w, "\tjmp %sb\n", label_cond)
	fmt.Fprintf(w, "%s:\n", label_done)
	fmt.Fprintf(w, "\tlea unit_lit, %%eax\n")
}

func (e *LessOrEqualExpr) genCollectLiterals(ints func(int32) int, strings func(string) int) {
	e.Left.genCollectLiterals(ints, strings)
	e.Right.genCollectLiterals(ints, strings)
}

func (e *LessOrEqualExpr) genCountVars(this int) int {
	vars := e.Left.genCountVars(this)
	if v := 1 + e.Right.genCountVars(this); v > vars {
		vars = v
	}
	return vars
}

func (e *LessOrEqualExpr) genCode(w io.Writer, label func() string, slot func() (int, func())) {
	label_true := label()
	label_false := label()
	label_done := label()

	e.genCodeJump(w, label_false+"f", label_true+"f", label, slot)
	fmt.Fprintf(w, "%s:\n", label_false)
	fmt.Fprintf(w, "\tlea boolean_false, %%eax\n")
	fmt.Fprintf(w, "\tjmp %sf\n", label_done)
	fmt.Fprintf(w, "%s:\n", label_true)
	fmt.Fprintf(w, "\tlea boolean_true, %%eax\n")
	fmt.Fprintf(w, "%s:\n", label_done)
}

func (e *LessOrEqualExpr) genCodeJump(w io.Writer, l0, l1 string, label func() string, slot func() (int, func())) {
	if raw, ok := e.Left.(ArithmeticExpr); ok {
		raw.genCodeRawInt(w, label, slot)
	} else {
		e.Left.genCode(w, label, slot)
		genGC(w, "%eax", label)
		fmt.Fprintf(w, "\tmovl offset_of_Int.value(%%eax), %%eax\n")
	}
	offset, unreserve := slot()
	fmt.Fprintf(w, "\tmovl %%eax, %d(%%ebp)\n", offset)
	if raw, ok := e.Right.(ArithmeticExpr); ok {
		raw.genCodeRawInt(w, label, slot)
	} else {
		e.Right.genCode(w, label, slot)
		genGC(w, "%eax", label)
		fmt.Fprintf(w, "\tmovl offset_of_Int.value(%%eax), %%eax\n")
	}
	fmt.Fprintf(w, "\tmovl %d(%%ebp), %%ebx\n", offset)
	unreserve()
	fmt.Fprintf(w, "\tcmpl %%eax, %%ebx\n")
	fmt.Fprintf(w, "\tjle %s\n", l1)
	fmt.Fprintf(w, "\tjmp %s\n", l0)
}

func (e *LessThanExpr) genCollectLiterals(ints func(int32) int, strings func(string) int) {
	e.Left.genCollectLiterals(ints, strings)
	e.Right.genCollectLiterals(ints, strings)
}

func (e *LessThanExpr) genCountVars(this int) int {
	vars := e.Left.genCountVars(this)
	if v := 1 + e.Right.genCountVars(this); v > vars {
		vars = v
	}
	return vars
}

func (e *LessThanExpr) genCode(w io.Writer, label func() string, slot func() (int, func())) {
	label_true := label()
	label_false := label()
	label_done := label()

	e.genCodeJump(w, label_false+"f", label_true+"f", label, slot)
	fmt.Fprintf(w, "%s:\n", label_false)
	fmt.Fprintf(w, "\tlea boolean_false, %%eax\n")
	fmt.Fprintf(w, "\tjmp %sf\n", label_done)
	fmt.Fprintf(w, "%s:\n", label_true)
	fmt.Fprintf(w, "\tlea boolean_true, %%eax\n")
	fmt.Fprintf(w, "%s:\n", label_done)
}

func (e *LessThanExpr) genCodeJump(w io.Writer, l0, l1 string, label func() string, slot func() (int, func())) {
	if raw, ok := e.Left.(ArithmeticExpr); ok {
		raw.genCodeRawInt(w, label, slot)
	} else {
		e.Left.genCode(w, label, slot)
		genGC(w, "%eax", label)
		fmt.Fprintf(w, "\tmovl offset_of_Int.value(%%eax), %%eax\n")
	}
	offset, unreserve := slot()
	fmt.Fprintf(w, "\tmovl %%eax, %d(%%ebp)\n", offset)
	if raw, ok := e.Right.(ArithmeticExpr); ok {
		raw.genCodeRawInt(w, label, slot)
	} else {
		e.Right.genCode(w, label, slot)
		genGC(w, "%eax", label)
		fmt.Fprintf(w, "\tmovl offset_of_Int.value(%%eax), %%eax\n")
	}
	fmt.Fprintf(w, "\tmovl %d(%%ebp), %%ebx\n", offset)
	unreserve()
	fmt.Fprintf(w, "\tcmpl %%eax, %%ebx\n")
	fmt.Fprintf(w, "\tjl %s\n", l1)
	fmt.Fprintf(w, "\tjmp %s\n", l0)
}

func genArithmetic(w io.Writer, label func() string, slot func() (int, func()), left, right Expr, compute func(), box bool) {
	if raw, ok := left.(ArithmeticExpr); ok {
		raw.genCodeRawInt(w, label, slot)
	} else {
		left.genCode(w, label, slot)
		genGC(w, "%eax", label)
		fmt.Fprintf(w, "\tmovl offset_of_Int.value(%%eax), %%eax\n")
	}
	offset, unreserve := slot()
	fmt.Fprintf(w, "\tmovl %%eax, %d(%%ebp)\n", offset)
	if raw, ok := right.(ArithmeticExpr); ok {
		raw.genCodeRawInt(w, label, slot)
	} else {
		right.genCode(w, label, slot)
		genGC(w, "%eax", label)
		fmt.Fprintf(w, "\tmovl offset_of_Int.value(%%eax), %%eax\n")
	}
	fmt.Fprintf(w, "\tmovl %%eax, %%ecx\n")
	fmt.Fprintf(w, "\tmovl %d(%%ebp), %%ebx\n", offset)
	compute()
	if box {
		fmt.Fprintf(w, "\tmovl %%eax, %d(%%ebp)\n", offset)
		fmt.Fprintf(w, "\tmovl $(size_of_Int + 4), %%eax\n")
		fmt.Fprintf(w, "\tmovl $tag_of_Int, %%ebx\n")
		fmt.Fprintf(w, "\tcall gc_alloc\n")
		fmt.Fprintf(w, "\tmovl %d(%%ebp), %%ecx\n", offset)
		fmt.Fprintf(w, "\tmovl %%ecx, offset_of_Int.value(%%eax)\n")
	}
	unreserve()
}

func (e *MultiplyExpr) genCollectLiterals(ints func(int32) int, strings func(string) int) {
	e.Left.genCollectLiterals(ints, strings)
	e.Right.genCollectLiterals(ints, strings)
}

func (e *MultiplyExpr) genCountVars(this int) int {
	vars := e.Left.genCountVars(this)
	if v := 1 + e.Right.genCountVars(this); v > vars {
		vars = v
	}
	return vars
}

func (e *MultiplyExpr) genCode(w io.Writer, label func() string, slot func() (int, func())) {
	genArithmetic(w, label, slot, e.Left, e.Right, func() {
		fmt.Fprintf(w, "\timul %%ebx, %%ecx\n")
		fmt.Fprintf(w, "\tmovl %%ecx, %%eax\n")
	}, true)
}

func (e *MultiplyExpr) genCodeRawInt(w io.Writer, label func() string, slot func() (int, func())) {
	genArithmetic(w, label, slot, e.Left, e.Right, func() {
		fmt.Fprintf(w, "\timul %%ebx, %%ecx\n")
		fmt.Fprintf(w, "\tmovl %%ecx, %%eax\n")
	}, false)
}

func (e *DivideExpr) genCollectLiterals(ints func(int32) int, strings func(string) int) {
	e.Left.genCollectLiterals(ints, strings)
	e.Right.genCollectLiterals(ints, strings)
}

func (e *DivideExpr) genCountVars(this int) int {
	vars := e.Left.genCountVars(this)
	if v := 1 + e.Right.genCountVars(this); v > vars {
		vars = v
	}
	return vars
}

func (e *DivideExpr) genCode(w io.Writer, label func() string, slot func() (int, func())) {
	genArithmetic(w, label, slot, e.Left, e.Right, func() {
		fmt.Fprintf(w, "\tmovl %%ebx, %%eax\n")
		fmt.Fprintf(w, "\tcdq\n")
		fmt.Fprintf(w, "\tidiv %%ecx\n")
	}, true)
}

func (e *DivideExpr) genCodeRawInt(w io.Writer, label func() string, slot func() (int, func())) {
	genArithmetic(w, label, slot, e.Left, e.Right, func() {
		fmt.Fprintf(w, "\tmovl %%ebx, %%eax\n")
		fmt.Fprintf(w, "\tcdq\n")
		fmt.Fprintf(w, "\tidiv %%ecx\n")
	}, false)
}

func (e *AddExpr) genCollectLiterals(ints func(int32) int, strings func(string) int) {
	e.Left.genCollectLiterals(ints, strings)
	e.Right.genCollectLiterals(ints, strings)
}

func (e *AddExpr) genCountVars(this int) int {
	vars := e.Left.genCountVars(this)
	if v := 1 + e.Right.genCountVars(this); v > vars {
		vars = v
	}
	return vars
}

func (e *AddExpr) genCode(w io.Writer, label func() string, slot func() (int, func())) {
	genArithmetic(w, label, slot, e.Left, e.Right, func() {
		fmt.Fprintf(w, "\taddl %%ebx, %%ecx\n")
		fmt.Fprintf(w, "\tmovl %%ecx, %%eax\n")
	}, true)
}

func (e *AddExpr) genCodeRawInt(w io.Writer, label func() string, slot func() (int, func())) {
	genArithmetic(w, label, slot, e.Left, e.Right, func() {
		fmt.Fprintf(w, "\taddl %%ebx, %%ecx\n")
		fmt.Fprintf(w, "\tmovl %%ecx, %%eax\n")
	}, false)
}

func (e *SubtractExpr) genCollectLiterals(ints func(int32) int, strings func(string) int) {
	e.Left.genCollectLiterals(ints, strings)
	e.Right.genCollectLiterals(ints, strings)
}

func (e *SubtractExpr) genCountVars(this int) int {
	vars := e.Left.genCountVars(this)
	if v := 1 + e.Right.genCountVars(this); v > vars {
		vars = v
	}
	return vars
}

func (e *SubtractExpr) genCode(w io.Writer, label func() string, slot func() (int, func())) {
	genArithmetic(w, label, slot, e.Left, e.Right, func() {
		fmt.Fprintf(w, "\tsubl %%ecx, %%ebx\n")
		fmt.Fprintf(w, "\tmovl %%ebx, %%eax\n")
	}, true)
}

func (e *SubtractExpr) genCodeRawInt(w io.Writer, label func() string, slot func() (int, func())) {
	genArithmetic(w, label, slot, e.Left, e.Right, func() {
		fmt.Fprintf(w, "\tsubl %%ecx, %%ebx\n")
		fmt.Fprintf(w, "\tmovl %%ebx, %%eax\n")
	}, false)
}

func (e *MatchExpr) genCollectLiterals(ints func(int32) int, strings func(string) int) {
	e.Left.genCollectLiterals(ints, strings)
	for _, c := range e.Cases {
		c.genCollectLiterals(ints, strings)
	}
}

func (e *MatchExpr) genCountVars(this int) int {
	vars := e.Left.genCountVars(this)
	for _, c := range e.Cases {
		if v := c.Body.genCountVars(this); v > vars {
			vars = v
		}
	}
	return vars + 1
}

func (e *MatchExpr) genCode(w io.Writer, label func() string, slot func() (int, func())) {
	label_null := label()
	label_done := label()

	e.Left.genCode(w, label, slot)
	offset, unreserve := slot()
	e.Offset = offset
	fmt.Fprintf(w, "\tmovl %%eax, %d(%%ebp)\n", offset)
	fmt.Fprintf(w, "\ttest %%eax, %%eax\n")
	fmt.Fprintf(w, "\tjz %sf\n", label_null)
	fmt.Fprintf(w, "\tmovl tag_offset(%%eax), %%eax\n")
	fmt.Fprintf(w, "%s:\n", label_null)

	labels := make([]string, len(e.Cases))
	for i := range labels {
		labels[i] = label()
	}

	for i, c := range e.Cases {
		if c.Type.Class.Order == c.Type.Class.MaxOrder {
			fmt.Fprintf(w, "\tcmpl $%d, %%eax\n", c.Type.Class.Order)
			fmt.Fprintf(w, "\tje %sf\n", labels[i])
		} else {
			label_skip := label()
			fmt.Fprintf(w, "\tcmpl $%d, %%eax\n", c.Type.Class.Order)
			fmt.Fprintf(w, "\tjl %sf\n", label_skip)
			fmt.Fprintf(w, "\tcmpl $%d, %%eax\n", c.Type.Class.MaxOrder)
			fmt.Fprintf(w, "\tjle %sf\n", labels[i])
			fmt.Fprintf(w, "%s:\n", label_skip)
		}
	}
	fmt.Fprintf(w, "\tjmp runtime.case_panic\n")

	for i, c := range e.Cases {
		fmt.Fprintf(w, "%s:\n", labels[i])
		c.genCode(w, label, slot)
		fmt.Fprintf(w, "\tjmp %sf\n", label_done)
	}

	fmt.Fprintf(w, "%s:\n", label_done)
	fmt.Fprintf(w, "\tmovl %d(%%ebp), %%ebx\n", offset)
	genGC(w, "%ebx", label)
	unreserve()
}

func (e *MatchExpr) genCodeRawInt(w io.Writer, label func() string, slot func() (int, func())) {
	label_null := label()
	label_done := label()

	e.Left.genCode(w, label, slot)
	offset, unreserve := slot()
	e.Offset = offset
	fmt.Fprintf(w, "\tmovl %%eax, %d(%%ebp)\n", offset)
	fmt.Fprintf(w, "\ttest %%eax, %%eax\n")
	fmt.Fprintf(w, "\tjz %sf\n", label_null)
	fmt.Fprintf(w, "\tmovl tag_offset(%%eax), %%eax\n")
	fmt.Fprintf(w, "%s:\n", label_null)

	labels := make([]string, len(e.Cases))
	for i := range labels {
		labels[i] = label()
	}

	for i, c := range e.Cases {
		if c.Type.Class.Order == c.Type.Class.MaxOrder {
			fmt.Fprintf(w, "\tcmpl $%d, %%eax\n", c.Type.Class.Order)
			fmt.Fprintf(w, "\tje %sf\n", labels[i])
		} else {
			label_skip := label()
			fmt.Fprintf(w, "\tcmpl $%d, %%eax\n", c.Type.Class.Order)
			fmt.Fprintf(w, "\tjl %sf\n", label_skip)
			fmt.Fprintf(w, "\tcmpl $%d, %%eax\n", c.Type.Class.MaxOrder)
			fmt.Fprintf(w, "\tjle %sf\n", labels[i])
			fmt.Fprintf(w, "%s:\n", label_skip)
		}
	}
	fmt.Fprintf(w, "\tjmp runtime.case_panic\n")

	for i, c := range e.Cases {
		fmt.Fprintf(w, "%s:\n", labels[i])
		c.genCodeRawInt(w, label, slot)
		fmt.Fprintf(w, "\tjmp %sf\n", label_done)
	}

	fmt.Fprintf(w, "%s:\n", label_done)
	fmt.Fprintf(w, "\tmovl %d(%%ebp), %%ebx\n", offset)
	genGC(w, "%ebx", label)
	unreserve()
}

func (e *MatchExpr) genCodeJump(w io.Writer, l0, l1 string, label func() string, slot func() (int, func())) {
	label_null := label()
	label_true := label()
	label_false := label()

	e.Left.genCode(w, label, slot)
	offset, unreserve := slot()
	e.Offset = offset
	fmt.Fprintf(w, "\tmovl %%eax, %d(%%ebp)\n", offset)
	fmt.Fprintf(w, "\ttest %%eax, %%eax\n")
	fmt.Fprintf(w, "\tjz %sf\n", label_null)
	fmt.Fprintf(w, "\tmovl tag_offset(%%eax), %%eax\n")
	fmt.Fprintf(w, "%s:\n", label_null)

	labels := make([]string, len(e.Cases))
	for i := range labels {
		labels[i] = label()
	}

	for i, c := range e.Cases {
		if c.Type.Class.Order == c.Type.Class.MaxOrder {
			fmt.Fprintf(w, "\tcmpl $%d, %%eax\n", c.Type.Class.Order)
			fmt.Fprintf(w, "\tje %sf\n", labels[i])
		} else {
			label_skip := label()
			fmt.Fprintf(w, "\tcmpl $%d, %%eax\n", c.Type.Class.Order)
			fmt.Fprintf(w, "\tjl %sf\n", label_skip)
			fmt.Fprintf(w, "\tcmpl $%d, %%eax\n", c.Type.Class.MaxOrder)
			fmt.Fprintf(w, "\tjle %sf\n", labels[i])
			fmt.Fprintf(w, "%s:\n", label_skip)
		}
	}
	fmt.Fprintf(w, "\tjmp runtime.case_panic\n")

	for i, c := range e.Cases {
		fmt.Fprintf(w, "%s:\n", labels[i])
		c.genCodeJump(w, label_false+"f", label_true+"f", label, slot)
	}

	fmt.Fprintf(w, "%s:\n", label_false)
	fmt.Fprintf(w, "\tmovl %d(%%ebp), %%ebx\n", offset)
	genGC(w, "%ebx", label)
	fmt.Fprintf(w, "\tjmp %s\n", l0)

	fmt.Fprintf(w, "%s:\n", label_true)
	fmt.Fprintf(w, "\tmovl %d(%%ebp), %%ebx\n", offset)
	genGC(w, "%ebx", label)
	fmt.Fprintf(w, "\tjmp %s\n", l1)

	unreserve()
}

func (e *DynamicCallExpr) genCollectLiterals(ints func(int32) int, strings func(string) int) {
	e.Recv.genCollectLiterals(ints, strings)
	for _, a := range e.Args {
		a.genCollectLiterals(ints, strings)
	}
}

func (e *DynamicCallExpr) genCountVars(this int) int {
	vars := e.Recv.genCountVars(this)
	for _, a := range e.Args {
		if v := a.genCountVars(this); v > vars {
			vars = v
		}
	}
	return vars
}

func (e *DynamicCallExpr) genCode(w io.Writer, label func() string, slot func() (int, func())) {
	e.Recv.genCode(w, label, slot)
	fmt.Fprintf(w, "\ttest %%eax, %%eax\n")
	fmt.Fprintf(w, "\tjz runtime.null_panic\n")
	fmt.Fprintf(w, "\tpush %%eax\n")
	for _, a := range e.Args {
		a.genCode(w, label, slot)
		fmt.Fprintf(w, "\tpush %%eax\n")
	}
	fmt.Fprintf(w, "\tmovl %d(%%esp), %%eax\n", len(e.Args)*4)
	fmt.Fprintf(w, "\tmovl tag_offset(%%eax), %%eax\n")
	fmt.Fprintf(w, "\tshll $2, %%eax\n")
	fmt.Fprintf(w, "\tmovl method_tables(%%eax), %%eax\n")
	fmt.Fprintf(w, "\tmovl %d(%%eax), %%eax\n", e.Name.Method.Order*4)
	fmt.Fprintf(w, "\tcall *%%eax\n")
}

func (e *SuperCallExpr) genCollectLiterals(ints func(int32) int, strings func(string) int) {
	for _, a := range e.Args {
		a.genCollectLiterals(ints, strings)
	}
}

func (e *SuperCallExpr) genCountVars(this int) int {
	e.This = this
	vars := 0
	for _, a := range e.Args {
		if v := a.genCountVars(this); v > vars {
			vars = v
		}
	}
	return vars
}

func (e *SuperCallExpr) genCode(w io.Writer, label func() string, slot func() (int, func())) {
	fmt.Fprintf(w, "\tmovl %d(%%ebp), %%eax\n", e.This)
	genRef(w, "%eax", label)
	fmt.Fprintf(w, "\tpush %%eax\n")
	for _, a := range e.Args {
		a.genCode(w, label, slot)
		fmt.Fprintf(w, "\tpush %%eax\n")
	}
	fmt.Fprintf(w, "\tcall %s.%s\n", e.Name.Method.Parent.Type.Name, e.Name.Method.Name.Name)
}

func (e *StaticCallExpr) genCollectLiterals(ints func(int32) int, strings func(string) int) {
	e.Recv.genCollectLiterals(ints, strings)
	for _, a := range e.Args {
		a.genCollectLiterals(ints, strings)
	}
}

func (e *StaticCallExpr) genCountVars(this int) int {
	vars := e.Recv.genCountVars(this)
	for _, a := range e.Args {
		if v := a.genCountVars(this); v > vars {
			vars = v
		}
	}
	return vars
}

func (e *StaticCallExpr) genCode(w io.Writer, label func() string, slot func() (int, func())) {
	e.Recv.genCode(w, label, slot)
	fmt.Fprintf(w, "\tpush %%eax\n")
	for _, a := range e.Args {
		a.genCode(w, label, slot)
		fmt.Fprintf(w, "\tpush %%eax\n")
	}
	fmt.Fprintf(w, "\tcall %s.%s\n", e.Name.Method.Parent.Type.Name, e.Name.Method.Name.Name)
}

func (e *AllocExpr) genCollectLiterals(ints func(int32) int, strings func(string) int) {
}

func (e *AllocExpr) genCountVars(this int) int {
	return 0
}

func (e *AllocExpr) genCode(w io.Writer, label func() string, slot func() (int, func())) {
	fmt.Fprintf(w, "\tmovl $size_of_%s, %%eax\n", e.Type.Name)
	fmt.Fprintf(w, "\tmovl $tag_of_%s, %%ebx\n", e.Type.Name)
	fmt.Fprintf(w, "\tcall gc_alloc\n")
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
				fmt.Fprintf(w, "\tmovl $%s, offset_of_%s.%s(%%eax)\n", sym, c.Type.Name, a.Name.Name)
			}
		}
	}
	gen(e.Type.Class)
}

func (e *AssignExpr) genCollectLiterals(ints func(int32) int, strings func(string) int) {
	e.Expr.genCollectLiterals(ints, strings)
}

func (e *AssignExpr) genCountVars(this int) int {
	e.This = this
	return e.Expr.genCountVars(this)
}

func (e *AssignExpr) genCode(w io.Writer, label func() string, slot func() (int, func())) {
	var rawInt bool
	if e.Name.Object.RawInt() {
		if a, ok := e.Expr.(ArithmeticExpr); ok {
			rawInt = true
			a.genCodeRawInt(w, label, slot)
		}
	}
	if !rawInt {
		e.Expr.genCode(w, label, slot)
	}
	fmt.Fprintf(w, "\tmovl %s, %%edx\n", e.Name.Object.Base(e.This))
	if e.Name.Object.Stack() {
		if !e.Name.Object.RawInt() {
			fmt.Fprintf(w, "\tmovl %s(%%edx), %%ebx\n", e.Name.Object.Offs())
			genGC(w, "%ebx", label)
		}
	} else if !rawInt {
		genGC(w, "%eax", label)
	}
	if e.Name.Object.RawInt() && !rawInt {
		if e.Name.Object.Stack() {
			genGC(w, "%eax", label)
		}
		fmt.Fprintf(w, "\tmovl offset_of_Int.value(%%eax), %%eax\n")
	}
	fmt.Fprintf(w, "\tmovl %%eax, %s(%%edx)\n", e.Name.Object.Offs())
	fmt.Fprintf(w, "\tleal unit_lit, %%eax\n")
}

func (e *VarExpr) genCollectLiterals(ints func(int32) int, strings func(string) int) {
	e.Init.genCollectLiterals(ints, strings)
	e.Body.genCollectLiterals(ints, strings)
}

func (e *VarExpr) genCountVars(this int) int {
	vars := e.Init.genCountVars(this)
	if v := e.Body.genCountVars(this); v > vars {
		vars = v
	}
	return vars + 1
}

func (e *VarExpr) genCodeShared(w io.Writer, label func() string, slot func() (int, func()), returnRawInt bool) {
	var rawInt bool
	if e.RawInt() {
		if a, ok := e.Init.(ArithmeticExpr); ok {
			rawInt = true
			a.genCodeRawInt(w, label, slot)
		}
	}
	if !rawInt {
		e.Init.genCode(w, label, slot)
		if e.RawInt() {
			genGC(w, "%eax", label)
			fmt.Fprintf(w, "\tmovl offset_of_Int.value(%%eax), %%eax\n")
		}
	}
	offset, unreserve := slot()
	e.Offset = offset
	fmt.Fprintf(w, "\tmovl %%eax, %d(%%ebp)\n", offset)
	if returnRawInt {
		if a, ok := e.Body.(ArithmeticExpr); ok {
			a.genCodeRawInt(w, label, slot)
		} else {
			e.Body.genCode(w, label, slot)
			genGC(w, "%eax", label)
			fmt.Fprintf(w, "\tmovl offset_of_Int.value(%%eax), %%eax\n")
		}
	} else {
		e.Body.genCode(w, label, slot)
	}
	fmt.Fprintf(w, "\tmovl %d(%%ebp), %%ebx\n", offset)
	if !e.RawInt() {
		genGC(w, "%ebx", label)
	}
	unreserve()
}

func (e *VarExpr) genCode(w io.Writer, label func() string, slot func() (int, func())) {
	e.genCodeShared(w, label, slot, false)
}

func (e *VarExpr) genCodeRawInt(w io.Writer, label func() string, slot func() (int, func())) {
	e.genCodeShared(w, label, slot, true)
}

func (e *ChainExpr) genCollectLiterals(ints func(int32) int, strings func(string) int) {
	e.Pre.genCollectLiterals(ints, strings)
	e.Expr.genCollectLiterals(ints, strings)
}

func (e *ChainExpr) genCountVars(this int) int {
	vars := e.Pre.genCountVars(this)
	if v := e.Expr.genCountVars(this); v > vars {
		vars = v
	}
	return vars
}

func (e *ChainExpr) genCode(w io.Writer, label func() string, slot func() (int, func())) {
	e.Pre.genCode(w, label, slot)
	genGC(w, "%eax", label)
	e.Expr.genCode(w, label, slot)
}

func (e *ChainExpr) genCodeRawInt(w io.Writer, label func() string, slot func() (int, func())) {
	e.Pre.genCode(w, label, slot)
	genGC(w, "%eax", label)
	if a, ok := e.Expr.(ArithmeticExpr); ok {
		a.genCodeRawInt(w, label, slot)
	} else {
		e.Expr.genCode(w, label, slot)
		genGC(w, "%eax", label)
		fmt.Fprintf(w, "\tmovl offset_of_Int.value(%%eax), %%eax\n")
	}
}

func (e *ThisExpr) genCollectLiterals(ints func(int32) int, strings func(string) int) {
}

func (e *ThisExpr) genCountVars(this int) int {
	e.Offset = this
	return 0
}

func (e *ThisExpr) genCode(w io.Writer, label func() string, slot func() (int, func())) {
	fmt.Fprintf(w, "\tmovl %d(%%ebp), %%eax\n", e.Offset)
	genRef(w, "%eax", label)
}

func (e *NullExpr) genCollectLiterals(ints func(int32) int, strings func(string) int) {
}

func (e *NullExpr) genCountVars(this int) int {
	return 0
}

func (e *NullExpr) genCode(w io.Writer, label func() string, slot func() (int, func())) {
	fmt.Fprintf(w, "\tmovl $0, %%eax\n")
}

func (e *UnitExpr) genCollectLiterals(ints func(int32) int, strings func(string) int) {
}

func (e *UnitExpr) genCountVars(this int) int {
	return 0
}

func (e *UnitExpr) genCode(w io.Writer, label func() string, slot func() (int, func())) {
	fmt.Fprintf(w, "\tleal unit_lit, %%eax\n")
}

func (e *NameExpr) genCollectLiterals(ints func(int32) int, strings func(string) int) {
}

func (e *NameExpr) genCountVars(this int) int {
	e.This = this
	return 0
}

func (e *NameExpr) genCode(w io.Writer, label func() string, slot func() (int, func())) {
	if e.Name.Object.RawInt() {
		fmt.Fprintf(w, "\tmovl $(size_of_Int + 4), %%eax\n")
		fmt.Fprintf(w, "\tmovl $tag_of_Int, %%ebx\n")
		fmt.Fprintf(w, "\tcall gc_alloc\n")
		fmt.Fprintf(w, "\tmovl %s, %%edx\n", e.Name.Object.Base(e.This))
		fmt.Fprintf(w, "\tmovl %s(%%edx), %%ebx\n", e.Name.Object.Offs())
		fmt.Fprintf(w, "\tmovl %%ebx, offset_of_Int.value(%%eax)\n")
	} else {
		fmt.Fprintf(w, "\tmovl %s, %%edx\n", e.Name.Object.Base(e.This))
		fmt.Fprintf(w, "\tmovl %s(%%edx), %%eax\n", e.Name.Object.Offs())
		genRef(w, "%eax", label)
	}
}

func (e *NameExpr) genCodeRawInt(w io.Writer, label func() string, slot func() (int, func())) {
	fmt.Fprintf(w, "\tmovl %s, %%edx\n", e.Name.Object.Base(e.This))
	fmt.Fprintf(w, "\tmovl %s(%%edx), %%eax\n", e.Name.Object.Offs())
	if !e.Name.Object.RawInt() {
		fmt.Fprintf(w, "\tmovl offset_of_Int.value(%%eax), %%eax\n")
	}
}

func (e *StringExpr) genCollectLiterals(ints func(int32) int, strings func(string) int) {
	e.Lit.LitID = strings(e.Lit.Str)
}

func (e *StringExpr) genCountVars(this int) int {
	return 0
}

func (e *StringExpr) genCode(w io.Writer, label func() string, slot func() (int, func())) {
	fmt.Fprintf(w, "\tleal string_lit_%d, %%eax\n", e.Lit.LitID)
}

func (e *BoolExpr) genCollectLiterals(ints func(int32) int, strings func(string) int) {
}

func (e *BoolExpr) genCountVars(this int) int {
	return 0
}

func (e *BoolExpr) genCode(w io.Writer, label func() string, slot func() (int, func())) {
	if e.Lit.Bool {
		fmt.Fprintf(w, "\tleal boolean_true, %%eax\n")
	} else {
		fmt.Fprintf(w, "\tleal boolean_false, %%eax\n")
	}
}

func (e *BoolExpr) genCodeJump(w io.Writer, l0, l1 string, label func() string, slot func() (int, func())) {
	if e.Lit.Bool {
		fmt.Fprintf(w, "\tjmp %s\n", l1)
	} else {
		fmt.Fprintf(w, "\tjmp %s\n", l0)
	}
}

func (e *IntExpr) genCollectLiterals(ints func(int32) int, strings func(string) int) {
	e.Lit.LitID = ints(e.Lit.Int)
}

func (e *IntExpr) genCountVars(this int) int {
	return 0
}

func (e *IntExpr) genCode(w io.Writer, label func() string, slot func() (int, func())) {
	fmt.Fprintf(w, "\tleal int_lit_%d, %%eax\n", e.Lit.LitID)
}

func (e *IntExpr) genCodeRawInt(w io.Writer, label func() string, slot func() (int, func())) {
	fmt.Fprintf(w, "\tmovl $%d, %%eax\n", e.Lit.Int)
}

func (e *NativeExpr) genCollectLiterals(ints func(int32) int, strings func(string) int) {
}

func (e *NativeExpr) genCountVars(this int) int {
	panic("NativeExpr.genCountVars should never be called")
}

func (e *NativeExpr) genCode(w io.Writer, label func() string, slot func() (int, func())) {
	panic("NativeExpr.genCode should never be called")
}

func (c *Case) genCollectLiterals(ints func(int32) int, strings func(string) int) {
	c.Body.genCollectLiterals(ints, strings)
}

func (c *Case) genCountVars(this int) int {
	return c.Body.genCountVars(this)
}

func (c *Case) genCode(w io.Writer, label func() string, slot func() (int, func())) {
	c.Body.genCode(w, label, slot)
}

func (c *Case) genCodeRawInt(w io.Writer, label func() string, slot func() (int, func())) {
	if raw, ok := c.Body.(ArithmeticExpr); ok {
		raw.genCodeRawInt(w, label, slot)
	} else {
		c.Body.genCode(w, label, slot)
		genGC(w, "%eax", label)
		fmt.Fprintf(w, "\tmovl offset_of_Int.value(%%eax), %%eax\n")
	}
}

func (c *Case) genCodeJump(w io.Writer, l0, l1 string, label func() string, slot func() (int, func())) {
	if raw, ok := c.Body.(JumpExpr); ok {
		raw.genCodeJump(w, l0, l1, label, slot)
	} else {
		c.Body.genCode(w, label, slot)
		fmt.Fprintf(w, "\tcmpl $boolean_false, %%eax\n")
		fmt.Fprintf(w, "\tje %s\n", l0)
		fmt.Fprintf(w, "\tjmp %s\n", l1)
	}
}
