# go-tus [![Build Status](https://travis-ci.org/eventials/go-tus.svg?branch=master)](https://travis-ci.org/eventials/go-tus) [![Go Report Card](https://goreportcard.com/badge/github.com/eventials/go-tus)](https://goreportcard.com/report/github.com/eventials/go-tus) [![GoDoc](https://godoc.org/github.com/eventials/go-tus?status.svg)](http://godoc.org/github.com/eventials/go-tus)

A pure Go client for the [tus resumable upload protocol](http://tus.io/)

## Example

```go
c, err := NewClient("http://localhost:1080/files/", "/videos/my-video.mp4", nil)

if err != nil {
    panic(err)
}

err = c.Upload()

if err != nil {
    panic(err)
}
```

## Features

> This is not a full protocol client implementation.

Checksum, Termination and Concatenation extensions are not implemented yet.

This client allows to resume an upload if a Storage is used.

## Built in Storages

Storages are used to save the progress of an upload.

| Name | Backend | Dependencies |
|:----:|:-------:|:------------:|
| MemoryStorage | In-Memory | None |

## Future Work

- [ ] SQLite storage
- [ ] Redis storage
- [ ] Memcached storage
- [ ] Checksum extension
- [ ] Termination extension
- [ ] Concatenation extension
