package main_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func init() {
	if err := exec.Command("go", "build").Run(); err != nil {
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
