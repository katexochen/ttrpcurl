package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/katexochen/ttrpcurl/hack/testserver/grpctest"
	"github.com/rogpeppe/go-internal/testscript"
)

const replaceWithGrpcurl = "REPLACE_WITH_GRPCURL"

var (
	update        = flag.Bool("u", false, "update testscript output files")
	runTestscript = flag.Bool("testscript", false, "run testscript tests")
)

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"ttrpcurl":   ttrpcurlMain,
		"testserver": grpctest.ScriptMain,
	}))
}

func TestGrpcurlCompat(t *testing.T) {
	if !*runTestscript {
		t.Skip("skipping testscript tests; use -testscript to enable")
	}

	conds := map[string]bool{
		"update": *update,
	}

	env := map[string]string{
		replaceWithGrpcurl: fmt.Sprintf("%t", *update),
	}

	testscript.Run(t, testscript.Params{
		Dir:                 filepath.Join("testdata", "script", "grpcurl"),
		Condition:           conditionsFromMap(conds),
		Setup:               setupEnv(env),
		UpdateScripts:       *update,
		RequireUniqueNames:  true,
		RequireExplicitExec: true,
	})
}

func TestUI(t *testing.T) {
	if !*runTestscript {
		t.Skip("skipping testscript tests; use -testscript to enable")
	}

	testscript.Run(t, testscript.Params{
		Dir:                 filepath.Join("testdata", "script", "ui"),
		UpdateScripts:       *update,
		RequireUniqueNames:  true,
		RequireExplicitExec: true,
	})
}

func TestClientServer(t *testing.T) {
	if !*runTestscript {
		t.Skip("skipping testscript tests; use -testscript to enable")
	}

	testscript.Run(t, testscript.Params{
		Dir:                 filepath.Join("testdata", "script", "clientserver"),
		UpdateScripts:       *update,
		RequireUniqueNames:  true,
		RequireExplicitExec: true,
	})
}

func setupEnv(envVars map[string]string) func(e *testscript.Env) error {
	return func(e *testscript.Env) error {
		for k, v := range envVars {
			e.Vars = append(e.Vars, fmt.Sprintf("%s=%s", k, v))
		}
		return nil
	}
}

func conditionsFromMap(m map[string]bool) func(string) (bool, error) {
	return func(cond string) (bool, error) {
		val, ok := m[cond]
		if !ok {
			return false, fmt.Errorf("unknown condition %q", cond)
		}
		return val, nil
	}
}

func ttrpcurlMain() int {
	if os.Getenv(replaceWithGrpcurl) == "true" {
		return grpcurlMain()
	}
	if err := run(); err != nil {
		return 1
	}
	return 0
}

func grpcurlMain() int {
	cmd := exec.Command("grpcurl", os.Args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
		return 1
	}
	return 0
}
