package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"docroid/pkg/app"
	"docroid/pkg/rootfs"
)

func setupTestServer(t *testing.T) (*Server, func()) {
	tempDir, err := os.MkdirTemp("", "docroid_web_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	store, err := app.NewStore(tempDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	manager := app.NewAppManager(store)
	rootfsMgr := rootfs.NewManager(store)
	server := NewServer(manager, rootfsMgr)

	cleanup := func() {
		_ = os.RemoveAll(tempDir)
	}

	return server, cleanup
}

func TestServeEmbeddedStaticFiles(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	server.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Docroid") {
		t.Errorf("expected HTML body to contain 'Docroid', got: %s", body)
	}
}

func TestGetAppsAPI(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/apps", nil)
	rec := httptest.NewRecorder()
	server.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var apps []AppWithStatsJSON
	if err := json.Unmarshal(rec.Body.Bytes(), &apps); err != nil {
		t.Fatalf("failed to parse apps JSON: %v", err)
	}

	if len(apps) != 0 {
		t.Errorf("expected 0 apps, got %d", len(apps))
	}
}

func TestGetSystemAPI(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/system", nil)
	rec := httptest.NewRecorder()
	server.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &data); err != nil {
		t.Fatalf("failed to parse system JSON: %v", err)
	}

	if _, ok := data["checks"]; !ok {
		t.Errorf("expected 'checks' field in system response")
	}
}

func TestCreateAppAPI(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	payload := `{"name":"api-test","command":"echo","args":["hello"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/apps", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	var created app.App
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to parse created app: %v", err)
	}

	if created.Name != "api-test" {
		t.Errorf("expected name 'api-test', got '%s'", created.Name)
	}
}

func TestGetImagesAPI(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/images", nil)
	rec := httptest.NewRecorder()
	server.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestImageBuildAndManageAPI(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// 1. Build an image via /api/images/build
	buildPayload := `{
		"tag": "mytest:1.0",
		"dockerfile": "FROM local\nWORKDIR /app\nRUN echo 'built via api' > test.txt\nCMD echo 'api container'"
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/images/build", strings.NewReader(buildPayload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201 from /api/images/build, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	var buildRes map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &buildRes); err != nil {
		t.Fatalf("failed to parse build response: %v", err)
	}
	if buildRes["success"] != true {
		t.Errorf("expected success=true, got %v", buildRes["success"])
	}

	// 2. Query /api/images
	req = httptest.NewRequest(http.MethodGet, "/api/images", nil)
	rec = httptest.NewRecorder()
	server.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 from /api/images, got %d", rec.Code)
	}

	var images []map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &images)
	found := false
	for _, img := range images {
		if img["name"] == "mytest" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected built image 'mytest' in /api/images, got: %s", rec.Body.String())
	}

	// 3. Inspect via /api/images/mytest:1.0
	req = httptest.NewRequest(http.MethodGet, "/api/images/mytest:1.0", nil)
	rec = httptest.NewRecorder()
	server.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200 from /api/images/mytest:1.0, got %d", rec.Code)
	}

	// 4. Delete via DELETE /api/images/mytest:1.0
	req = httptest.NewRequest(http.MethodDelete, "/api/images/mytest:1.0", nil)
	rec = httptest.NewRecorder()
	server.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200 from DELETE /api/images/mytest:1.0, got %d", rec.Code)
	}
}

func TestExportAppAPI(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// 1. Create an app first
	created, err := server.manager.CreateApp("export-test", app.EngineNative, "echo", []string{"hi"}, nil, "", 0, nil, "", "no")
	if err != nil {
		t.Fatalf("failed to create test app: %v", err)
	}

	// 2. Call export endpoint
	req := httptest.NewRequest(http.MethodGet, "/api/apps/"+created.ID+"/export", nil)
	rec := httptest.NewRecorder()
	server.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/gzip" {
		t.Errorf("expected Content-Type 'application/gzip', got '%s'", contentType)
	}
}

