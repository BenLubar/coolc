.include "basic_defs.s"

.data

.align 2
symbol:
	.long 0

.globl boolean_false
.align 2
boolean_false:
	.long tag_of_Boolean
	.long size_of_Boolean
	.long gc_tag_root

.globl boolean_true
.align 2
boolean_true:
	.long tag_of_Boolean
	.long size_of_Boolean
	.long gc_tag_root

.globl unit_lit
.align 2
unit_lit:
	.long tag_of_Unit
	.long size_of_Unit
	.long gc_tag_root

.text

.globl _start
_start:
	call gc_init

	call main

	push $0
	call runtime.exit

.globl Any.toString
.type Any.toString, @function
Any.toString:
	.cfi_startproc
	push %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register ebp
	subl $0, %esp

	movl 8(%ebp), %ebx
	movl tag_offset(%ebx), %eax
	shll $2, %eax
	movl class_names(%eax), %eax

	cmpl $0, gc_offset(%ebx)
	jle 1f
	decl gc_offset(%ebx)
1:

	leave
	.cfi_def_cfa esp, 4
	ret $4
	.cfi_endproc
	.size Any.toString, .-Any.toString

.globl Any.equals
.type Any.equals, @function
Any.equals:
	.cfi_startproc
	push %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register ebp
	subl $0, %esp

	movl 8(%ebp), %eax
	movl 12(%ebp), %ebx

	cmpl $0, gc_offset(%eax)
	jle 3f
	decl gc_offset(%eax)
3:
	cmpl $0, gc_offset(%ebx)
	jle 4f
	decl gc_offset(%ebx)
4:

	cmpl %eax, %ebx
	jne 1f

	leal boolean_true, %eax

	jmp 2f

1:
	leal boolean_false, %eax

2:
	leave
	.cfi_def_cfa esp, 4
	ret $8
	.cfi_endproc
	.size Any.equals, .-Any.equals

.globl IO.abort
.type IO.abort, @function
IO.abort:
	.cfi_startproc
	push %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register ebp
	subl $0, %esp

	movl 8(%ebp), %eax
	push %eax
	call runtime.output

	push $1
	call runtime.exit
	.cfi_endproc
	.size IO.abort, .-IO.abort

.globl IO.out
.type IO.out, @function
IO.out:
	.cfi_startproc
	push %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register ebp
	subl $0, %esp

	movl 8(%ebp), %eax
	test %eax, %eax
	jz runtime.null_panic
	push %eax
	call runtime.output

	movl 8(%ebp), %eax
	cmpl $0, gc_offset(%eax)
	jle 1f
	decl gc_offset(%eax)
1:
	movl 12(%ebp), %eax

	leave
	.cfi_def_cfa esp, 4
	ret $8
	.cfi_endproc
	.size IO.out, .-IO.out

.globl IO.in
.type IO.in, @function
IO.in:
	.cfi_startproc
	push %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register ebp
	subl $4, %esp

	movl $(size_of_Int + 4), %eax
	movl $tag_of_Int, %ebx
	call gc_alloc
	movl $runtime_input_max, offset_of_Int.value(%eax)
	movl %eax, -4(%ebp)

	movl $size_of_String, %eax
	addl $runtime_input_max, %eax
	movl $tag_of_String, %ebx
	call gc_alloc

	movl -4(%ebp), %ebx
	movl %ebx, offset_of_String.length(%eax)
	decl gc_offset(%ebx)

	push %eax
	call runtime.input

	movl 8(%ebp), %ebx
	cmpl $0, gc_offset(%ebx)
	jle 1f
	decl gc_offset(%ebx)
1:

	leave
	.cfi_def_cfa esp, 4
	ret $4
	.cfi_endproc
	.size IO.in, .-IO.in

.globl IO.symbol
.type IO.symbol, @function
IO.symbol:
	.cfi_startproc
	push %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register ebp
	subl $12, %esp

	// there's no state in IO, so release `this`
	movl 12(%ebp), %eax
	cmpl $0, gc_offset(%eax)
	jle 1f
	decl gc_offset(%eax)
