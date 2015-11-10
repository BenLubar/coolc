package ast

import "io"

type Options struct {
	Errors io.Writer

	Benchmark int
	Coroutine bool

	OptInt      bool
	OptJump     bool
	OptUnused   bool
	OptDispatch bool
	OptFold     bool
	OptInline   bool
}
