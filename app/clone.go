// TODO: we fully buffer the response body before doing any parsing, stream in the response instead
package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

// utility

func readUint32BigEndian(bytes []byte) uint32 {
	return uint32(bytes[0])<<24 | uint32(bytes[1])<<16 | uint32(bytes[2])<<8 | uint32(bytes[3])
}

func readPktLine(blob []byte) (int, []byte, error) {
	pktLength := blob[:4]
	blob = blob[4:]

	dst := [2]byte{}
	_, err := hex.Decode(dst[:], pktLength)
	if err != nil {
		return 0, nil, err
	}

	size := uint16(dst[0])<<8 | uint16(dst[1])

	// pkt-line 0000
	if size == 0 {
		return 4, []byte{}, nil
	}

	if len(blob) < int(size)-4 {
		return 4, nil, errors.New("Error reading pkt line")
	}

	data := blob[:size-4]

	// strip trailing linefeed, if it exists
	if data[len(data)-1] == '\n' {
		data = data[:len(data)-1]
	}
	return int(size), data, nil
}

func getObjectName(pktLines [][]byte) (string, error) {
	// skip the first pktLine (001e# service=git-upload-pack)
	for _, pktLine := range pktLines[1:] {
		if len(pktLine) == 0 {
			continue
		}

		var hash, ref string
		fmt.Sscanf(string(pktLine), "%s %s", &hash, &ref)
		if ref == "refs/heads/master" {
			return hash, nil
		}
	}
	return "", errors.New("Invalid pktLines")
}

func getPackfile(cloneUrl string) ([]byte, error) {
	response, err := http.Get(fmt.Sprintf("%s/info/refs?service=git-upload-pack", cloneUrl))
	if err != nil {
		return nil, err
	}

	discoveryBuffer := bytes.Buffer{}
	io.Copy(&discoveryBuffer, response.Body)
	discovery := discoveryBuffer.Bytes()

	pktLines := [][]byte{}

	for len(discovery) > 0 {
		n, data, err := readPktLine(discovery)
		if err != nil {
			return nil, err
		}

		discovery = discovery[n:]

		pktLines = append(pktLines, data)
	}

	objectName, err := getObjectName(pktLines)
	if err != nil {
		return nil, err
	}

	buffer := bytes.NewBufferString(fmt.Sprintf("0032want %s\n00000009done\n", objectName))
	response, err = http.Post(fmt.Sprintf("%s/git-upload-pack", cloneUrl), "application/x-git-upload-pack-request", buffer)
	if err != nil {
		return nil, err
	}

	packfileBuffer := bytes.Buffer{}
	io.Copy(&packfileBuffer, response.Body)
	packfile := packfileBuffer.Bytes()
	n, _, err := readPktLine(packfile) // read 0008NAK
	if err != nil {
		return nil, err
	}

	packfile = packfile[n:]
	return packfile, nil
}

type ObjectType int

const (
	OBJ_COMMIT    ObjectType = 1
	OBJ_TREE                 = 2
	OBJ_BLOB                 = 3
	OBJ_TAG                  = 4
	OBJ_OFS_DELTA            = 6
	OBJ_REF_DELTA            = 7
)

func writeObjectWithType(object []byte, objectType string) error {
	blob := bytes.Buffer{}

	fmt.Fprintf(&blob, "%s %d", objectType, len(object))
	blob.WriteByte(0)
	blob.Write(object)

	hash := sha1.Sum(blob.Bytes())

	// Write to disk
	err := WriteObject(hash, blob)
	return err
}

func readObjectHeader(packfile []byte) (size uint64, objectType ObjectType, used int, err error) {
	data := packfile[used]
	used++

	objectType = ObjectType((data >> 4) & 0x7)
	size = uint64(data & 0xF)
	shift := 4

	for data&0x80 != 0 {
		if len(packfile) <= used || 64 <= shift {
			return 0, ObjectType(0), 0, errors.New("Bad object header")
		}

		data = packfile[used]
		used++

		size += uint64(data&0x7F) << shift
		shift += 7
	}
	return size, objectType, used, nil
}

