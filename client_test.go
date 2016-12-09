package tus

import (
    "fmt"
    "net/http"
    "net/http/httptest"
    "os"
    "sync"
    "testing"
    "time"

    "github.com/stretchr/testify/suite"
    "github.com/tus/tusd"
    "github.com/tus/tusd/filestore"
)

type ClientTestSuite struct {
    suite.Suite

    ts *httptest.Server
    url string
}

func (s *ClientTestSuite) SetupSuite() {
    store := filestore.FileStore{
        Path: os.TempDir(),
    }

    composer := tusd.NewStoreComposer()

    store.UseIn(composer)

    handler, err := tusd.NewHandler(tusd.Config{
        BasePath:      "/uploads/",
        StoreComposer: composer,
        MaxSize: 0,
        NotifyCompleteUploads:   false,
        NotifyTerminatedUploads: false,
        RespectForwardedHeaders: true,
    })

    if err != nil {
        panic(err)
    }

    s.ts = httptest.NewServer(http.StripPrefix("/uploads/", handler))
    s.url = fmt.Sprintf("%s/uploads/", s.ts.URL)
}

func (s *ClientTestSuite) TearDownSuite() {
    s.ts.Close()
}

func (s *ClientTestSuite) TestSmallUpload() {
    file := fmt.Sprintf("%s/%d", os.TempDir(), time.Now().Unix())

    f, err := os.Create(file)
    s.Nil(err)

    err = f.Truncate(1048576) // 1 MB
    s.Nil(err)

    c, err := NewClient(s.url, file, nil)
    s.Nil(err)

    err = c.Upload()
    s.Nil(err)
}

func (s *ClientTestSuite) TestLargeUpload() {
    file := fmt.Sprintf("%s/%d", os.TempDir(), time.Now().Unix())

    f, err := os.Create(file)
    s.Nil(err)

    err = f.Truncate(1048576 * 150) // 150 MB
    s.Nil(err)

    c, err := NewClient(s.url, file, nil)
    s.Nil(err)

    err = c.Upload()
    s.Nil(err)
}

func (s *ClientTestSuite) TestOverridePatchMethod() {
    file := fmt.Sprintf("%s/%d", os.TempDir(), time.Now().Unix())

    f, err := os.Create(file)
    s.Nil(err)

    err = f.Truncate(1048576) // 1 MB
    s.Nil(err)

    conf := DefaultConfig()
    conf.OverridePatchMethod = true

    c, err := NewClient(s.url, file, conf)
    s.Nil(err)

    err = c.Upload()
    s.Nil(err)
}

func (s *ClientTestSuite) TestConcurrentUploads() {
    var wg sync.WaitGroup

    file := fmt.Sprintf("%s/%d", os.TempDir(), time.Now().Unix())

    f, err := os.Create(file)
    s.Nil(err)

    err = f.Truncate(1048576) // 1 MB
    s.Nil(err)

    conf := DefaultConfig()
    conf.Resume = false

    for i := 0; i < 10; i ++ {
        wg.Add(1)

        go func() {
            defer wg.Done()

            c, err := NewClient(s.url, file, conf)
            s.Nil(err)

            err = c.Upload()
            s.Nil(err)
        }()
    }

    wg.Wait()
}

func (s *ClientTestSuite) TestResumeUpload() {
    file := fmt.Sprintf("%s/%d", os.TempDir(), time.Now().Unix())

    f, err := os.Create(file)
    s.Nil(err)

    err = f.Truncate(1048576 * 150) // 150 MB
    s.Nil(err)

    c, err := NewClient(s.url, file, nil)
    s.Nil(err)

    // This will stop the first upload.
    go func() {
        time.Sleep(250 * time.Millisecond)
        c.Abort()
    }()

    err = c.Upload()
    s.Nil(err)

    // This should resume upload.
    err = c.Upload()
    s.Nil(err)
}

func (s *ClientTestSuite) TestUploadDir() {
    c, err := NewClient(s.url, os.TempDir(), nil)
    s.Nil(c)
    s.NotNil(err)
}

func (s *ClientTestSuite) TestUploadInvalidFile() {
    c, err := NewClient(s.url, "randomfile.txt", nil)
    s.Nil(c)
    s.NotNil(err)
}

func TestClientTestSuite(t *testing.T) {
    suite.Run(t, new(ClientTestSuite))
}
