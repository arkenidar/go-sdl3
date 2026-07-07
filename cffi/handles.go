// Package cffi exposes internal/widgets as a flat C ABI (via
// -buildmode=c-shared) for use from ctypes, LuaJIT FFI, and similar.
// See docs/ffi-design.md for the full design.
package main

import "sync"

type handleTable struct {
	mu   sync.Mutex
	next uint64
	objs map[uint64]any
}

var handles = &handleTable{objs: make(map[uint64]any)}

func (t *handleTable) put(v any) uint64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.next++
	h := t.next
	t.objs[h] = v
	return h
}

func (t *handleTable) get(h uint64) (any, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	v, ok := t.objs[h]
	return v, ok
}

func (t *handleTable) delete(h uint64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.objs, h)
}
