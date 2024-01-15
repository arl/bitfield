package main_test

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func generate(t *testing.T, args ...string) {
	t.Helper()

	a := append([]string{"run", "main.go"}, args...)
	cmd := exec.Command("go", a...)
	if buf, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("generate failed:\n%s\n%s\n\noutput:\n%s", strings.Join(a, " "), err, buf)
	}
}

func test(t *testing.T, args ...string) {
	t.Helper()

	a := append([]string{"test", "-C", "testpkg"}, args...)
	cmd := exec.Command("go", a...)
	if buf, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("generate failed:\n%s\n%s\n\noutput:\n%s", strings.Join(a, " "), err, buf)
	}
}

func TestBitfield(t *testing.T) {
	generate(t, "-in", filepath.Join("testpkg", "mystruct.go"))
	test(t, "mystruct_test.go", "mystruct_bits.go")
}
