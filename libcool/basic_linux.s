.include "basic_defs.s"

.text

.globl runtime.exit
runtime.exit:
	enter $0, $0

	movl $1, %eax
	movl 8(%ebp), %ebx
	int $0x80

.globl runtime.output
runtime.output:
	enter $0, $0

	movl 8(%ebp), %eax

	movl offset_of_String.length(%eax), %edx
	movl offset_of_Int.value(%edx), %edx
	leal offset_of_String.str_field(%eax), %ecx

	movl $4, %eax
	movl $1, %ebx
	int $0x80

	leave
	ret $4

.globl runtime.heap_get
runtime.heap_get:
	enter $0, $0

	movl $45, %eax
	movl $0, %ebx
	int $0x80

	leave
	ret $0

.globl runtime.heap_set
runtime.heap_set:
	enter $0, $0

	movl $45, %eax
	movl 8(%ebp), %ebx
	int $0x80

	leave
	ret $4

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
runtime.case_panic:
	enter $0, $0

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

.data

.align 2
null_panic_before:
	.ascii "Null pointer dereference\n"
.set null_panic_before_length, .-null_panic_before

.text

.globl runtime.null_panic
runtime.null_panic:
	enter $0, $0

	movl $4, %eax
	movl $1, %ebx
	leal null_panic_before, %ecx
	movl $null_panic_before_length, %edx
	int $0x80

	movl $1, %eax
	movl $1, %ebx
	int $0x80

.data

.align 2
bounds_panic_before:
	.ascii "Index out of bounds\n"
.set bounds_panic_before_length, .-bounds_panic_before

.text

.globl runtime.bounds_panic
runtime.bounds_panic:
	enter $0, $0

	movl $4, %eax
	movl $1, %ebx
	leal bounds_panic_before, %ecx
	movl $bounds_panic_before_length, %edx
	int $0x80

	movl $1, %eax
	movl $1, %ebx
	int $0x80

.globl runtime.TODO
runtime.TODO:
	enter $0, $0

	int $3

	movl $1, %eax
	movl $1, %ebx
	int $0x80
