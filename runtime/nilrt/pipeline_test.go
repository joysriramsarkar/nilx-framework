package nilrt

import (
	"strings"
	"testing"
)

func TestCompletePipelineExecution(t *testing.T) {
	nilLangCode := `
component MainView {
    build() {
        Column {
            Text("Alap on Onuron")
                .fontSize(22)
                .color("#FFFFFF")

            Button("Click to Trigger GPU") {
                onClick => {
                    print("Button clicked on Wayland surface")
                }
            }
            .backgroundColor("#176BFF")
            .borderRadius(8)
        }
        .backgroundColor("#1C1C1E")
    }
}
`
	engine := New()
	res, err := engine.ExecuteFullPipeline(nilLangCode, 1080, 1920)
	if err != nil {
		t.Fatalf("pipeline execution failed: %v", err)
	}

	// 1. Verify NABC Bytecode
	if res.NABCSize == 0 || len(res.NABCBytes) == 0 {
		t.Errorf("expected non-empty NABC bytecode, got %d bytes", res.NABCSize)
	}

	// 2. Verify NilUI Tree
	if res.UITree == nil || res.UITree.Type != "MainView" {
		t.Errorf("expected root component node 'MainView', got: %v", res.UITree)
	}
	if len(res.UITree.Children) == 0 || res.UITree.Children[0].Type != "Column" {
		t.Errorf("expected child UI node 'Column', got: %v", res.UITree.Children)
	}

	// 3. Verify Layout Calculation
	if res.UITree.Bounds.Width != 1080 || res.UITree.Bounds.Height != 1920 {
		t.Errorf("expected root layout bounds 1080x1920, got %fx%f",
			res.UITree.Bounds.Width, res.UITree.Bounds.Height)
	}

	// 4. Verify Vulkan GPU Draw Commands
	if len(res.VulkanFrame.Commands) < 2 {
		t.Errorf("expected at least 2 Vulkan draw commands, got %d", len(res.VulkanFrame.Commands))
	}
	if !strings.Contains(res.VulkanJSON, "Alap on Onuron") {
		t.Errorf("expected Vulkan frame JSON to contain text, got:\n%s", res.VulkanJSON)
	}

	// 5. Verify Onuron & Wayland
	if res.WaylandTitle != "AlapApp" {
		t.Errorf("expected Wayland title 'NilXApp', got %q", res.WaylandTitle)
	}
	if !strings.Contains(res.OnuronKernel, "Onuron") && !strings.Contains(res.NilOSKernel, "Onuron") {
		t.Errorf("expected Onuron kernel identification, got %q", res.NilOSKernel)
	}

	t.Logf("Pipeline Execution Success! Total time: %s", res.TotalElapsed)
	for stage, dur := range res.StageTimes {
		t.Logf("  Stage [%s]: %s", stage, dur)
	}
}
