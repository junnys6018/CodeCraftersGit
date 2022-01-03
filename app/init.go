package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

func Init() {
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
}
