package tus

import (
	"sync"
	"testing"
)

func TestUploaderAbort(t *testing.T) {
	tests := []struct {
		name  string
		count int
		want  bool
	}{
		{
			name:  "zero abort",
			count: 0,
			want:  false,
		},
		{
			name:  "one abort",
			count: 1,
			want:  true,
		},
		{
			name:  "ten abort",
			count: 10,
			want:  true,
		},
		{
			name:  "1000 aborts",
			count: 1000,
			want:  true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var upload Upload
			var url string
			var c Client
			u := NewUploader(&c, url, &upload, 0)
			var wg sync.WaitGroup
			wg.Add(tc.count)
			for i := 0; i < tc.count; i++ {
				go func() {
					defer wg.Done()
					u.Abort()
				}()
			}
			wg.Wait()
			if got, want := u.IsAborted(), tc.want; got != want {
				t.Fatalf("unexpected result:\n- want: %t\n-  got: %t",
					want, got)
			}
		})
	}
}
