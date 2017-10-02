package tus

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

type Metadata map[string]string

type Upload struct {
	stream   io.ReadSeeker
	size     int64
	offset   int64
	finished bool

	Fingerprint string
	Metadata    Metadata
}

// Updates the Upload information based on offset
func (u *Upload) UpdateProgress(offset int64) {
	u.offset = offset
	if u.size == offset {
		u.finished = true
	}
}

// Returns whether this upload is finished
func (u *Upload) Finished() bool {
	return u.finished
}

// Returns the progress in a percentage
func (u *Upload) Progress() int64 {
	return (u.offset * 100) / u.size
}

// Offset returns the current offset
func (u *Upload) Offset() int64 {
	return u.offset
}

// Size returns the size of the upload body.
func (u *Upload) Size() int64 {
	return u.size
}

// EncodedMetadata encodes the upload metadata.
func (u *Upload) EncodedMetadata() string {
	var buffer bytes.Buffer

	for k, v := range u.Metadata {
		buffer.WriteString(fmt.Sprintf("%s %s,", k, b64encode(v)))
	}

	return buffer.String()
}

// NewUploadFromFile creates a new Upload from an os.File.
func NewUploadFromFile(f *os.File) (*Upload, error) {
	fi, err := f.Stat()

	if err != nil {
		return nil, err
	}

	metadata := map[string]string{
		"filename": fi.Name(),
	}

	fingerprint := fmt.Sprintf("%s-%d-%d", fi.Name(), fi.Size(), fi.ModTime())

	return NewUpload(f, fi.Size(), metadata, fingerprint), nil
}

// NewUploadFromBytes creates a new upload from a byte array.
func NewUploadFromBytes(b []byte) *Upload {
	buffer := bytes.NewReader(b)
	return NewUpload(buffer, buffer.Size(), nil, "")
}

// NewUpload creates a new upload from an io.Reader.
func NewUpload(reader io.Reader, size int64, metadata Metadata, fingerprint string) *Upload {
	stream, ok := reader.(io.ReadSeeker)

	if !ok {
		buf := new(bytes.Buffer)
		buf.ReadFrom(reader)
		stream = bytes.NewReader(buf.Bytes())
	}

	if metadata == nil {
		metadata = make(Metadata)
	}

	return &Upload{
		stream: stream,
		size:   size,

		Fingerprint: fingerprint,
		Metadata:    metadata,
	}
}

func b64encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}
