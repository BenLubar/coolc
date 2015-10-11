.include "basic_defs.s"

.data

.align 2
gc_heap_start:
	.long 0
.align 2
gc_heap_end:
	.long 0

.align 2
symbol:
	.long 0

.set gc_increase_heap_size, 0x1000
.set max_line_length, 0x400

.text

gc_init:
	call runtime.heap_get
	movl %eax, gc_heap_start
	movl %eax, gc_heap_end
	call gc_increase_heap
	ret $0

gc_increase_heap:
	movl gc_heap_end, %eax
	addl $gc_increase_heap_size, %eax
	push %eax
	call runtime.heap_set
	movl %eax, gc_heap_end
	ret $0

.globl gc_alloc
gc_alloc:
	movl %eax, %ecx

	test $3, %ecx
	jz 1f

	andl $-4, %ecx
	addl $4, %ecx

1:
	movl gc_heap_start, %eax
	movl gc_heap_end, %edx
	subl %ecx, %edx
	subl $data_offset, %edx

2:
	cmpl %edx, %eax
	jle 3f

	cmpl $0, tag_offset(%eax)
	je 4f
	cmpl $-1, tag_offset(%eax)
	jne 5f

	cmpl %ecx, size_offset(%eax)
	jl 6f

	addl $data_offset, %ecx
	cmpl %ecx, size_offset(%eax)
	jl 8f

	push %ebx
	movl size_offset(%eax), %ebx
	subl %ecx, %ebx
	addl %ecx, %eax
	movl $-1, tag_offset(%eax)
	movl %ebx, size_offset(%eax)
	movl $-1, ref_offset(%eax)
	subl %ecx, %eax
	subl $data_offset, %ecx
	pop %ebx

	jmp 4f

8:
	movl size_offset(%eax), %ecx
	jmp 4f

6:
	push %eax
	addl size_offset(%eax), %eax
	addl $data_offset, %eax
	cmpl $-1, tag_offset(%eax)
	je 7f
	cmpl $0, tag_offset(%eax)
	je 7f
	pop %eax

5:
	addl size_offset(%eax), %eax
	addl $data_offset, %eax

	jmp 2b

7:
	push %ebx
	movl size_offset(%eax), %ebx
	addl $data_offset, %ebx

	movl 4(%esp), %eax
	addl %ebx, size_offset(%eax)
	addl 0(%esp), %ebx
	addl $8, %esp

	jmp 2b

3:
	push %eax
	push %ebx
	call gc_increase_heap
	pop %ebx
	pop %eax
	addl $gc_increase_heap_size, %edx

	jmp 2b

4:
	movl %ebx, tag_offset(%eax)
	movl %ecx, size_offset(%eax)
	movl $0, ref_offset(%eax)
	leal data_offset(%eax), %edi
	movl %eax, %ebx
	movb $0, %al
	cld
	rep stosb
	movl %ebx, %eax
	ret $0

.globl _start
_start:
	call gc_init

	call main

	push $0
	call runtime.exit

.globl Any.toString
Any.toString:
	enter $0, $0

	movl 8(%ebp), %eax
	movl tag_offset(%eax), %eax
	shll $2, %eax
	movl class_names(%eax), %eax

	leave
	ret $4

.globl Any.equals
Any.equals:
	enter $0, $0

	movl 8(%ebp), %eax
	movl 12(%ebp), %ebx

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
	push %eax
	call runtime.output

	movl 12(%ebp), %eax

	leave
	ret $8

.globl IO.in
IO.in:
	enter $0, $0

	movl $(size_of_Int + 4), %eax
	movl $tag_of_Int, %ebx
	call gc_alloc
	movl $(size_of_String + max_line_length), offset_of_Int.value(%eax)
	push %eax

	movl $(size_of_String + max_line_length), %eax
	movl $tag_of_String, %ebx
	call gc_alloc

	pop %ebx
	movl %ebx, offset_of_String.length(%eax)

	push %eax
	call runtime.input

	leave
	ret $4

.globl IO.symbol
IO.symbol:
	enter $0, $0

	movl 8(%ebp), %eax
	test %eax, %eax
	jz runtime.null_panic

	lea symbol, %ebx
	movl symbol, %ecx
	movl $0, %edx

