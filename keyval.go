package main

import "sync"

type KeyVal struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func NewKeyVal() *KeyVal {
	return &KeyVal{
		data: map[string][]byte{},
	}
}

func (kv *KeyVal) Set(key, val []byte) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	kv.data[string(key)] = []byte(val)

	return nil

}

func (kv *KeyVal) Get(key []byte) ([]byte, bool) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()

	val, ok := kv.data[string(key)]

	return val, ok

}
