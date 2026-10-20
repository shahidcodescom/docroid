package registry

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Client handles OCI and Docker Registry V2 HTTP API interactions
type Client struct {
	httpClient *http.Client
	tokenMu    sync.Mutex
	tokens     map[string]string // cache key -> bearer token
}

// NewClient creates a new registry client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 15 * time.Minute},
		tokens:     make(map[string]string),
	}
}

// Pull downloads an image from an OCI/Docker registry and unpacks it into destDir
func (c *Client) Pull(ref *Reference, destDir string, out io.Writer) (*PullResult, error) {
	tagOrDigest := ref.Tag
	if ref.Digest != "" {
		tagOrDigest = ref.Digest
	}
	if tagOrDigest == "" {
		tagOrDigest = "latest"
	}

	fmt.Fprintf(out, "Pulling from %s/%s:%s\n", ref.Registry, ref.Repository, tagOrDigest)

	// Step 1: Fetch Manifest
	manifestURL := fmt.Sprintf("%s/v2/%s/manifests/%s", ref.BaseURL(), ref.Repository, tagOrDigest)
	req, err := http.NewRequest(http.MethodGet, manifestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create manifest request: %w", err)
	}

	req.Header.Set("Accept", strings.Join([]string{
		MediaTypeDockerV2Manifest,
		MediaTypeOCIV1Manifest,
		MediaTypeDockerV2ManifestList,
		MediaTypeOCIV1Index,
	}, ", "))

	resp, err := c.doWithAuth(req, ref, "pull")
	if err != nil {
		return nil, fmt.Errorf("manifest request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("registry returned %s: %s", resp.Status, string(body))
	}

	contentType := resp.Header.Get("Content-Type")
	manifestBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var targetManifest ImageManifest

	// Check if response is a Manifest List or OCI Image Index
	if strings.Contains(contentType, "manifest.list") || strings.Contains(contentType, "image.index") {
		var index IndexManifest
		if err := json.Unmarshal(manifestBytes, &index); err != nil {
			return nil, fmt.Errorf("failed to parse index manifest: %w", err)
		}

		// Pick matching platform
		selectedDigest := ""
		hostArch := runtime.GOARCH
		for _, m := range index.Manifests {
			if m.Platform != nil {
				if m.Platform.Architecture == hostArch && (m.Platform.OS == "linux" || m.Platform.OS == "") {
					selectedDigest = m.Digest
					break
				}
			}
		}
		// Fallback: if no exact arch match, pick first or amd64/arm64
		if selectedDigest == "" && len(index.Manifests) > 0 {
			for _, m := range index.Manifests {
				if m.Platform != nil && (m.Platform.Architecture == "arm64" || m.Platform.Architecture == "amd64") {
					selectedDigest = m.Digest
					break
				}
			}
			if selectedDigest == "" {
				selectedDigest = index.Manifests[0].Digest
			}
		}

		if selectedDigest == "" {
			return nil, fmt.Errorf("no suitable image platform found in index for %s", hostArch)
		}

		// Fetch the architecture-specific manifest
		subURL := fmt.Sprintf("%s/v2/%s/manifests/%s", ref.BaseURL(), ref.Repository, selectedDigest)
		subReq, err := http.NewRequest(http.MethodGet, subURL, nil)
		if err != nil {
			return nil, err
		}
		subReq.Header.Set("Accept", strings.Join([]string{MediaTypeDockerV2Manifest, MediaTypeOCIV1Manifest}, ", "))

		subResp, err := c.doWithAuth(subReq, ref, "pull")
		if err != nil {
			return nil, fmt.Errorf("sub-manifest request failed: %w", err)
		}
		defer subResp.Body.Close()

		if subResp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("sub-manifest returned %s", subResp.Status)
		}

		if err := json.NewDecoder(subResp.Body).Decode(&targetManifest); err != nil {
			return nil, fmt.Errorf("failed to decode platform manifest: %w", err)
		}
	} else {
		if err := json.Unmarshal(manifestBytes, &targetManifest); err != nil {
			return nil, fmt.Errorf("failed to parse image manifest: %w", err)
		}
	}

	payloadDir := filepath.Join(destDir, "payload")
	if err := os.MkdirAll(payloadDir, 0755); err != nil {
		return nil, err
	}

	// Step 2: Download & Parse Config Blob
	var imageConfig ImageConfig
	if targetManifest.Config.Digest != "" {
		configURL := fmt.Sprintf("%s/v2/%s/blobs/%s", ref.BaseURL(), ref.Repository, targetManifest.Config.Digest)
		configReq, err := http.NewRequest(http.MethodGet, configURL, nil)
		if err == nil {
			configResp, err := c.doWithAuth(configReq, ref, "pull")
			if err == nil && configResp.StatusCode == http.StatusOK {
				_ = json.NewDecoder(configResp.Body).Decode(&imageConfig)
				configResp.Body.Close()
			}
		}
	}

	// Step 3: Download and extract layers sequentially
	totalLayers := len(targetManifest.Layers)
	for idx, layer := range targetManifest.Layers {
		shortDigest := layer.Digest
		if len(shortDigest) > 19 {
			shortDigest = shortDigest[:19]
		}
		fmt.Fprintf(out, "[%d/%d] Pulling layer %s (%.1f MB)...\n", idx+1, totalLayers, shortDigest, float64(layer.Size)/(1024*1024))

		blobURL := fmt.Sprintf("%s/v2/%s/blobs/%s", ref.BaseURL(), ref.Repository, layer.Digest)
		layerReq, err := http.NewRequest(http.MethodGet, blobURL, nil)
		if err != nil {
			return nil, err
		}

		layerResp, err := c.doWithAuth(layerReq, ref, "pull")
		if err != nil {
			return nil, fmt.Errorf("failed to download layer %s: %w", shortDigest, err)
		}

		if layerResp.StatusCode != http.StatusOK {
			layerResp.Body.Close()
			return nil, fmt.Errorf("layer download failed with status: %s", layerResp.Status)
		}

		err = extractLayer(layerResp.Body, payloadDir)
		layerResp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to extract layer %s: %w", shortDigest, err)
		}
		fmt.Fprintf(out, "       Extracted layer %s\n", shortDigest)
	}

	// Build Docroid Manifest and Result
	result := &PullResult{
		Name:         ref.FullName(),
		Tag:          tagOrDigest,
		Architecture: imageConfig.Architecture,
		Engine:       "rootfs",
		Env:          make(map[string]string),
	}
	if result.Architecture == "" {
		result.Architecture = runtime.GOARCH
	}

	// Transfer Environment Variables
	for _, envStr := range imageConfig.Config.Env {
		parts := strings.SplitN(envStr, "=", 2)
		if len(parts) == 2 {
			result.Env[parts[0]] = parts[1]
		}
	}

	// Transfer WorkDir
	if imageConfig.Config.WorkingDir != "" {
		result.WorkDir = imageConfig.Config.WorkingDir
	}

	// Transfer Command & Args
	var fullCmd []string
	if len(imageConfig.Config.Entrypoint) > 0 {
		fullCmd = append(fullCmd, imageConfig.Config.Entrypoint...)
	}
	if len(imageConfig.Config.Cmd) > 0 {
		fullCmd = append(fullCmd, imageConfig.Config.Cmd...)
	}

	if len(fullCmd) > 0 {
		result.Command = fullCmd[0]
		if len(fullCmd) > 1 {
			result.Args = fullCmd[1:]
		}
	} else {
		// Default to /bin/sh if present
		result.Command = "/bin/sh"
	}

	// Transfer Exposed Ports
	for pStr := range imageConfig.Config.ExposedPorts {
		numStr := strings.Split(pStr, "/")[0]
		if p, err := strconv.Atoi(numStr); err == nil && result.Port == 0 {
			result.Port = p
		}
	}

	manifestMap := map[string]interface{}{
		"format_version": "1.0",
		"name":           result.Name,
		"version":        result.Tag,
		"architecture":   result.Architecture,
		"engine":         result.Engine,
		"command":        result.Command,
		"args":           result.Args,
		"port":           result.Port,
		"env":            result.Env,
		"work_dir":       result.WorkDir,
		"description":    fmt.Sprintf("Pulled from %s/%s:%s", ref.Registry, ref.Repository, tagOrDigest),
		"created_at":     time.Now().UTC(),
	}

	manifestPath := filepath.Join(destDir, "manifest.json")
	manifestJSON, err := json.MarshalIndent(manifestMap, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to encode manifest: %w", err)
	}
	if err := os.WriteFile(manifestPath, manifestJSON, 0644); err != nil {
		return nil, fmt.Errorf("failed to save image manifest: %w", err)
	}

	fmt.Fprintf(out, "✓ Successfully pulled %s\n", ref.TaggedName())
	return result, nil
}

