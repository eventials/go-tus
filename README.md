# go-tus [![Build Status](https://travis-ci.org/eventials/go-tus.svg?branch=master)](https://travis-ci.org/eventials/go-tus) [![Go Report Card](https://goreportcard.com/badge/github.com/eventials/go-tus)](https://goreportcard.com/report/github.com/eventials/go-tus) [![GoDoc](https://godoc.org/github.com/eventials/go-tus?status.svg)](http://godoc.org/github.com/eventials/go-tus)

A pure Go client for the [tus resumable upload protocol](http://tus.io/)

## Example

```go
package main

import (
    "os"
    "github.com/eventials/go-tus"
)

func main() {
    f, err := os.Open("my-file.txt")

    if err != nil {
        panic(err)
    }

    defer f.Close()

    // create the tus client.
    client, _ := tus.NewClient("https://tus.example.org/files", nil)

    // create an upload from a file.
    upload, _ := tus.NewUploadFromFile(f)

    // create the uploader.
    uploader, _ := client.CreateUpload(upload)

    // start the uploading process.
    uploader.Upload()
}
```

## Example with resume

```go
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/eventials/go-tus"
	"github.com/eventials/go-tus/memorystore"
)

func main() {
	// Open the file
	file, err := os.Open(`/path/to/your/file.txt`)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	// Create a MemoryStore also can be LeveldbStore or you can implement your own store
	store, err := memorystore.NewMemoryStore()
	if err != nil {
		panic(err)
	}
	// Create a Client
	client, err := tus.NewClient("http://tus.example.org/files/", &tus.Config{
		Resume:    true,    // Important to resume uploads
		Store:     store,   // Important to resume uploads
		ChunkSize: 4 << 20, // 4 Mb
	})
	if err != nil {
		panic(err)
	}
	// (Optional) Create a chan to notify upload status
	uploadChan := make(chan tus.Upload, 1)
	go func() {
		for uploadStatus := range uploadChan {
			// Print the upload status
			fmt.Printf("Completed %v%% %v Bytes of %v Bytes\n",
				uploadStatus.Progress(),
				uploadStatus.Offset(),
				uploadStatus.Size())
		}
	}()
	// Create new upload
	upload, err := tus.NewUploadFromFile(file)
	if err != nil {
		panic(err)
	}
	// Declare number of attempts
	const attemps = 50
	for i := 1; i <= attemps; i++ {
		fmt.Printf("Attemp %v of %v\n", i, attemps)
		// Create an uploader
		uploader, err := client.CreateOrResumeUpload(upload)
		if err != nil {
			fmt.Println("Error", err)
			fmt.Println("Trying again in 10 seg")
			time.Sleep(time.Second * 10)
			continue
		}
		// (Optional) Notify Upload Status
		uploader.NotifyUploadProgress(uploadChan)
		// Start upload to server
		err = uploader.Upload()
		if err != nil {
			fmt.Println("Error", err)
			fmt.Println("Trying again in 10 seg")
			time.Sleep(time.Second * 10)
			continue
		}
		break
	}
	// If after all attemps there's an error panic!
	if err != nil {
		panic(err)
	}
	fmt.Println("Finished!")
}
```

## Features

> This is not a full protocol client implementation.

Checksum, Termination and Concatenation extensions are not implemented yet.

This client allows to resume an upload if a Store is used.

## Built in Store

Store is used to map an upload's fingerprint with the corresponding upload URL.

| Name | Backend | Dependencies |
|:----:|:-------:|:------------:|
| MemoryStore  | In-Memory | None |
| LeveldbStore | LevelDB   | [goleveldb](https://github.com/syndtr/goleveldb) |

## Future Work

- [ ] SQLite store
- [ ] Redis store
- [ ] Memcached store
- [ ] Checksum extension
- [ ] Termination extension
- [ ] Concatenation extension
