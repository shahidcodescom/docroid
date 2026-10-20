package engine

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"docroid/pkg/app"
)

func TestNativeEngineExecRun(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "docroid_exec_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store, err := app.NewStore(tempDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	testApp := &app.App{
		ID:      "test-app-1",
		Name:    "test-app",
		Engine:  app.EngineNative,
		Command: "sh",
		Args:    []string{"-c", "echo HELLO_EXEC"},
		Env: map[string]string{
			"BASE_VAR": "base_val",
		},
		Status: app.StatusRunning,
	}
	if err := store.SaveApp(testApp); err != nil {
		t.Fatalf("failed to save app: %v", err)
	}

	eng := NewNativeEngine()

	// 1. Prepare with environment and user
	var stdoutBuf, stderrBuf bytes.Buffer
	execApp := &app.App{
		ID:      testApp.ID,
		Name:    testApp.Name + "-exec",
		Engine:  testApp.Engine,
		Command: "sh",
		Args:    []string{"-c", "echo $BASE_VAR-$OVERRIDE_VAR"},
		Env: map[string]string{
			"BASE_VAR":     "base_val",
			"OVERRIDE_VAR": "override_val",
		},
		WorkDir: tempDir,
	}

	cfg := &RunConfig{
		App:         execApp,
		Store:       store,
		Stdout:      &stdoutBuf,
		Stderr:      &stderrBuf,
		Interactive: true,
		User:        "testuser",
	}

	cmd, err := eng.Prepare(cfg)
	if err != nil {
		t.Fatalf("Prepare failed: %v", err)
	}

	// Verify cmd.Dir
	if cmd.Dir != tempDir {
		t.Errorf("expected cmd.Dir to be %s, got %s", tempDir, cmd.Dir)
	}

	// Verify environment variables contain user and custom envs
	hasBase := false
	hasOverride := false
	hasUser := false
	for _, env := range cmd.Env {
		if env == "BASE_VAR=base_val" {
			hasBase = true
		}
		if env == "OVERRIDE_VAR=override_val" {
			hasOverride = true
		}
		if env == "USER=testuser" {
			hasUser = true
		}
	}
	if !hasBase || !hasOverride || !hasUser {
		t.Errorf("missing expected environment variables (base:%v, override:%v, user:%v)", hasBase, hasOverride, hasUser)
	}

	// 2. Run command
	proc, err := eng.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	state, err := proc.Wait()
	if err != nil {
		t.Fatalf("Wait failed: %v", err)
	}
	if !state.Success() {
		t.Errorf("expected success exit code, got: %d", state.ExitCode())
	}

	out := strings.TrimSpace(stdoutBuf.String())
	if out != "base_val-override_val" {
		t.Errorf("expected stdout 'base_val-override_val', got '%s'", out)
	}
}
