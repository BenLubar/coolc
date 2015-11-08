package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/token"
	"io/ioutil"
	"os"
	"strings"

	"github.com/BenLubar/coolc/internal/ast"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "[ -o fileout ] file1.cool file2.cool ... filen.cool")
		flag.PrintDefaults()
		os.Exit(1)
	}

	var opt ast.Options

	flagOutput := flag.String("o", "", "output filename")
	flag.IntVar(&opt.Benchmark, "benchmark", 1, "repeat the program this many times")
	flag.BoolVar(&opt.Coroutine, "coroutine", false, "enable coroutine support")
	flag.BoolVar(&opt.OptInt, "opt-int", true, "optimization: use raw integers")
	flag.BoolVar(&opt.OptJump, "opt-jump", true, "optimization: convert conditions to jumps")
	flag.BoolVar(&opt.OptUnused, "opt-unused", true, "optimization: skip computing unused values")
	flag.BoolVar(&opt.OptDispatch, "opt-dispatch", true, "optimization: convert dynamic dispatch to a known method to static dispatch")
	flag.BoolVar(&opt.OptFold, "opt-fold", true, "optimization: precompute the values of constant arithmetic expressions")
	flag.BoolVar(&opt.OptInline, "opt-inline", true, "optimization: inline methods that are sufficiently simple")

	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
	}

	if *flagOutput == "" {
		*flagOutput = strings.TrimSuffix(flag.Arg(0), ".cool") + ".s"
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

		haveErrors = prog.Parse(f, bytes.NewReader(basicCool))
	}

	if opt.Coroutine {
		f := fset.AddFile("coroutine.cool", -1, len(coroutineCool))
		f.SetLinesForContent(coroutineCool)

		haveErrors = prog.Parse(f, bytes.NewReader(coroutineCool))
	}

	for _, name := range flag.Args() {
		b, err := ioutil.ReadFile(name)
		if err != nil {
			fmt.Printf("%s: %v", name, err)
			haveErrors = true
			continue
		}

		f := fset.AddFile(name, -1, len(b))
		f.SetLinesForContent(b)

		haveErrors = prog.Parse(f, bytes.NewReader(b)) || haveErrors
	}

	if haveErrors {
		os.Exit(2)
	}

	if prog.Semant(opt, fset) {
		os.Exit(2)
	}

	f, err := os.Create(*flagOutput)
	if err != nil {
		fmt.Printf("%s: %v\n", *flagOutput, err)
		os.Exit(2)
	}
	defer f.Close()

	err = prog.CodeGen(opt, fset, f)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
}
