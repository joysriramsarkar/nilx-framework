// Package state implements reactive state reconciliation and UI tree diffing for Alap.
package state

import (
	"reflect"
	"sync"

	"github.com/joysriramsarkar/alap-framework/ui/engine"
)

// MutationKind describes the type of DOM/UI modification.
type MutationKind int

const (
	MutationAddNode MutationKind = iota
	MutationRemoveNode
	MutationUpdateProps
	MutationUpdateText
)

// Mutation represents a granular UI patch operation.
type Mutation struct {
	Kind     MutationKind           `json:"kind"`
	NodeID   string                 `json:"nodeId"`
	ParentID string                 `json:"parentId,omitempty"`
	Props    map[string]interface{} `json:"props,omitempty"`
	Text     string                 `json:"text,omitempty"`
}

// Reconciler tracks tree state, performs VDOM diffing, and triggers re-renders.
type Reconciler struct {
	mu           sync.RWMutex
	currentTree  *engine.WidgetNode
	renderTarget func(mutations []Mutation)
}

// NewReconciler creates a reactive UI reconciler.
func NewReconciler(onUpdate func(mutations []Mutation)) *Reconciler {
	return &Reconciler{
		renderTarget: onUpdate,
	}
}

// SetTree sets the initial UI tree.
func (r *Reconciler) SetTree(tree *engine.WidgetNode) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.currentTree = tree
}

// DiffAndPatch computes minimal changes between old and new tree and notifies render target.
func (r *Reconciler) DiffAndPatch(newTree *engine.WidgetNode) []Mutation {
	r.mu.Lock()
	defer r.mu.Unlock()

	var mutations []Mutation
	if r.currentTree == nil {
		if newTree != nil {
			mutations = append(mutations, Mutation{
				Kind:   MutationAddNode,
				NodeID: newTree.ID,
				Props:  newTree.Props,
			})
		}
		r.currentTree = newTree
		if r.renderTarget != nil && len(mutations) > 0 {
			r.renderTarget(mutations)
		}
		return mutations
	}

	diffNodes("", r.currentTree, newTree, &mutations)
	r.currentTree = newTree

	if r.renderTarget != nil && len(mutations) > 0 {
		r.renderTarget(mutations)
	}

	return mutations
}

func diffNodes(parentID string, oldNode, newNode *engine.WidgetNode, mutations *[]Mutation) {
	if oldNode == nil && newNode == nil {
		return
	}

	if oldNode == nil && newNode != nil {
		*mutations = append(*mutations, Mutation{
			Kind:     MutationAddNode,
			NodeID:   newNode.ID,
			ParentID: parentID,
			Props:    newNode.Props,
		})
		return
	}

	if oldNode != nil && newNode == nil {
		*mutations = append(*mutations, Mutation{
			Kind:   MutationRemoveNode,
			NodeID: oldNode.ID,
		})
		return
	}

	// Compare Props
	changedProps := make(map[string]interface{})
	for k, newVal := range newNode.Props {
		if oldVal, ok := oldNode.Props[k]; !ok || !reflect.DeepEqual(oldVal, newVal) {
			changedProps[k] = newVal
		}
	}

	if len(changedProps) > 0 {
		*mutations = append(*mutations, Mutation{
			Kind:   MutationUpdateProps,
			NodeID: newNode.ID,
			Props:  changedProps,
		})
	}

	// Recurse children
	maxLen := len(oldNode.Children)
	if len(newNode.Children) > maxLen {
		maxLen = len(newNode.Children)
	}

	for i := 0; i < maxLen; i++ {
		var oldChild, newChild *engine.WidgetNode
		if i < len(oldNode.Children) {
			oldChild = oldNode.Children[i]
		}
		if i < len(newNode.Children) {
			newChild = newNode.Children[i]
		}
		diffNodes(newNode.ID, oldChild, newChild, mutations)
	}
}

// BindSignalToComponent re-renders component when signal value changes.
func BindSignalToComponent(sig *Signal, reconciler *Reconciler, rebuild func() *engine.WidgetNode) func() {
	return sig.Subscribe(func(newVal interface{}) {
		newTree := rebuild()
		reconciler.DiffAndPatch(newTree)
	})
}
