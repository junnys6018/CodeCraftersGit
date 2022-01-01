package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

// Usage: your_git.sh run <image> <command> <arg1> <arg2> ...
func main() {
	switch command := os.Args[1]; command {
	case "init":
		for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
			if err := os.Mkdir(dir, 0755); err != nil {
				fmt.Printf("Error creating directory: %s\n", err)
			}
		}
		headFileContents := []byte("ref: refs/heads/master\n")
		if err := ioutil.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
			fmt.Printf("Error writing file: %s\n", err)
		}
		fmt.Println("Initialized git directory")

	case "cat-file":
		sha := os.Args[3]

		if file, err := os.Open(fmt.Sprintf(".git/objects/%s/%s", sha[:2], sha[2:])); err == nil {
			if reader, err := zlib.NewReader(file); err == nil {
				if data, err := io.ReadAll(reader); err == nil {
					// blob {size}\0{content}
					idx := 0
					for i, val := range data {
						if val == 0 {
							idx = i
							break
						}
					}

					os.Stdout.Write(data[idx+1:])
				}
			}
		}

	case "hash-object":
		path := os.Args[3]

		if data, err := os.ReadFile(path); err == nil {
			blob := bytes.Buffer{}

			fmt.Fprintf(&blob, "blob %d", len(data))
			blob.WriteByte(0)
			blob.Write(data)

			hash := sha1.Sum(blob.Bytes())

			// Write to disk
			err := os.Mkdir(fmt.Sprintf(".git/objects/%x/", hash[:1]), 0755)
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

			fmt.Printf("%x", hash)
		}

	default:
		fmt.Printf("Unknown command %s\n", command)
		os.Exit(1)
	}
}
