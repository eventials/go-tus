package tus

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfingMissingStore(t *testing.T) {
	c := Config{
		ChunkSize:           1048576 * 15, // 15 MB
		Resume:              true,
		OverridePatchMethod: false,
		Store:               nil,
		Logger:              log.New(os.Stdout, "[tus] ", 0),
	}

	assert.NotNil(t, c.Validate())
}

func TestConfingChunkSizeZero(t *testing.T) {
	c := Config{
		ChunkSize:           0,
		Resume:              false,
		OverridePatchMethod: false,
		Store:               nil,
		Logger:              log.New(os.Stdout, "[tus] ", 0),
	}

	assert.NotNil(t, c.Validate())
}

func TestConfingMissingLogger(t *testing.T) {
	c := Config{
		ChunkSize:           1048576 * 15, // 15 MB
		Resume:              false,
		OverridePatchMethod: false,
		Store:               nil,
		Logger:              nil,
	}

	assert.NotNil(t, c.Validate())
}

func TestConfingValid(t *testing.T) {
	c := DefaultConfig()
	assert.Nil(t, c.Validate())
}
