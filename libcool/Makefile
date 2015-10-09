noop: basic.o basic_linux.o noop.o
	ld -melf_i386 -o $@ $^

%.o: %.s
	as -32 -o $@ $<

%.s: %.cool ../coolc
	../coolc -o $@ $<

.PHONY: ../coolc
../coolc:
	-(cd .. && go build)
