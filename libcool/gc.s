.include "basic_defs.s"

.data

// tag_offset
.globl tag_of_garbage
.set tag_of_garbage, -1

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
	enter $8, $0

	movl %eax, -4(%ebp)
	movl %ebx, -8(%ebp)
	call gc_collect
	movl -4(%ebp), %eax
	movl -8(%ebp), %ebx

	// make sure it's aligned
	addl $3, %eax
	andl $-4, %eax

	movl %eax, %ecx
	movl %ebx, -4(%ebp)

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
	jle 3f

	// are we in clean memory?
	cmpl $0, tag_offset(%eax)
	je 4f

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

	call gc_check

	leave
	ret $0

3:
	// we ran out of space. make more space and start over.
	call gc_increase_heap
	jmp 1b

gc_collect:
	enter $4, $0

	movl $0, %eax
	call gc_check

1:
	// while we're finding new references:
	movl $0, %ebx
	movl gc_heap_start, %eax

2:
	// mark references
	cmpl $0, tag_offset(%eax)
	je 8f
	cmpl $gc_tag_garbage, gc_offset(%eax)
	je 7f
	cmpl $gc_tag_none, gc_offset(%eax)
	je 7f
	movl tag_offset(%eax), %ecx
	shll $2, %ecx
	movl gc_sizes(%ecx), %ecx
	cmpl $tag_of_ArrayAny, tag_offset(%eax)
	jne 3f
	movl offset_of_ArrayAny.length(%eax), %edx
	addl offset_of_Int.value(%edx), %ecx

3:
	leal (data_offset - 4)(%eax), %edx
	jmp 6f

4:
	cmpl $0, (%edx)
	je 6f
	movl %edx, -4(%ebp)
	movl (%edx), %edx
	cmpl $gc_tag_none, gc_offset(%edx)
	jne 5f
	movl $gc_tag_live, gc_offset(%edx)
	movl $1, %ebx
5:
	movl -4(%ebp), %edx

6:
	addl $4, %edx
	test %ecx, %ecx
	jz 7f
	subl $1, %ecx
	jmp 4b

7:
	addl size_offset(%eax), %eax
	leal data_offset(%eax), %eax
	jmp 2b

8:
	test %ebx, %ebx
	jnz 1b

	// now we have:
	// gc_tag_garbage, gc_tag_root -> unchanged, don't touch
	// gc_tag_live -> visible from some root, don't touch
	// gc_tag_none -> garbage, free
	// > gc_tag_none -> on a stack somewhere, don't touch

	movl gc_heap_start, %eax

9:
	cmpl $0, tag_offset(%eax)
	je 11f
	cmpl $gc_tag_none, gc_offset(%eax)
	jne 10f
	movl $1, %ebx
	movl $gc_tag_garbage, gc_offset(%eax)
	movl $tag_of_garbage, tag_offset(%eax)

10:
	addl size_offset(%eax), %eax
	leal data_offset(%eax), %eax
	jmp 9b

11:
	movl gc_heap_start, %eax

12:
	// clear GC flags: live->none
	cmpl $0, tag_offset(%eax)
	je 14f
	cmpl $gc_tag_live, gc_offset(%eax)
	jne 13f
	movl $gc_tag_none, gc_offset(%eax)

13:
	addl size_offset(%eax), %eax
	leal data_offset(%eax), %eax
	jmp 12b

14:
	// if we freed anything, go back to the beginning.
	test %ebx, %ebx
	jnz 1b

	movl $0, %eax
	call gc_check

	leave
	ret $0

.globl gc_check
gc_check:
	enter $4, $0
	movl %eax, -4(%ebp)

	movl gc_heap_start, %eax
1:
	cmpl $0, tag_offset(%eax)
	je 18f
	jg 6f

	// it is garbage
	cmpl $tag_of_garbage, tag_offset(%eax)
	je 2f
	// garbage has invalid tag
	int $3
2:
	cmpl $gc_tag_garbage, gc_offset(%eax)
	je 3f
	// garbage has invalid GC tag
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
	ret $0
