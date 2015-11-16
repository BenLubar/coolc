package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func init() {
	mk := exec.Command("make", "-s", "libcool.a", "libcoolsched.a")
	mk.Dir = "libcool"
	if output, err := mk.CombinedOutput(); err != nil {
		fmt.Println(string(output))
		panic(err)
	} else if len(output) != 0 {
		fmt.Println(string(output))
		panic("unexpected output from make")
	}
}

func runCompiler(args []string) (out []byte, exit int) {
	var buf bytes.Buffer

	exit = compiler(args, &buf)
	out = buf.Bytes()

	return
}

func testBad(t testing.TB, prefix string, args ...string) {
	prefix = filepath.Join("testdata", prefix)
	expected := prefix + ".expected"
	source := prefix + ".cool"

	expect, err := ioutil.ReadFile(expected)
	if err != nil {
		t.Fatalf("error reading %q: %v", expected, err)
	}

	out, exit := runCompiler(append(append([]string{"coolc", "-o", os.DevNull}, args...), source))
	if exit != 2 {
		t.Errorf("exit status for %q was unexpected: %v", source, exit)
	}

	if !bytes.Equal(expect, out) {
		t.Errorf("for %q:\nExpected output:\n%s\nActual output:\n%s", source, expect, out)
	}
}

func testGood(t testing.TB, prefix, lib string, args ...string) {
	prefix = filepath.Join("testdata", prefix)
	expected := prefix + ".expected"
	source := prefix + ".cool"
	asm := prefix + ".s"
	obj := prefix + ".o"
	// use .exe regardless of platform to make .gitignore easier
	exe := prefix + strings.Join(args, "") + ".exe"

	expect, err := ioutil.ReadFile(expected)
	if err != nil {
		t.Fatalf("error reading %q: %v", expected, err)
	}

	if out, exit := runCompiler(append(append([]string{"coolc", "-o", asm}, args...), source)); exit != 0 {
		t.Fatalf("unexpected compiler exit status for %q: %v\n%s", source, exit, out)
	} else if len(out) != 0 {
		t.Errorf("unexpected compiler ouput for %q:\n%s", source, out)
	}

	if output, err := exec.Command("as", "-32", "-g", "--fatal-warnings", "-o", obj, asm).CombinedOutput(); err != nil {
		t.Fatalf("unexpected compiler error for %q: %v\n%s", source, err, output)
	}

	if output, err := exec.Command("ld", "-melf_i386", "-o", exe, "--start-group", filepath.Join("testdata", lib), obj).CombinedOutput(); err != nil {
		t.Fatalf("unexpected compiler error for %q: %v\n%s", source, err, output)
	}

	out, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Errorf("error running %q: %v", prefix, err)
	}

	if !bytes.Equal(expect, out) {
		t.Errorf("for %q:\nExpected output:\n%s\nActual output:\n%s", source, expect, out)
	}
}

func benchmarkGood(b *testing.B, prefix, lib string, args ...string) {
	prefix = filepath.Join("testdata", prefix)
	source := prefix + ".cool"
	asm := prefix + ".s"
	obj := prefix + ".o"
	// use .exe regardless of platform to make .gitignore easier
	exe := prefix + strings.Join(args, "") + ".exe"

	if out, exit := runCompiler(append(append([]string{"coolc", "-o", asm, "-benchmark", strconv.Itoa(b.N)}, args...), source)); exit != 0 {
		b.Fatalf("unexpected compiler exit status for %q: %v\n%s", source, exit, out)
	} else if len(out) != 0 {
		b.Errorf("unexpected compiler ouput for %q:\n%s", source, out)
	}

	if output, err := exec.Command("as", "-32", "-g", "--fatal-warnings", "-o", obj, asm).CombinedOutput(); err != nil {
		b.Fatalf("unexpected compiler error for %q: %v\n%s", source, err, output)
	}

	if output, err := exec.Command("ld", "-melf_i386", "-o", exe, "--start-group", filepath.Join("testdata", lib), obj).CombinedOutput(); err != nil {
		b.Fatalf("unexpected compiler error for %q: %v\n%s", source, err, output)
	}

	b.ResetTimer()

	if err := exec.Command(exe).Run(); err != nil {
		b.Errorf("error running %q: %v", prefix, err)
	}
}