func TestRollingRestartAPI(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	created, err := server.manager.CreateApp("restart-test", app.EngineNative, "sleep", []string{"10"}, nil, "", 0, nil, "", "no")
	if err != nil {
		t.Fatalf("failed to create test app: %v", err)
	}

	// Start it first
	startReq := httptest.NewRequest(http.MethodPost, "/api/apps/"+created.ID+"/start", nil)
	startRec := httptest.NewRecorder()
	server.mux.ServeHTTP(startRec, startReq)
	if startRec.Code != http.StatusOK {
		t.Fatalf("failed to start app: %d (%s)", startRec.Code, startRec.Body.String())
	}

	// Test zero-downtime rolling restart
	rollReq := httptest.NewRequest(http.MethodPost, "/api/apps/"+created.ID+"/rolling-restart", nil)
	rollRec := httptest.NewRecorder()
	server.mux.ServeHTTP(rollRec, rollReq)

	if rollRec.Code != http.StatusOK {
		t.Fatalf("expected status 200 on rolling-restart, got %d (%s)", rollRec.Code, rollRec.Body.String())
	}

	var res map[string]interface{}
	if err := json.Unmarshal(rollRec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse rolling-restart json: %v", err)
	}
	if status, ok := res["status"].(string); !ok || status != "completed" {
		t.Errorf("expected status 'completed', got %v", res["status"])
	}

	// Cleanup running app
	_ = server.manager.StopApp(created.ID)
}

func TestUpdateAPI(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	created, err := server.manager.CreateApp("update-test", app.EngineNative, "sleep", []string{"10"}, nil, "", 0, nil, "", "no")
	if err != nil {
		t.Fatalf("failed to create test app: %v", err)
	}

	// Start app
	startReq := httptest.NewRequest(http.MethodPost, "/api/apps/"+created.ID+"/start", nil)
	startRec := httptest.NewRecorder()
	server.mux.ServeHTTP(startRec, startReq)
	if startRec.Code != http.StatusOK {
		t.Fatalf("failed to start app: %d (%s)", startRec.Code, startRec.Body.String())
	}

	// Call update endpoint to change command and args
	payload := `{"command":"sleep","args":["15"],"env":{"VERSION":"2.0"},"drain_timeout":0.2,"probe_timeout":3.0}`
	updateReq := httptest.NewRequest(http.MethodPost, "/api/apps/"+created.ID+"/update", strings.NewReader(payload))
	updateReq.Header.Set("Content-Type", "application/json")
	updateRec := httptest.NewRecorder()
	server.mux.ServeHTTP(updateRec, updateReq)

	if updateRec.Code != http.StatusOK {
		t.Fatalf("expected status 200 on update, got %d (%s)", updateRec.Code, updateRec.Body.String())
	}

	var res map[string]interface{}
	if err := json.Unmarshal(updateRec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse update json: %v", err)
	}
	if status, ok := res["status"].(string); !ok || status != "completed" {
		t.Errorf("expected status 'completed', got %v", res["status"])
	}

	// Verify app metadata was updated in store
	updatedApp, err := server.manager.Store().FindApp(created.ID)
	if err != nil {
		t.Fatalf("failed to find updated app: %v", err)
	}
	if len(updatedApp.Args) == 0 || updatedApp.Args[0] != "15" {
		t.Errorf("expected updated args ['15'], got %v", updatedApp.Args)
	}
	if updatedApp.Env["VERSION"] != "2.0" {
		t.Errorf("expected env VERSION=2.0, got %v", updatedApp.Env["VERSION"])
	}

	// Cleanup
	_ = server.manager.StopApp(created.ID)
}

func TestComposeParseAPI(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	yamlData := `
version: "1.0"
services:
  api:
    name: my-api-service
    command: python3 -m http.server 9000
    ports:
      - "9000:9000"
    restart: always
    environment:
      ENV: testing
`
	req := httptest.NewRequest(http.MethodPost, "/api/compose/parse", strings.NewReader(yamlData))
	req.Header.Set("Content-Type", "text/yaml")
	rec := httptest.NewRecorder()
	server.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d (%s)", rec.Code, rec.Body.String())
	}

	var parsed struct {
		Version  string `json:"version"`
		Services []struct {
			Name        string            `json:"name"`
			Command     string            `json:"command"`
			PortMapping string            `json:"port_mapping"`
			Environment map[string]string `json:"environment"`
		} `json:"services"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}

	if len(parsed.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(parsed.Services))
	}

	s := parsed.Services[0]
	if s.Name != "my-api-service" || s.Command != "python3" || s.PortMapping != "9000:9000" {
		t.Errorf("unexpected parsed service fields: %+v", s)
	}
	if s.Environment["ENV"] != "testing" {
		t.Errorf("expected ENV=testing, got %v", s.Environment["ENV"])
	}
}

func TestComposeGenerateAPI(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	payload := `{
		"name": "gen-service",
		"engine": "native",
		"command": "./server",
		"args": ["--port", "8080"],
		"port_mapping": "8080:8080",
		"restart_policy": "always",
		"env": {"DEBUG": "true"}
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/compose/generate", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d (%s)", rec.Code, rec.Body.String())
	}

	cd := rec.Header().Get("Content-Disposition")
	if !strings.Contains(cd, "docroid-compose-gen-service.yml") {
		t.Errorf("expected attachment filename with gen-service, got: %s", cd)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "gen-service") || !strings.Contains(body, "8080:8080") {
		t.Errorf("expected YAML body with gen-service and 8080:8080, got:\n%s", body)
	}
}