1:

	// 8(%ebp) = the string we're looking for
	// -4(%ebp) = the address of the "next" field of the previous symbol
	// -8(%ebp) = the address of the symbol we're currently looking at
	// -12(%ebp) = our current hash value
	movl $symbol, -4(%ebp)
	movl symbol, %eax
	movl %eax, -8(%ebp)
	movl $0, -12(%ebp)

	// make sure we actually have a string
	movl 8(%ebp), %eax
	test %eax, %eax
	jz runtime.null_panic

2:
	// add 1 to our hash
	incl -12(%ebp)

	// are we at the end?
	movl -8(%ebp), %eax
	test %eax, %eax
	jz 3f

	movl offset_of_Symbol.name(%eax), %eax
	cmpl $0, gc_offset(%eax)
	jl 6f
	incl gc_offset(%eax)
6:
	push %eax
	movl 8(%ebp), %eax
	cmpl $0, gc_offset(%eax)
	jl 7f
	incl gc_offset(%eax)
7:
	push %eax
	call String.equals
	cmpl $boolean_true, %eax
	je 4f

	// go to the next one
	movl -8(%ebp), %eax
	leal offset_of_Symbol.next(%eax), %eax
	movl %eax, -4(%ebp)
	movl (%eax), %eax
	movl %eax, -8(%ebp)
	jmp 2b

3:
	// box the hash
	movl $(size_of_Int + 4), %eax
	movl $tag_of_Int, %ebx
	call gc_alloc
	movl $gc_tag_root, gc_offset(%eax)
	movl -12(%ebp), %ebx
	movl %ebx, offset_of_Int.value(%eax)
	movl %eax, -12(%ebp)

	// mark the string and its length as a root
	movl 8(%ebp), %eax
	movl $gc_tag_root, gc_offset(%eax)
	movl offset_of_String.length(%eax), %eax
	movl $gc_tag_root, gc_offset(%eax)

	// make a new symbol object
	movl $size_of_Symbol, %eax
	movl $tag_of_Symbol, %ebx
	call gc_alloc
	movl $gc_tag_root, gc_offset(%eax)
	movl 8(%ebp), %ebx
	movl %ebx, offset_of_Symbol.name(%eax)
	movl -12(%ebp), %ebx
	movl %ebx, offset_of_Symbol.hash(%eax)

	// save it at the end of the list
	movl -4(%ebp), %ebx
	movl %eax, (%ebx)

	jmp 6f

4:
	// let the string be garbage collected
	movl 8(%ebp), %eax
	cmpl $0, gc_offset(%eax)
	jle 5f
	decl gc_offset(%eax)
5:
	// grab the symbol we found
	movl -8(%ebp), %eax

6:
	leave
	.cfi_def_cfa esp, 4
	ret $8
	.cfi_endproc
	.size IO.symbol, .-IO.symbol

.globl IO.symbol_name
.type IO.symbol_name, @function
IO.symbol_name:
	.cfi_startproc
	push %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register ebp
	subl $0, %esp

	movl 8(%ebp), %eax
	test %eax, %eax
	jz runtime.null_panic
	movl offset_of_Symbol.name(%eax), %eax

	leave
	.cfi_def_cfa esp, 4
	ret $8
	.cfi_endproc
	.size IO.symbol_name, .-IO.symbol_name

.data

int_lit_min_int_length:
	.long tag_of_Int
	.long size_of_Int + 4
	.long gc_tag_root
	.long string_lit_min_int_length

string_lit_min_int:
	.long tag_of_String
	.long size_of_String + string_lit_min_int_length
	.long gc_tag_root
	.long int_lit_min_int_length
string_lit_min_int_start:
	.ascii "-2147483648"
	.set string_lit_min_int_length, .-string_lit_min_int_start

.text

.globl Int.toString
.type Int.toString, @function
Int.toString:
	.cfi_startproc
	push %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register ebp
	subl $4, %esp

	movl 8(%ebp), %eax
	cmpl $0, gc_offset(%eax)
	jle 1f
	decl gc_offset(%eax)