// Push uploads a local Docroid image to an OCI/Docker registry
func (c *Client) Push(ref *Reference, imageDir string, out io.Writer) (string, error) {
	manifestPath := filepath.Join(imageDir, "manifest.json")
	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		return "", fmt.Errorf("image manifest not found at %s: %w", manifestPath, err)
	}

	var manifest struct {
		Name         string            `json:"name"`
		Version      string            `json:"version"`
		Architecture string            `json:"architecture"`
		Engine       string            `json:"engine"`
		Command      string            `json:"command"`
		Args         []string          `json:"args"`
		WorkDir      string            `json:"work_dir"`
		Env          map[string]string `json:"env"`
	}
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return "", fmt.Errorf("failed to parse manifest at %s: %w", manifestPath, err)
	}

	payloadDir := filepath.Join(imageDir, "payload")
	if _, err := os.Stat(payloadDir); os.IsNotExist(err) {
		return "", fmt.Errorf("image payload not found at %s", payloadDir)
	}

	tag := ref.Tag
	if tag == "" {
		tag = "latest"
	}

	fmt.Fprintf(out, "The push refers to repository [%s/%s]\n", ref.Registry, ref.Repository)

	// Step 1: Pack payload directory into tar.gz layer
	fmt.Fprintf(out, "Preparing image layer...\n")
	var layerBuf bytes.Buffer
	gz := gzip.NewWriter(&layerBuf)
	tw := tar.NewWriter(gz)

	err = filepath.Walk(payloadDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(payloadDir, path)
		if err != nil || rel == "." {
			return nil
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(rel)

		if info.IsDir() {
			header.Typeflag = tar.TypeDir
			return tw.WriteHeader(header)
		}

		if info.Mode()&os.ModeSymlink != 0 {
			linkTarget, err := os.Readlink(path)
			if err == nil {
				header.Typeflag = tar.TypeSymlink
				header.Linkname = linkTarget
			}
		}

		if info.Mode().IsRegular() {
			header.Typeflag = tar.TypeReg
			header.Size = info.Size()
		}

		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			_, err = io.Copy(tw, f)
			f.Close()
			return err
		}
		return nil
	})
	if err != nil {
		_ = tw.Close()
		_ = gz.Close()
		return "", fmt.Errorf("failed to tar payload: %w", err)
	}
	_ = tw.Close()
	_ = gz.Close()

	layerBytes := layerBuf.Bytes()
	layerDigest := fmt.Sprintf("sha256:%x", sha256.Sum256(layerBytes))
	layerSize := int64(len(layerBytes))

	// Step 2: Upload Layer Blob
	shortDigest := layerDigest
	if len(shortDigest) > 19 {
		shortDigest = shortDigest[:19]
	}

	// Check if layer already exists
	headURL := fmt.Sprintf("%s/v2/%s/blobs/%s", ref.BaseURL(), ref.Repository, layerDigest)
	headReq, _ := http.NewRequest(http.MethodHead, headURL, nil)
	headResp, err := c.doWithAuth(headReq, ref, "push")
	if err == nil && headResp.StatusCode == http.StatusOK {
		headResp.Body.Close()
		fmt.Fprintf(out, "%s: Layer already exists\n", shortDigest)
	} else {
		if headResp != nil {
			headResp.Body.Close()
		}
		fmt.Fprintf(out, "%s: Pushing (%.1f MB)...\n", shortDigest, float64(layerSize)/(1024*1024))
		if err := c.uploadBlob(ref, layerDigest, layerBytes, MediaTypeOCIV1Layer); err != nil {
			return "", fmt.Errorf("failed to upload layer: %w", err)
		}
		fmt.Fprintf(out, "%s: Pushed\n", shortDigest)
	}

	// Step 3: Generate and Upload OCI Image Config Blob
	var envList []string
	for k, v := range manifest.Env {
		envList = append(envList, fmt.Sprintf("%s=%s", k, v))
	}

	cfg := ImageConfig{
		Architecture: manifest.Architecture,
		OS:           "linux",
		Config: ContainerConfig{
			Env:        envList,
			Entrypoint: []string{manifest.Command},
			Cmd:        manifest.Args,
			WorkingDir: manifest.WorkDir,
		},
	}
	if cfg.Architecture == "" {
		cfg.Architecture = runtime.GOARCH
	}

	configBytes, _ := json.MarshalIndent(cfg, "", "  ")
	configDigest := fmt.Sprintf("sha256:%x", sha256.Sum256(configBytes))
	configSize := int64(len(configBytes))

	if err := c.uploadBlob(ref, configDigest, configBytes, MediaTypeOCIV1Config); err != nil {
		return "", fmt.Errorf("failed to upload config blob: %w", err)
	}

	// Step 4: Upload OCI Image Manifest
	ociManifest := ImageManifest{
		SchemaVersion: 2,
		MediaType:     MediaTypeOCIV1Manifest,
		Config: Descriptor{
			MediaType: MediaTypeOCIV1Config,
			Size:      configSize,
			Digest:    configDigest,
		},
		Layers: []Descriptor{
			{
				MediaType: MediaTypeOCIV1Layer,
				Size:      layerSize,
				Digest:    layerDigest,
			},
		},
	}

	manifestBytes, _ = json.MarshalIndent(ociManifest, "", "  ")
	manifestDigest := fmt.Sprintf("sha256:%x", sha256.Sum256(manifestBytes))

	putManifestURL := fmt.Sprintf("%s/v2/%s/manifests/%s", ref.BaseURL(), ref.Repository, tag)
	putReq, err := http.NewRequest(http.MethodPut, putManifestURL, bytes.NewReader(manifestBytes))
	if err != nil {
		return "", err
	}
	putReq.Header.Set("Content-Type", MediaTypeOCIV1Manifest)

	putResp, err := c.doWithAuth(putReq, ref, "push")
	if err != nil {
		return "", fmt.Errorf("failed to put manifest: %w", err)
	}
	defer putResp.Body.Close()

	if putResp.StatusCode != http.StatusOK && putResp.StatusCode != http.StatusCreated && putResp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(putResp.Body)
		return "", fmt.Errorf("registry rejected manifest (%s): %s", putResp.Status, string(body))
	}

	fmt.Fprintf(out, "%s: digest: %s size: %d\n", tag, manifestDigest, len(manifestBytes))
	fmt.Fprintf(out, "✓ Successfully pushed %s\n", ref.TaggedName())
	return manifestDigest, nil
}