func readSize(packfile []byte) (size uint64, used int, err error) {
	data := packfile[used]
	used++

	size = uint64(data & 0x7F)
	shift := 7

	for data&0x80 != 0 {
		if len(packfile) <= used || 64 <= shift {
			return 0, 0, errors.New("Bad size")
		}

		data = packfile[used]
		used++

		size += uint64(data&0x7F) << shift
		shift += 7
	}
	return size, used, nil
}

func readObject(packfile []byte) (int, []byte, error) {
	b := bytes.NewReader(packfile)
	r, err := zlib.NewReader(b)
	if err != nil {
		return 0, nil, err
	}

	object, err := io.ReadAll(r)
	if err != nil {
		return 0, nil, err
	}

	bytesRead := int(b.Size()) - b.Len()
	return bytesRead, object, nil
}

func writePackfile(packfile []byte, dir string) error {
	head := packfile[:len(packfile)-20]

	// verify packfile
	if len(packfile) < 32 {
		return errors.New("Bad packfile")
	}

	checksum := packfile[len(packfile)-20:]
	expected := sha1.Sum(packfile[:len(packfile)-20])
	if bytes.Compare(checksum, expected[:]) != 0 {
		return errors.New("Invalid packfile checksum")
	}

	if bytes.Compare(head[:4], []byte("PACK")) != 0 {
		return errors.New("Invalid packfile header")
	}
	head = head[4:]

	version := readUint32BigEndian(head)
	head = head[4:]

	if version != 2 && version != 3 {
		return errors.New("Invalid packfile version")
	}

	numObjects := readUint32BigEndian(head)
	head = head[4:]

	var objectsRead uint32
	for len(head) > 0 {
		objectsRead++

		size, objectType, used, err := readObjectHeader(head)
		head = head[used:]

		if err != nil {
			return err
		}

		switch objectType {
		case OBJ_COMMIT:
			read, object, err := readObject(head)
			if err != nil {
				return err
			}

			if int(size) != len(object) {
				return errors.New("Bad object header length")
			}

			head = head[read:]

			err = writeObjectWithType(object, "commit")
			if err != nil {
				return err
			}

		case OBJ_TREE:
			read, object, err := readObject(head)
			if err != nil {
				return err
			}

			if int(size) != len(object) {
				return errors.New("Bad object header length")
			}

			head = head[read:]

			err = writeObjectWithType(object, "tree")
			if err != nil {
				return err
			}

		case OBJ_BLOB:
			read, object, err := readObject(head)
			if err != nil {
				return err
			}

			if int(size) != len(object) {
				return errors.New("Bad object header length")
			}

			head = head[read:]

			err = writeObjectWithType(object, "blob")
			if err != nil {
				return err
			}

		case OBJ_TAG:
			read, object, err := readObject(head)
			if err != nil {
				return err
			}

			if int(size) != len(object) {
				return errors.New("Bad object header length")
			}

			head = head[read:]

			err = writeObjectWithType(object, "tag")
			if err != nil {
				return err
			}

		case OBJ_OFS_DELTA:
			// read offset
			_, used, err := readSize(head)
			if err != nil {
				return err
			}
			head = head[used:]

			read, object, err := readObject(head)
			if err != nil {
				return err
			}

			if int(size) != len(object) {
				return errors.New("Bad object header length")
			}

			head = head[read:]

		case OBJ_REF_DELTA:
			fmt.Println("refdelta")

			// hash := head[:20]
			head = head[20:]
			read, object, err := readObject(head)
			if err != nil {
				return err
			}

			if int(size) != len(object) {
				return errors.New("Bad object header length")
			}

			head = head[read:]

		}
	}

	if numObjects != objectsRead {
		return errors.New("Bad object count")
	}

	return nil
}

func Clone(cloneUrl, dir string) {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		panic(err)
	}

	err = os.Chdir(dir)
	if err != nil {
		panic(err)
	}

	Init()

	packfile, err := getPackfile(cloneUrl)
	if err != nil {
		panic(err)
	}
	err = writePackfile(packfile, dir)
	if err != nil {
		panic(err)
	}
}
