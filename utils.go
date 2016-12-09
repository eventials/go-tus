package tus

import (
	"encoding/base64"
)

func b64encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}
