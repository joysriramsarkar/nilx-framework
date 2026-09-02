// Package abi implements the ABI boundary for the Alap Runtime.
// It allows embedding Alap in Android (JNI/NDK), iOS (Swift/Metal), and desktop apps.
package abi

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/joysriramsarkar/alap-framework/compiler/codegen"
	"github.com/joysriramsarkar/alap-framework/runtime/vm"
	"github.com/joysriramsarkar/alap-framework/ui/engine"
)

// RuntimeContext represents an active Alap VM and UI runtime instance.
type RuntimeContext struct {
	VM      *vm.VM
	Module  *codegen.Module
	UITree  *engine.UITree
	Version string
}

var (
	handlesMu  sync.RWMutex
	handles    = make(map[uintptr]*RuntimeContext)
	nextHandle uintptr = 1
)

func RegisterContext(r *RuntimeContext) uintptr {
	handlesMu.Lock()
	defer handlesMu.Unlock()
	id := nextHandle
	nextHandle++
	handles[id] = r
	return id
}

func GetContext(id uintptr) *RuntimeContext {
	handlesMu.RLock()
	defer handlesMu.RUnlock()
	return handles[id]
}

func RemoveContext(id uintptr) {
	handlesMu.Lock()
	defer handlesMu.Unlock()
	delete(handles, id)
}

// CreateRuntime creates and registers a new runtime context.
func CreateRuntime() uintptr {
	ctx := &RuntimeContext{
		Version: "0.1.0",
		UITree:  engine.NewUITree(),
	}
	return RegisterContext(ctx)
}

// DestroyRuntime frees the runtime context.
func DestroyRuntime(handle uintptr) {
	RemoveContext(handle)
}

// ExecuteBytecode deserializes and runs NABC bytecode within the context.
func ExecuteBytecode(handle uintptr, nabc []byte) error {
	ctx := GetContext(handle)
	if ctx == nil {
		return fmt.Errorf("invalid runtime handle %d", handle)
	}

	mod, err := codegen.Deserialize(nabc)
	if err != nil {
		return fmt.Errorf("deserialization error: %w", err)
	}

	ctx.Module = mod
	ctx.VM = vm.New(mod)
	if err := ctx.VM.Run(); err != nil {
		return fmt.Errorf("execution error: %w", err)
	}
	ctx.UITree = ctx.VM.GetUITree()
	return nil
}

// DispatchTouchEvent routes touch coordinates to the active UI hierarchy.
func DispatchTouchEvent(handle uintptr, x, y float64) bool {
	ctx := GetContext(handle)
	if ctx == nil || ctx.UITree == nil {
		return false
	}
	return ctx.UITree.DispatchTouch(x, y)
}

// GetUIRootJSON exports the current UI component tree as a JSON string for native host rendering.
func GetUIRootJSON(handle uintptr) (string, error) {
	ctx := GetContext(handle)
	if ctx == nil || ctx.UITree == nil || ctx.UITree.Root == nil {
		return "{}", nil
	}
	data, err := json.Marshal(ctx.UITree.Root)
	if err != nil {
		return "{}", err
	}
	return string(data), nil
}

// RunBytecodeInGo is a one-shot helper for running bytecode in pure Go tests/tools.
func RunBytecodeInGo(nabc []byte) (string, error) {
	h := CreateRuntime()
	defer DestroyRuntime(h)

	if err := ExecuteBytecode(h, nabc); err != nil {
		return "", err
	}
	return GetUIRootJSON(h)
}
