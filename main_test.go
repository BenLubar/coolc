package main_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func init() {
	if output, err := exec.Command("go", "build").CombinedOutput(); err != nil {
		fmt.Println(string(output))
		panic(err)
	}

	mk := exec.Command("make", "libcool.a")
	mk.Dir = "libcool"
	if output, err := mk.CombinedOutput(); err != nil {
		fmt.Println(string(output))
		panic(err)
	}
}

func TestBad(t *testing.T) {
	const expected = 6

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	cases, err := filepath.Glob(filepath.Join("testdata", "bad????.expected"))
	if err != nil {
		t.Fatal(err)
	}

	if len(cases) != expected {
		t.Errorf("expected %d cases but there are %d", expected, len(cases))
	}

	for _, expectName := range cases {
		expect, err := ioutil.ReadFile(expectName)
		if err != nil {
			t.Errorf("error reading %q: %v", expectName, err)
			continue
		}

		in := strings.TrimSuffix(expectName, ".expected") + ".cool"

		out, err := exec.Command(filepath.Join(cwd, "coolc"), "-o", os.DevNull, in).CombinedOutput()
		if _, ok := err.(*exec.ExitError); !ok {
			t.Errorf("compiler error for %q was not exit status: %v", in, err)
		}

		if !bytes.Equal(expect, out) {
			t.Errorf("for %q:\nExpected output:\n%s\nActual output:\n%s", in, expect, out)
		}
	}
}

func TestGood(t *testing.T) {
	const expected = 2

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	cases, err := filepath.Glob(filepath.Join("testdata", "good????.expected"))
	if err != nil {
		t.Fatal(err)
	}

	if len(cases) != expected {
		t.Errorf("expected %d cases but there are %d", expected, len(cases))
	}

	for _, expectName := range cases {
		expect, err := ioutil.ReadFile(expectName)
		if err != nil {
			t.Errorf("error reading %q: %v", expectName, err)
			continue
		}

		base := strings.TrimSuffix(expectName, ".expected")
		in := base + ".cool"

		if output, err := exec.Command(filepath.Join(cwd, "coolc"), "-o", base+".s", in).CombinedOutput(); err != nil {
			t.Errorf("unexpected compiler error for %q: %v\n%s", in, err, output)
			continue
		}

		if output, err := exec.Command("as", "-32", "-g", "--fatal-warnings", "-o", base+".o", base+".s").CombinedOutput(); err != nil {
			t.Errorf("unexpected compiler error for %q: %v\n%s", in, err, output)
			continue
		}

		// use .exe regardless of platform to make .gitignore easier
		if output, err := exec.Command("ld", "-melf_i386", "-o", base+".exe", "--start-group", filepath.Join("testdata", "libcool.a"), base+".o").CombinedOutput(); err != nil {
			t.Errorf("unexpected compiler error for %q: %v\n%s", in, err, output)
			continue
		}

		out, err := exec.Command(base + ".exe").CombinedOutput()
		if err != nil {
			t.Errorf("error running %q: %v", base, err)
		}

		if !bytes.Equal(expect, out) {
			t.Errorf("for %q:\nExpected output:\n%s\nActual output:\n%s", in, expect, out)
		}
	}
}
