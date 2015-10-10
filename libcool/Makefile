.PHONY: all
all: noop hello array

%: basic.o basic_linux.o %.o
	ld -melf_i386 -o $@ --start-group $^

%.o: %.s
	as -32 -g --fatal-warnings -o $@ $<

%.s: %.cool ../coolc
	../coolc -o $@ $<

.PHONY: ../coolc
../coolc: ../internal/ast/y.go
	(cd .. && go build)

../internal/ast/y.go: ../internal/ast/syntax.y
	(cd ../internal/ast && go generate)
