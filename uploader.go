package tus

import (
	"bytes"
	"context"
	"sync"
)

type Uploader struct {
	client     *Client
	url        string
	upload     *Upload
	offset     int64
	lock       sync.RWMutex
	aborted    bool
	cancel     context.CancelFunc
	uploadSubs []chan Upload
	notifyChan chan bool
}

// Subscribes to progress updates.
func (u *Uploader) NotifyUploadProgress(c chan Upload) {
	u.uploadSubs = append(u.uploadSubs, c)
}

// Abort aborts the upload process.
func (u *Uploader) Abort() {
	u.cancel()
	u.lock.Lock()
	u.aborted = true
	u.lock.Unlock()
}

// IsAborted returns true if the upload was aborted.
func (u *Uploader) IsAborted() bool {
	u.lock.RLock()
	defer u.lock.RUnlock()
	return u.aborted
}

// Url returns the upload url.
func (u *Uploader) Url() string {
	return u.url
}

// Offset returns the current offset uploaded.
func (u *Uploader) Offset() int64 {
	return u.offset
}

// Upload uploads the entire body to the server.
func (u *Uploader) Upload() error {
	for u.offset < u.upload.size {
		if u.IsAborted() {
			return ErrUploadAborted
		}
		if err := u.UploadChunck(); err != nil {
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

	ctx, cancel := context.WithCancel(context.Background())
	u.cancel = cancel
	newOffset, err := u.client.uploadChunck(ctx, u.url, body, int64(size), u.offset)
	if err != nil {
		return err
	}

	u.offset = newOffset

	u.upload.updateProgress(u.offset)

	u.notifyChan <- true

	return nil
}

// Waits for a signal to broadcast to all subscribers
func (u *Uploader) broadcastProgress() {
	for _ = range u.notifyChan {
		for _, c := range u.uploadSubs {
			c <- *u.upload
		}
	}
}

// NewUploader creates a new Uploader.
func NewUploader(client *Client, url string, upload *Upload, offset int64) *Uploader {
	notifyChan := make(chan bool)

	uploader := &Uploader{
		client:     client,
		url:        url,
		upload:     upload,
		offset:     offset,
		notifyChan: notifyChan,
		cancel:     func() {},
	}

	go uploader.broadcastProgress()

	return uploader
}
