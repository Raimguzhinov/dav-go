package etag

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"io"
)

func FromData(data []byte) (string, error) {
	h := sha1.New()
	if _, err := io.Copy(h, bytes.NewReader(data)); err != nil {
		return "", err
	}
	csum := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(csum[:]), nil
}
