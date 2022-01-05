package main

import (
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"sort"
)

func readTreeEntry(data []byte) (used int, mode string, name string, hash [20]byte) {
	idx := findNull(data)

	entry := string(data[:idx])
	fmt.Sscanf(entry, "%s %s", &mode, &name)

	copy(hash[:], data[idx+1:idx+21])

	return idx + 21, mode, name, hash
}

func LsTree(sha string) {
	if file, err := os.Open(fmt.Sprintf(".git/objects/%s/%s", sha[:2], sha[2:])); err == nil {
		if reader, err := zlib.NewReader(file); err == nil {
			if data, err := io.ReadAll(reader); err == nil {
				idx := findNull(data)
				data = data[idx+1:]

				entries := []string{}
				for len(data) != 0 {
					used, _, name, _ := readTreeEntry(data)
					data = data[used:]
					entries = append(entries, name)
				}

				sort.Strings(entries)

				for _, entry := range entries {
					fmt.Println(entry)
				}
			}
		}
	}
}
