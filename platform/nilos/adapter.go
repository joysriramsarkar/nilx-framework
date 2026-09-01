// Package nilos implements the native platform adapter for NilOS.
// Connects NilX apps directly to NilOS services:
//   - nilui / nilui-gpu (Vulkan renderer & Wayland surface manager)
//   - nilbus-client (distributed IPC)
//   - nilhal (hardware & sensor abstraction)
//   - nilpkg / nilrt (signed sandbox & app lifecycle supervisor)
package nilos

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/joysriramsarkar/nilx-framework/compiler/codegen"
	"github.com/joysriramsarkar/nilx-framework/platform/nilos/nilbus"
	"github.com/joysriramsarkar/nilx-framework/platform/nilos/nilhal"
	"github.com/joysriramsarkar/nilx-framework/platform/nilos/nilui"
	"github.com/joysriramsarkar/nilx-framework/runtime/vm"
)

// Adapter is the full NilOS native platform adapter.
type Adapter struct {
	mu            sync.RWMutex
	KernelVersion string
	DisplayServer string // "wayland"
	GPUBackend    string // "vulkan"
	BusClient     *nilbus.Client
	Renderer      *nilui.Renderer
	HAL           *nilhal.HAL
	initialized   bool
	activeVM      *vm.VM
	capabilities  map[string]bool
}

// Event represents a platform input event.
type Event struct {
	Type    EventType `json:"type"`
	X       float64   `json:"x"`
	Y       float64   `json:"y"`
	Key     string    `json:"key,omitempty"`
	Char    rune      `json:"char,omitempty"`
	ScrollX float64   `json:"scrollX,omitempty"`
	ScrollY float64   `json:"scrollY,omitempty"`
}

type EventType int

const (
	EventTouchDown EventType = iota
	EventTouchMove
	EventTouchUp
	EventKeyDown
	EventKeyUp
	EventCharInput
	EventScroll
	EventResize
	EventQuit
	EventFocus
	EventBlur
	EventBackButton
)

// New creates a new NilOS platform adapter with full system service clients.
func New() *Adapter {
	return &Adapter{
		KernelVersion: "NilOS-0.1-vulkan",
		DisplayServer: "wayland",
		GPUBackend:    "vulkan",
		BusClient:     nilbus.NewClient(""),
		Renderer:      nilui.NewRenderer("NilXApp", 1080, 1920),
		HAL:           nilhal.NewHAL(),
		capabilities:  make(map[string]bool),
	}
}

// Init connects to the NilOS system services (NilBus, NilUI, NilHAL).
func (a *Adapter) Init() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.initialized {
		return nil
	}

	// 1. Connect to NilBus IPC
	if err := a.BusClient.Connect(); err != nil {
		return fmt.Errorf("failed connecting to nilbus: %w", err)
	}

	// 2. Grant default runtime capabilities
	a.capabilities["storage.read"] = true
	a.capabilities["notifications"] = true
	a.capabilities["sensors"] = true
	a.capabilities["network"] = true

	a.initialized = true
	return nil
}

// Shutdown disconnects from all NilOS system services.
func (a *Adapter) Shutdown() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.initialized {
		if a.BusClient != nil {
			a.BusClient.Close()
		}
		a.initialized = false
	}
}

// GetKernelVersion returns the active NilOS kernel version string.
func (a *Adapter) GetKernelVersion() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.KernelVersion
}

// CreateWindow creates a Wayland xdg_surface with Vulkan swapchain.
func (a *Adapter) CreateWindow(title string, width, height int) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Renderer = nilui.NewRenderer(title, width, height)
	return nil
}

// CheckCapability verifies if the app has been granted a specific permission.
func (a *Adapter) CheckCapability(capName string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.capabilities[capName]
}

// SendNotification dispatches a system notification via NilBus.
func (a *Adapter) SendNotification(title, body string) error {
	if !a.CheckCapability("notifications") {
		return fmt.Errorf("permission denied: notifications")
	}

	payload := []byte(fmt.Sprintf(`{"title":%q,"body":%q}`, title, body))
	_, err := a.BusClient.Call("org.nilos.NotificationService", "Notify", payload)
	return err
}

// ReadSensorData reads hardware telemetry via NilHAL.
func (a *Adapter) ReadSensorData(sensorID int) ([]float64, error) {
	if !a.CheckCapability("sensors") {
		return nil, fmt.Errorf("permission denied: sensors")
	}
	return a.HAL.ReadSensor(nilhal.SensorKind(sensorID))
}

