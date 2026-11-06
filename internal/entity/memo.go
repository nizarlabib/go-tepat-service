package entity

import (
	"sync"
)

type Memo interface {
	HasMemo(key string) bool
	GetMemo(key string) any
	GetMemoString(key string) string
	GetMemoBool(key string) bool
	SetMemo(key string, value any)
	RemoveMemo(key string)
	ResetMemo()
}

type memo struct {
	mu    sync.RWMutex
	value map[string]any `json:"-" gorm:"-"`
}

func (b *memo) HasMemo(key string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.value == nil {
		return false
	}
	_, ok := b.value[key]
	return ok
}

func (b *memo) GetMemo(key string) any {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.value == nil {
		return nil
	}
	return b.value[key]
}

func (b *memo) GetMemoString(key string) string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.value == nil {
		return ""
	}
	v, ok := b.value[key]
	if !ok {
		return ""
	}
	if val, ok := v.(string); ok {
		return val
	}
	return ""
}

func (b *memo) GetMemoInt(key string) int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.value == nil {
		return -1
	}
	v, ok := b.value[key]
	if !ok {
		return -1
	}
	if val, ok := v.(int); ok {
		return val
	}
	return -1
}

func (b *memo) GetMemoBool(key string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.value == nil {
		return false
	}
	v, ok := b.value[key]
	if !ok {
		return false
	}
	if val, ok := v.(bool); ok {
		return val
	}
	return false
}

func (b *memo) SetMemo(key string, value any) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.value == nil {
		b.value = make(map[string]any)
	}
	b.value[key] = value
}

func (b *memo) RemoveMemo(key string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.value == nil {
		return
	}
	delete(b.value, key)
}

func (b *memo) ResetMemo() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.value = nil
}
