package util

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"hash"
	"io"
	"io/ioutil"
	"os"
)

// UnzipBytes decode zip bytes
func UnzipBytes(data []byte) ([]byte, error) {
	b := bytes.NewReader(data)
	z, err := gzip.NewReader(b)
	if err != nil {
		return nil, err
	}
	defer z.Close()
	p, err := ioutil.ReadAll(z)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func MD5File(f *os.File) (hash.Hash, error) {
	h := md5.New()
	_, err := io.Copy(h, f)
	if err != nil {
		return nil, err
	}
	return h, nil
}

func MD5String(data []byte) string {
	h := md5.New()
	_, err := h.Write(data)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(h.Sum(nil))
}
