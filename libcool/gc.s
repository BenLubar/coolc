.include "basic_defs.s"

.data

// tag_offset
.globl tag_of_garbage
.set tag_of_garbage, -1
.globl tag_of_raw
.set tag_of_raw, -2

// gc_offset
.globl gc_tag_live
.set gc_tag_live, -3
.globl gc_tag_root
.set gc_tag_root, -2
.globl gc_tag_garbage
.set gc_tag_garbage, -1
.globl gc_tag_none
.set gc_tag_none, 0

.align 2
gc_heap_start:
	.long 0
.align 2
gc_heap_end:
	.long 0

.set gc_increase_heap_size, 0x1000

.text

.globl gc_init
.type gc_init, @function
gc_init:
	.cfi_startproc
	call runtime.heap_get
	movl %eax, gc_heap_start
	movl %eax, gc_heap_end
	call gc_increase_heap
	movl gc_heap_start, %eax
	movl $0, tag_offset(%eax)
	ret $0
	.cfi_endproc
	.size gc_init, .-gc_init

.type gc_increase_heap, @function
gc_increase_heap:
	.cfi_startproc
	movl gc_heap_end, %eax
	addl $gc_increase_heap_size, %eax
	push %eax
	.cfi_adjust_cfa_offset -4
	call runtime.heap_set
	.cfi_adjust_cfa_offset 4
	movl %eax, gc_heap_end
	ret $0
	.cfi_endproc
	.size gc_increase_heap, .-gc_increase_heap

.globl gc_alloc
.type gc_alloc, @function
gc_alloc:
	.cfi_startproc
	push %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register ebp
	subl $12, %esp

	// make sure it's aligned
	addl $3, %eax
	andl $-4, %eax

	movl %eax, %ecx
	movl %ebx, -4(%ebp)
	movl $0, -12(%ebp)

1:
	// eax = current pointer
	// ebx = scratch
	// ecx = requested size
	// edx = highest address we can return
	movl gc_heap_start, %eax
	movl gc_heap_end, %edx
	subl %ecx, %edx
	subl $data_offset, %edx

2:
	// did we run out of space?
	cmpl %eax, %edx
	jg 3f

	// only collect garbage once. otherwise, increase the size of the heap.
	cmpl $0, -12(%ebp)
	jne 9f

	movl $1, -12(%ebp)

	// collect garbage
	movl %ecx, -8(%ebp)
	call runtime.gc_collect
	movl -8(%ebp), %ecx
	jmp 1b

9:
	// we ran out of space. make more space and start over.
	call gc_increase_heap
	jmp 1b

3:
	// are we in clean memory?
	cmpl $0, tag_offset(%eax)
	je 8f

	// if we're not looking at garbage, skip to the next one.
	cmpl $tag_of_garbage, tag_offset(%eax)
	jne 5f

	// check if it's big enough.
	cmpl %ecx, size_offset(%eax)
	jl 6f

	// check if it's too big.
	addl $data_offset, %ecx
	cmpl %ecx, size_offset(%eax)
	jge 7f

	// use up the whole space if we're one or two words off.
	movl size_offset(%eax), %ecx
	jmp 4f

7:
	// it's too big. split it up.
	movl %ecx, -8(%ebp)
	movl %eax, %ebx
	addl %ecx, %ebx
	movl $tag_of_garbage, tag_offset(%ebx)
	movl size_offset(%eax), %ecx
	movl %ecx, size_offset(%ebx)
	movl -8(%ebp), %ecx
	subl %ecx, size_offset(%ebx)
	movl $gc_tag_garbage, gc_offset(%ebx)
	subl $data_offset, %ecx

	jmp 4f

6:
	// is the next object also garbage?
	movl %eax, %ebx
	addl size_offset(%eax), %ebx
	addl $data_offset, %ebx
	cmpl $tag_of_garbage, tag_offset(%ebx)
	jne 5f

	// it is. join them.
	movl %ecx, -8(%ebp)
	movl size_offset(%ebx), %ecx
	addl $data_offset, %ecx
	addl %ecx, size_offset(%eax)
	movl -8(%ebp), %ecx

	jmp 2b

5:
	// advance %eax by one object and try again.
	addl size_offset(%eax), %eax
	addl $data_offset, %eax
	jmp 2b

8:
	leal data_offset(%eax), %edx
	addl %ecx, %edx
	movl $0, tag_offset(%edx)

