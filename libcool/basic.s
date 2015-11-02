.include "basic_defs.s"

.data

.align 2
symbol:
	.long 0

.set max_line_length, 0x400

.text

.globl _start
_start:
	call gc_init

	call main

	push $0
	call runtime.exit

.globl Any.toString
Any.toString:
	enter $0, $0

	movl 8(%ebp), %ebx
	movl tag_offset(%ebx), %eax
	shll $2, %eax
	movl class_names(%eax), %eax

	cmpl $0, gc_offset(%ebx)
	jle 1f
	decl gc_offset(%ebx)
1:

	leave
	ret $4

.globl Any.equals
Any.equals:
	enter $0, $0

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

	leave
	ret $8

1:
	leal boolean_false, %eax

	leave
	ret $8

.globl IO.abort
IO.abort:
	enter $0, $0

	movl 8(%ebp), %eax
	push %eax
	call runtime.output

	push $1
	call runtime.exit

.globl IO.out
IO.out:
	enter $0, $0

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
	ret $8

.globl IO.in
IO.in:
	enter $4, $0

	movl $(size_of_Int + 4), %eax
	movl $tag_of_Int, %ebx
	call gc_alloc
	movl $(size_of_String + max_line_length), offset_of_Int.value(%eax)
	movl %eax, -4(%ebp)

	movl $(size_of_String + max_line_length), %eax
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
	ret $4

.globl IO.symbol
IO.symbol:
	enter $12, $0

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

	leave
	ret $8

4:
	// let the string be garbage collected
	movl 8(%ebp), %eax
	cmpl $0, gc_offset(%eax)
	jle 5f
	decl gc_offset(%eax)
5:
	// grab the symbol we found
	movl -8(%ebp), %eax

	leave
	ret $8

.globl IO.symbol_name
IO.symbol_name:
	enter $0, $0

	movl 8(%ebp), %eax
	test %eax, %eax
	jz runtime.null_panic
	movl offset_of_Symbol.name(%eax), %eax

	leave
	ret $8

.globl Int.equals
Int.equals:
	enter $0, $0

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

	leave
	ret $8

1:
	leal boolean_false, %eax

	leave
	ret $8

.globl String.equals
String.equals:
	enter $0, $0

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
	ret $8

.globl String.concat
String.concat:
	enter $4, $0

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
	ret $8

String._check_bounds:
	movl offset_of_String.length(%eax), %ecx
	movl offset_of_Int.value(%ecx), %ecx
	movl offset_of_Int.value(%ebx), %edx

	cmpl %edx, %ecx
	jbe runtime.bounds_panic

	ret

.globl String.substring
String.substring:
	enter $4, $0

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
	ret $12

.globl String.charAt
String.charAt:
	enter $0, $0

	movl 12(%ebp), %eax
	movl 8(%ebp), %ebx
	call String._check_bounds

	addl %edx, %eax
	movl $0, %edx
	movb offset_of_String.str_field(%eax), %dl

	shll $2, %edx
	movl byte_ints(%edx), %eax

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
	ret $8

ArrayAny._check_bounds:
	movl offset_of_ArrayAny.length(%eax), %ecx
	movl offset_of_Int.value(%ecx), %ecx
	movl offset_of_Int.value(%ebx), %edx

	cmpl %edx, %ecx
	jbe runtime.bounds_panic

	shll $2, %edx
	addl %edx, %eax
	leal offset_of_ArrayAny.array_field(%eax), %eax

	ret

.globl ArrayAny.get
ArrayAny.get:
	enter $0, $0

	movl 12(%ebp), %eax
	movl 8(%ebp), %ebx
	call ArrayAny._check_bounds

	movl (%eax), %eax
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
	ret $8

.globl ArrayAny.set
ArrayAny.set:
	enter $0, $0

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
	ret $12

.globl ArrayAny.ArrayAny
ArrayAny.ArrayAny:
	enter $0, $0

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
3:

	leave
	ret $8

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
