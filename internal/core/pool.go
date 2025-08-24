package core

import (
	"sync"
)

// ObjectPool provides object pooling for performance optimization
type ObjectPool struct {
	pool sync.Pool
}

// NewObjectPool creates a new object pool
func NewObjectPool(newFunc func() interface{}) *ObjectPool {
	return &ObjectPool{
		pool: sync.Pool{
			New: newFunc,
		},
	}
}

// Get retrieves an object from the pool
func (p *ObjectPool) Get() interface{} {
	return p.pool.Get()
}

// Put returns an object to the pool
func (p *ObjectPool) Put(obj interface{}) {
	p.pool.Put(obj)
}

// Global pools for frequently used objects
var (
	StringBuilderPool *ObjectPool
	ByteBufferPool    *ObjectPool
	CommandPool       *ObjectPool
)

// InitializePool initializes global object pools
func InitializePool() {
	// String builder pool for efficient string concatenation
	StringBuilderPool = NewObjectPool(func() interface{} {
		return make([]byte, 0, 256)
	})

	// Byte buffer pool for I/O operations
	ByteBufferPool = NewObjectPool(func() interface{} {
		return make([]byte, 4096)
	})

	// Command pool for parsed commands
	CommandPool = NewObjectPool(func() interface{} {
		return &struct {
			Name   string
			Args   []string
			Buffer []string
		}{
			Args:   make([]string, 0, 8),
			Buffer: make([]string, 0, 8),
		}
	})

	// Initialize caches as well
	InitializeCache()
}

// ResetByteSlice resets a byte slice for reuse
func ResetByteSlice(slice []byte) []byte {
	return slice[:0]
}

// ResetStringSlice resets a string slice for reuse
func ResetStringSlice(slice []string) []string {
	for i := range slice {
		slice[i] = ""
	}
	return slice[:0]
}