1:
	movl offset_of_Int.value(%eax), %eax

	cmpl $-2147483648, %eax
	jne 2f

	movl $string_lit_min_int, %eax
	leave
	ret $4
2:
	movl %eax, 8(%ebp)

	movl $(size_of_Int + 4), %eax
	movl $tag_of_Int, %ebx
	call gc_alloc
	movl %eax, -4(%ebp)

	movl $(size_of_String + string_lit_min_int_length), %eax
	movl $tag_of_String, %ebx
	call gc_alloc
	movl -4(%ebp), %edx
	decl gc_offset(%edx)
	movl %edx, offset_of_String.length(%eax)
	movl %eax, -4(%ebp)

	movl 8(%ebp), %ebx
	leal offset_of_String.str_field(%eax), %ecx
	movl $string_lit_min_int_length, offset_of_Int.value(%edx)
	cmpl $0, %ebx
	jl 3f
	jg 4f

	movb $0x30, (%ecx)
	movl $1, offset_of_Int.value(%edx)

	jmp 15f

3:
	movb $0x2D, (%ecx)
	incl %ecx
	incl offset_of_Int.value(%edx)
	negl %ebx
4:
	decl offset_of_Int.value(%edx)
	cmpl $1000000000, %ebx
	jge 5f
	decl offset_of_Int.value(%edx)
	cmpl $100000000, %ebx
	jge 6f
	decl offset_of_Int.value(%edx)
	cmpl $10000000, %ebx
	jge 7f
	decl offset_of_Int.value(%edx)
	cmpl $1000000, %ebx
	jge 8f
	decl offset_of_Int.value(%edx)
	cmpl $100000, %ebx
	jge 9f
	decl offset_of_Int.value(%edx)
	cmpl $10000, %ebx
	jge 10f
	decl offset_of_Int.value(%edx)
	cmpl $1000, %ebx
	jge 11f
	decl offset_of_Int.value(%edx)
	cmpl $100, %ebx
	jge 12f
	decl offset_of_Int.value(%edx)
	cmpl $10, %ebx
	jge 13f
	decl offset_of_Int.value(%edx)
	jmp 14f

5:
	movl %ebx, %eax
	movl $0, %edx
	movl $1000000000, %ebx
	divl %ebx
	addl $0x30, %eax
	movb %al, (%ecx)
	incl %ecx
	movl %edx, %ebx
6:
	movl %ebx, %eax
	movl $0, %edx
	movl $100000000, %ebx
	divl %ebx
	addl $0x30, %eax
	movb %al, (%ecx)
	incl %ecx
	movl %edx, %ebx
7:
	movl %ebx, %eax
	movl $0, %edx
	movl $10000000, %ebx
	divl %ebx
	addl $0x30, %eax
	movb %al, (%ecx)
	incl %ecx
	movl %edx, %ebx
8:
	movl %ebx, %eax
	movl $0, %edx
	movl $1000000, %ebx
	divl %ebx
	addl $0x30, %eax
	movb %al, (%ecx)
	incl %ecx
	movl %edx, %ebx
9:
	movl %ebx, %eax
	movl $0, %edx
	movl $100000, %ebx
	divl %ebx
	addl $0x30, %eax
	movb %al, (%ecx)
	incl %ecx
	movl %edx, %ebx
10:
	movl %ebx, %eax
	movl $0, %edx
	movl $10000, %ebx
	divl %ebx
	addl $0x30, %eax
	movb %al, (%ecx)
	incl %ecx
	movl %edx, %ebx
11:
	movl %ebx, %eax
	movl $0, %edx
	movl $1000, %ebx
	divl %ebx
	addl $0x30, %eax
	movb %al, (%ecx)
	incl %ecx
	movl %edx, %ebx
12:
	movl %ebx, %eax
	movl $0, %edx
	movl $100, %ebx
	divl %ebx
	addl $0x30, %eax
	movb %al, (%ecx)
	incl %ecx
	movl %edx, %ebx
13:
	movl %ebx, %eax
	movl $0, %edx
	movl $10, %ebx
	divl %ebx
	addl $0x30, %eax
	movb %al, (%ecx)
	incl %ecx
	movl %edx, %ebx