1:
	test %ecx, %ecx
	jz 2f

	push %ebx
	push %ecx
	push %edx
	movl 8(%ebp), %eax
	push %eax
	movl offset_of_Symbol.name(%ecx), %eax
	push %eax
	call String.equals
	addl $8, %esp
	lea boolean_true, %ebx
	cmpl %ebx, %eax
	je 3f
	pop %edx
	pop %ecx
	pop %ebx

	lea offset_of_Symbol.next(%ecx), %ebx
	movl offset_of_Symbol.next(%ecx), %ecx
	incl %edx
	jmp 1b

2:
	push %ebx
	push %edx
	movl $(size_of_Int + 4), %eax
	movl $tag_of_Int, %ebx
	call gc_alloc
	pop %edx
	movl %edx, offset_of_Int.value(%eax)
	push %eax
	movl $size_of_Symbol, %eax
	movl $tag_of_Symbol, %ebx
	call gc_alloc
	pop %edx
	movl %edx, offset_of_Symbol.hash(%eax)
	movl 8(%ebp), %ecx
	movl %ecx, offset_of_Symbol.name(%eax)
	pop %ebx
	movl %eax, (%ebx)

	leave
	ret $8

3:
	pop %edx
	pop %ecx
	pop %ebx
	movl %ecx, %eax

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

	leave
	ret $8

1:
	lea boolean_false, %eax

	leave
	ret $8

.globl String.concat
String.concat:
	enter $0, $0

	movl 8(%ebp), %eax
	test %eax, %eax
	jz runtime.null_panic

	movl 12(%ebp), %ebx
	movl offset_of_String.length(%eax), %eax
	movl offset_of_String.length(%ebx), %ebx
	movl offset_of_Int.value(%eax), %eax
	movl offset_of_Int.value(%ebx), %ebx
	addl %eax, %ebx
	push %ebx

	movl $(size_of_Int + 4), %eax
	movl $tag_of_Int, %ebx
	call gc_alloc

	pop %ebx
	movl %ebx, offset_of_Int.value(%eax)
	push %eax

	movl %ebx, %eax
	addl $size_of_String, %eax
	movl $tag_of_String, %ebx
	call gc_alloc

	pop %ebx
	movl %ebx, offset_of_String.length(%eax)
	leal offset_of_String.str_field(%eax), %edi

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
	enter $0, $0

	movl 16(%ebp), %eax
	movl 12(%ebp), %ebx
	call String._check_bounds

	movl %edx, %ecx
	movl 8(%ebp), %ebx
	movl offset_of_Int.value(%ebx), %edx

	subl %ecx, %edx
	jb runtime.bounds_panic
	push %ecx
	push %edx

	movl $(size_of_Int + 4), %eax
	movl $tag_of_Int, %ebx
	call gc_alloc
	pop %edx
	movl %edx, offset_of_Int.value(%eax)
	push %eax

	movl %edx, %eax
	addl $size_of_String, %eax
	movl $tag_of_String, %ebx
	call gc_alloc

	pop %ebx
	movl %ebx, offset_of_String.length(%eax)
	leal offset_of_String.str_field(%eax), %edi
	movl offset_of_Int.value(%ebx), %ecx
	movl 12(%ebp), %ebx
	movl offset_of_Int.value(%ebx), %ebx
	movl 16(%ebp), %edx
	addl %ebx, %edx
	leal offset_of_String.str_field(%edx), %esi

	cld
	rep movsb

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

	leave
	ret $12

.globl ArrayAny.ArrayAny
ArrayAny.ArrayAny:
	enter $0, $0

	movl 8(%ebp), %eax
	movl offset_of_Int.value(%eax), %eax
	cmpl $0, %eax
	jl runtime.bounds_panic
	shll $2, %eax
	addl $4, %eax
	cmpl $0, %eax
	jl runtime.bounds_panic

	movl 12(%ebp), %ebx
	movl size_offset(%ebx), %ebx

	cmpl %eax, %ebx
	jl 1f

	movl 12(%ebp), %eax

2:
	movl 8(%ebp), %ebx
	movl %ebx, offset_of_ArrayAny.length(%eax)

	leave
	ret $8

1:
	movl $tag_of_ArrayAny, %ebx
	call gc_alloc

	jmp 2b
