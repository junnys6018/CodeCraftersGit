package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"os"
)

func HashObject(path string) [20]byte {
	if data, err := os.ReadFile(path); err == nil {
		blob := bytes.Buffer{}

		fmt.Fprintf(&blob, "blob %d", len(data))
		blob.WriteByte(0)
		blob.Write(data)

		hash := sha1.Sum(blob.Bytes())

		// Write to disk
		WriteObject(hash, blob)

		return hash
	}
	return [20]byte{}
}
