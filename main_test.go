//go:generate go run main_test_gen.go

package main

import (
	"testing"
)

func TestBad0000(t *testing.T) {
	testBad(t, "bad0000")
}

func TestBad0001(t *testing.T) {
	testBad(t, "bad0001")
}

func TestBad0002(t *testing.T) {
	testBad(t, "bad0002")
}

func TestBad0003(t *testing.T) {
	testBad(t, "bad0003")
}

func TestBad0004(t *testing.T) {
	testBad(t, "bad0004")
}

func TestBad0005(t *testing.T) {
	testBad(t, "bad0005")
}

func TestGood0000(t *testing.T) {
	testGood(t, "good0000", "libcool.a")
}
func BenchmarkGood0000(b *testing.B) {
	benchmarkGood(b, "good0000", "libcool.a")
}
func TestGood0000Co(t *testing.T) {
	testGood(t, "good0000", "libcoolsched.a", "-coroutine")
}
func BenchmarkGood0000Co(b *testing.B) {
	benchmarkGood(b, "good0000", "libcoolsched.a", "-coroutine")
}

func TestGood0001(t *testing.T) {
	testGood(t, "good0001", "libcool.a")
}
func BenchmarkGood0001(b *testing.B) {
	benchmarkGood(b, "good0001", "libcool.a")
}
func TestGood0001Co(t *testing.T) {
	testGood(t, "good0001", "libcoolsched.a", "-coroutine")
}
func BenchmarkGood0001Co(b *testing.B) {
	benchmarkGood(b, "good0001", "libcoolsched.a", "-coroutine")
}

func TestGood0002(t *testing.T) {
	testGood(t, "good0002", "libcool.a")
}
func BenchmarkGood0002(b *testing.B) {
	benchmarkGood(b, "good0002", "libcool.a")
}
func TestGood0002Co(t *testing.T) {
	testGood(t, "good0002", "libcoolsched.a", "-coroutine")
}
func BenchmarkGood0002Co(b *testing.B) {
	benchmarkGood(b, "good0002", "libcoolsched.a", "-coroutine")
}

func TestGood0003(t *testing.T) {
	testGood(t, "good0003", "libcool.a")
}
func BenchmarkGood0003(b *testing.B) {
	benchmarkGood(b, "good0003", "libcool.a")
}
func TestGood0003Co(t *testing.T) {
	testGood(t, "good0003", "libcoolsched.a", "-coroutine")
}
func BenchmarkGood0003Co(b *testing.B) {
	benchmarkGood(b, "good0003", "libcoolsched.a", "-coroutine")
}

func TestCoroutine0000Co(t *testing.T) {
	testGood(t, "coroutine0000", "libcoolsched.a", "-coroutine")
}
func BenchmarkCoroutine0000Co(b *testing.B) {
	benchmarkGood(b, "coroutine0000", "libcoolsched.a", "-coroutine")
}

func TestCoroutine0001Co(t *testing.T) {
	testGood(t, "coroutine0001", "libcoolsched.a", "-coroutine")
}
func BenchmarkCoroutine0001Co(b *testing.B) {
	benchmarkGood(b, "coroutine0001", "libcoolsched.a", "-coroutine")
}

func TestCoroutine0002Co(t *testing.T) {
	testGood(t, "coroutine0002", "libcoolsched.a", "-coroutine")
}
func BenchmarkCoroutine0002Co(b *testing.B) {
	benchmarkGood(b, "coroutine0002", "libcoolsched.a", "-coroutine")
}
