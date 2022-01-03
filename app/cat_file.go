package main

import (
	"compress/zlib"
	"fmt"
	"io"
	"os"
)

func CatFile(sha string) {
	if file, err := os.Open(fmt.Sprintf(".git/objects/%s/%s", sha[:2], sha[2:])); err == nil {
		if reader, err := zlib.NewReader(file); err == nil {
			if data, err := io.ReadAll(reader); err == nil {
				// blob {size}\0{content}
				idx := findNull(data)

				os.Stdout.Write(data[idx+1:])
			}
		}
	}
}
