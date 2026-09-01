// Package gc provides memory allocation tracking and statistics for NilRT.
package gc

import (
	"runtime"
	"sync"
	"sync/atomic"
)

// Stats tracks NilRT memory allocations and object counts.
type Stats struct {
	AllocatedObjects uint64 `json:"allocatedObjects"`
	AllocatedBytes   uint64 `json:"allocatedBytes"`
	LiveObjects      uint64 `json:"liveObjects"`
	GCCycles         uint64 `json:"gcCycles"`
}

// MemoryManager monitors and manages NilRT heap memory.
type MemoryManager struct {
	mu          sync.RWMutex
	liveObjects uint64
	totalAllocs uint64
	totalBytes  uint64
	gcCount     uint64
}

// New creates a new memory manager instance.
func New() *MemoryManager {
	return &MemoryManager{}
}

// TrackAlloc records a new heap allocation.
func (m *MemoryManager) TrackAlloc(sizeBytes uint64) {
	atomic.AddUint64(&m.totalAllocs, 1)
	atomic.AddUint64(&m.totalBytes, sizeBytes)
	atomic.AddUint64(&m.liveObjects, 1)
}

// TrackFree records object deallocation.
func (m *MemoryManager) TrackFree() {
	atomic.AddUint64(&m.liveObjects, ^uint64(0))
}

// Collect triggers a memory collection cycle.
func (m *MemoryManager) Collect() {
	runtime.GC()
	atomic.AddUint64(&m.gcCount, 1)
}

// GetStats returns current memory telemetry.
func (m *MemoryManager) GetStats() Stats {
	return Stats{
		AllocatedObjects: atomic.LoadUint64(&m.totalAllocs),
		AllocatedBytes:   atomic.LoadUint64(&m.totalBytes),
		LiveObjects:      atomic.LoadUint64(&m.liveObjects),
		GCCycles:         atomic.LoadUint64(&m.gcCount),
	}
}
