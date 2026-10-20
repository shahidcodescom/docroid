package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"docroid/pkg/app"
)

func TestAppExecAPI(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// 1. Create a dummy app
	runningApp := &app.App{
		ID:      "exec-test-app",
		Name:    "exec-test-app",
		Engine:  app.EngineNative,
		Command: "sleep",
		Args:    []string{"100"},
		Env: map[string]string{
			"APP_ENV_KEY": "app_env_val",
		},
		Status: app.StatusRunning,
	}
	if err := server.manager.Store().SaveApp(runningApp); err != nil {
		t.Fatalf("failed to save app: %v", err)
	}

	// 2. Exec with valid command and env
	payload := map[string]interface{}{
		"command": "sh",
		"args":    []string{"-c", "echo $APP_ENV_KEY-$CUSTOM_KEY"},
		"env":     []string{"CUSTOM_KEY=custom_val"},
	}
	bodyBytes, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/apps/exec-test-app/exec", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()
	server.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	var res struct {
		ExitCode int    `json:"exit_code"`
		Stdout   string `json:"stdout"`
		Stderr   string `json:"stderr"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if res.ExitCode != 0 {
		t.Errorf("expected exit_code 0, got %d", res.ExitCode)
	}
	if strings.TrimSpace(res.Stdout) != "app_env_val-custom_val" {
		t.Errorf("expected stdout 'app_env_val-custom_val', got '%s'", res.Stdout)
	}

	// 3. Test non-zero exit code
	failPayload := map[string]interface{}{
		"command": "sh",
		"args":    []string{"-c", "exit 42"},
	}
	failBytes, _ := json.Marshal(failPayload)
	reqFail := httptest.NewRequest(http.MethodPost, "/api/apps/exec-test-app/exec", bytes.NewReader(failBytes))
	recFail := httptest.NewRecorder()
	server.mux.ServeHTTP(recFail, reqFail)

	if recFail.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recFail.Code)
	}
	var resFail struct {
		ExitCode int `json:"exit_code"`
	}
	_ = json.Unmarshal(recFail.Body.Bytes(), &resFail)
	if resFail.ExitCode != 42 {
		t.Errorf("expected exit_code 42, got %d", resFail.ExitCode)
	}

	// 4. Test stopped container without force
	stoppedApp := &app.App{
		ID:     "stopped-app",
		Name:   "stopped-app",
		Engine: app.EngineNative,
		Status: app.StatusStopped,
	}
	_ = server.manager.Store().SaveApp(stoppedApp)

	reqStopped := httptest.NewRequest(http.MethodPost, "/api/apps/stopped-app/exec", bytes.NewReader(bodyBytes))
	recStopped := httptest.NewRecorder()
	server.mux.ServeHTTP(recStopped, reqStopped)

	if recStopped.Code != http.StatusConflict {
		t.Errorf("expected 409 Conflict for stopped app, got %d", recStopped.Code)
	}

	// 5. Test stopped container WITH force: true
	forcePayload := map[string]interface{}{
		"command": "sh",
		"args":    []string{"-c", "echo FORCED"},
		"force":   true,
	}
	forceBytes, _ := json.Marshal(forcePayload)
	reqForce := httptest.NewRequest(http.MethodPost, "/api/apps/stopped-app/exec", bytes.NewReader(forceBytes))
	recForce := httptest.NewRecorder()
	server.mux.ServeHTTP(recForce, reqForce)

	if recForce.Code != http.StatusOK {
		t.Errorf("expected status 200 with force, got %d", recForce.Code)
	}
}
