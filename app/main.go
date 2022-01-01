package main

import (
	"compress/zlib"
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
	default:
		fmt.Printf("Unknown command %s\n", command)
		os.Exit(1)
	}
}
