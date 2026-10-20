package compose

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"docroid/pkg/app"
	"docroid/pkg/bundle"
	"docroid/pkg/engine"
	"docroid/pkg/netproxy"
	"docroid/pkg/network"
	"docroid/pkg/secrets"
	"docroid/pkg/supervisor"
	"docroid/pkg/utils"

	"gopkg.in/yaml.v3"
)

// ParseYAML decodes YAML content into a ComposeConfig
func ParseYAML(data []byte) (*ComposeConfig, error) {
	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	var cfg ComposeConfig
	// Check if "services" exists
	if _, hasServices := raw["services"]; hasServices {
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("invalid compose structure: %w", err)
		}
	} else {
		// Single service format written at root
		var single ServiceSpec
		if err := yaml.Unmarshal(data, &single); err != nil {
			return nil, fmt.Errorf("invalid single-service compose: %w", err)
		}
		serviceName := "app"
		if single.Name != "" {
			serviceName = single.Name
		}
		cfg.Version = "1.0"
		cfg.Services = map[string]ServiceSpec{
			serviceName: single,
		}
	}

	if len(cfg.Services) == 0 {
		return nil, fmt.Errorf("no services defined in compose file")
	}

	return &cfg, nil
}

// NormalizeService converts raw YAML fields into strongly typed, normalized properties
func NormalizeService(serviceKey string, s *ServiceSpec) (*ResolvedService, error) {
	res := &ResolvedService{
		ServiceName: serviceKey,
		Name:        serviceKey,
		Engine:      string(app.EngineNative),
		Environment: make(map[string]string),
		Volumes:     make([]string, 0),
		RestartPolicy: "no",
		Detached:    true,
	}

	if s.Name != "" {
		res.Name = s.Name
	}
	res.Image = s.Image
	res.Rootfs = s.Rootfs
	res.WorkingDir = s.WorkingDir
	res.Secrets = s.Secrets

	if s.Detached != nil {
		res.Detached = *s.Detached
	}

	// Network extraction
	if s.Networks != nil {
		switch netVal := s.Networks.(type) {
		case []interface{}:
			for _, n := range netVal {
				if ns, ok := n.(string); ok && ns != "" {
					res.Network = ns
					break
				}
			}
		case []string:
			if len(netVal) > 0 {
				res.Network = netVal[0]
			}
		case string:
			res.Network = netVal
		}
	}
	res.NetworkAliases = append(res.NetworkAliases, serviceKey)
	if s.Name != "" && s.Name != serviceKey {
		res.NetworkAliases = append(res.NetworkAliases, s.Name)
	}

	// Engine determination
	if s.Engine != "" {
		res.Engine = strings.ToLower(s.Engine)
	} else if s.Rootfs != "" {
		res.Engine = string(app.EngineRootfs)
	}

	// Command & Args parsing
	switch v := s.Command.(type) {
	case string:
		v = strings.TrimSpace(v)
		if v != "" {
			parts := strings.Fields(v)
			if len(parts) > 0 {
				res.Command = parts[0]
				if len(parts) > 1 {
					res.Args = append(res.Args, parts[1:]...)
				}
			}
		}
	case []interface{}:
		for i, item := range v {
			str := fmt.Sprintf("%v", item)
			if i == 0 {
				res.Command = str
			} else {
				res.Args = append(res.Args, str)
			}
		}
	case []string:
		if len(v) > 0 {
			res.Command = v[0]
			if len(v) > 1 {
				res.Args = append(res.Args, v[1:]...)
			}
		}
	}

	if len(s.Args) > 0 {
		res.Args = append(res.Args, s.Args...)
	}

	// Ports parsing
	if len(s.Ports) > 0 {
		res.PortMapping = strings.TrimSpace(s.Ports[0])
		if pm, err := netproxy.ParsePortMapping(res.PortMapping); err == nil {
			res.Port = pm.ContainerPort
		}
	}

	// Environment parsing
	switch v := s.Environment.(type) {
	case map[string]interface{}:
		for k, val := range v {
			res.Environment[k] = fmt.Sprintf("%v", val)
		}
	case map[string]string:
		for k, val := range v {
			res.Environment[k] = val
		}
	case []interface{}:
		for _, item := range v {
			line := fmt.Sprintf("%v", item)
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				res.Environment[parts[0]] = parts[1]
			} else if len(parts) == 1 && parts[0] != "" {
				res.Environment[parts[0]] = ""
			}
		}
	case []string:
		for _, line := range v {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				res.Environment[parts[0]] = parts[1]
			}
		}
	}

	// Volumes
	res.Volumes = append(res.Volumes, s.Volumes...)

	// Restart policy
	if s.Restart != "" {
		pol := strings.ToLower(strings.TrimSpace(s.Restart))
		if pol == "unless-stopped" {
			pol = "always"
		}
		res.RestartPolicy = pol
	}

	return res, nil
}

