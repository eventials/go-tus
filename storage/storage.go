package storage

import (
	"os"
)

type Storage interface {
	Get(f *os.File) (string, bool)
	Set(f *os.File, url string)
	Delete(f *os.File)
	Close()
}