// uploadBlob executes the OCI 2-step blob upload protocol
func (c *Client) uploadBlob(ref *Reference, digest string, data []byte, mediaType string) error {
	initURL := fmt.Sprintf("%s/v2/%s/blobs/uploads/", ref.BaseURL(), ref.Repository)
	postReq, err := http.NewRequest(http.MethodPost, initURL, nil)
	if err != nil {
		return err
	}

	postResp, err := c.doWithAuth(postReq, ref, "push")
	if err != nil {
		return fmt.Errorf("failed to initiate blob upload: %w", err)
	}
	defer postResp.Body.Close()

	if postResp.StatusCode != http.StatusAccepted && postResp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(postResp.Body)
		return fmt.Errorf("initiate blob upload returned %s: %s", postResp.Status, string(b))
	}

	location := postResp.Header.Get("Location")
	if location == "" {
		return fmt.Errorf("missing Location header in blob upload initiation")
	}

	// Resolve location if relative
	if !strings.HasPrefix(location, "http://") && !strings.HasPrefix(location, "https://") {
		baseURL, _ := url.Parse(ref.BaseURL())
		locURL, err := baseURL.Parse(location)
		if err == nil {
			location = locURL.String()
		}
	}

	// Append digest parameter
	sep := "?"
	if strings.Contains(location, "?") {
		sep = "&"
	}
	uploadURL := fmt.Sprintf("%s%sdigest=%s", location, sep, digest)

	putReq, err := http.NewRequest(http.MethodPut, uploadURL, bytes.NewReader(data))
	if err != nil {
		return err
	}
	putReq.Header.Set("Content-Type", mediaType)
	putReq.Header.Set("Content-Length", strconv.Itoa(len(data)))

	putResp, err := c.doWithAuth(putReq, ref, "push")
	if err != nil {
		return fmt.Errorf("failed to complete blob upload: %w", err)
	}
	defer putResp.Body.Close()

	if putResp.StatusCode != http.StatusCreated && putResp.StatusCode != http.StatusOK && putResp.StatusCode != http.StatusNoContent {
		b, _ := io.ReadAll(putResp.Body)
		return fmt.Errorf("blob upload failed (%s): %s", putResp.Status, string(b))
	}

	return nil
}

