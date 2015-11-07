.include "basic_defs.s"

.data

.globl runtime_input_max
.set runtime_input_max, 0x400

.align 2
runtime_input_buf:
	.skip runtime_input_max

.align 2
runtime_input_remaining:
	.long 0

.text

.globl runtime.exit
.type runtime.exit, @function
runtime.exit:
	.cfi_startproc
	push %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register ebp
	subl $0, %esp

	movl $1, %eax
	movl 8(%ebp), %ebx
	int $0x80
	.cfi_endproc
	.size runtime.exit, .-runtime.exit

.globl runtime.output
.type runtime.output, @function
runtime.output:
	.cfi_startproc
	push %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register ebp
	subl $0, %esp

	movl 8(%ebp), %eax

	movl offset_of_String.length(%eax), %edx
	movl offset_of_Int.value(%edx), %edx
	leal offset_of_String.str_field(%eax), %ecx

	movl $4, %eax
	movl $1, %ebx
	int $0x80

	leave
	.cfi_def_cfa esp, 4
	ret $4
	.cfi_endproc
	.size runtime.output, .-runtime.output

.globl runtime.input
.type runtime.input, @function
runtime.input:
	.cfi_startproc
	push %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register ebp
	subl $0, %esp

1:
	movl runtime_input_remaining, %ecx
	test %ecx, %ecx
	jz 2f

	leal runtime_input_buf, %edi
	movl $10, %eax
	repne scasb
	jne 2f

	movl %edi, %ecx
	subl $runtime_input_buf, %ecx
3:
	leal runtime_input_buf, %esi
	movl 8(%ebp), %eax
	movl offset_of_String.length(%eax), %ebx
	movl %ecx, offset_of_Int.value(%ebx)
	leal offset_of_String.str_field(%eax), %edi
	cld
	rep movsb

	movl runtime_input_remaining, %ecx
	subl offset_of_Int.value(%ebx), %ecx
	movl %ecx, runtime_input_remaining

	cmpl $0, offset_of_Int.value(%ebx)
	je 4f

	cmpb $10, -1(%edi)
	jne 4f

	decl offset_of_Int.value(%ebx)

4:
	leal runtime_input_buf, %edi
	cld
	rep movsb

	jmp 5f

2:
	call runtime.fill_buf

	test %eax, %eax
	jnz 1b

	movl runtime_input_remaining, %ecx
	cmpl $0, %ecx
	jne 3b

	movl 8(%ebp), %eax
	decl gc_offset(%eax)
	movl $0, %eax

5:
	leave
	.cfi_def_cfa esp, 4
	ret $4
	.cfi_endproc
	.size runtime.input, .-runtime.input

.type runtime.fill_buf, @function
runtime.fill_buf:
	.cfi_startproc
	push %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register ebp
	subl $0, %esp

	movl $3, %eax
	movl $0, %ebx
	movl runtime_input_remaining, %ecx
	movl $runtime_input_max, %edx
	subl %ecx, %edx
	leal runtime_input_buf(%ecx), %ecx
	int $0x80

	cmpl $0, %eax
	jge 1f

	movl $0, %eax

1:
	addl %eax, runtime_input_remaining

	leave
	.cfi_def_cfa esp, 4
	ret $0
	.cfi_endproc
	.size runtime.fill_buf, .-runtime.fill_buf

.globl runtime.heap_get
.type runtime.heap_get, @function
runtime.heap_get:
	.cfi_startproc
	push %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register ebp
	subl $0, %esp

	movl $45, %eax
	movl $0, %ebx
	int $0x80

	leave
	.cfi_def_cfa esp, 4
	ret $0
	.cfi_endproc
	.size runtime.heap_get, .-runtime.heap_get

.globl runtime.heap_set
.type runtime.heap_set, @function
runtime.heap_set:
	.cfi_startproc
	push %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register ebp
	subl $0, %esp

	movl $45, %eax
	movl 8(%ebp), %ebx
	int $0x80

	leave
	.cfi_def_cfa esp, 4
	ret $4
	.cfi_endproc
	.size runtime.heap_set, .-runtime.heap_set

.data

.align 2
case_panic_before:
	.ascii "Unhandled type in match expression: "
.set case_panic_before_length, .-case_panic_before

.align 2
case_panic_after:
	.ascii "\n"
.set case_panic_after_length, .-case_panic_after

.text

.globl runtime.case_panic
.type runtime.case_panic, @function
runtime.case_panic:
	.cfi_startproc
	push %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register ebp
	subl $0, %esp

	movl class_names(%eax), %eax
	push %eax

	movl $4, %eax
	movl $1, %ebx
	leal case_panic_before, %ecx
	movl $case_panic_before_length, %edx
	int $0x80

	call runtime.output

	movl $4, %eax
	movl $1, %ebx
	leal case_panic_after, %ecx
	movl $case_panic_after_length, %edx
	int $0x80

	movl $1, %eax
	movl $1, %ebx
	int $0x80

	.cfi_endproc
	.size runtime.case_panic, .-runtime.case_panic

.data

.align 2
null_panic_before:
	.ascii "Null pointer dereference\n"
.set null_panic_before_length, .-null_panic_before

.text

.globl runtime.null_panic
.type runtime.null_panic, @function
runtime.null_panic:
	.cfi_startproc
	push %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register ebp
	subl $0, %esp

	movl $4, %eax
	movl $1, %ebx
	leal null_panic_before, %ecx
	movl $null_panic_before_length, %edx
	int $0x80

	movl $1, %eax
	movl $1, %ebx
	int $0x80

	.cfi_endproc
	.size runtime.null_panic, .-runtime.null_panic

.data

.align 2
bounds_panic_before:
	.ascii "Index out of bounds\n"
.set bounds_panic_before_length, .-bounds_panic_before

.text

.globl runtime.bounds_panic
.type runtime.bounds_panic, @function
runtime.bounds_panic:
	.cfi_startproc
	push %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register ebp
	subl $0, %esp

	movl $4, %eax
	movl $1, %ebx
	leal bounds_panic_before, %ecx
	movl $bounds_panic_before_length, %edx
	int $0x80

	movl $1, %eax
	movl $1, %ebx
	int $0x80

	.cfi_endproc
	.size runtime.bounds_panic, .-runtime.bounds_panic

.data

.align 2
deadlock_panic_before:
	.ascii "Deadlock\n"
.set deadlock_panic_before_length, .-deadlock_panic_before

.text

.globl runtime.deadlock_panic
.type runtime.deadlock_panic, @function
runtime.deadlock_panic:
	.cfi_startproc
	push %ebp
	.cfi_def_cfa_offset 8
	.cfi_offset ebp, -8
	movl %esp, %ebp
	.cfi_def_cfa_register ebp
	subl $0, %esp

	movl $4, %eax
	movl $1, %ebx
	leal deadlock_panic_before, %ecx
	movl $deadlock_panic_before_length, %edx
	int $0x80

	movl $1, %eax
	movl $1, %ebx
	int $0x80

	.cfi_endproc
	.size runtime.deadlock_panic, .-runtime.deadlock_panic
