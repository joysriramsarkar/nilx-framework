// Package nilos implements the NilX platform adapter for NilOS.
// This adapter connects NilX apps to NilOS system services:
//   - nilui / nilui-gpu (Vulkan renderer)
//   - nilbus-client (distributed IPC)
//   - nilhal (hardware abstraction)
//   - Wayland compositor
package nilos

import (
	"fmt"
	"os"
	"path/filepath"
)

// Adapter is the NilOS platform adapter.
type Adapter struct {
	KernelVersion string
	DisplayServer string // "wayland"
	GPUBackend    string // "vulkan"
	DisplayHandle uintptr
	WindowHandle  uintptr
	initialized   bool
}

// Platform interface — every platform adapter implements this.
type Platform interface {
	Init() error
	Shutdown()
	CreateWindow(title string, width, height int) error
	ShowWindow()
	PollEvents() []Event
	SwapBuffers()
	// Hardware access
	GetKernelVersion() string
	TriggerSensor(sensorID int) error
	ReadSensorData(sensorID int) ([]float64, error)
	// File system
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte) error
	// Notifications
	SendNotification(title, body string) error
}

// Event represents a platform input event.
type Event struct {
	Type    EventType
	X, Y    float32
	Key     string
	Char    rune
	ScrollX float32
	ScrollY float32
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

// New creates a NilOS platform adapter.
func New() *Adapter {
	return &Adapter{
		KernelVersion: "NilOS-0.1",
		DisplayServer: "wayland",
		GPUBackend:    "vulkan",
	}
}

// Init initializes the NilOS platform.
// In production: connects to Wayland compositor, initializes Vulkan via nilui-gpu.
func (a *Adapter) Init() error {
	fmt.Println("[NilOS] Initializing NilX platform adapter...")
	fmt.Printf("[NilOS] Kernel: %s | Display: %s | GPU: %s\n",
		a.KernelVersion, a.DisplayServer, a.GPUBackend)
	a.initialized = true
	return nil
}

func (a *Adapter) Shutdown() {
	if a.initialized {
		fmt.Println("[NilOS] Shutting down NilX platform adapter...")
		a.initialized = false
	}
}

func (a *Adapter) CreateWindow(title string, width, height int) error {
	fmt.Printf("[NilOS] Creating Wayland window: %q %dx%d\n", title, width, height)
	// In production: calls nilshell Wayland compositor API
	// xdg_surface, xdg_toplevel, wl_surface
	return nil
}

func (a *Adapter) ShowWindow() {
	fmt.Println("[NilOS] Showing window via Wayland...")
}

func (a *Adapter) PollEvents() []Event {
	// In production: reads Wayland event queue
	return nil
}

func (a *Adapter) SwapBuffers() {
	// In production: calls Vulkan present (vkQueuePresentKHR)
}

func (a *Adapter) GetKernelVersion() string {
	return a.KernelVersion
}

func (a *Adapter) TriggerSensor(sensorID int) error {
	fmt.Printf("[NilOS] Triggering hardware sensor %d via nilhal...\n", sensorID)
	// In production: calls NilOS HAL sensor API
	return nil
}

func (a *Adapter) ReadSensorData(sensorID int) ([]float64, error) {
	// Stub
	return []float64{0.0, 9.81, 0.0}, nil
}

func (a *Adapter) ReadFile(path string) ([]byte, error) {
	fmt.Printf("[NilOS] Reading file via nilfs: %s\n", path)
	return nil, nil
}

func (a *Adapter) WriteFile(path string, data []byte) error {
	fmt.Printf("[NilOS] Writing file via nilfs: %s (%d bytes)\n", path, len(data))
	return nil
}

func (a *Adapter) SendNotification(title, body string) error {
	fmt.Printf("[NilOS] Notification: %s — %s\n", title, body)
	return nil
}

// NilOSAPI provides direct NilOS kernel API access (unique advantage of NilOS).
// No bridge overhead — direct kernel calls.
type NilOSAPI struct {
	adapter *Adapter
}

func NewAPI(a *Adapter) *NilOSAPI {
	return &NilOSAPI{adapter: a}
}

func (n *NilOSAPI) GetKernelVersion() string {
	return n.adapter.GetKernelVersion()
}

func (n *NilOSAPI) AccessMemoryRegion(addr uintptr) ([]byte, error) {
	fmt.Printf("[NilOS] Direct memory access at 0x%X\n", addr)
	return make([]byte, 16), nil
}

func (n *NilOSAPI) TriggerCustomHardwareSensor(id int) error {
	return n.adapter.TriggerSensor(id)
}

func (n *NilOSAPI) DistributedBusCall(service string, method string, args []byte) ([]byte, error) {
	fmt.Printf("[NilOS] nilbus-client: %s.%s(%d bytes)\n", service, method, len(args))
	// In production: uses nilbus-client IPC
	return nil, nil
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
Permissions = storage.read, notifications, network
DisplayServer = wayland
GPUBackend = vulkan
KernelMinVersion = 0.1.0
`
	if err := os.WriteFile(filepath.Join(nilosDir, "app.nilxmanifest"), []byte(manifest), 0644); err != nil {
		return fmt.Errorf("failed writing manifest: %w", err)
	}

	if len(bytecode) > 0 {
		if err := os.WriteFile(filepath.Join(binDir, "main.nabc"), bytecode, 0644); err != nil {
			return fmt.Errorf("failed writing bytecode: %w", err)
		}
	}

	return nil
}