// doWithAuth executes an HTTP request, automatically handling 401 Bearer challenges
func (c *Client) doWithAuth(req *http.Request, ref *Reference, scopeAction string) (*http.Response, error) {
	tokenKey := fmt.Sprintf("%s:%s:%s", ref.Registry, ref.Repository, scopeAction)

	c.tokenMu.Lock()
	cachedToken := c.tokens[tokenKey]
	c.tokenMu.Unlock()

	if cachedToken != "" {
		req.Header.Set("Authorization", "Bearer "+cachedToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// If 401 Unauthorized, parse Www-Authenticate header and obtain token
	if resp.StatusCode == http.StatusUnauthorized {
		authHeader := resp.Header.Get("Www-Authenticate")
		resp.Body.Close()

		if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			token, err := c.fetchBearerToken(authHeader, ref, scopeAction)
			if err != nil {
				return nil, fmt.Errorf("authentication error: %w", err)
			}

			c.tokenMu.Lock()
			c.tokens[tokenKey] = token
			c.tokenMu.Unlock()

			// Clone request with auth header
			clonedReq := req.Clone(req.Context())
			clonedReq.Header.Set("Authorization", "Bearer "+token)
			if req.GetBody != nil {
				body, err := req.GetBody()
				if err == nil {
					clonedReq.Body = body
				}
			}

			return c.httpClient.Do(clonedReq)
		}
	}

	return resp, nil
}

// fetchBearerToken calls the auth service to retrieve a bearer token
func (c *Client) fetchBearerToken(authHeader string, ref *Reference, scopeAction string) (string, error) {
	params := parseAuthHeader(authHeader)
	realm := params["realm"]
	service := params["service"]
	scope := params["scope"]

	if realm == "" {
		return "", fmt.Errorf("auth challenge missing realm")
	}

	if scope == "" {
		scope = fmt.Sprintf("repository:%s:%s", ref.Repository, scopeAction)
	}

	authURL, err := url.Parse(realm)
	if err != nil {
		return "", err
	}

	q := authURL.Query()
	if service != "" {
		q.Set("service", service)
	}
	q.Set("scope", scope)
	authURL.RawQuery = q.Encode()

	authResp, err := c.httpClient.Get(authURL.String())
	if err != nil {
		return "", fmt.Errorf("token request failed: %w", err)
	}
	defer authResp.Body.Close()

	if authResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token endpoint returned status %s", authResp.Status)
	}

	var tokenPayload struct {
		Token       string `json:"token"`
		AccessToken string `json:"access_token"`
	}

	if err := json.NewDecoder(authResp.Body).Decode(&tokenPayload); err != nil {
		return "", err
	}

	if tokenPayload.Token != "" {
		return tokenPayload.Token, nil
	}
	if tokenPayload.AccessToken != "" {
		return tokenPayload.AccessToken, nil
	}

	return "", fmt.Errorf("empty token received from auth service")
}

