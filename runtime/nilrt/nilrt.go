// Package nilrt implements the official NilRT Runtime Supervisor.
// It executes the unified pipeline:
// NilLang → nilc → NABC → NilRT → NilUI → Layout → Vulkan → Wayland → Onuron
package nilrt

import (
	"fmt"
	"time"

	"github.com/joysriramsarkar/alap-framework/compiler/codegen"
	"github.com/joysriramsarkar/alap-framework/compiler/lexer"
	"github.com/joysriramsarkar/alap-framework/compiler/parser"
	"github.com/joysriramsarkar/alap-framework/compiler/types"
	"github.com/joysriramsarkar/alap-framework/platform/onuron"
	"github.com/joysriramsarkar/alap-framework/platform/onuron/nilui"
	"github.com/joysriramsarkar/alap-framework/runtime/gc"
	"github.com/joysriramsarkar/alap-framework/runtime/scheduler"
	"github.com/joysriramsarkar/alap-framework/runtime/vm"
	"github.com/joysriramsarkar/alap-framework/ui/engine"
)

// PipelineResult encapsulates telemetry and frame artifacts for all 8 stages.
type PipelineResult struct {
	SourceCode   string             `json:"sourceCode"`
	NABCBytes    []byte             `json:"nabcBytes"`
	NABCSize     int                `json:"nabcSize"`
	UITree       *engine.WidgetNode `json:"uiTree"`
	LayoutWidth  float64            `json:"layoutWidth"`
	LayoutHeight float64            `json:"layoutHeight"`
	VulkanFrame  *nilui.FramePacket `json:"vulkanFrame"`
	VulkanJSON   string             `json:"vulkanJson"`
	WaylandTitle string             `json:"waylandTitle"`
	OnuronKernel string             `json:"onuronKernel"`
	NilOSKernel  string             `json:"nilosKernel"`
	StageTimes   map[string]string  `json:"stageTimes"`
	TotalElapsed string             `json:"totalElapsed"`
}

// Engine is the central NilRT supervisor.
type Engine struct {
	Scheduler *scheduler.Scheduler
	Memory    *gc.MemoryManager
	Adapter   *onuron.Adapter
}

// New creates a new NilRT runtime engine.
func New() *Engine {
	adapter := onuron.New()
	_ = adapter.Init()
	return &Engine{
		Scheduler: scheduler.New(4),
		Memory:    gc.New(),
		Adapter:   adapter,
	}
}

// ExecuteFullPipeline runs the full 8-step pipeline from NilLang source to NilOS screen.
func (e *Engine) ExecuteFullPipeline(src string, width, height int) (*PipelineResult, error) {
	totalStart := time.Now()
	stageTimes := make(map[string]string)

	// ─── Stage 1 & 2: NilLang → nilc (Lexer, Parser, TypeCheck, Codegen) ───
	s1 := time.Now()
	l := lexer.New("pipeline.nil", src)
	tokens := l.Tokenize()
	if len(l.Errors()) > 0 {
		return nil, fmt.Errorf("nilc lexer error: %v", l.Errors())
	}

	p := parser.New("pipeline.nil", tokens)
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		return nil, fmt.Errorf("nilc parser error: %v", p.Errors())
	}

	checker := types.New()
	checker.CheckProgram(prog)

	gen := codegen.New("pipeline_app")
	gen.GenerateProgram(prog)
	if len(gen.Errors()) > 0 {
		return nil, fmt.Errorf("nilc codegen error: %v", gen.Errors())
	}
	mod := gen.Module()
	stageTimes["1_nilc_compile"] = time.Since(s1).String()

	// ─── Stage 3: NABC Bytecode Serialization & Deserialization ───────────
	s2 := time.Now()
	nabcBytes := codegen.Serialize(mod)
	deserializedMod, err := codegen.Deserialize(nabcBytes)
	if err != nil {
		return nil, fmt.Errorf("NABC deserialization error: %w", err)
	}
	stageTimes["2_nabc_bytecode"] = time.Since(s2).String()

	// ─── Stage 4: NilRT VM Execution ──────────────────────────────────────
	s3 := time.Now()
	runner := vm.New(deserializedMod)
	e.Memory.TrackAlloc(uint64(len(nabcBytes)))

	if err := runner.Run(); err != nil {
		return nil, fmt.Errorf("NilRT VM execution error: %w", err)
	}
	stageTimes["3_nilrt_vm"] = time.Since(s3).String()

	// ─── Stage 5: NilUI Component Tree Binding ────────────────────────────
	s4 := time.Now()
	uiTree := runner.GetUITree()
	if uiTree == nil || uiTree.Root == nil {
		return nil, fmt.Errorf("NilUI tree not generated")
	}
	stageTimes["4_nilui_tree"] = time.Since(s4).String()

	// ─── Stage 6: Layout Calculation (Flexbox Measurement & Bounds) ───────
	s5 := time.Now()
	runner.ComputeUILayout(float64(width), float64(height))
	stageTimes["5_layout_flexbox"] = time.Since(s5).String()

	// ─── Stage 7: Vulkan GPU Command Buffer Compilation (nilui-gpu) ──────
	s6 := time.Now()
	vulkanRenderer := nilui.NewRenderer("AlapApp", width, height)
	framePacket := vulkanRenderer.RenderTree(uiTree.Root)
	frameJSON, err := vulkanRenderer.ExportFrameJSON(framePacket)
	if err != nil {
		return nil, fmt.Errorf("Vulkan frame export error: %w", err)
	}
	stageTimes["6_vulkan_gpu"] = time.Since(s6).String()

	// ─── Stage 8: Wayland Surface Presentation on NilOS ───────────────────
	s7 := time.Now()
	_ = e.Adapter.CreateWindow("AlapApp", width, height)
	_ = e.Adapter.SendNotification("NilRT", "Frame rendered on Wayland surface")
	stageTimes["7_wayland_surface"] = time.Since(s7).String()
	stageTimes["8_onuron_host"] = e.Adapter.GetKernelVersion()

	return &PipelineResult{
		SourceCode:   src,
		NABCBytes:    nabcBytes,
		NABCSize:     len(nabcBytes),
		UITree:       uiTree.Root,
		LayoutWidth:  float64(width),
		LayoutHeight: float64(height),
		VulkanFrame:  framePacket,
		VulkanJSON:   frameJSON,
		WaylandTitle: "AlapApp",
		OnuronKernel: e.Adapter.GetKernelVersion(),
		NilOSKernel:  e.Adapter.GetKernelVersion(),
		StageTimes:   stageTimes,
		TotalElapsed: time.Since(totalStart).String(),
	}, nil
}
