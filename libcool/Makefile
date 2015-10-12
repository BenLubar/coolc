.PHONY: all
all: libcool.a fizzbuzz.s fizzbuzz

fizzbuzz: libcool.a fizzbuzz.o
	ld -melf_i386 -o $@ --start-group $^

libcool.a: basic.o basic_linux.o
	ar rcs $@ $^

%.o: %.s
	as -32 -g --fatal-warnings -o $@ $<

%.s: %.cool ../coolc
	../coolc -o $@ $<

.PHONY: ../coolc
../coolc: ../internal/ast/y.go
	(cd .. && go build)

../internal/ast/y.go: ../internal/ast/syntax.y
	(cd ../internal/ast && go generate)