func parseAuthHeader(header string) map[string]string {
	res := make(map[string]string)
	trimmed := strings.TrimPrefix(header, "Bearer ")
	trimmed = strings.TrimPrefix(trimmed, "bearer ")

	pairs := strings.Split(trimmed, ",")
	for _, p := range pairs {
		kv := strings.SplitN(strings.TrimSpace(p), "=", 2)
		if len(kv) == 2 {
			k := strings.ToLower(strings.TrimSpace(kv[0]))
			v := strings.Trim(strings.TrimSpace(kv[1]), "\"")
			res[k] = v
		}
	}
	return res
}

// extractLayer unpacks a tar or gzip stream into destDir
func extractLayer(r io.Reader, destDir string) error {
	// Wrap in gzip if gzipped
	br := r
	gz, err := gzip.NewReader(r)
	if err == nil {
		br = gz
		defer gz.Close()
	}

	tr := tar.NewReader(br)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		cleanPath := filepath.Clean(header.Name)
		// Security: prevent Zip Slip
		target := filepath.Join(destDir, cleanPath)
		if !strings.HasPrefix(target, filepath.Clean(destDir)+string(os.PathSeparator)) && target != destDir {
			continue
		}

		// Handle OCI/Docker whiteout deletions (.wh.<file>)
		baseName := filepath.Base(cleanPath)
		if strings.HasPrefix(baseName, ".wh.") {
			realName := strings.TrimPrefix(baseName, ".wh.")
			targetToDelete := filepath.Join(filepath.Dir(target), realName)
			_ = os.RemoveAll(targetToDelete)
			continue
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, header.FileInfo().Mode().Perm())
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		case tar.TypeSymlink:
			_ = os.Remove(target)
			_ = os.MkdirAll(filepath.Dir(target), 0755)
			_ = os.Symlink(header.Linkname, target)
		}
	}
	return nil
}
