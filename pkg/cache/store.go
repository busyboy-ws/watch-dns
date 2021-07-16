package store

import "sync"



type MapStore struct {
	lock  sync.Mutex
	items map[string]string

}


func NewStore() *MapStore {
	return &MapStore{

		items: map[string]string{},
	}
}

func (m *MapStore)Add(key string, value string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.items[key] = value
}

func (m *MapStore)Update(key string, value string)  {
	m.Add(key,value)
}

func (m *MapStore)Delete(key string)  {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.items, key)
}

func (m *MapStore)Get(key string) (string, bool) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if value, ok := m.items[key]; ok{
		return value, true
	}
	return "", false
}

func (m *MapStore)List() []interface{}  {
	m.lock.Lock()
	defer m.lock.Unlock()
	list := make([]interface{}, 0, len(m.items))
	for item  := range m.items{
		list = append(list, item)
	}
	return list
}