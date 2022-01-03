package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"os"
	"sort"
)

type TreeEntry struct {
	mode string
	name string
	hash [20]byte
}

func (te TreeEntry) String() string {
	return fmt.Sprintf("%s %s %x", te.mode, te.name, te.hash)
}

type TreeEntrys []TreeEntry

func (te TreeEntrys) Len() int {
	return len(te)
}

func (te TreeEntrys) Less(i, j int) bool {
	return te[i].name < te[j].name
}

func (te TreeEntrys) Swap(i, j int) {
	te[i], te[j] = te[j], te[i]
}

func (te TreeEntrys) SerialisedSize() (size int) {
	for _, entry := range te {
		size += len(entry.mode)
		size += 1
		size += len(entry.name)
		size += 1
		size += 20
	}
	return
}

func WriteTree(path string) [20]byte {
	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	treeEntries := make(TreeEntrys, 0)

	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != ".git" {
			hash := WriteTree(fmt.Sprintf("%s/%s", path, entry.Name()))
			treeEntries = append(treeEntries, TreeEntry{"40000", entry.Name(), hash})

		} else if !entry.IsDir() {
			hash := HashObject(fmt.Sprintf("%s/%s", path, entry.Name()))
			treeEntries = append(treeEntries, TreeEntry{"100644", entry.Name(), hash})
		}
	}

	// Serialize
	sort.Sort(treeEntries)

	blob := bytes.Buffer{}
	fmt.Fprintf(&blob, "tree %d", treeEntries.SerialisedSize())
	blob.WriteByte(0)

	for _, te := range treeEntries {
		fmt.Fprintf(&blob, "%s %s", te.mode, te.name)
		blob.WriteByte(0)
		blob.Write(te.hash[:])
	}

	hash := sha1.Sum(blob.Bytes())

	// Write to disk
	err = os.MkdirAll(fmt.Sprintf(".git/objects/%x/", hash[:1]), 0755)
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
