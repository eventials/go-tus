package leveldbstore

import (
	"github.com/eventials/go-tus"
	"github.com/syndtr/goleveldb/leveldb"
)

type LeveldbStore struct {
	db *leveldb.DB
}

func NewLeveldbStore(path string) (tus.Store, error) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}

	store := &LeveldbStore{db: db}
	return store, err
}

func (s *LeveldbStore) Get(fingerprint string) (string, bool) {
	url, err := s.db.Get([]byte(fingerprint), nil)
	ok := true
	if err != nil {
		ok = false
	}
	return string(url), ok
}

func (s *LeveldbStore) Set(fingerprint, url string) {
	s.db.Put([]byte(fingerprint), []byte(url), nil)
}

func (s *LeveldbStore) Delete(fingerprint string) {
	s.db.Delete([]byte(fingerprint), nil)
}

func (s *LeveldbStore) Close() {
	s.Close()
}
