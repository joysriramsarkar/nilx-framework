package state

import (
	"fmt"
	"testing"

	"github.com/joysriramsarkar/alap-framework/ui/engine"
)

func TestReactiveReconciliation(t *testing.T) {
	countSig := NewSignal(0)
	var lastMutations []Mutation

	reconciler := NewReconciler(func(muts []Mutation) {
		lastMutations = muts
	})

	render := func() *engine.WidgetNode {
		val := countSig.Get().(int)
		root := engine.NewWidgetNode("col_1", "Column")
		txt := engine.NewWidgetNode("txt_1", "Text")
		txt.Props["text"] = fmt.Sprintf("Count: %d", val)
		root.AddChild(txt)
		return root
	}

	// Initial render
	initialTree := render()
	reconciler.SetTree(initialTree)

	// Bind signal
	unsubscribe := BindSignalToComponent(countSig, reconciler, render)
	defer unsubscribe()

	// Mutate state
	countSig.Set(1)

	if len(lastMutations) == 0 {
		t.Fatalf("expected mutations on state change, got 0")
	}

	foundTextUpdate := false
	for _, m := range lastMutations {
		if m.Kind == MutationUpdateProps && m.Props["text"] == "Count: 1" {
			foundTextUpdate = true
			break
		}
	}

	if !foundTextUpdate {
		t.Errorf("expected text update mutation for 'Count: 1', got %v", lastMutations)
	}
}
