package main

import (
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"sort"
)

func LsTree(sha string) {
	if file, err := os.Open(fmt.Sprintf(".git/objects/%s/%s", sha[:2], sha[2:])); err == nil {
		if reader, err := zlib.NewReader(file); err == nil {
			if data, err := io.ReadAll(reader); err == nil {
				idx := findNull(data)
				data = data[idx+1:]

				readEntry := func(data []byte) ([]byte, string) {
					idx := findNull(data)

					entry := string(data[:idx])

					for i, ch := range entry {
						if ch == ' ' {
							entry = entry[i+1:]
							break
						}
					}

					return data[idx+21:], entry
				}

				entries := []string{}
				for len(data) != 0 {
					var entry string
					data, entry = readEntry(data)
					entries = append(entries, entry)
				}

				sort.Strings(entries)

				for _, entry := range entries {
					fmt.Println(entry)
				}
			}
		}
	}
}
