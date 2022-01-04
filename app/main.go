package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"os"
)

func findNull(bytes []byte) int {
	for i, val := range bytes {
		if val == 0 {
			return i
		}
	}
	return len(bytes)
}

func WriteObject(hash [20]byte, blob bytes.Buffer) {
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
}

// Usage: your_git.sh run <image> <command> <arg1> <arg2> ...
// Assumes command is run in the root of the git repo... not hard to fix but cbf
// Only implements the happy path, does not handle io errors etc...
func main() {
	switch command := os.Args[1]; command {
	case "init":
		Init()

	case "cat-file":
		sha := os.Args[3]
		CatFile(sha)

	case "hash-object":
		path := os.Args[3]
		fmt.Printf("%x\n", HashObject(path))

	case "ls-tree":
		sha := os.Args[3]
		LsTree(sha)

	case "write-tree":
		fmt.Printf("%x\n", WriteTree("."))

	case "commit-tree":
		treeSha := os.Args[2]
		commitSha := os.Args[4]
		message := os.Args[6]
		fmt.Printf("%x\n", CommitTree(treeSha, commitSha, message))

	case "clone":
		url := os.Args[2]
		dir := os.Args[3]
		Clone(url, dir)

	default:
		fmt.Printf("Unknown command %s\n", command)
		os.Exit(1)
	}
}
