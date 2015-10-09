noop: basic.o basic_linux.o noop.o
	ld -melf_i386 -o $@ $^

%.o: %.s
	as -32 -o $@ $<

%.s: %.cool ../coolc
	../coolc -o $@ $<

.PHONY: ../coolc
../coolc: ../internal/ast/y.go
	(cd .. && go build)

../internal/ast/y.go: ../internal/ast/syntax.y
	(cd ../internal/ast && go generate)
