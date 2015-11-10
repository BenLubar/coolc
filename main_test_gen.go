// +build ignore

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	f, err := os.Create("main_test.go")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	fmt.Fprintln(f, "//go:generate go run main_test_gen.go")
	fmt.Fprintln(f, "")
	fmt.Fprintln(f, "package main")
	fmt.Fprintln(f, "")
	fmt.Fprintln(f, "import (")
	fmt.Fprintln(f, "\t\"testing\"")
	fmt.Fprintln(f, ")")

	err = os.Chdir("testdata")
	if err != nil {
		panic(err)
	}

	bad, err := filepath.Glob("bad????.cool")
	if err != nil {
		panic(err)
	}
	for _, name := range bad {
		fmt.Fprintf(f, `
func TestBad%[1]s(t *testing.T) {
	testBad(t, %[2]q)
}
`, name[len("bad"):][:4], name[:len("bad")+4])
	}
	good, err := filepath.Glob("good????.cool")
	if err != nil {
		panic(err)
	}
	for _, name := range good {
		fmt.Fprintf(f, `
func TestGood%[1]s(t *testing.T) {
	testGood(t, %[2]q, "libcool.a")
}
func BenchmarkGood%[1]s(b *testing.B) {
	benchmarkGood(b, %[2]q, "libcool.a")
}
func TestGood%[1]sCo(t *testing.T) {
	testGood(t, %[2]q, "libcoolsched.a", "-coroutine")
}
func BenchmarkGood%[1]sCo(b *testing.B) {
	benchmarkGood(b, %[2]q, "libcoolsched.a", "-coroutine")
}
`, name[len("good"):][:4], name[:len("good")+4])
	}
	coroutine, err := filepath.Glob("coroutine????.cool")
	if err != nil {
		panic(err)
	}
	for _, name := range coroutine {
		fmt.Fprintf(f, `
func TestCoroutine%[1]sCo(t *testing.T) {
	testGood(t, %[2]q, "libcoolsched.a", "-coroutine")
}
func BenchmarkCoroutine%[1]sCo(b *testing.B) {
	benchmarkGood(b, %[2]q, "libcoolsched.a", "-coroutine")
}
`, name[len("coroutine"):][:4], name[:len("coroutine")+4])
	}
}
