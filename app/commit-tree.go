package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
)

func CommitTree(treeSha, commitSha, message string) [20]byte {
	blob := bytes.Buffer{}

	fmt.Fprintf(&blob, "tree %s\n", treeSha)
	fmt.Fprintf(&blob, "parent %s\n", commitSha)
	fmt.Fprint(&blob, "author Jun Lim <jun@gmail.com> 1243040974 -0700\n")
	fmt.Fprint(&blob, "committer Jun Lim <jun@gmail.com> 1243040974 -0700\n\n")
	fmt.Fprintf(&blob, "%s\n", message)

	newBlob := bytes.Buffer{}

	fmt.Fprintf(&newBlob, "commit %d", blob.Len())
	newBlob.WriteByte(0)
	newBlob.Write(blob.Bytes())

	hash := sha1.Sum(newBlob.Bytes())

	WriteObject(hash, newBlob)

	return hash
}