func TestStaticSitesAPI(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// 1. Initial list should be empty
	reqList := httptest.NewRequest(http.MethodGet, "/api/sites", nil)
	recList := httptest.NewRecorder()
	server.mux.ServeHTTP(recList, reqList)
	if recList.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recList.Code)
	}

	// 2. Create site with JSON
	autoStart := false
	jsonPayload := `{
		"name": "my-json-site",
		"port": 9191,
		"domain": "json.local",
		"spa": true,
		"html": "<h1>Direct JSON HTML</h1>",
		"auto_start": false
	}`
	_ = autoStart
	reqCreate := httptest.NewRequest(http.MethodPost, "/api/sites", strings.NewReader(jsonPayload))
	reqCreate.Header.Set("Content-Type", "application/json")
	recCreate := httptest.NewRecorder()
	server.mux.ServeHTTP(recCreate, reqCreate)

	if recCreate.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d (%s)", recCreate.Code, recCreate.Body.String())
	}

	var created map[string]interface{}
	if err := json.Unmarshal(recCreate.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	siteID := created["id"].(string)

	// 3. Inspect files
	reqFiles := httptest.NewRequest(http.MethodGet, "/api/sites/"+siteID+"/files", nil)
	recFiles := httptest.NewRecorder()
	server.mux.ServeHTTP(recFiles, reqFiles)
	if recFiles.Code != http.StatusOK {
		t.Fatalf("expected 200 for files, got %d", recFiles.Code)
	}

	// 4. Inspect raw file content
	reqContent := httptest.NewRequest(http.MethodGet, "/api/sites/"+siteID+"/files?path=index.html&content=true", nil)
	recContent := httptest.NewRecorder()
	server.mux.ServeHTTP(recContent, reqContent)
	if recContent.Code != http.StatusOK || !strings.Contains(recContent.Body.String(), "Direct JSON HTML") {
		t.Fatalf("expected file content, got %s", recContent.Body.String())
	}

	// 5. Export ZIP
	reqExport := httptest.NewRequest(http.MethodGet, "/api/sites/"+siteID+"/export", nil)
	recExport := httptest.NewRecorder()
	server.mux.ServeHTTP(recExport, reqExport)
	if recExport.Code != http.StatusOK {
		t.Fatalf("expected 200 for export, got %d", recExport.Code)
	}

	// 6. Start site via API
	reqStart := httptest.NewRequest(http.MethodPost, "/api/sites/"+siteID+"/start", nil)
	recStart := httptest.NewRecorder()
	server.mux.ServeHTTP(recStart, reqStart)
	if recStart.Code != http.StatusOK {
		t.Fatalf("expected 200 for start, got %d (%s)", recStart.Code, recStart.Body.String())
	}

	// 7. Stop site via API (MUST NOT kill process / must succeed)
	reqStop := httptest.NewRequest(http.MethodPost, "/api/sites/"+siteID+"/stop", nil)
	recStop := httptest.NewRecorder()
	server.mux.ServeHTTP(recStop, reqStop)
	if recStop.Code != http.StatusOK {
		t.Fatalf("expected 200 for stop, got %d (%s)", recStop.Code, recStop.Body.String())
	}

	var stoppedSite map[string]interface{}
	if err := json.Unmarshal(recStop.Body.Bytes(), &stoppedSite); err != nil {
		t.Fatalf("failed to decode stop response: %v", err)
	}
	if stoppedSite["status"] != "stopped" {
		t.Errorf("expected status 'stopped', got '%v'", stoppedSite["status"])
	}

	// 8. Delete site
	reqDel := httptest.NewRequest(http.MethodDelete, "/api/sites/"+siteID, nil)
	recDel := httptest.NewRecorder()
	server.mux.ServeHTTP(recDel, reqDel)
	if recDel.Code != http.StatusOK {
		t.Fatalf("expected 200 for delete, got %d", recDel.Code)
	}
}

