package util

import "sync"

func MapLoadOrCreateMutex(m *sync.Map, key string) *sync.Mutex {
	if _, hasKey := m.Load(key); !hasKey {
		m.Store(key, &sync.Mutex{})
	}

	return MapMustLoad(m, key).(*sync.Mutex)
}

func MapMustLoad(m *sync.Map, key string) interface{} {
	val, hasKey := m.Load(key)
	if !hasKey {
		panic("key must be set in map")
	}

	return val
}
