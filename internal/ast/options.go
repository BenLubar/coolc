package ast

type Options struct {
	Benchmark int
	Coroutine bool

	OptInt      bool
	OptJump     bool
	OptUnused   bool
	OptDispatch bool
}
