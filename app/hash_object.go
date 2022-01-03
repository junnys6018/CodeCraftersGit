package main

import (
	"bytes"
	"compress/zlib"
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
		err := os.MkdirAll(fmt.Sprintf(".git/objects/%x/", hash[:1]), 0755)
		if err != nil {
			fmt.Println(err)
		}

		compressed := bytes.Buffer{}
		writer := zlib.NewWriter(&compressed)
		writer.Write(blob.Bytes())
		writer.Close()

		err = os.WriteFile(fmt.Sprintf(".git/objects/%x/%x", hash[:1], hash[1:]), compressed.Bytes(), 0644)
		if err != nil {
			fmt.Println(err)
		}

		return hash
	}
	return [20]byte{}
}
