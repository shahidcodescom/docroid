package supervisor

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"docroid/pkg/app"
	"docroid/pkg/engine"
)

func TestSupervisorAutoRestart(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "docroid_supervisor_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store, err := app.NewStore(tempDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	// Create script that exits immediately
	scriptPath := filepath.Join(tempDir, "exit_script.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/sh\nexit 1\n"), 0755); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	testApp, err := store.ListApps()
	_ = testApp

	createdApp := &app.App{
		ID:            "sup12345",
		Name:          "crash-test",
		Engine:        app.EngineNative,
		Command:       "/bin/sh",
		Args:          []string{scriptPath},
		RestartPolicy: "always",
		Status:        app.StatusRunning,
		CreatedAt:     time.Now(),
	}

	if err := store.SaveApp(createdApp); err != nil {
		t.Fatalf("failed to save app: %v", err)
	}

	eng := engine.NewNativeEngine()
	runCfg := &engine.RunConfig{
		App:      createdApp,
		Store:    store,
		Detached: true,
	}

	proc, err := eng.Run(t.Context(), runCfg)
	if err != nil {
		t.Fatalf("failed to launch test process: %v", err)
	}

	createdApp.PID = proc.Pid
	_ = store.SaveApp(createdApp)

	sup := GetSupervisor()
	sup.Watch(createdApp, store, eng)
	defer sup.Cancel(createdApp.ID)

	// Wait up to 3 seconds for supervisor to detect exit and restart
	restarted := false
	for i := 0; i < 30; i++ {
		time.Sleep(100 * time.Millisecond)
		fresh, err := store.LoadApp("sup12345")
		if err == nil && fresh.RestartCount > 0 {
			restarted = true
			break
		}
	}

	if !restarted {
		t.Errorf("expected supervisor to restart application, RestartCount was not incremented")
	}
}