14:
	addl $0x30, %ebx
	movb %bl, (%ecx)

	movl -4(%ebp), %eax

15:
	leave
	.cfi_def_cfa esp, 4
	ret $4
	.cfi_endproc
	.size Int.toString, .-Int.toString

.globl Int.equals
.type Int.equals, @function
Int.equals:
	.cfi_startproc
	push %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register ebp
	subl $0, %esp

	movl 8(%ebp), %eax
	test %eax, %eax
	jz 1f
	movl tag_offset(%eax), %ebx
	cmpl $tag_of_Int, %ebx
	jne 1f
	movl 12(%ebp), %ebx

	cmpl $0, gc_offset(%eax)
	jle 2f
	decl gc_offset(%eax)
2:
	cmpl $0, gc_offset(%ebx)
	jle 3f
	decl gc_offset(%ebx)
3:
	movl offset_of_Int.value(%eax), %eax
	movl offset_of_Int.value(%ebx), %ebx
	cmpl %eax, %ebx
	jne 1f

	leal boolean_true, %eax

	jmp 4f

1:
	leal boolean_false, %eax

4:
	leave
	.cfi_def_cfa esp, 4
	ret $8
	.cfi_endproc
	.size Int.equals, .-Int.equals

.globl String.equals
.type String.equals, @function
String.equals:
	.cfi_startproc
	push %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register ebp
	subl $0, %esp

	movl 8(%ebp), %eax
	test %eax, %eax
	jz 1f
	cmpl $tag_of_String, tag_offset(%eax)
	jne 1f

	movl offset_of_String.length(%eax), %ebx
	movl offset_of_Int.value(%ebx), %ebx
	movl 12(%ebp), %edx
	movl offset_of_String.length(%edx), %ecx
	cmpl %ebx, offset_of_Int.value(%ecx)
	jne 1f

	movl %ebx, %ecx
	leal offset_of_String.str_field(%eax), %esi
	leal offset_of_String.str_field(%edx), %edi
	cld
	repe cmpsb
	jne 1f

	lea boolean_true, %eax

	jmp 2f

1:
	lea boolean_false, %eax

2:
	movl 8(%ebp), %ebx
	cmpl $0, gc_offset(%ebx)
	jle 3f
	decl gc_offset(%ebx)
3:
	movl 12(%ebp), %ebx
	cmpl $0, gc_offset(%ebx)
	jle 4f
	decl gc_offset(%ebx)
4:
	leave
	.cfi_def_cfa esp, 4
	ret $8
	.cfi_endproc
	.size String.equals, .-String.equals

.globl String.concat
.type String.concat, @function
String.concat:
	.cfi_startproc
	push %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register ebp
	subl $4, %esp

	movl 8(%ebp), %eax
	test %eax, %eax
	jz runtime.null_panic

	movl 12(%ebp), %ebx
	movl offset_of_String.length(%eax), %eax
	movl offset_of_String.length(%ebx), %ebx
	movl offset_of_Int.value(%eax), %eax
	movl offset_of_Int.value(%ebx), %ebx
	addl %eax, %ebx
	movl %ebx, -4(%ebp)

	movl $(size_of_Int + 4), %eax
	movl $tag_of_Int, %ebx
	call gc_alloc

	movl -4(%ebp), %ebx
	movl %ebx, offset_of_Int.value(%eax)
	movl %eax, -4(%ebp)

	movl %ebx, %eax
	addl $size_of_String, %eax
	movl $tag_of_String, %ebx
	call gc_alloc

	movl -4(%ebp), %ebx
	movl %ebx, offset_of_String.length(%eax)
	leal offset_of_String.str_field(%eax), %edi
	decl gc_offset(%ebx)

	movl 12(%ebp), %ebx
	movl offset_of_String.length(%ebx), %ecx
	leal offset_of_String.str_field(%ebx), %esi
	movl offset_of_Int.value(%ecx), %ecx
	cld
	rep movsb

	movl 8(%ebp), %ebx
	movl offset_of_String.length(%ebx), %ecx
	leal offset_of_String.str_field(%ebx), %esi
	movl offset_of_Int.value(%ecx), %ecx
	cld
	rep movsb

	movl 8(%ebp), %ebx
	cmpl $0, gc_offset(%ebx)
	jle 1f
	decl gc_offset(%ebx)
