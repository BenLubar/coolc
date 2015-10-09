.data

.align 2
gc_heap_start:
	.word 0
gc_heap_alloc:
	.word 0
gc_heap_end:
	.word 0

.globl this
this:
	.word 0

gc_increase_heap_size:
	.word 0x1000

.globl tag_offset
tag_offset:
	.word 0
.globl size_offset
size_offset:
	.word 4
.globl ref_offset
ref_offset:
	.word 8
.globl data_offset
data_offset:
	.word 12

.text

gc_init:
	call runtime.heap_get
	movl %eax, gc_heap_start
	movl %eax, gc_heap_end
	movl %eax, gc_heap_alloc
	call gc_increase_heap
	ret

gc_increase_heap:
	movl gc_heap_end, %eax
	addl $gc_increase_heap_size, %eax
	push %eax
	call runtime.heap_set
	addl 4, %esp
	movl %eax, gc_heap_end
	ret

.globl gc_alloc
gc_alloc:
	movl %eax, %ecx
	movl gc_heap_start, %eax
	movl gc_heap_alloc, %edx
	subl %ecx, %edx
	subl $data_offset, %edx

gc_alloc.loop:
	cmpl %edx, %eax
	jle gc_alloc.expand

	cmpl $0, tag_offset(%eax)
	je gc_alloc.found
	cmpl $-1, tag_offset(%eax)
	jne gc_alloc.next

	cmpl %ecx, size_offset(%eax)
	jl gc_alloc.small

	addl $data_offset, %ecx
	cmpl %ecx, size_offset(%eax)
	jl gc_alloc.use

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

	jmp gc_alloc.found

gc_alloc.use:
	movl size_offset(%eax), %ecx
	jmp gc_alloc.found

gc_alloc.next:
	addl size_offset(%eax), %eax
	addl $data_offset, %eax

	jmp gc_alloc.loop

gc_alloc.expand:
	push %eax
	push %ebx
	call gc_increase_heap
	pop %ebx
	pop %eax
	addl $gc_increase_heap_size, %edx

	jmp gc_alloc.loop

gc_alloc.found:
	movl %ebx, tag_offset(%eax)
	movl %ecx, size_offset(%eax)
	movl $0, ref_offset(%eax)
	movl %eax, %edx
	movl %eax, %ebx
	addl $data_offset, %edx
	movb $0, %al
	rep stosb
	movl %ebx, %eax
	ret

.globl _start
_start:
	call gc_init
	movl size_of_Main, %eax
	movl tag_of_Main, %ebx
	call gc_alloc
	movl %eax, this
	call Main.Main
	push $0
	call runtime.exit

.globl Any.toString
Any.toString:
	leal this, %eax
	movl tag_offset(%eax), %eax
	addl class_names, %eax
	movl (%eax), %eax
	ret

.globl Any.equals
Any.equals:
	movl 8(%ebp), %eax

	test %eax, this
	je Any.equals.false

	leal boolean_true, %eax
	ret

Any.equals.false:
	leal boolean_false, %eax
	ret

.globl IO.abort
IO.abort:
	movl 8(%ebp), %eax
	push %eax
	call runtime.output
	addl $4, %esp

	push $1
	call runtime.exit

.globl IO.out
IO.out:
	movl 8(%ebp), %eax
	push %eax
	call runtime.output
	addl $4, %esp

	movl this, %eax
	ret

.globl IO.in
IO.in:
.globl IO.symbol
IO.symbol:
.globl IO.symbol_name
IO.symbol_name:
.globl Int.toString
Int.toString:
.globl Int.equals
Int.equals:
.globl Boolean.equals
Boolean.equals:
.globl String.equals
String.equals:
.globl String.concat
String.concat:
.globl String.substring
String.substring:
.globl String.charAt
String.charAt:
.globl ArrayAny.get
ArrayAny.get:
.globl ArrayAny.set
ArrayAny.set:
.globl ArrayAny.ArrayAny
ArrayAny.ArrayAny:
