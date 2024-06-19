package storage

import (
	"errors"
	"fmt"
	"sync"

	"github.com/kode4food/respect/pkg/resp"
)

type (
	memNode struct {
		children map[resp.BulkString]*memNode
		value    resp.Value
		version  int
		sync.RWMutex
	}

	keySeen struct{}
)

// Error messages
const (
	ErrKeyNotFound = "key not found: %s"
)

// compile-time checks for interface implementation
var (
	_ Storage = (*memNode)(nil)
)

func NewMemory() Storage {
	return &memNode{}
}

func (m *memNode) Get(key Key) (resp.Value, error) {
	if len(key) == 0 {
		return nil, fmt.Errorf(ErrEmptyKey)
	}
	m.RLock()
	if child := m.fetchNested(key); child != nil {
		defer child.RUnlock()
		if child.value != nil {
			return child.value, nil
		}
	}
	return nil, fmt.Errorf(ErrKeyNotFound, key)
}

func (m *memNode) fetchNested(key Key) *memNode {
	if len(key) == 0 {
		return m
	}
	child, ok := m.getChild(key[0])
	if !ok {
		m.RUnlock()
		return nil
	}
	m.transferRLockTo(child)
	return child.fetchNested(key[1:])
}

func (m *memNode) getChild(comp resp.BulkString) (*memNode, bool) {
	if m.children == nil {
		return nil, false
	}
	child, ok := m.children[comp]
	return child, ok
}

func (m *memNode) Set(key Key, value resp.Value) (resp.Value, error) {
	if len(key) == 0 {
		return nil, fmt.Errorf(ErrEmptyKey)
	}
	m.Lock()
	return m.set(key, value), nil
}

func (m *memNode) set(k Key, v resp.Value) resp.Value {
	if len(k) == 0 {
		old := m.value
		m.value = v
		m.version++
		m.Unlock()
		return old
	}
	child := m.ensureNested(k[0])
	m.transferLockTo(child)
	return child.set(k[1:], v)
}

func (m *memNode) ensureNested(comp resp.BulkString) *memNode {
	if m.children == nil {
		m.children = map[resp.BulkString]*memNode{}
	}
	child, ok := m.children[comp]
	if !ok {
		child = &memNode{}
		m.children[comp] = child
		m.version++
	}
	return child
}

func (m *memNode) Delete(key Key) (resp.Value, error) {
	if len(key) == 0 {
		return nil, fmt.Errorf(ErrEmptyKey)
	}
	m.Lock()
	res, ok := m.delete(key)
	if !ok {
		return nil, fmt.Errorf(ErrKeyNotFound, key)
	}
	return res, nil
}

func (m *memNode) delete(k Key) (resp.Value, bool) {
	if len(k) > 0 {
		return m.deleteNested(k)
	}
	defer m.Unlock()
	old := m.value
	if old == nil {
		return nil, false
	}
	m.value = nil
	m.version++
	return old, true
}

func (m *memNode) deleteNested(k Key) (resp.Value, bool) {
	comp := k[0]
	child, ok := m.getChild(comp)
	if !ok {
		m.Unlock()
		return nil, false
	}
	m.transferLockTo(child)
	v, ok := child.delete(k[1:])
	if !ok {
		return nil, false
	}

	m.attemptToPrune(comp)
	return v, true
}

func (m *memNode) attemptToPrune(comp resp.BulkString) {
	m.Lock()
	defer m.Unlock()
	if child, ok := m.getChild(comp); ok && child.canBePruned() {
		delete(m.children, comp)
	}
}

func (m *memNode) canBePruned() bool {
	return m.value == nil && len(m.children) == 0
}

func (m *memNode) Exists(key Key) (bool, error) {
	if len(key) == 0 {
		return false, fmt.Errorf(ErrEmptyKey)
	}
	m.RLock()
	if child := m.fetchNested(key); child != nil {
		defer child.RUnlock()
		if child.value != nil {
			return true, nil
		}
	}
	return false, fmt.Errorf(ErrKeyNotFound, key)
}

func (m *memNode) IterateKeys(pfx Key, accept Accept[Key]) error {
	m.RLock()
	if child := m.fetchNested(pfx); child != nil {
		err := child.forEach(pfx, accept)
		if err == nil || errors.Is(err, StopIteration) {
			return nil
		}
		return err
	}
	return fmt.Errorf(ErrKeyNotFound, pfx)
}

func (m *memNode) forEach(pfx Key, accept Accept[Key]) error {
	defer m.RUnlock()
	keys := m.getKeys()
	ver := m.version
	cur := 0

	for cur < len(keys) {
		if m.version != ver {
			old := keys[:cur]
			keys = append(old, m.getNewKeys(old)...)
			ver = m.version
			continue
		}
		if child, ok := m.children[keys[cur]]; ok {
			ck := append(pfx, keys[cur])
			child.RLock()
			if err := m.doRUnlocked(func() error {
				return child.forEach(ck, accept)
			}); err != nil {
				return err
			}
		}
		cur++
	}

	if m.value != nil {
		return m.doRUnlocked(func() error {
			return accept(pfx)
		})
	}
	return nil
}

func (m *memNode) getKeys() []resp.BulkString {
	res := make([]resp.BulkString, 0, len(m.children))
	for k := range m.children {
		res = append(res, k)
	}
	return res
}

func (m *memNode) getNewKeys(old []resp.BulkString) []resp.BulkString {
	seen := make(map[resp.BulkString]keySeen, len(old))
	for _, k := range old {
		seen[k] = keySeen{}
	}
	res := make([]resp.BulkString, 0, len(m.children)-len(seen))
	for k := range m.children {
		if _, ok := seen[k]; !ok {
			res = append(res, k)
		}
	}
	return res
}

func (m *memNode) transferLockTo(child *memNode) {
	child.Lock()
	m.Unlock()
}

func (m *memNode) transferRLockTo(child *memNode) {
	child.RLock()
	m.RUnlock()
}

func (m *memNode) doRUnlocked(fn func() error) error {
	m.RUnlock()
	defer m.RLock()
	return fn()
}