// ToYAML serializes a ComposeConfig into formatted YAML
func ToYAML(cfg *ComposeConfig) ([]byte, error) {
	if cfg.Version == "" {
		cfg.Version = "1.0"
	}
	return yaml.Marshal(cfg)
}

// GenerateComposeYAML builds a ComposeConfig from single application parameters and outputs YAML bytes
func GenerateComposeYAML(
	name, engineType, rootfs, command string,
	args []string,
	portMapping, restartPolicy, workDir string,
	env map[string]string,
	volumes []string,
	secretsList []string,
	detached bool,
) ([]byte, error) {
	serviceName := name
	if serviceName == "" {
		serviceName = "app"
	}

	var ports []string
	if portMapping != "" {
		ports = []string{portMapping}
	}

	cfg := &ComposeConfig{
		Version: "1.0",
		Services: map[string]ServiceSpec{
			serviceName: {
				Name:        name,
				Engine:      engineType,
				Rootfs:      rootfs,
				Command:     command,
				Args:        args,
				Ports:       ports,
				Environment: env,
				Volumes:     volumes,
				Restart:     restartPolicy,
				WorkingDir:  workDir,
				Secrets:     secretsList,
				Detached:    &detached,
			},
		},
	}

	return ToYAML(cfg)
}

// ServiceToApp converts a resolved compose service into a Docroid App model
func ServiceToApp(resolved *ResolvedService, store *app.Store) (*app.App, error) {
	volMounts := make([]app.VolumeMount, 0)
	for _, v := range resolved.Volumes {
		parts := strings.Split(v, ":")
		if len(parts) >= 2 {
			readOnly := false
			if len(parts) >= 3 && parts[2] == "ro" {
				readOnly = true
			}
			volMounts = append(volMounts, app.VolumeMount{
				HostPath:      parts[0],
				ContainerPath: parts[1],
				ReadOnly:      readOnly,
			})
		}
	}

	application := &app.App{
		ID:            app.GenerateID(),
		Name:          resolved.Name,
		Engine:        app.EngineType(resolved.Engine),
		Command:       resolved.Command,
		Args:          resolved.Args,
		Env:           resolved.Environment,
		WorkDir:       resolved.WorkingDir,
		Port:          resolved.Port,
		PortMapping:   resolved.PortMapping,
		Volumes:       volMounts,
		Rootfs:        resolved.Rootfs,
		RestartPolicy: resolved.RestartPolicy,
		Network:       resolved.Network,
		NetworkAliases: resolved.NetworkAliases,
		Status:        app.StatusCreated,
	}

	return application, nil
}

