.include "basic_defs.s"

.set offset_of_Coroutine.prev, data_offset + size_of_Coroutine + 0
.set offset_of_Coroutine.next, data_offset + size_of_Coroutine + 4
.set offset_of_Coroutine.segment, data_offset + size_of_Coroutine + 8
.set offset_of_Coroutine.stack, data_offset + size_of_Coroutine + 12
.set real_size_of_Coroutine, size_of_Coroutine + 16

.set min_stack_size, 0x1000

.data

.align 2
current_coroutine:
	.long 0
coroutine_deadlock:
	.long 0
coroutine_gc_stack_pointer:
	.long 0

.text

.globl runtime.morestack
runtime.morestack:
	.cfi_startproc

	// no-op if we're not inside a coroutine.
	cmpl $0, current_coroutine
	jne 1f
	ret
1:

	push %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register ebp
	subl $4, %esp

	// add enough for a function call to something that immediately calls
	// runtime.morestack, and for a function call to gc_alloc.
	addl $(4 + 4 + 4 + 12 + 4 + 4 + 4), %eax

	// since we can't actually change where the stack starts from, as that
	// would make all the saved %ebp values point to garbage, we have to
	// do some magic. if the stack is too small, we inject a call to
	// runtime.lessstack between the actual return address and the new
	// stack's return address. we also have to copy the arguments to the
	// function (the count is stored in %ebx) before the return address,
	// and pop the arguments off the stack before we return.

	movl current_coroutine, %ecx
	movl %esp, %edx
	subl offset_of_Coroutine.segment(%ecx), %edx

	cmpl %eax, %edx
	jg 2f

	// make a new stack
	movl %ebx, -4(%ebp)
	addl $min_stack_size, %eax
	movl $tag_of_raw, %ebx
	call gc_alloc
	movl $gc_tag_root, gc_offset(%eax)

	leal data_offset(%eax), %ebx
	addl size_offset(%eax), %ebx

	// push the old segment
	movl current_coroutine, %ecx
	movl offset_of_Coroutine.segment(%ecx), %edx
	subl $4, %ebx
	movl %edx, (%ebx)

	// push the number of arguments to pop
	movl -4(%ebp), %edx
	subl $4, %ebx
	movl %edx, (%ebx)

	// push our parent's return address
	movl 8(%ebp), %edx
	subl $4, %ebx
	movl %edx, (%ebx)

	// push our parent's %ebp
	movl (%ebp), %edx
	subl $4, %ebx
	movl %edx, (%ebx)

	// store the segment
	movl %eax, offset_of_Coroutine.segment(%ecx)

	// copy arguments
	movl -4(%ebp), %ecx
	leal 12(%ebp), %esi
	leal -4(%ebx), %edi
	std
	rep movsd
	movl %edi, %ebx

	// inject runtime.lessstack between our parent and its parent.
	movl $runtime.lessstack, (%ebx)

	// we continue from the pc in %eax
	movl 4(%ebp), %eax

	// magic part:
	leal 8(%ebx), %ebp
	movl %ebx, %esp
	jmp *%eax

2:

	leave
	.cfi_def_cfa esp, 4
	ret
	.cfi_endproc
	.size runtime.morestack, .-runtime.morestack

runtime.lessstack:
	.cfi_startproc
	.cfi_def_cfa_register ebp

	// we can't touch %eax or we'll lose the return value of the function.

	movl current_coroutine, %ebx
	movl offset_of_Coroutine.segment(%ebx), %ecx
	movl $tag_of_garbage, tag_offset(%ecx)
	movl $gc_tag_garbage, gc_offset(%ecx)
	movl 12(%ebp), %ecx
	movl %ecx, offset_of_Coroutine.segment(%ebx)
	movl 8(%ebp), %ecx
	incl %ecx
	shll $2, %ecx
	movl 4(%ebp), %ebx

	leave
	.cfi_def_cfa esp, 4
	subl %ecx, %esp
	jmp *%ebx
	.cfi_endproc
	.size runtime.lessstack, .-runtime.lessstack

.globl runtime.sched
runtime.sched:
	.cfi_startproc
	push %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register ebp
	subl $0, %esp

	movl current_coroutine, %eax
	movl %ebp, offset_of_Coroutine.stack(%eax)
	movl coroutine_deadlock, %ebx
	test %ebx, %ebx
	jz 1f
	cmpl %eax, %ebx
	je runtime.deadlock_panic
	jmp 2f
1:
	movl %eax, coroutine_deadlock
