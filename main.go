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

	haveError := false

	var prog ast.Program

	for _, name := range flag.Args() {
		b, err := ioutil.ReadFile(name)
		if err != nil {
			fmt.Printf("%s: %v", name, err)
			haveError = true
			continue
		}

		f := fset.AddFile(name, -1, len(b))
		f.SetLinesForContent(b)

		haveError = prog.Parse(f, bytes.NewReader(b)) || haveError
	}

	if haveError {
		os.Exit(2)
	}
}
