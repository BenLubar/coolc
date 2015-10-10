.text

.globl runtime.exit
runtime.exit:
	movl $1, %eax
	movl 8(%ebp), %ebx
	int $0x80

.globl runtime.output
runtime.output:
	movl 8(%ebp), %eax

	movl offset_of_String.length(%eax), %edx
	movl offset_of_Int.value(%edx), %edx
	leal offset_of_String.str_field(%eax), %ecx

	movl $4, %eax
	movl $1, %ebx
	int $0x80

	ret

.globl runtime.heap_get
runtime.heap_get:
	movl $45, %eax
	movl $0, %ebx
	int $0x80
	ret

.globl runtime.heap_set
runtime.heap_set:
	movl $45, %eax
	movl 8(%ebp), %ebx
	int $0x80

	ret

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
