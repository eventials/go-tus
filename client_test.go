package tus

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/tus/tusd/pkg/filestore"
	tusd "github.com/tus/tusd/pkg/handler"
)

type MockStore struct {
	m map[string]string
}

func NewMockStore() Store {
	return &MockStore{
		make(map[string]string),
	}
}

func (s *MockStore) Get(fingerprint string) (string, bool) {
	url, ok := s.m[fingerprint]
	return url, ok
}

func (s *MockStore) Set(fingerprint, url string) {
	s.m[fingerprint] = url
}

func (s *MockStore) Delete(fingerprint string) {
	delete(s.m, fingerprint)
}

func (s *MockStore) Close() {
	for k := range s.m {
		delete(s.m, k)
	}
}

type UploadTestSuite struct {
	suite.Suite

	ts    *httptest.Server
	store filestore.FileStore
	url   string
}

func (s *UploadTestSuite) SetupSuite() {
	store := filestore.FileStore{
		Path: os.TempDir(),
	}

	composer := tusd.NewStoreComposer()

	store.UseIn(composer)

	handler, err := tusd.NewHandler(tusd.Config{
		BasePath:                "/uploads/",
		StoreComposer:           composer,
		MaxSize:                 0,
		NotifyCompleteUploads:   false,
		NotifyTerminatedUploads: false,
		RespectForwardedHeaders: true,
	})

	if err != nil {
		panic(err)
	}

	s.store = store
	s.ts = httptest.NewServer(http.StripPrefix("/uploads/", handler))
	s.url = fmt.Sprintf("%s/uploads/", s.ts.URL)
}

func (s *UploadTestSuite) TearDownSuite() {
	s.ts.Close()
}

func (s *UploadTestSuite) TestSmallUploadFromFile() {
	file := fmt.Sprintf("%s/%d", os.TempDir(), time.Now().Unix())

	f, err := os.Create(file)
	s.Nil(err)

	defer f.Close()

	err = f.Truncate(1048576) // 1 MB
	s.Nil(err)

	client, err := NewClient(s.url, nil)
	s.Nil(err)

	upload, err := NewUploadFromFile(f)
	s.Nil(err)

	uploader, err := client.CreateUpload(upload)
	s.Nil(err)
	s.NotNil(uploader)

	err = uploader.Upload()
	s.Nil(err)

	up, err := s.store.GetUpload(context.Background(), uploadIdFromUrl(uploader.url))
	s.Nil(err)
	fi, err := up.GetInfo(context.Background())
	s.Nil(err)

	s.EqualValues(1048576, fi.Size)
}

func (s *UploadTestSuite) TestLargeUpload() {
	file := fmt.Sprintf("%s/%d", os.TempDir(), time.Now().Unix())

	f, err := os.Create(file)
	s.Nil(err)

	defer f.Close()

	err = f.Truncate(1048576 * 150) // 150 MB
	s.Nil(err)

	client, err := NewClient(s.url, nil)
	s.Nil(err)

	upload, err := NewUploadFromFile(f)
	s.Nil(err)

	uploader, err := client.CreateUpload(upload)
	s.Nil(err)
	s.NotNil(uploader)

	err = uploader.Upload()
	s.Nil(err)

	up, err := s.store.GetUpload(context.Background(), uploadIdFromUrl(uploader.url))
	s.Nil(err)
	fi, err := up.GetInfo(context.Background())
	s.Nil(err)

	s.EqualValues(1048576*150, fi.Size)
}

func (s *UploadTestSuite) TestUploadFromBytes() {
	client, err := NewClient(s.url, nil)
	s.Nil(err)

	upload := NewUploadFromBytes([]byte("1234567890"))
	s.Nil(err)

	uploader, err := client.CreateUpload(upload)
	s.Nil(err)
	s.NotNil(uploader)

	err = uploader.Upload()
	s.Nil(err)

	up, err := s.store.GetUpload(context.Background(), uploadIdFromUrl(uploader.url))
	s.Nil(err)
	fi, err := up.GetInfo(context.Background())
	s.Nil(err)

	s.EqualValues(10, fi.Size)
}

func (s *UploadTestSuite) TestOverridePatchMethod() {
	client, err := NewClient(s.url, nil)
	s.Nil(err)

	client.Config.OverridePatchMethod = true

	upload := NewUploadFromBytes([]byte("1234567890"))
	s.Nil(err)

	uploader, err := client.CreateUpload(upload)
	s.Nil(err)
	s.NotNil(uploader)

	err = uploader.Upload()
	s.Nil(err)

	up, err := s.store.GetUpload(context.Background(), uploadIdFromUrl(uploader.url))
	s.Nil(err)
	fi, err := up.GetInfo(context.Background())
	s.Nil(err)

	s.EqualValues(10, fi.Size)
}

func (s *UploadTestSuite) TestConcurrentUploads() {
	var wg sync.WaitGroup

	client, err := NewClient(s.url, nil)
	s.Nil(err)

	for i := 0; i < 20; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			file := fmt.Sprintf("%s/%d", os.TempDir(), time.Now().UnixNano())

			f, err := os.Create(file)
			s.Nil(err)

			defer f.Close()

			err = f.Truncate(1048576 * 5) // 5 MB
			s.Nil(err)

			upload, err := NewUploadFromFile(f)
			s.Nil(err)

			uploader, err := client.CreateUpload(upload)
			s.Nil(err)
			s.NotNil(uploader)

			err = uploader.Upload()
			s.Nil(err)

			up, err := s.store.GetUpload(context.Background(), uploadIdFromUrl(uploader.url))
			s.Nil(err)
			fi, err := up.GetInfo(context.Background())
			s.Nil(err)

			s.EqualValues(1048576*5, fi.Size)
		}()
	}

	wg.Wait()
}

func (s *UploadTestSuite) TestResumeUpload() {
	file := fmt.Sprintf("%s/%d", os.TempDir(), time.Now().Unix())

	f, err := os.Create(file)
	s.Nil(err)

	defer f.Close()

	err = f.Truncate(1048576 * 150) // 150 MB
	s.Nil(err)

	cfg := &Config{
		ChunkSize:           2 * 1024 * 1024,
		Resume:              true,
		OverridePatchMethod: false,
		Store:               NewMockStore(),
		Header: map[string][]string{
			"X-Extra-Header": []string{"somevalue"},
		},
	}

	client, err := NewClient(s.url, cfg)
	s.Nil(err)

	upload, err := NewUploadFromFile(f)
	s.Nil(err)

	uploader, err := client.CreateUpload(upload)
	s.Nil(err)
	s.NotNil(uploader)

	// This will stop the first upload.
	go func() {
		time.Sleep(250 * time.Millisecond)
		uploader.Abort()
	}()

	err = uploader.Upload()
	s.Nil(err)

	s.True(uploader.aborted)

	uploader, err = client.ResumeUpload(upload)
	s.Nil(err)
	s.NotNil(uploader)

	err = uploader.Upload()
	s.Nil(err)

	up, err := s.store.GetUpload(context.Background(), uploadIdFromUrl(uploader.url))
	s.Nil(err)
	fi, err := up.GetInfo(context.Background())
	s.Nil(err)

	s.EqualValues(1048576*150, fi.Size)
}

func TestUploadTestSuite(t *testing.T) {
	suite.Run(t, new(UploadTestSuite))
}

func uploadIdFromUrl(url string) string {
	parts := strings.Split(url, "/")
	return parts[len(parts)-1]
}