func TestClusterAPI(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// 1. Initial cluster status
	reqStatus := httptest.NewRequest(http.MethodGet, "/api/cluster/status", nil)
	recStatus := httptest.NewRecorder()
	server.mux.ServeHTTP(recStatus, reqStatus)
	if recStatus.Code != http.StatusOK {
		t.Fatalf("expected 200 for cluster status, got %d", recStatus.Code)
	}

	// 2. Initialize cluster
	initPayload := `{"name": "test-web-mesh", "port": 9090}`
	reqInit := httptest.NewRequest(http.MethodPost, "/api/cluster/init", strings.NewReader(initPayload))
	reqInit.Header.Set("Content-Type", "application/json")
	recInit := httptest.NewRecorder()
	server.mux.ServeHTTP(recInit, reqInit)
	if recInit.Code != http.StatusOK {
		t.Fatalf("expected 200 for cluster init, got %d: %s", recInit.Code, recInit.Body.String())
	}

	// 3. Check nodes list
	reqNodes := httptest.NewRequest(http.MethodGet, "/api/cluster/nodes", nil)
	recNodes := httptest.NewRecorder()
	server.mux.ServeHTTP(recNodes, reqNodes)
	if recNodes.Code != http.StatusOK {
		t.Fatalf("expected 200 for cluster nodes, got %d", recNodes.Code)
	}
	if !strings.Contains(recNodes.Body.String(), "control-plane") {
		t.Fatalf("expected control-plane node in response, got %s", recNodes.Body.String())
	}

	// 4. Create deployment
	depPayload := `{
		"name": "web-cluster-app",
		"replicas": 2,
		"template": {
			"name": "web-worker",
			"command": "echo",
			"args": ["running pod"]
		}
	}`
	reqDep := httptest.NewRequest(http.MethodPost, "/api/cluster/deployments", strings.NewReader(depPayload))
	reqDep.Header.Set("Content-Type", "application/json")
	recDep := httptest.NewRecorder()
	server.mux.ServeHTTP(recDep, reqDep)
	if recDep.Code != http.StatusCreated {
		t.Fatalf("expected 201 for deployment create, got %d: %s", recDep.Code, recDep.Body.String())
	}

	// Wait briefly for scale goroutine
	time.Sleep(200 * time.Millisecond)

	// 5. Scale deployment to 3
	scalePayload := `{"deployment": "web-cluster-app", "replicas": 3}`
	reqScale := httptest.NewRequest(http.MethodPost, "/api/cluster/scale", strings.NewReader(scalePayload))
	reqScale.Header.Set("Content-Type", "application/json")
	recScale := httptest.NewRecorder()
	server.mux.ServeHTTP(recScale, reqScale)
	if recScale.Code != http.StatusOK {
		t.Fatalf("expected 200 for scale, got %d: %s", recScale.Code, recScale.Body.String())
	}

	// 6. List pods
	reqPods := httptest.NewRequest(http.MethodGet, "/api/cluster/pods", nil)
	recPods := httptest.NewRecorder()
	server.mux.ServeHTTP(recPods, reqPods)
	if recPods.Code != http.StatusOK {
		t.Fatalf("expected 200 for pods list, got %d", recPods.Code)
	}
}

