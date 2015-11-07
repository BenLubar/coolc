package main_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
)

func init() {
	if output, err := exec.Command("go", "build").CombinedOutput(); err != nil {
		fmt.Println(string(output))
		panic(err)
	}

	mk := exec.Command("make", "libcool.a", "libcoolsched.a")
	mk.Dir = "libcool"
	if output, err := mk.CombinedOutput(); err != nil {
		fmt.Println(string(output))
		panic(err)
	}
}

func testBad(t *testing.T, prefix string) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	expect, err := ioutil.ReadFile(prefix + ".expected")
	if err != nil {
		t.Errorf("error reading %q: %v", prefix+".expected", err)
		return
	}

	out, err := exec.Command(filepath.Join(cwd, "coolc"), "-o", os.DevNull, prefix+".cool").CombinedOutput()
	if _, ok := err.(*exec.ExitError); !ok {
		t.Errorf("compiler error for %q was not exit status: %v", prefix+".cool", err)
	}

	if !bytes.Equal(expect, out) {
		t.Errorf("for %q:\nExpected output:\n%s\nActual output:\n%s", prefix+".cool", expect, out)
	}
}

func TestBad(t *testing.T) {
	const expected = 6

	cases, err := filepath.Glob(filepath.Join("testdata", "bad????.expected"))
	if err != nil {
		t.Fatal(err)
	}

	if len(cases) != expected {
		t.Errorf("expected %d cases but there are %d", expected, len(cases))
	}
}

func TestBad0000(t *testing.T) { testBad(t, filepath.Join("testdata", "bad0000")) }
func TestBad0001(t *testing.T) { testBad(t, filepath.Join("testdata", "bad0001")) }
func TestBad0002(t *testing.T) { testBad(t, filepath.Join("testdata", "bad0002")) }
func TestBad0003(t *testing.T) { testBad(t, filepath.Join("testdata", "bad0003")) }
func TestBad0004(t *testing.T) { testBad(t, filepath.Join("testdata", "bad0004")) }
func TestBad0005(t *testing.T) { testBad(t, filepath.Join("testdata", "bad0005")) }

func testGood(t *testing.T, prefix string) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	expect, err := ioutil.ReadFile(prefix + ".expected")
	if err != nil {
		t.Errorf("error reading %q: %v", prefix+".expected", err)
		return
	}

	if output, err := exec.Command(filepath.Join(cwd, "coolc"), "-o", prefix+".s", prefix+".cool").CombinedOutput(); err != nil {
		t.Errorf("unexpected compiler error for %q: %v\n%s", prefix+".cool", err, output)
		return
	}

	if output, err := exec.Command("as", "-32", "-g", "--fatal-warnings", "-o", prefix+".o", prefix+".s").CombinedOutput(); err != nil {
		t.Errorf("unexpected compiler error for %q: %v\n%s", prefix+".cool", err, output)
		return
	}

	// use .exe regardless of platform to make .gitignore easier
	if output, err := exec.Command("ld", "-melf_i386", "-o", prefix+".exe", "--start-group", filepath.Join("testdata", "libcool.a"), prefix+".o").CombinedOutput(); err != nil {
		t.Errorf("unexpected compiler error for %q: %v\n%s", prefix+".cool", err, output)
		return
	}

	out, err := exec.Command(prefix + ".exe").CombinedOutput()
	if err != nil {
		t.Errorf("error running %q: %v", prefix, err)
	}

	if !bytes.Equal(expect, out) {
		t.Errorf("for %q:\nExpected output:\n%s\nActual output:\n%s", prefix+".cool", expect, out)
	}
}

func benchmarkGood(b *testing.B, prefix string) {
	cwd, err := os.Getwd()
	if err != nil {
		b.Fatal(err)
	}

	if output, err := exec.Command(filepath.Join(cwd, "coolc"), "-o", prefix+".s", "-benchmark", strconv.Itoa(b.N), prefix+".cool").CombinedOutput(); err != nil {
		b.Errorf("unexpected compiler error for %q: %v\n%s", prefix+".cool", err, output)
		return
	}

	if output, err := exec.Command("as", "-32", "-g", "--fatal-warnings", "-o", prefix+".o", prefix+".s").CombinedOutput(); err != nil {
		b.Errorf("unexpected compiler error for %q: %v\n%s", prefix+".cool", err, output)
		return
	}

	if output, err := exec.Command("ld", "-melf_i386", "-o", prefix+".exe", "--start-group", filepath.Join("testdata", "libcool.a"), prefix+".o").CombinedOutput(); err != nil {
		b.Errorf("unexpected compiler error for %q: %v\n%s", prefix+".cool", err, output)
		return
	}

	b.ResetTimer()

	if err := exec.Command(prefix + ".exe").Run(); err != nil {
		b.Errorf("error running %q: %v", prefix, err)
	}
}

func TestGood(t *testing.T) {
	const expected = 4

	cases, err := filepath.Glob(filepath.Join("testdata", "good????.expected"))
	if err != nil {
		t.Fatal(err)
	}

	if len(cases) != expected {
		t.Errorf("expected %d cases but there are %d", expected, len(cases))
	}
}