1:
	movl 12(%ebp), %ebx
	cmpl $0, gc_offset(%ebx)
	jle 2f
	decl gc_offset(%ebx)
2:

	leave
	.cfi_def_cfa esp, 4
	ret $8
	.cfi_endproc
	.size String.concat, .-String.concat

.type String._check_bounds, @function
String._check_bounds:
	.cfi_startproc
	movl offset_of_String.length(%eax), %ecx
	movl offset_of_Int.value(%ecx), %ecx
	movl offset_of_Int.value(%ebx), %edx

	cmpl %edx, %ecx
	jbe runtime.bounds_panic

	ret
	.cfi_endproc
	.size String._check_bounds, .-String._check_bounds

.globl String.substring
.type String.substring, @function
String.substring:
	.cfi_startproc
	push %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register ebp
	subl $4, %esp

	movl 16(%ebp), %eax
	movl 12(%ebp), %ebx
	call String._check_bounds

	movl %edx, %ecx
	movl 8(%ebp), %ebx
	movl offset_of_Int.value(%ebx), %edx

	subl %ecx, %edx
	jb runtime.bounds_panic
	movl %edx, -4(%ebp)

	movl $(size_of_Int + 4), %eax
	movl $tag_of_Int, %ebx
	call gc_alloc
	movl -4(%ebp), %edx
	movl %edx, offset_of_Int.value(%eax)
	movl %eax, -4(%ebp)

	movl %edx, %eax
	addl $size_of_String, %eax
	movl $tag_of_String, %ebx
	call gc_alloc

	movl -4(%ebp), %ebx
	movl %ebx, offset_of_String.length(%eax)
	decl gc_offset(%ebx)

	leal offset_of_String.str_field(%eax), %edi
	movl offset_of_Int.value(%ebx), %ecx
	movl 12(%ebp), %ebx
	movl offset_of_Int.value(%ebx), %ebx
	movl 16(%ebp), %edx
	addl %ebx, %edx
	leal offset_of_String.str_field(%edx), %esi

	cld
	rep movsb

	movl 8(%ebp), %ebx
	cmpl $0, gc_offset(%ebx)
	jle 1f
	decl gc_offset(%ebx)
1:
	movl 12(%ebp), %ebx
	cmpl $0, gc_offset(%ebx)
	jle 2f
	decl gc_offset(%ebx)
2:
	movl 16(%ebp), %ebx
	cmpl $0, gc_offset(%ebx)
	jle 3f
	decl gc_offset(%ebx)
3:

	leave
	.cfi_def_cfa esp, 4
	ret $12
	.cfi_endproc
	.size String.substring, .-String.substring

.globl String.charAt
.type String.charAt, @function
String.charAt:
	.cfi_startproc
	push %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register ebp
	subl $4, %esp

	movl 12(%ebp), %eax
	movl 8(%ebp), %ebx
	call String._check_bounds

	addl %edx, %eax
	movl $0, %edx
	movb offset_of_String.str_field(%eax), %dl

	movl %edx, -4(%ebp)

	movl 8(%ebp), %ebx
	cmpl $0, gc_offset(%ebx)
	jle 1f
	decl gc_offset(%ebx)
1:
	movl 12(%ebp), %ebx
	cmpl $0, gc_offset(%ebx)
	jle 2f
	decl gc_offset(%ebx)
2:

	movl $(size_of_Int + 4), %eax
	movl $tag_of_Int, %ebx
	call gc_alloc
	movl -4(%ebp), %ebx
	movl %ebx, offset_of_Int.value(%eax)

	leave
	.cfi_def_cfa esp, 4
	ret $8
	.cfi_endproc
	.size String.charAt, .-String.charAt