// Up deploys and launches all services defined in a ComposeConfig
func Up(ctx context.Context, store *app.Store, manager *app.AppManager, cfg *ComposeConfig, detach bool) ([]*app.App, error) {
	var launched []*app.App

	// Initialize / ensure project network (Docker Compose standard behavior)
	projectName := cfg.Name
	if projectName == "" {
		projectName = "default"
	}
	defaultNetName := projectName + "_default"

	if netMgr, err := network.NewNetworkManager(store.BaseDir()); err == nil && netMgr != nil {
		if len(cfg.Networks) > 0 {
			for netName, nSpec := range cfg.Networks {
				if existing, _ := netMgr.GetNetwork(netName); existing == nil {
					driver := network.DriverBridge
					if nSpec.Driver != "" {
						driver = network.DriverType(nSpec.Driver)
					}
					_, _ = netMgr.CreateNetwork(netName, driver, "", nSpec.Labels)
				}
			}
		} else {
			if existing, _ := netMgr.GetNetwork(defaultNetName); existing == nil {
				_, _ = netMgr.CreateNetwork(defaultNetName, network.DriverBridge, "", map[string]string{"compose": projectName})
			}
		}
	}

	for serviceKey, s := range cfg.Services {
		resolved, err := NormalizeService(serviceKey, &s)
		if err != nil {
			return launched, fmt.Errorf("service %s: %w", serviceKey, err)
		}
		if resolved.Network == "" {
			resolved.Network = defaultNetName
		}

		// Stop existing instance with same name if running
		if existing, _ := store.FindApp(resolved.Name); existing != nil {
			_ = manager.StopApp(existing.ID)
			_ = manager.RemoveApp(existing.ID, true)
		}

		// Check if target is an installed image
		if resolved.Image != "" {
			if img, _ := bundle.GetImage(store, resolved.Image); img != nil {
				deployed, err := bundle.DeployFromImage(store, manager, resolved.Image, resolved.Name, detach)
				if err != nil {
					return launched, fmt.Errorf("failed to deploy service %s from image %s: %w", serviceKey, resolved.Image, err)
				}
				if resolved.PortMapping != "" {
					deployed.PortMapping = resolved.PortMapping
					if pm, err := netproxy.ParsePortMapping(resolved.PortMapping); err == nil {
						deployed.Port = pm.ContainerPort
					}
					_ = store.SaveApp(deployed)
				}
				launched = append(launched, deployed)
				continue
			}
		}

		// Check port mapping
		var pMap *netproxy.PortMapping
		if resolved.PortMapping != "" {
			if pm, err := netproxy.ParsePortMapping(resolved.PortMapping); err == nil {
				pMap = pm
				if pMap.HostPort == pMap.ContainerPort {
					if freePort, err := utils.GetFreePort(); err == nil {
						oldPortStr := strconv.Itoa(pMap.HostPort)
						pMap.ContainerPort = freePort
						newPortStr := strconv.Itoa(freePort)
						for i, a := range resolved.Args {
							if a == oldPortStr {
								resolved.Args[i] = newPortStr
							}
						}
					}
				}
				resolved.Port = pMap.ContainerPort
				resolved.PortMapping = pMap.String()
			}
		}

		// Inject secrets if specified
		if len(resolved.Secrets) > 0 {
			if secStore, err := secrets.NewStore(store.BaseDir()); err == nil {
				for _, secKey := range resolved.Secrets {
					parts := strings.SplitN(secKey, "=", 2)
					envKey := parts[0]
					storeKey := parts[0]
					if len(parts) == 2 {
						envKey = parts[0]
						storeKey = parts[1]
					}
					if val, err := secStore.Get(storeKey); err == nil {
						resolved.Environment[envKey] = val
					}
				}
			}
		}

		application, err := ServiceToApp(resolved, store)
		if err != nil {
			return launched, fmt.Errorf("service %s: %w", serviceKey, err)
		}

		if err := store.SaveApp(application); err != nil {
			return launched, fmt.Errorf("failed to save app %s: %w", resolved.Name, err)
		}

		// Choose engine
		var eng engine.Engine
		if application.Engine == app.EngineRootfs {
			eng = engine.NewRootfsEngine()
		} else {
			eng = engine.NewNativeEngine()
		}

		runCfg := &engine.RunConfig{
			App:      application,
			Store:    store,
			Detached: detach,
		}

		proc, err := eng.Run(ctx, runCfg)
		if err != nil {
			application.Status = app.StatusFailed
			_ = store.SaveApp(application)
			return launched, fmt.Errorf("failed to start service %s: %w", serviceKey, err)
		}

		application.PID = proc.Pid
		application.Status = app.StatusRunning
		_ = store.SaveApp(application)

		// Start background supervisor process to hold proxy and monitor restarts if detached
		if detach && (application.RestartPolicy != "no" || (pMap != nil && pMap.HostPort != pMap.ContainerPort)) {
			selfPath, err := os.Executable()
			if err != nil {
				selfPath = os.Args[0]
			}
			supCmd := exec.Command(selfPath, "_supervise", application.ID)
			supCmd.Stdin = nil
			supCmd.Stdout = nil
			supCmd.Stderr = nil
			if err := supCmd.Start(); err == nil {
				_ = os.WriteFile(filepath.Join(store.AppDir(application.ID), "supervisor.pid"), []byte(strconv.Itoa(supCmd.Process.Pid)), 0644)
				go func() {
					_ = supCmd.Process.Release()
				}()
			}
		} else {
			// In-process supervision and proxy
			appDir := store.AppDir(application.ID)
			if pMap != nil && pMap.HostPort != pMap.ContainerPort {
				_ = netproxy.GetProxyManager().StartProxyWithControl(application.ID, pMap, appDir)
			}
			if application.RestartPolicy != "" && application.RestartPolicy != "no" {
				supervisor.GetSupervisor().Watch(application, store, eng)
			}
		}

		launched = append(launched, application)
	}

	return launched, nil
}

// Down stops and removes all services defined in a ComposeConfig
func Down(ctx context.Context, store *app.Store, manager *app.AppManager, cfg *ComposeConfig) error {
	for serviceKey, s := range cfg.Services {
		name := serviceKey
		if s.Name != "" {
			name = s.Name
		}

		if target, _ := store.FindApp(name); target != nil {
			_ = manager.StopApp(target.ID)
			_ = manager.RemoveApp(target.ID, true)
		}
	}
	return nil
}

// Status returns runtime states of all services in a ComposeConfig
func Status(store *app.Store, cfg *ComposeConfig) ([]*app.App, error) {
	var list []*app.App
	for serviceKey, s := range cfg.Services {
		name := serviceKey
		if s.Name != "" {
			name = s.Name
		}
		if target, _ := store.FindApp(name); target != nil {
			list = append(list, target)
		}
	}
	return list, nil
}

var _ = os.Getenv
var _ = filepath.Base
