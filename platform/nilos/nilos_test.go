package nilos

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/joysriramsarkar/nilx-framework/platform/nilos/nilbus"
	"github.com/joysriramsarkar/nilx-framework/platform/nilos/nilhal"
	"github.com/joysriramsarkar/nilx-framework/platform/nilos/nilui"
	"github.com/joysriramsarkar/nilx-framework/ui/engine"
)

func TestNilOSAdapterLifecycle(t *testing.T) {
	adapter := New()
	if err := adapter.Init(); err != nil {
		t.Fatalf("adapter.Init failed: %v", err)
	}
	defer adapter.Shutdown()

	if !adapter.CheckCapability("notifications") {
		t.Errorf("expected notifications capability to be granted")
	}

	if err := adapter.SendNotification("Welcome", "NilOS Adapter is online"); err != nil {
		t.Errorf("expected notification to succeed: %v", err)
	}

	data, err := adapter.ReadSensorData(int(nilhal.SensorAccelerometer))
	if err != nil || len(data) != 3 {
		t.Errorf("expected accelerometer 3-axis readings, got: %v (err: %v)", data, err)
	}
}

func TestNilBusIPC(t *testing.T) {
	client := nilbus.NewClient("")
	if err := client.Connect(); err != nil {
		t.Fatalf("failed connecting nilbus client: %v", err)
	}
	defer client.Close()

	// Test Notification service RPC
	res, err := client.Call("org.nilos.NotificationService", "Notify", []byte(`{"title":"Test"}`))
	if err != nil {
		t.Fatalf("nilbus Call failed: %v", err)
	}

	if !strings.Contains(string(res), "posted") {
		t.Errorf("expected response to contain 'posted', got: %s", string(res))
	}

	// Test Pub/Sub
	receivedChan := make(chan bool, 1)
	client.Subscribe("nilos.sensors.accel", func(evt []byte) {
		receivedChan <- true
	})
	client.Publish("nilos.sensors.accel", []byte("10.0,0.0,0.0"))

	select {
	case <-receivedChan:
		// OK
	case <-time.After(1 * time.Second):
		t.Errorf("expected subscriber to receive published event")
	}
}

func TestNilUIRenderer(t *testing.T) {
	renderer := nilui.NewRenderer("NilXApp", 400, 800)

	tree := engine.NewTree()
	tree.BeginNode("Column")
	tree.SetProp("backgroundColor", "#1A1A1A")

	tree.BeginNode("Text")
	tree.SetProp("text", "NilOS Vulkan UI")
	tree.SetProp("color", "#FFFFFF")
	tree.SetProp("fontSize", 20.0)
	tree.EndNode()

	tree.BeginNode("Button")
	tree.SetProp("backgroundColor", "#176BFF")
	tree.SetProp("borderRadius", 8.0)
	tree.EndNode()

	tree.EndNode()

	packet := renderer.RenderTree(tree.Root)
	if len(packet.Commands) < 2 {
		t.Fatalf("expected at least 2 GPU draw commands, got %d", len(packet.Commands))
	}

	jsonStr, err := renderer.ExportFrameJSON(packet)
	if err != nil {
		t.Fatalf("failed exporting frame JSON: %v", err)
	}

	if !strings.Contains(jsonStr, "NilOS Vulkan UI") || !strings.Contains(jsonStr, "#176BFF") {
		t.Errorf("expected frame JSON to contain text and button color, got:\n%s", jsonStr)
	}
}

func TestNilHAL(t *testing.T) {
	hal := nilhal.NewHAL()

	accel, err := hal.ReadSensor(nilhal.SensorAccelerometer)
	if err != nil || len(accel) != 3 {
		t.Errorf("expected 3-axis accelerometer data, got: %v", accel)
	}

	bat := hal.GetBatteryInfo()
	if bat.Level <= 0 || bat.Level > 100 {
		t.Errorf("expected valid battery level (0-100), got %d", bat.Level)
	}

	if err := hal.TriggerVibration(100); err != nil {
		t.Errorf("expected vibration trigger to succeed: %v", err)
	}
}

func TestNilOSProjectGenerator(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "nilos_bundle_*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	adapter := New()
	err = adapter.GenerateProject(tempDir, []byte("NABC_MOCK_BYTECODE"))
	if err != nil {
		t.Fatalf("adapter.GenerateProject failed: %v", err)
	}

	manifestBytes, err := os.ReadFile(filepath.Join(tempDir, "nilos", "app.nilxmanifest"))
	if err != nil {
		t.Fatalf("expected app.nilxmanifest to exist: %v", err)
	}

	if !strings.Contains(string(manifestBytes), "DisplayServer = wayland") ||
		!strings.Contains(string(manifestBytes), "GPUBackend = vulkan") {
		t.Errorf("expected manifest to specify wayland and vulkan, got:\n%s", string(manifestBytes))
	}

	launcherBytes, err := os.ReadFile(filepath.Join(tempDir, "nilos", "nilos-launcher.sh"))
	if err != nil {
		t.Fatalf("expected nilos-launcher.sh to exist: %v", err)
	}

	if !strings.Contains(string(launcherBytes), "wayland-0") {
		t.Errorf("expected launcher to configure Wayland display, got:\n%s", string(launcherBytes))
	}
}