.type ArrayAny._check_bounds, @function
ArrayAny._check_bounds:
	.cfi_startproc
	movl offset_of_ArrayAny.length(%eax), %ecx
	movl offset_of_Int.value(%ecx), %ecx
	movl offset_of_Int.value(%ebx), %edx

	cmpl %edx, %ecx
	jbe runtime.bounds_panic

	shll $2, %edx
	addl %edx, %eax
	leal offset_of_ArrayAny.array_field(%eax), %eax

	ret
	.cfi_endproc
	.size ArrayAny._check_bounds, .-ArrayAny._check_bounds

.globl ArrayAny.get
.type ArrayAny.get, @function
ArrayAny.get:
	.cfi_startproc
	push %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register ebp
	subl $0, %esp

	movl 12(%ebp), %eax
	movl 8(%ebp), %ebx
	call ArrayAny._check_bounds

	movl (%eax), %eax
	test %eax, %eax
	jz 1f
	cmpl $0, gc_offset(%eax)
	jl 1f
	incl gc_offset(%eax)
1:
	movl 8(%ebp), %ebx
	cmpl $0, gc_offset(%ebx)
	jle 2f
	decl gc_offset(%ebx)
2:
	movl 12(%ebp), %ebx
	cmpl $0, gc_offset(%ebx)
	jle 3f
	decl gc_offset(%ebx)
3:

	leave
	.cfi_def_cfa esp, 4
	ret $8
	.cfi_endproc
	.size ArrayAny.get, .-ArrayAny.get

.globl ArrayAny.set
.type ArrayAny.set, @function
ArrayAny.set:
	.cfi_startproc
	push %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register ebp
	subl $0, %esp

	movl 16(%ebp), %eax
	movl 12(%ebp), %ebx
	call ArrayAny._check_bounds

	movl 8(%ebp), %ecx
	movl (%eax), %ebx
	movl %ecx, (%eax)
	movl %ebx, %eax
	test %eax, %eax
	jz 1f
	cmpl $0, gc_offset(%eax)
	jl 1f
	incl gc_offset(%eax)
1:
	movl 8(%ebp), %ebx
	test %ebx, %ebx
	jz 2f
	cmpl $0, gc_offset(%ebx)
	jle 2f
	decl gc_offset(%ebx)
2:
	movl 12(%ebp), %ebx
	cmpl $0, gc_offset(%ebx)
	jle 3f
	decl gc_offset(%ebx)
3:
	movl 16(%ebp), %ebx
	cmpl $0, gc_offset(%ebx)
	jle 4f
	decl gc_offset(%ebx)
4:

	leave
	.cfi_def_cfa esp, 4
	ret $12
	.cfi_endproc
	.size ArrayAny.set, .-ArrayAny.set

.globl ArrayAny.ArrayAny
.type ArrayAny.ArrayAny, @function
ArrayAny.ArrayAny:
	.cfi_startproc
	push %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register ebp
	subl $0, %esp

	// top 3 bits of length must be unset
	movl 8(%ebp), %eax
	movl offset_of_Int.value(%eax), %eax
	test $0xE0000000, %eax
	jnz runtime.bounds_panic

	// make sure length doesn't overflow
	shll $2, %eax
	addl $size_of_ArrayAny, %eax
	jc runtime.bounds_panic

	// get the actual size
	movl 12(%ebp), %ebx
	movl size_offset(%ebx), %ecx

	// check if we're big enough
	cmpl %eax, %ecx
	jl 1f

	// we're big enough, so let's keep going
	movl 12(%ebp), %eax

2:
	movl 8(%ebp), %ebx
	movl %ebx, offset_of_ArrayAny.length(%eax)

	cmpl $0, gc_offset(%ebx)
	jle 3f
	decl gc_offset(%ebx)

	jmp 3f
1:
	// let the old one be garbage collected
	movl 12(%ebp), %ecx
	cmpl $0, gc_offset(%ecx)
	jle 4f
	decl gc_offset(%ecx)
4:

	// %eax already holds the size we need
	movl $tag_of_ArrayAny, %ebx
	call gc_alloc

	jmp 2b
3:

	leave
	.cfi_def_cfa esp, 4
	ret $8
	.cfi_endproc
	.size ArrayAny.ArrayAny, .-ArrayAny.ArrayAny
