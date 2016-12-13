// Package tus provides a client to tus protocol version 1.0.0.
//
// tus is a protocol based on HTTP for resumable file uploads. Resumable means that
// an upload can be interrupted at any moment and can be resumed without
// re-uploading the previous data again. An interruption may happen willingly, if
// the user wants to pause, or by accident in case of an network issue or server
// outage (http://tus.io).
package tus

import (
	"bytes"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
)

// Client represents the tus client.
// You can use it in goroutines to create parallels uploads.
type Client struct {
	config          *Config
	client          *http.Client
	aborted         bool
	filename        string
	url             string
	protocolVersion string
}

// NewClient creates a new tus client.
func NewClient(url, filename string, config *Config) (*Client, error) {
	fi, err := os.Stat(filename)

	if err != nil {
		return nil, err
	}

	if fi.IsDir() {
		return nil, fmt.Errorf("'%s' is a directory.", filename)
	}

	if config == nil {
		config = DefaultConfig()
	} else {
		if err = config.Validate(); err != nil {
			return nil, err
		}
	}

	return &Client{
		config:          config,
		client:          &http.Client{},
		aborted:         false,
		filename:        filename,
		url:             url,
		protocolVersion: "1.0.0",
	}, nil
}

// Upload start the uploading process.
// If resume is enabled and part of the file was uploaded, it will resume the upload.
func (c *Client) Upload() error {
	c.config.Logger.Printf("processing upload of '%s'.\n", c.filename)

	c.aborted = false

	f, err := os.Open(c.filename)

	if err != nil {
		return err
	}

	defer f.Close()

	defer func() {
		if c.config.Resume && !c.aborted {
			c.config.Storage.Delete(f)
		}
	}()

	if c.config.Resume {
		c.config.Logger.Printf("checking if can resume upload of '%s'.\n", c.filename)

		url, ok := c.config.Storage.Get(f)

		if ok {
			c.config.Logger.Printf("resuming upload of '%s'.\n", c.filename)

			offset, err := c.uploadOffset(f, url)

			if err != nil {
				return err
			}

			if offset != -1 {
				return c.upload(f, url, offset)
			}
		}
	}

	url, err := c.create(f)

	if err != nil {
		return err
	}

	c.config.Logger.Printf("starting upload of '%s'.\n", c.filename)

	if c.config.Resume {
		c.config.Storage.Set(f, url)
	}

	err = c.upload(f, url, 0)

	if err == nil {
		c.config.Logger.Printf("upload of '%s' completed.\n", c.filename)
	} else {
		c.config.Logger.Printf("upload of '%s' failed.\n", c.filename)
	}

	return err
}

// Abort stop the upload process.
// If resume is enabled you can continue the upload later.
func (c *Client) Abort() {
	c.config.Logger.Printf("aborting upload of '%s'.\n", c.filename)
	c.aborted = true
}

func (c *Client) create(f *os.File) (string, error) {
	fileInfo, err := f.Stat()

	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", c.url, nil)

	if err != nil {
		return "", fmt.Errorf("failed to create upload of '%s': %s", err)
	}

	req.Header.Set("Content-Length", "0")
	req.Header.Set("Upload-Length", strconv.FormatInt(fileInfo.Size(), 10))
	req.Header.Set("Tus-Resumable", c.protocolVersion)
	req.Header.Set("Upload-Metadata", fmt.Sprintf("filename %s", b64encode(fileInfo.Name())))

	res, err := c.client.Do(req)

	if err != nil {
		return "", fmt.Errorf("failed to create upload of '%s': %s", fileInfo.Name(), err)
	}

	switch res.StatusCode {
	case 201:
		return res.Header.Get("Location"), nil
	case 412:
		return "", fmt.Errorf("failed to create upload of '%s': this client is incompatible with Tus sever version %s.", fileInfo.Name(), res.Header.Get("Tus-Version"))
	case 413:
		return "", fmt.Errorf("failed to create upload of '%s': upload file is to large.", fileInfo.Name())
	default:
		return "", fmt.Errorf("failed to create upload of '%s': %d", fileInfo.Name(), res.StatusCode)
	}
}