4:
	// we found enough space! set the meta-fields, then zero out the rest.
	movl -4(%ebp), %ebx
	movl %ebx, tag_offset(%eax)
	movl %ecx, size_offset(%eax)
	movl $1, gc_offset(%eax)
	movl %eax, -4(%ebp)
	leal data_offset(%eax), %edi
	movl $0, %eax
	cld
	rep stosb

	movl -4(%ebp), %eax

	//call gc_check

	leave
	.cfi_def_cfa esp, 4
	ret $0
	.cfi_endproc
	.size gc_alloc, .-gc_alloc

.globl gc_collect
.type gc_collect, @function
gc_collect:
	.cfi_startproc
	push %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register ebp
	subl $12, %esp

	//movl $0, %eax
	//call gc_check

1:
	// starting state: all objects on the heap are one of:
	// - gc_tag_garbage (garbage)
	// - gc_tag_root (a garbage collection root)
	// - [a positive integer] (referenced on the stack)
	// - gc_tag_none (not referenced on the stack)
	//
	// The first is not touched, the second and third make up the starting
	// gray set, and the last is the starting white set. We now make all
	// the gray objects black by making white objects they point to gray.
	//
	// Objects not on the heap are always roots and never point to non-root
	// objects, so they do not need to be touched.

	movl gc_heap_start, %eax
2:
	// check the tag
	movl tag_offset(%eax), %ebx
	cmpl $0, %ebx
	je 6f
	jl 5f

	// skip anything that isn't a root or on the stack.
	movl gc_offset(%eax), %ecx
	cmpl $gc_tag_live, %ecx
	je 5f
	cmpl $gc_tag_none, %ecx
	je 5f

	// get the number of pointers
	movl %ebx, %ecx
	shll $2, %ecx
	movl gc_sizes(%ecx), %ecx
	leal data_offset(%eax), %edx

	// save eax
	movl %eax, -4(%ebp)

	// ArrayAny is a special case with more pointers
	cmpl $tag_of_ArrayAny, %ebx
	jne 3f
	movl offset_of_ArrayAny.length(%eax), %ebx
	test %ebx, %ebx
	jz 3f
	addl offset_of_Int.value(%ebx), %ecx
3:
	// for each pointer...
	test %ecx, %ecx
	jz 4f

	// save ecx and edx
	movl %ecx, -8(%ebp)
	movl %edx, -12(%ebp)

	// mark!
	push (%edx)
	call gc_mark

	// restore ecx and edx
	movl -8(%ebp), %ecx
	movl -12(%ebp), %edx

	// go to the next pointer
	decl %ecx
	addl $4, %edx
	jmp 3b

4:
	// restore eax
	movl -4(%ebp), %eax

5:
	// move to the next object
	addl size_offset(%eax), %eax
	addl $data_offset, %eax
	jmp 2b

6:
	// sweep phase: a single pass through the heap. anything unmarked is
	// marked as garbage. anything marked as live is unmarked. if anything
	// was marked as garbage, start over from the beginning.

	movl gc_heap_start, %eax
	movl $0, %ebx
7:
	cmpl $0, tag_offset(%eax)
	je 10f

	// collect garbage
	cmpl $gc_tag_none, gc_offset(%eax)
	jne 8f

	movl $tag_of_garbage, tag_offset(%eax)
	movl $gc_tag_garbage, gc_offset(%eax)
	movl $1, %ebx

	jmp 9f
8:
	// unmark live
	cmpl $gc_tag_live, gc_offset(%eax)
	jne 9f

	movl $gc_tag_none, gc_offset(%eax)

9:
	// move to the next object
	addl size_offset(%eax), %eax
	addl $data_offset, %eax
	jmp 7b

10:
	test %ebx, %ebx
	jnz 1b

	//movl $0, %eax
	//call gc_check

	leave
	.cfi_def_cfa esp, 4
	ret $0
	.cfi_endproc
	.size gc_collect, .-gc_collect

.type gc_mark, @function
gc_mark:
	.cfi_startproc
	push %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register ebp
	subl $8, %esp

	// grab the argument
	movl 8(%ebp), %eax

	// don't touch it if it's null
	test %eax, %eax
	jz 2f

	// don't touch it if it's not a white object
	cmpl $gc_tag_none, gc_offset(%eax)
	jne 2f

	// mark it live. we consider it gray until this function returns,
	// then it is black.
	movl $gc_tag_live, gc_offset(%eax)

	// get the number of pointers
	movl tag_offset(%eax), %ebx
	movl %ebx, %ecx
	shll $2, %ecx
	movl gc_sizes(%ecx), %ecx
	leal data_offset(%eax), %edx

	// ArrayAny is a special case with more pointers
	cmpl $tag_of_ArrayAny, %ebx
	jne 1f
	movl offset_of_ArrayAny.length(%eax), %ebx
	test %ebx, %ebx
	jz 1f
	addl offset_of_Int.value(%ebx), %ecx