2:
	movl offset_of_Coroutine.next(%eax), %eax
	movl %eax, current_coroutine

runtime.sched0:
	movl current_coroutine, %eax
	movl offset_of_Coroutine.stack(%eax), %ebp
	movl %ebp, %esp

	leave
	.cfi_def_cfa esp, 4
	ret
	.cfi_endproc
	.size runtime.sched, .-runtime.sched

.globl Coroutine.Coroutine
Coroutine.Coroutine:
	.cfi_startproc

	movl $8, %eax
	movl $2, %ebx
	call runtime.morestack

	push %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register ebp
	subl $0, %esp

	// get the Coroutine
	movl 12(%ebp), %eax

	// check if we're big enough
	cmpl $real_size_of_Coroutine, size_offset(%eax)
	jge 2f

	// let the old one be garbage collected
	cmpl $0, gc_offset(%eax)
	jle 1f
	decl gc_offset(%eax)
1:
	// make a new Coroutine
	movl $real_size_of_Coroutine, %eax
	movl $tag_of_Coroutine, %ebx
	call gc_alloc
2:
	// add a reference
	cmpl $0, gc_offset(%eax)
	jl 3f
	incl gc_offset(%eax)
3:
	// put the runnable inside
	movl 8(%ebp), %ebx
	test %ebx, %ebx
	jz runtime.null_panic
	cmpl $0, gc_offset(%ebx)
	jle 4f
	decl gc_offset(%ebx)
4:
	movl %ebx, offset_of_Coroutine.runnable(%eax)

	// put the Coroutine back where it was
	movl %eax, 12(%ebp)

	// allocate the stack
	movl $min_stack_size, %eax
	movl $tag_of_raw, %ebx
	call gc_alloc
	movl $gc_tag_root, gc_offset(%eax)
	movl %eax, %ebx

	// put the stack in the Coroutine
	movl 12(%ebp), %eax
	movl %ebx, offset_of_Coroutine.segment(%eax)
	addl size_offset(%ebx), %ebx
	addl $data_offset, %ebx

	// set up the stack
	movl $0, -4(%ebx)
	movl $0, -8(%ebx)
	movl $0, -12(%ebx)
	movl $Coroutine._run, -16(%ebx)
	movl $0, -20(%ebx)
	leal -20(%ebx), %ecx
	movl %ecx, offset_of_Coroutine.stack(%eax)

	// check if we're Main.
	movl current_coroutine, %ecx
	test %ecx, %ecx
	jnz 5f

	// save the stack and frame pointers so we can come back.
	movl %ebp, -4(%ebx)
	movl %esp, -8(%ebx)

	movl %eax, offset_of_Coroutine.prev(%eax)
	movl %eax, offset_of_Coroutine.next(%eax)
	movl %eax, current_coroutine

	// save %esp so the garbage collector can use the non-heap stack.
	movl %esp, coroutine_gc_stack_pointer

	// jump into the scheduler.
	jmp runtime.sched0

runtime.sched1:
	// Main terminated. throw away the coroutines and return.
	movl current_coroutine, %eax
	movl $0, current_coroutine
	jmp 6f

5:
	movl offset_of_Coroutine.next(%ecx), %ebx
	movl %eax, offset_of_Coroutine.prev(%ebx)
	movl %eax, offset_of_Coroutine.next(%ecx)
	movl %ecx, offset_of_Coroutine.prev(%eax)
	movl %ebx, offset_of_Coroutine.next(%eax)

	// add a reference so that the Coroutine can be saved even after the
	// only reference is the coroutine linked list.
	cmpl $0, gc_offset(%eax)
	jl 6f
	incl gc_offset(%eax)
6:
	leave
	.cfi_def_cfa esp, 4
	ret
	.cfi_endproc
	.size Coroutine.Coroutine, .-Coroutine.Coroutine

Coroutine._run:
	.cfi_startproc
	push %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register ebp
	subl $0, %esp

	movl current_coroutine, %eax
	movl offset_of_Coroutine.runnable(%eax), %ebx
	cmpl $0, gc_offset(%ebx)
	jl 1f
	incl gc_offset(%ebx)