func (c *Client) upload(f *os.File, url string, offset int64) error {
	fileInfo, err := f.Stat()

	if err != nil {
		return err
	}

	fileSize := fileInfo.Size()
	totalParts := math.Ceil(float64(fileSize) / float64(c.config.ChunkSize))

	for offset < fileSize && !c.aborted {
		currentPart := math.Ceil(float64(offset) / float64(c.config.ChunkSize))
		c.config.Logger.Printf("uploading file '%s' (%g/%g).\n", c.filename, currentPart+1, totalParts)

		_, err := f.Seek(offset, 0)

		if err != nil {
			return fmt.Errorf("failed to upload '%s': %s", fileInfo.Name(), err)
		}

		data := make([]byte, c.config.ChunkSize)
		size, err := f.Read(data)

		if err != nil {
			return fmt.Errorf("failed to upload '%s': %s", fileInfo.Name(), err)
		}

		method := "PATCH"

		if c.config.OverridePatchMethod {
			method = "POST"
		}

		req, err := http.NewRequest(method, url, bytes.NewBuffer(data[:size]))

		if err != nil {
			return fmt.Errorf("failed to upload '%s': %s", fileInfo.Name(), err)
		}

		req.Header.Set("Content-Type", "application/offset+octet-stream")
		req.Header.Set("Content-Length", strconv.Itoa(size))
		req.Header.Set("Upload-Offset", strconv.FormatInt(offset, 10))
		req.Header.Set("Tus-Resumable", c.protocolVersion)

		if c.config.OverridePatchMethod {
			req.Header.Set("X-HTTP-Method-Override", "PATCH")
		}

		res, err := c.client.Do(req)

		if err != nil {
			return fmt.Errorf("failed to upload '%s': %s", err)
		}

		switch res.StatusCode {
		case 204:
			offset, err = strconv.ParseInt(res.Header.Get("Upload-Offset"), 10, 64)

			if err != nil {
				return fmt.Errorf("failed to upload '%s': can't parse upload offset.", fileInfo.Name())
			}
		case 409:
			return fmt.Errorf("failed to upload '%s': upload offset doesn't match.", fileInfo.Name())
		case 412:
			return fmt.Errorf("failed to upload '%s': this client is incompatible with Tus server version %s.", fileInfo.Name(), res.Header.Get("Tus-Version"))
		case 413:
			return fmt.Errorf("failed to upload '%s': upload file is to large.", fileInfo.Name())
		default:
			return fmt.Errorf("failed to upload '%s': %d", fileInfo.Name(), res.StatusCode)
		}
	}

	return nil
}

func (c *Client) uploadOffset(f *os.File, url string) (int64, error) {
	fileInfo, err := f.Stat()

	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest("HEAD", url, nil)

	if err != nil {
		return 0, fmt.Errorf("failed to resume upload of '%s': %s", err)
	}

	req.Header.Set("Tus-Resumable", c.protocolVersion)

	res, err := c.client.Do(req)

	if err != nil {
		return 0, fmt.Errorf("failed to resume upload of '%s': %s", fileInfo.Name(), err)
	}

	switch res.StatusCode {
	case 200:
		i, err := strconv.ParseInt(res.Header.Get("Upload-Offset"), 10, 64)

		if err == nil {
			return i, nil
		} else {
			return 0, fmt.Errorf("failed to resume upload of '%s': can't parse upload offset.", fileInfo.Name())
		}
	case 403, 404, 410:
		// file doesn't exists.
		return -1, nil
	case 412:
		return 0, fmt.Errorf("failed to resume upload of '%s': this client is incompatible with Tus server version %s.", fileInfo.Name(), res.Header.Get("Tus-Version"))
	default:
		return 0, fmt.Errorf("failed to resume upload of '%s': %d", fileInfo.Name(), res.StatusCode)
	}
}
