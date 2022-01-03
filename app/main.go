package main

import (
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

// Usage: your_git.sh run <image> <command> <arg1> <arg2> ...
// Assumes command is run in the root of the git repo... not hard to fix but cbf
func main() {
	switch command := os.Args[1]; command {
	case "init":
		Init()

	case "cat-file":
		sha := os.Args[3]
		CatFile(sha)

	case "hash-object":
		path := os.Args[3]
		fmt.Println(HashObject(path))

	case "ls-tree":
		sha := os.Args[3]
		LsTree(sha)

	case "write-tree":
		fmt.Println(WriteTree("."))

	default:
		fmt.Printf("Unknown command %s\n", command)
		os.Exit(1)
	}
}
