package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
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

func TestBitfieldCLI(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("GetWd error: %v", err)
	}
	params := testscript.Params{
		Dir: "testdata",
		Setup: func(env *testscript.Env) error {
			env.Setenv("BITFIELD_DIR", wd)
			return nil
		},
		TestWork: testing.Verbose(),
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			"bitfield": func(ts *testscript.TestScript, neg bool, args []string) {
				cfg, out, err := parseFlags(args)
				if err != nil {
					ts.Fatalf("parseFlags error: %s\n\noutput: %s", err, out)
				}
				err = run(cfg)
				fmt.Fprint(ts.Stderr(), err, "\n")
				if (err != nil) != neg {
					ts.Fatalf("unexpected error: %v", err)
				}
			},
		},
	}
	testscript.Run(t, params)
}