package log

import "sync"

type Exist struct {
	lock *sync.RWMutex
	data map[string]struct{}
}

func NewExist(initCapacity int) *Exist {
	return &Exist{
		lock: &sync.RWMutex{},
		data: make(map[string]struct{}, initCapacity)}
}

func (e *Exist) Exist(key string) bool {
	e.lock.RLock()
	defer e.lock.RUnlock()

	_, exist := e.data[key]

	return exist
}

func (e *Exist) Set(key string) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.data[key] = struct{}{}
}

func (e *Exist) Copy() *Exist {
	result := NewExist(len(e.data))

	e.lock.RLock()
	defer e.lock.RUnlock()

	for key := range e.data {
		result.data[key] = struct{}{}
	}

	return result
}
