package engine

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
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

func TestRootfsEnginePrepare(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "docroid_rootfs_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store, err := app.NewStore(tempDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	// Create simulated alpine rootfs directory
	alpineDir := store.RootfsDir("alpine")
	if err := os.MkdirAll(alpineDir, 0755); err != nil {
		t.Fatalf("failed to create rootfs dir: %v", err)
	}

	testApp := &app.App{
		ID:      "test-rootfs-1",
		Name:    "test-rootfs",
		Engine:  app.EngineRootfs,
		Rootfs:  "alpine",
		Command: "/bin/sh",
		Args:    []string{"-c", "echo hello"},
		Port:    8080,
		Env: map[string]string{
			"MY_CUSTOM_VAR": "custom_val",
		},
		Status: app.StatusRunning,
	}

	eng := &RootfsEngine{prootPath: "/data/data/com.termux/files/usr/bin/proot"}

	cfg := &RunConfig{
		App:         testApp,
		Store:       store,
		Interactive: false,
		User:        "appuser",
	}

	// Temporarily set LD_PRELOAD in environment to test that it is excluded
	os.Setenv("LD_PRELOAD", "/data/data/com.termux/files/usr/lib/libtermux-exec.so")
	defer os.Unsetenv("LD_PRELOAD")

	cmd, err := eng.Prepare(cfg)
	if err != nil {
		t.Fatalf("Prepare failed: %v", err)
	}

	// 1. Verify prootArgs NEVER contains unknown "-E" flag
	for i, arg := range cmd.Args {
		if arg == "-E" {
			t.Fatalf("found invalid proot argument '-E' at index %d", i)
		}
	}

	// 2. Verify prootArgs contains --link2symlink when Termux proot is used
	hasLink2Symlink := false
	for _, arg := range cmd.Args {
		if arg == "--link2symlink" {
			hasLink2Symlink = true
			break
		}
	}
	if !hasLink2Symlink {
		t.Errorf("expected '--link2symlink' in proot args for Termux proot")
	}

	// 3. Verify environment variables in cmd.Env
	envMap := make(map[string]string)
	for _, env := range cmd.Env {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	if envMap["PORT"] != "8080" {
		t.Errorf("expected PORT=8080 in cmd.Env, got: %s", envMap["PORT"])
	}
	if envMap["MY_CUSTOM_VAR"] != "custom_val" {
		t.Errorf("expected MY_CUSTOM_VAR=custom_val in cmd.Env, got: %s", envMap["MY_CUSTOM_VAR"])
	}
	if envMap["USER"] != "appuser" {
		t.Errorf("expected USER=appuser in cmd.Env, got: %s", envMap["USER"])
	}
	if envMap["HOME"] != "/home/appuser" {
		t.Errorf("expected HOME=/home/appuser in cmd.Env, got: %s", envMap["HOME"])
	}
	if envMap["PATH"] != "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin" {
		t.Errorf("expected rootfs PATH in cmd.Env, got: %s", envMap["PATH"])
	}
	if _, hasLD := envMap["LD_PRELOAD"]; hasLD {
		t.Errorf("LD_PRELOAD should have been filtered out of cmd.Env")
	}
}

func TestRootfsEngineDockerImageResolution(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "docroid_docker_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store, err := app.NewStore(tempDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	// Create simulated pulled image ~/.docroid/images/bkimminich/juice-shop:latest
	imageDir := filepath.Join(store.BaseDir(), "images", "bkimminich", "juice-shop:latest")
	payloadDir := filepath.Join(imageDir, "payload")
	if err := os.MkdirAll(filepath.Join(payloadDir, "bin"), 0755); err != nil {
		t.Fatalf("failed to create payload bin: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(payloadDir, "etc"), 0755); err != nil {
		t.Fatalf("failed to create payload etc: %v", err)
	}
	if err := os.WriteFile(filepath.Join(payloadDir, "bin", "sh"), []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatalf("failed to create dummy sh: %v", err)
	}

	manifestJSON := []byte(`{
		"format_version": "1.0",
		"name": "bkimminich/juice-shop",
		"version": "latest",
		"engine": "rootfs",
		"command": "/nodejs/bin/node",
		"args": ["/juice-shop/build/app.js"],
		"port": 3000,
		"env": {
			"PATH": "/nodejs/bin:/usr/local/bin:/usr/bin",
			"NODE_ENV": "production"
		},
		"work_dir": "/juice-shop"
	}`)
	if err := os.WriteFile(filepath.Join(imageDir, "manifest.json"), manifestJSON, 0644); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}

	testApp := &app.App{
		ID:      "test-app-juice",
		Name:    "juice-shop",
		Engine:  app.EngineRootfs,
		Command: "bkimminich/juice-shop:latest",
		Status:  app.StatusRunning,
	}

	eng := &RootfsEngine{prootPath: "/data/data/com.termux/files/usr/bin/proot"}
	cfg := &RunConfig{
		App:         testApp,
		Store:       store,
		Interactive: false,
	}

	cmd, err := eng.Prepare(cfg)
	if err != nil {
		t.Fatalf("Prepare failed: %v", err)
	}

	// Verify command inside rootfs is the resolved node binary, NOT the docker image tag string
	cmdArgsStr := strings.Join(cmd.Args, " ")
	if strings.Contains(cmdArgsStr, "bkimminich/juice-shop:latest") {
		t.Errorf("cmd.Args should not contain image tag as binary: %v", cmd.Args)
	}
	if !strings.Contains(cmdArgsStr, "/nodejs/bin/node") {
		t.Errorf("cmd.Args should execute /nodejs/bin/node, got: %v", cmd.Args)
	}
	if !strings.Contains(cmdArgsStr, "/juice-shop/build/app.js") {
		t.Errorf("cmd.Args should pass /juice-shop/build/app.js, got: %v", cmd.Args)
	}

	// Verify -w was set to image work_dir "/juice-shop"
	hasWorkDir := false
	for i, arg := range cmd.Args {
		if arg == "-w" && i+1 < len(cmd.Args) && cmd.Args[i+1] == "/juice-shop" {
			hasWorkDir = true
			break
		}
	}
	if !hasWorkDir {
		t.Errorf("expected -w /juice-shop in proot args, got: %v", cmd.Args)
	}

	// Verify cmd.Env has the image PATH and environment
	hasNodePath := false
	for _, env := range cmd.Env {
		if env == "PATH=/nodejs/bin:/usr/local/bin:/usr/bin" {
			hasNodePath = true
			break
		}
	}
	if !hasNodePath {
		t.Errorf("expected image PATH in cmd.Env, got: %v", cmd.Env)
	}
}


