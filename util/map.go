package util

import (
	"math/rand"
	"sync"
)

type StringMap struct {
	s map[string]interface{}
	sync.RWMutex
}

func NewStringMap() *StringMap {
	return &StringMap{
		s: make(map[string]interface{}),
		// RWMutex: sync.RWMutex{},
	}
}

func (m *StringMap) Get(k string) interface{} {
	m.RLock()
	defer m.RUnlock()

	return m.s[k]
}

func (m *StringMap) RandGet() interface{} {

	// TODO: has there any smarter way?

	l := len(m.s)
	if l == 0 {
		return nil
	}

	i := 0
	index := rand.Intn(l)
	for _, v := range m.s {
		if i == index {
			return v
		}else{
			i++
		}
	}
	return nil
}

func (m *StringMap) Set(k string, v interface{}) {
	m.Lock()
	defer m.Unlock()

	m.s[k] = v
}

func (m *StringMap) Del(k string) bool {
	if m.Get(k) == nil {
		return false
	} else {
		m.Lock()
		delete(m.s, k)
		m.Unlock()
		return true
	}
}

func (m *StringMap) GetKeys() []string {
	m.RLock()
	defer m.RUnlock()
	s := make([]string, len(m.s))
	for k := range m.s {
		s = append(s, k)
	}
	return s
}
