.text

// we don't need anything special when there aren't coroutines.

.globl runtime.gc_collect
runtime.gc_collect:
	.cfi_startproc

	jmp gc_collect

	.cfi_endproc
	.size runtime.gc_collect, .-runtime.gc_collect