// DispatchTouchEvent routes Wayland touch events to the active UI engine.
func (a *Adapter) DispatchTouchEvent(x, y float64) bool {
	a.mu.RLock()
	activeVM := a.activeVM
	a.mu.RUnlock()

	if activeVM == nil {
		return false
	}

	tree := activeVM.GetUITree()
	if tree == nil {
		return false
	}

	return tree.DispatchTouch(x, y)
}

// RenderCurrentFrame computes UI layout and exports Vulkan render commands.
func (a *Adapter) RenderCurrentFrame() (string, error) {
	a.mu.RLock()
	activeVM := a.activeVM
	renderer := a.Renderer
	a.mu.RUnlock()

	if activeVM == nil || renderer == nil {
		return "{}", nil
	}

	tree := activeVM.GetUITree()
	if tree == nil || tree.Root == nil {
		return "{}", nil
	}

	activeVM.ComputeUILayout(float64(renderer.Width), float64(renderer.Height))
	packet := renderer.RenderTree(tree.Root)
	return renderer.ExportFrameJSON(packet)
}

// RunApp launches and supervises a native .nilapp package on NilOS.
func (a *Adapter) RunApp(bundlePath string) error {
	if err := a.Init(); err != nil {
		return err
	}

	nabcPath := filepath.Join(bundlePath, "bin", "main.nabc")
	if _, err := os.Stat(nabcPath); os.IsNotExist(err) {
		nabcPath = filepath.Join(bundlePath, "main.nabc")
	}

	data, err := os.ReadFile(nabcPath)
	if err != nil {
		return fmt.Errorf("failed reading .nilapp bytecode: %w", err)
	}

	mod, err := codegen.Deserialize(data)
	if err != nil {
		return fmt.Errorf("failed deserializing bytecode: %w", err)
	}

	runner := vm.New(mod)
	a.mu.Lock()
	a.activeVM = runner
	a.mu.Unlock()

	return runner.Run()
}

// GenerateProject builds a complete native NilOS .nilapp bundle structure.
func (a *Adapter) GenerateProject(outputDir string, bytecode []byte) error {
	nilosDir := filepath.Join(outputDir, "nilos")
	binDir := filepath.Join(nilosDir, "bin")
	resDir := filepath.Join(nilosDir, "res")

	dirs := []string{nilosDir, binDir, resDir}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("failed creating directory %s: %w", d, err)
		}
	}

	manifest := `[NilApp]
Name = NilXApp
Version = 0.1.0
Platform = nilos
Entry = bin/main.nabc
Permissions = storage.read, notifications, sensors, network
DisplayServer = wayland
GPUBackend = vulkan
KernelMinVersion = 0.1.0
`
	launcherSh := `#!/bin/sh
# nilos-launcher.sh — Native NilOS Wayland/Vulkan app launcher
DIR="$(cd "$(dirname "$0")" && pwd)"
export NILOS_DISPLAY="wayland-0"
export NILOS_GPU="vulkan"
exec nilc -in "$DIR/bin/main.nabc" -run "$@"
`
	serviceUnit := `[Unit]
Description=NilX Native Application on NilOS
After=nilbus.service nilui.service

[Service]
Type=simple
ExecStart=/usr/bin/nilc -in /app/bin/main.nabc -run
Restart=on-failure
Environment="NILBUS_SOCKET=/run/nilbus/system.sock"

[Install]
WantedBy=graphical-session.target
`

	if err := os.WriteFile(filepath.Join(nilosDir, "app.nilxmanifest"), []byte(manifest), 0644); err != nil {
		return fmt.Errorf("failed writing manifest: %w", err)
	}
	if err := os.WriteFile(filepath.Join(nilosDir, "nilos-launcher.sh"), []byte(launcherSh), 0755); err != nil {
		return fmt.Errorf("failed writing launcher: %w", err)
	}
	if err := os.WriteFile(filepath.Join(nilosDir, "nilx-app.service"), []byte(serviceUnit), 0644); err != nil {
		return fmt.Errorf("failed writing service unit: %w", err)
	}

	if len(bytecode) > 0 {
		if err := os.WriteFile(filepath.Join(binDir, "main.nabc"), bytecode, 0644); err != nil {
			return fmt.Errorf("failed writing bytecode: %w", err)
		}
	}

	return nil
}
