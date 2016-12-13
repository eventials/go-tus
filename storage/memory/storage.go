package memory

import (
	"github.com/eventials/go-tus/storage"
	"os"
)

// MemoryStorage implements an in-memory Storage.
type MemoryStorage struct {
	m map[string]string
}

// NewMemoryStorage creates a new MemoryStorage.
func NewMemoryStorage() (storage.Storage, error) {
	return &MemoryStorage{
		make(map[string]string),
	}, nil
}

func (s *MemoryStorage) Get(f *os.File) (string, bool) {
	url, ok := s.m[f.Name()]
	return url, ok
}

func (s *MemoryStorage) Set(f *os.File, url string) {
	s.m[f.Name()] = url
}

func (s *MemoryStorage) Delete(f *os.File) {
	delete(s.m, f.Name())
}

func (s *MemoryStorage) Close() {
	for k := range s.m {
		delete(s.m, k)
	}
}