1:
	// for each pointer...
	test %ecx, %ecx
	jz 2f

	// save ecx and edx
	movl %ecx, -4(%ebp)
	movl %edx, -8(%ebp)

	// mark!
	push (%edx)
	call gc_mark

	// restore ecx and edx
	movl -4(%ebp), %ecx
	movl -8(%ebp), %edx

	// go to the next pointer
	decl %ecx
	addl $4, %edx
	jmp 1b

2:

	leave
	.cfi_def_cfa esp, 4
	ret $4
	.cfi_endproc
	.size gc_mark, .-gc_mark

.globl gc_check
.type gc_check, @function
gc_check:
	.cfi_startproc
	push %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register ebp
	subl $4, %esp
	movl %eax, -4(%ebp)

	movl gc_heap_start, %eax
1:
	cmpl $0, tag_offset(%eax)
	je 18f
	jg 6f

	// it is garbage
	cmpl $tag_of_garbage, tag_offset(%eax)
	je 2f

	// or raw?
	cmpl $tag_of_raw, tag_offset(%eax)
	je 19f

	// garbage has invalid tag
	int $3
2:
	cmpl $gc_tag_garbage, gc_offset(%eax)
	je 3f
	// garbage has invalid GC tag
	int $3
19:
	cmpl $gc_tag_root, gc_offset(%eax)
	je 3f
	// raw has invalid GC tag
	int $3
3:
	movl size_offset(%eax), %ebx
	cmpl $0, %ebx
	jge 4f
	// garbage has negative size
	int $3
4:
	test $3, %ebx
	jz 5f
	// garbage has non-aligned size
	int $3
5:
	// go to the next object
	addl size_offset(%eax), %eax
	addl $data_offset, %eax
	jmp 1b

6:
	cmpl $max_tag, tag_offset(%eax)
	jle 7f
	// invalid tag
	int $3
7:
	cmpl $gc_tag_root, gc_offset(%eax)
	je 8f
	cmpl $0, gc_offset(%eax)
	jge 8f
	// invalid GC tag
	int $3
8:
	movl size_offset(%eax), %ebx
	cmpl $0, %ebx
	jge 9f
	// object has negative size
	int $3
9:
	test $3, %ebx
	jz 10f
	// object has non-aligned size
	int $3
10:
	movl tag_offset(%eax), %ebx
	shll $2, %ebx
	movl gc_sizes(%ebx), %ebx
	movl %ebx, %ecx
	shll $2, %ecx
	cmpl $tag_of_Int, tag_offset(%eax)
	je 12f
	cmpl $tag_of_String, tag_offset(%eax)
	je 13f
	cmpl $tag_of_ArrayAny, tag_offset(%eax)
	je 14f
11:
	movl %eax, %edx
	addl $data_offset, %edx
	cmpl %ecx, size_offset(%eax)
	jge 15f
	// object is too small for contents
	int $3
	jmp 15f
12:
	// special case: Int has 4 extra bytes
	addl $4, %ecx
	jmp 11b
13:
	// special case: String has (length) extra bytes
	movl offset_of_String.length(%eax), %edx
	test %edx, %edx
	jz 11b
	addl offset_of_Int.value(%edx), %ecx
	jmp 11b
14:
	// special case: ArrayAny has (length) extra pointers
	movl offset_of_ArrayAny.length(%eax), %edx
	test %edx, %edx
	jz 11b
	addl offset_of_Int.value(%edx), %ebx
	movl %ebx, %ecx
	shll $2, %ecx
	jmp 11b
15:
	// check each pointer
	test %ebx, %ebx
	jz 5b
	decl %ebx

	movl (%edx), %ecx
	test %ecx, %ecx
	// don't follow null pointer
	jz 17f

	cmpl $0, tag_offset(%ecx)
	jg 16f
	// pointer is garbage
	int $3
16:
	cmpl $max_tag, tag_offset(%ecx)
	jle 17f
	// pointer has invalid tag (probably garbage)
	int $3
17:
	// go to the next pointer
	addl $4, %edx
	jmp 15b

18:
	movl -4(%ebp), %eax
	leave
	.cfi_def_cfa esp, 4
	ret $0
	.cfi_endproc
	.size gc_check, .-gc_check
