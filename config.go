package tus

import (
	"fmt"
	"log"
	"os"

	"github.com/eventials/go-tus/storage"
	memstore "github.com/eventials/go-tus/storage/memory"
)

// Config provides a way to configure the Client depending on your needs.
type Config struct {
	// ChunkSize divide the file into chunks.
	ChunkSize           int64
	// Resume enables resumable upload.
	Resume              bool
	// OverridePatchMethod allow to by pass proxies sendind a POST request instead of PATCH.
	OverridePatchMethod bool
	// Storage is the backend to save upload progress.
	// If Resume is true the Storage is required.
	Storage             storage.Storage
	// Logger is the logger to use internally, mostly for upload progress.
	Logger              *log.Logger
}

// DefaultConfig return the default Client configuration.
func DefaultConfig() *Config {
	s, _ := memstore.NewMemoryStorage()

	return &Config{
		ChunkSize:           1048576 * 15, // 15 MB
		Resume:              true,
		OverridePatchMethod: false,
		Storage:             s,
		Logger:              log.New(os.Stdout, "[tus] ", 0),
	}
}

// Validate validates the custom configuration.
func (c *Config) Validate() error {
	if c.ChunkSize < 1 {
		return fmt.Errorf("invalid configuration: ChunkSize must be greater than zero.")
	}

	if c.Logger == nil {
		return fmt.Errorf("invalid configuration: Logger can't be nil.")
	}

	if c.Resume && c.Storage == nil {
		return fmt.Errorf("invalid configuration: Storage can't be nil if Resume is enable.")
	}

	return nil
}
