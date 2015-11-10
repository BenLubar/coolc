package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/token"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/BenLubar/coolc/internal/ast"
)

func main() {
	os.Exit(compiler(os.Args, os.Stderr))
}

func compiler(args []string, errors io.Writer) int {
	var opt ast.Options

	opt.Errors = errors

	flagSet := flag.NewFlagSet("coolc", flag.ContinueOnError)
	flagSet.SetOutput(errors)
	flagSet.Usage = func() {
		fmt.Fprintln(opt.Errors, "Usage:", args[0], "[ -o fileout ] file1.cool file2.cool ... filen.cool")
		flagSet.PrintDefaults()
	}

	flagOutput := flagSet.String("o", "", "output filename")
	flagSet.IntVar(&opt.Benchmark, "benchmark", 1, "repeat the program this many times")
	flagSet.BoolVar(&opt.Coroutine, "coroutine", false, "enable coroutine support")
	flagSet.BoolVar(&opt.OptInt, "opt-int", true, "optimization: use raw integers")
	flagSet.BoolVar(&opt.OptJump, "opt-jump", true, "optimization: convert conditions to jumps")
	flagSet.BoolVar(&opt.OptUnused, "opt-unused", true, "optimization: skip computing unused values")
	flagSet.BoolVar(&opt.OptDispatch, "opt-dispatch", true, "optimization: convert dynamic dispatch to a known method to static dispatch")
	flagSet.BoolVar(&opt.OptFold, "opt-fold", true, "optimization: precompute the values of constant arithmetic expressions")
	flagSet.BoolVar(&opt.OptInline, "opt-inline", true, "optimization: inline methods that are sufficiently simple")

	if err := flagSet.Parse(args[1:]); err != nil {
		flagSet.Usage()
		return 1
	}

	if flagSet.NArg() == 0 {
		flagSet.Usage()
		return 1
	}

	if *flagOutput == "" {
		*flagOutput = strings.TrimSuffix(flagSet.Arg(0), ".cool") + ".s"
	}

	if opt.Benchmark < 1 {
		opt.Benchmark = 1
	}

	fset := token.NewFileSet()

	var haveErrors bool
	var prog ast.Program

	{
		f := fset.AddFile("basic.cool", -1, len(basicCool))
		f.SetLinesForContent(basicCool)

		haveErrors = prog.Parse(f, opt, bytes.NewReader(basicCool))
	}

	if opt.Coroutine {
		f := fset.AddFile("coroutine.cool", -1, len(coroutineCool))
		f.SetLinesForContent(coroutineCool)

		haveErrors = prog.Parse(f, opt, bytes.NewReader(coroutineCool))
	}

	for _, name := range flagSet.Args() {
		b, err := ioutil.ReadFile(name)
		if err != nil {
			fmt.Fprintf(opt.Errors, "%s: %v", name, err)
			haveErrors = true
			continue
		}

		f := fset.AddFile(name, -1, len(b))
		f.SetLinesForContent(b)

		haveErrors = prog.Parse(f, opt, bytes.NewReader(b)) || haveErrors
	}

	if haveErrors {
		return 2
	}

	if prog.Semant(opt, fset) {
		return 2
	}

	f, err := os.Create(*flagOutput)
	if err != nil {
		fmt.Fprintf(opt.Errors, "%s: %v\n", *flagOutput, err)
		return 2
	}
	defer f.Close()

	err = prog.CodeGen(opt, fset, f)
	if err != nil {
		fmt.Fprintf(opt.Errors, "error during code generation: %v\n", err)
		return 2
	}

	return 0
}
