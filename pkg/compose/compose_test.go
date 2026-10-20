package compose

import (
	"os"
	"strings"
	"testing"

	"docroid/pkg/app"
)

func TestParseMultiServiceCompose(t *testing.T) {
	yamlData := `
version: "1.0"
services:
  web:
    name: web-frontend
    engine: native
    command: python3 -m http.server 8080
    ports:
      - "8080:8080"
    environment:
      APP_ENV: production
      PORT: "8080"
    restart: always
    volumes:
      - "/sdcard/Download:/downloads"
    secrets:
      - DB_PASS

  worker:
    name: bg-worker
    command: ["./worker", "--threads", "4"]
    environment:
      - QUEUE=priority
      - RETRIES=3
    restart: on-failure
`

	cfg, err := ParseYAML([]byte(yamlData))
	if err != nil {
		t.Fatalf("failed to parse compose YAML: %v", err)
	}

	if len(cfg.Services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(cfg.Services))
	}

	// Test web service
	webService, ok := cfg.Services["web"]
	if !ok {
		t.Fatalf("service 'web' not found in parsed config")
	}

	resWeb, err := NormalizeService("web", &webService)
	if err != nil {
		t.Fatalf("failed to normalize web service: %v", err)
	}

	if resWeb.Name != "web-frontend" {
		t.Errorf("expected name 'web-frontend', got '%s'", resWeb.Name)
	}
	if resWeb.Command != "python3" {
		t.Errorf("expected command 'python3', got '%s'", resWeb.Command)
	}
	if len(resWeb.Args) != 3 || resWeb.Args[0] != "-m" || resWeb.Args[1] != "http.server" || resWeb.Args[2] != "8080" {
		t.Errorf("unexpected args: %v", resWeb.Args)
	}
	if resWeb.PortMapping != "8080:8080" {
		t.Errorf("expected port mapping '8080:8080', got '%s'", resWeb.PortMapping)
	}
	if resWeb.Environment["APP_ENV"] != "production" {
		t.Errorf("expected APP_ENV=production, got '%s'", resWeb.Environment["APP_ENV"])
	}
	if resWeb.RestartPolicy != "always" {
		t.Errorf("expected restart policy 'always', got '%s'", resWeb.RestartPolicy)
	}

	// Test worker service
	workerService, ok := cfg.Services["worker"]
	if !ok {
		t.Fatalf("service 'worker' not found in parsed config")
	}

	resWorker, err := NormalizeService("worker", &workerService)
	if err != nil {
		t.Fatalf("failed to normalize worker service: %v", err)
	}

	if resWorker.Name != "bg-worker" {
		t.Errorf("expected name 'bg-worker', got '%s'", resWorker.Name)
	}
	if resWorker.Command != "./worker" {
		t.Errorf("expected command './worker', got '%s'", resWorker.Command)
	}
	if len(resWorker.Args) != 2 || resWorker.Args[0] != "--threads" || resWorker.Args[1] != "4" {
		t.Errorf("unexpected args: %v", resWorker.Args)
	}
	if resWorker.Environment["QUEUE"] != "priority" || resWorker.Environment["RETRIES"] != "3" {
		t.Errorf("unexpected environment list parsing: %v", resWorker.Environment)
	}
	if resWorker.RestartPolicy != "on-failure" {
		t.Errorf("expected restart policy 'on-failure', got '%s'", resWorker.RestartPolicy)
	}
}

func TestParseSingleServiceRootCompose(t *testing.T) {
	yamlData := `
name: single-app
command: ./run.sh
ports:
  - "3000"
restart: unless-stopped
environment:
  MODE: standalone
`

	cfg, err := ParseYAML([]byte(yamlData))
	if err != nil {
		t.Fatalf("failed to parse single-service compose: %v", err)
	}

	if len(cfg.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(cfg.Services))
	}

	var service ServiceSpec
	for _, s := range cfg.Services {
		service = s
	}

	res, err := NormalizeService("app", &service)
	if err != nil {
		t.Fatalf("failed to normalize service: %v", err)
	}

	if res.Name != "single-app" {
		t.Errorf("expected name 'single-app', got '%s'", res.Name)
	}
	if res.Command != "./run.sh" {
		t.Errorf("expected command './run.sh', got '%s'", res.Command)
	}
	if res.RestartPolicy != "always" {
		t.Errorf("expected restart policy 'always' (mapped from unless-stopped), got '%s'", res.RestartPolicy)
	}
	if res.Environment["MODE"] != "standalone" {
		t.Errorf("expected MODE=standalone, got '%s'", res.Environment["MODE"])
	}
}

func TestGenerateAndRoundtripComposeYAML(t *testing.T) {
	yamlBytes, err := GenerateComposeYAML(
		"api-service",
		"native",
		"",
		"go",
		[]string{"run", "main.go"},
		"8080:8080",
		"always",
		"/data/app",
		map[string]string{"PORT": "8080", "RELEASE": "1.0"},
		[]string{"/sdcard/data:/data"},
		[]string{"API_KEY"},
		true,
	)
	if err != nil {
		t.Fatalf("failed to generate compose YAML: %v", err)
	}

	yamlStr := string(yamlBytes)
	if !strings.Contains(yamlStr, "api-service") {
		t.Errorf("expected generated YAML to contain 'api-service', got:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "8080:8080") {
		t.Errorf("expected generated YAML to contain '8080:8080', got:\n%s", yamlStr)
	}

	// Round-trip parse
	cfg, err := ParseYAML(yamlBytes)
	if err != nil {
		t.Fatalf("failed to round-trip parse generated YAML: %v", err)
	}

	svc, ok := cfg.Services["api-service"]
	if !ok {
		t.Fatalf("service 'api-service' missing in roundtrip config")
	}

	res, err := NormalizeService("api-service", &svc)
	if err != nil {
		t.Fatalf("failed to normalize round-trip service: %v", err)
	}

	if res.Name != "api-service" || res.Command != "go" || res.PortMapping != "8080:8080" {
		t.Errorf("mismatched parsed fields: %+v", res)
	}
}

func TestServiceToAppConversion(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "docroid_compose_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store, err := app.NewStore(tempDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	spec := &ServiceSpec{
		Name:        "converted-app",
		Command:     "./bin/server",
		Args:        []string{"--verbose"},
		Ports:       []string{"18080:18080"},
		Volumes:     []string{"/tmp/host:/container/path:ro"},
		Restart:     "on-failure",
		Environment: map[string]string{"ENV": "test"},
	}

	resolved, err := NormalizeService("test", spec)
	if err != nil {
		t.Fatalf("failed to normalize: %v", err)
	}

	application, err := ServiceToApp(resolved, store)
	if err != nil {
		t.Fatalf("failed to convert service to app: %v", err)
	}

	if application.Name != "converted-app" {
		t.Errorf("expected name 'converted-app', got '%s'", application.Name)
	}
	if len(application.Volumes) != 1 {
		t.Fatalf("expected 1 volume, got %d", len(application.Volumes))
	}
	if application.Volumes[0].HostPath != "/tmp/host" || !application.Volumes[0].ReadOnly {
		t.Errorf("unexpected volume mount: %+v", application.Volumes[0])
	}
	if application.RestartPolicy != "on-failure" {
		t.Errorf("expected restart policy 'on-failure', got '%s'", application.RestartPolicy)
	}
}
