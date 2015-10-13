.PHONY: all
all: libcool.a

%: libcool.a %.o
	ld -melf_i386 -o $@ --start-group $^

libcool.a: basic.o runtime_linux.o gc.o
	ar rcs $@ $^

%.o: %.s basic_defs.s
	as -32 -g --fatal-warnings -o $@ $<

%.s: %.cool ../coolc
	../coolc -o $@ $<

.PHONY: ../coolc
../coolc: ../internal/ast/y.go
	(cd .. && go build)

../internal/ast/y.go: ../internal/ast/syntax.y
	(cd ../internal/ast && go generate)