1:
	push %ebx

	// Call Runnable.run
	movl tag_offset(%ebx), %ecx
	shll $2, %ecx
	movl method_tables(%ecx), %ecx
	movl method_offset_Runnable.run(%ecx), %ecx
	call *%ecx

	// we terminated. clean up.
	movl current_coroutine, %eax

	// remove ourself from the linked list.
	movl offset_of_Coroutine.next(%eax), %ebx
	movl offset_of_Coroutine.prev(%eax), %ecx
	movl %ebx, offset_of_Coroutine.next(%ecx)
	movl %ecx, offset_of_Coroutine.prev(%ebx)
	movl $0, offset_of_Coroutine.next(%eax)
	movl $0, offset_of_Coroutine.prev(%eax)

	// save the next coroutine as the current one.
	movl %ebx, current_coroutine

	// free our stack
	movl offset_of_Coroutine.segment(%eax), %ebx
	movl $tag_of_garbage, tag_offset(%ebx)
	movl $gc_tag_garbage, gc_offset(%ebx)
	movl $0, offset_of_Coroutine.segment(%eax)
	movl $0, offset_of_Coroutine.stack(%eax)

	// remove the reference we have to ourself.
	cmpl $0, gc_offset(%eax)
	jle 2f
	decl gc_offset(%eax)
2:
	// check if we are Main.
	cmpl $0, 8(%ebp)
	je 3f

	// we are Main. go home.
	movl 8(%ebp), %esp
	movl 12(%ebp), %ebp
	movl $0, coroutine_gc_stack_pointer
	jmp runtime.sched1

3:
	// our stack is invalid now. runtime.sched0 will give us a new one.
	jmp runtime.sched0

	.cfi_def_cfa esp, 4
	.cfi_endproc
	.size Coroutine.Coroutine, .-Coroutine.Coroutine

.globl Channel.Channel
Channel.Channel:
	.cfi_startproc

	movl $8, %eax
	movl $1, %ebx
	call runtime.morestack

	push %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register ebp
	subl $0, %esp

	movl 8(%ebp), %eax
	cmpl $(size_of_Channel + 4), size_offset(%eax)
	jge 2f

	// garbage collect the old channel
	cmpl $0, gc_offset(%eax)
	jle 1f
	decl gc_offset(%eax)
1:
	// make a new one
	movl $(size_of_Channel + 4), %eax
	movl $tag_of_Channel, %ebx
	call gc_alloc
2:
	// set the pointer to -1, which is never a valid heap address.
	movl $-1, offset_of_Channel.channel_field(%eax)

	leave
	.cfi_def_cfa esp, 4
	ret
	.cfi_endproc
	.size Channel.Channel, .-Channel.Channel

.globl Channel.send
Channel.send:
	.cfi_startproc

	movl $8, %eax
	movl $2, %ebx
	call runtime.morestack

	push %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register ebp
	subl $0, %esp

1:
	movl 12(%ebp), %ebx
	cmpl $-1, offset_of_Channel.channel_field(%ebx)
	je 2f
	call runtime.sched
	jmp 1b
2:
	movl $0, coroutine_deadlock
	movl 8(%ebp), %eax
	movl %eax, offset_of_Channel.channel_field(%ebx)
3:
	movl 12(%ebp), %ebx
	movl 8(%ebp), %eax
	cmpl %eax, offset_of_Channel.channel_field(%ebx)
	jne 4f
	call runtime.sched
	jmp 3b
4:
	movl $0, coroutine_deadlock
	movl $unit_lit, %eax

	leave
	.cfi_def_cfa esp, 4
	ret
	.cfi_endproc
	.size Channel.send, .-Channel.send

.globl Channel.recv
Channel.recv:
	.cfi_startproc

	movl $8, %eax
	movl $1, %ebx
	call runtime.morestack

	push %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register ebp
	subl $0, %esp

1:
	movl 8(%ebp), %ebx
	movl offset_of_Channel.channel_field(%ebx), %eax
	cmpl $-1, %eax
	jne 2f
	call runtime.sched
	jmp 1b
2:
	movl $0, coroutine_deadlock
	movl $-1, offset_of_Channel.channel_field(%ebx)

	leave
	.cfi_def_cfa esp, 4
	ret
	.cfi_endproc
	.size Channel.recv, .-Channel.recv

// gc_collect can't run with a small stack, and we can't call the allocator
// from inside the garbage collector, so we pretend we're being called from
// main.

.globl runtime.gc_collect
runtime.gc_collect:
	.cfi_startproc

	movl coroutine_gc_stack_pointer, %eax
	test %eax, %eax
	jnz 1f

	// we're not in a coroutine
	jmp gc_collect

1:
	movl %esp, %ebx
	movl %eax, %esp
	push %ebx
	call gc_collect
	pop %esp
	ret

	.cfi_endproc
	.size runtime.gc_collect, .-runtime.gc_collect
