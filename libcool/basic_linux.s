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
	leal offset_of_String.str_value(%eax), %ecx

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
