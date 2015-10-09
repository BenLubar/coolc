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

	flagOutput := flag.String("o", "", "output filename")

	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
	}

	if *flagOutput == "" {
		*flagOutput = strings.TrimSuffix(flag.Arg(0), ".cool") + ".s"
	}

	fset := token.NewFileSet()

	var haveErrors bool
	var prog ast.Program

	{
		f := fset.AddFile("basic.cool", -1, len(basicCool))
		f.SetLinesForContent(basicCool)

		haveErrors = prog.Parse(f, bytes.NewReader(basicCool))
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

	if prog.Semant(fset) {
		os.Exit(2)
	}

	f, err := os.Create(*flagOutput)
	if err != nil {
		fmt.Printf("%s: %v\n", *flagOutput, err)
		os.Exit(2)
	}
	defer f.Close()

	err = prog.CodeGen(f)
	if err != nil {
		fmt.Printf("%s: %v\n", *flagOutput, err)
		os.Exit(2)
	}
}
