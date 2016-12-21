package tus

import (
	"bytes"
)

type Uploader struct {
	client  *Client
	url     string
	upload  *Upload
	offset  int64
	aborted bool
}

// Abort aborts the upload process.
// It doens't abort the current chunck, only the remaining.
func (u *Uploader) Abort() {
	u.aborted = true
}

// IsAborted returns true if the upload was aborted.
func (u *Uploader) IsAborted() bool {
	return u.aborted
}

// Upload uploads the entire body to the server.
func (u *Uploader) Upload() error {
	for u.offset < u.upload.size && !u.aborted {
		err := u.UploadChunck()

		if err != nil {
			return err
		}
	}

	return nil
}

// UploadChunck uploads a single chunck.
func (u *Uploader) UploadChunck() error {
	data := make([]byte, u.client.Config.ChunkSize)

	_, err := u.upload.stream.Seek(u.offset, 0)

	if err != nil {
		return err
	}

	size, err := u.upload.stream.Read(data)

	if err != nil {
		return err
	}

	body := bytes.NewBuffer(data[:size])

	newOffset, err := u.client.uploadChunck(u.url, body, int64(size), u.offset)

	if err != nil {
		return err
	}

	u.offset = newOffset

	return nil
}

// NewUploader creates a new Uploader.
func NewUploader(client *Client, url string, upload *Upload, offset int64) *Uploader {
	return &Uploader{
		client,
		url,
		upload,
		offset,
		false,
	}
}
