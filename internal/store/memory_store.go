package store

import "sync"

type MemoryStore struct {
	mu   sync.RWMutex
	data Snapshot
}

func NewMemoryStore() *MemoryStore {
	data := Snapshot{
		NextCategoryID: 1,
		NextArticleID:  1,
		NextTagID:      1,
	}
	NormalizeSnapshot(&data)
	return &MemoryStore{data: data}
}

func (s *MemoryStore) Read() Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return CloneSnapshot(s.data)
}

func (s *MemoryStore) Write(fn func(*Snapshot) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	next := CloneSnapshot(s.data)
	if err := fn(&next); err != nil {
		return err
	}
	NormalizeSnapshot(&next)
	s.data = next
	return nil
}