func TestGood0000(t *testing.T) { testGood(t, filepath.Join("testdata", "good0000")) }
func TestGood0001(t *testing.T) { testGood(t, filepath.Join("testdata", "good0001")) }
func TestGood0002(t *testing.T) { testGood(t, filepath.Join("testdata", "good0002")) }
func TestGood0003(t *testing.T) { testGood(t, filepath.Join("testdata", "good0003")) }

func BenchmarkGood0000(b *testing.B) { benchmarkGood(b, filepath.Join("testdata", "good0000")) }
func BenchmarkGood0001(b *testing.B) { benchmarkGood(b, filepath.Join("testdata", "good0001")) }
func BenchmarkGood0002(b *testing.B) { benchmarkGood(b, filepath.Join("testdata", "good0002")) }
func BenchmarkGood0003(b *testing.B) { benchmarkGood(b, filepath.Join("testdata", "good0003")) }

func testCoroutine(t *testing.T, prefix string) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	expect, err := ioutil.ReadFile(prefix + ".expected")
	if err != nil {
		t.Errorf("error reading %q: %v", prefix+".expected", err)
		return
	}

	if output, err := exec.Command(filepath.Join(cwd, "coolc"), "-coroutine", "-o", prefix+".s", prefix+".cool").CombinedOutput(); err != nil {
		t.Errorf("unexpected compiler error for %q: %v\n%s", prefix+".cool", err, output)
		return
	}

	if output, err := exec.Command("as", "-32", "-g", "--fatal-warnings", "-o", prefix+".o", prefix+".s").CombinedOutput(); err != nil {
		t.Errorf("unexpected compiler error for %q: %v\n%s", prefix+".cool", err, output)
		return
	}

	// use .exe regardless of platform to make .gitignore easier
	if output, err := exec.Command("ld", "-melf_i386", "-o", prefix+".exe", "--start-group", filepath.Join("testdata", "libcoolsched.a"), prefix+".o").CombinedOutput(); err != nil {
		t.Errorf("unexpected compiler error for %q: %v\n%s", prefix+".cool", err, output)
		return
	}

	out, err := exec.Command(prefix + ".exe").CombinedOutput()
	if err != nil {
		t.Errorf("error running %q: %v", prefix, err)
	}

	if !bytes.Equal(expect, out) {
		t.Errorf("for %q:\nExpected output:\n%s\nActual output:\n%s", prefix+".cool", expect, out)
	}
}

func benchmarkCoroutine(b *testing.B, prefix string) {
	cwd, err := os.Getwd()
	if err != nil {
		b.Fatal(err)
	}

	if output, err := exec.Command(filepath.Join(cwd, "coolc"), "-o", prefix+".s", "-coroutine", "-benchmark", strconv.Itoa(b.N), prefix+".cool").CombinedOutput(); err != nil {
		b.Errorf("unexpected compiler error for %q: %v\n%s", prefix+".cool", err, output)
		return
	}

	if output, err := exec.Command("as", "-32", "-g", "--fatal-warnings", "-o", prefix+".o", prefix+".s").CombinedOutput(); err != nil {
		b.Errorf("unexpected compiler error for %q: %v\n%s", prefix+".cool", err, output)
		return
	}

	if output, err := exec.Command("ld", "-melf_i386", "-o", prefix+".exe", "--start-group", filepath.Join("testdata", "libcoolsched.a"), prefix+".o").CombinedOutput(); err != nil {
		b.Errorf("unexpected compiler error for %q: %v\n%s", prefix+".cool", err, output)
		return
	}

	b.ResetTimer()

	if err := exec.Command(prefix + ".exe").Run(); err != nil {
		b.Errorf("error running %q: %v", prefix, err)
	}
}

func TestCoroutine(t *testing.T) {
	const expected = 1

	cases, err := filepath.Glob(filepath.Join("testdata", "coroutine????.expected"))
	if err != nil {
		t.Fatal(err)
	}

	if len(cases) != expected {
		t.Errorf("expected %d cases but there are %d", expected, len(cases))
	}
}

func TestCoroutineGood0000(t *testing.T) { testCoroutine(t, filepath.Join("testdata", "good0000")) }
func TestCoroutineGood0001(t *testing.T) { testCoroutine(t, filepath.Join("testdata", "good0001")) }
func TestCoroutineGood0002(t *testing.T) { testCoroutine(t, filepath.Join("testdata", "good0002")) }
func TestCoroutineGood0003(t *testing.T) { testCoroutine(t, filepath.Join("testdata", "good0003")) }
func TestCoroutine0000(t *testing.T)     { testCoroutine(t, filepath.Join("testdata", "coroutine0000")) }

func BenchmarkCoroutineGood0000(b *testing.B) {
	benchmarkCoroutine(b, filepath.Join("testdata", "good0000"))
}
func BenchmarkCoroutineGood0001(b *testing.B) {
	benchmarkCoroutine(b, filepath.Join("testdata", "good0001"))
}
func BenchmarkCoroutineGood0002(b *testing.B) {
	benchmarkCoroutine(b, filepath.Join("testdata", "good0002"))
}
func BenchmarkCoroutineGood0003(b *testing.B) {
	benchmarkCoroutine(b, filepath.Join("testdata", "good0003"))
}
func BenchmarkCoroutine0000(b *testing.B) {
	benchmarkCoroutine(b, filepath.Join("testdata", "coroutine0000"))
}
