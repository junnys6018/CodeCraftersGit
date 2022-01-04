package main

import "fmt"

func Clone(url, dir string) {
	fmt.Println(url, dir)
	// response, err := http.Get("http://github.com/junnys6018/Emulator-Hub.git/info/refs?service=git-upload-pack")
	// if err != nil {
	// 	panic(err)
	// }

	// io.Copy(os.Stdout, response.Body)

	// buffer := bytes.NewBufferString("0032want 573b5511c9eb2218ccba8b069c984d22741bcc71\n0000")
	// response, err = http.Post("http://github.com/junnys6018/Emulator-Hub.git/git-upload-pack", "application/x-git-upload-pack-request", buffer)
	// if err != nil {
	// 	panic(err)
	// }

	// io.Copy(os.Stdout, response.Body)
	// fmt.Println(response)
}