func TestNetworksAPI(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// 1. List networks (should have bridge & host)
	reqList := httptest.NewRequest(http.MethodGet, "/api/networks", nil)
	recList := httptest.NewRecorder()
	server.mux.ServeHTTP(recList, reqList)
	if recList.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", recList.Code)
	}
	if !strings.Contains(recList.Body.String(), "bridge") {
		t.Errorf("expected bridge network in list")
	}

	// 2. Create custom network
	createPayload := `{"name": "test-custom-net", "driver": "bridge", "subnet": "10.98.0.0/16"}`
	reqCreate := httptest.NewRequest(http.MethodPost, "/api/networks", strings.NewReader(createPayload))
	reqCreate.Header.Set("Content-Type", "application/json")
	recCreate := httptest.NewRecorder()
	server.mux.ServeHTTP(recCreate, reqCreate)
	if recCreate.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", recCreate.Code, recCreate.Body.String())
	}

	// 3. Inspect network
	reqInspect := httptest.NewRequest(http.MethodGet, "/api/networks/test-custom-net", nil)
	recInspect := httptest.NewRecorder()
	server.mux.ServeHTTP(recInspect, reqInspect)
	if recInspect.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", recInspect.Code)
	}
	if !strings.Contains(recInspect.Body.String(), "10.98.0.0/16") {
		t.Errorf("expected subnet 10.98.0.0/16 in inspect")
	}

	// 4. Create an app and connect to network
	appInst, err := server.manager.CreateApp("db-test-app", app.EngineNative, "echo", []string{"hello"}, nil, "", 5432, nil, "", "no")
	if err != nil {
		t.Fatalf("failed to create test app: %v", err)
	}

	connPayload := fmt.Sprintf(`{"container": "%s", "aliases": ["db", "postgres"]}`, appInst.ID)
	reqConn := httptest.NewRequest(http.MethodPost, "/api/networks/test-custom-net/connect", strings.NewReader(connPayload))
	reqConn.Header.Set("Content-Type", "application/json")
	recConn := httptest.NewRecorder()
	server.mux.ServeHTTP(recConn, reqConn)
	if recConn.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for connect, got %d: %s", recConn.Code, recConn.Body.String())
	}
	if !strings.Contains(recConn.Body.String(), "10.98.0.2") {
		t.Errorf("expected IP 10.98.0.2 for connected container")
	}

	// 5. Disconnect container
	disconnPayload := fmt.Sprintf(`{"container": "%s"}`, appInst.ID)
	reqDisconn := httptest.NewRequest(http.MethodPost, "/api/networks/test-custom-net/disconnect", strings.NewReader(disconnPayload))
	reqDisconn.Header.Set("Content-Type", "application/json")
	recDisconn := httptest.NewRecorder()
	server.mux.ServeHTTP(recDisconn, reqDisconn)
	if recDisconn.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for disconnect, got %d", recDisconn.Code)
	}

	// 6. Delete network
	reqDel := httptest.NewRequest(http.MethodDelete, "/api/networks/test-custom-net", nil)
	recDel := httptest.NewRecorder()
	server.mux.ServeHTTP(recDel, reqDel)
	if recDel.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for delete, got %d", recDel.Code)
	}
}

func TestTunnelsAPI(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// 1. Check tunnel status endpoint
	reqStatus := httptest.NewRequest(http.MethodGet, "/api/tunnels/status", nil)
	recStatus := httptest.NewRecorder()
	server.mux.ServeHTTP(recStatus, reqStatus)
	if recStatus.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /api/tunnels/status, got %d", recStatus.Code)
	}

	// 2. Discover targets
	reqTargets := httptest.NewRequest(http.MethodGet, "/api/tunnels/targets", nil)
	recTargets := httptest.NewRecorder()
	server.mux.ServeHTTP(recTargets, reqTargets)
	if recTargets.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /api/tunnels/targets, got %d", recTargets.Code)
	}

	// 3. Create tunnel without auto-start
	createPayload := `{"name":"api-tunnel","target":"8080","target_port":8080,"mode":"quick","auto_start":false}`
	reqCreate := httptest.NewRequest(http.MethodPost, "/api/tunnels", strings.NewReader(createPayload))
	reqCreate.Header.Set("Content-Type", "application/json")
	recCreate := httptest.NewRecorder()
	server.mux.ServeHTTP(recCreate, reqCreate)
	if recCreate.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for /api/tunnels, got %d: %s", recCreate.Code, recCreate.Body.String())
	}

	var created struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Mode string `json:"mode"`
		Port int    `json:"target_port"`
	}
	if err := json.NewDecoder(recCreate.Body).Decode(&created); err != nil {
		t.Fatalf("failed to decode created tunnel: %v", err)
	}
	if created.ID == "" || created.Name != "api-tunnel" || created.Port != 8080 {
		t.Fatalf("unexpected tunnel properties: %+v", created)
	}

	// 4. List tunnels
	reqList := httptest.NewRequest(http.MethodGet, "/api/tunnels", nil)
	recList := httptest.NewRecorder()
	server.mux.ServeHTTP(recList, reqList)
	if recList.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /api/tunnels, got %d", recList.Code)
	}

	// 5. Get tunnel by ID
	reqGet := httptest.NewRequest(http.MethodGet, "/api/tunnels/"+created.ID, nil)
	recGet := httptest.NewRecorder()
	server.mux.ServeHTTP(recGet, reqGet)
	if recGet.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for get tunnel, got %d", recGet.Code)
	}

	// 6. Delete tunnel
	reqDel := httptest.NewRequest(http.MethodDelete, "/api/tunnels/"+created.ID, nil)
	recDel := httptest.NewRecorder()
	server.mux.ServeHTTP(recDel, reqDel)
	if recDel.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for delete tunnel, got %d", recDel.Code)
	}
}




