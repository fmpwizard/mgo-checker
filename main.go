package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	filepath.Walk("./seeddata", walk)
}

func walk(path string, info os.FileInfo, err error) error {
	fmt.Println("reading: ", path)
	if strings.HasSuffix(info.Name(), ".go") {
		findDirective(path)
	}
	return err
}

func findDirective(fpath string) error {
	f, err := os.Open(fpath)
	if err != nil {
		return err
	}
	input := bufio.NewReader(f)
	for {
		var buf []byte
		buf, err = input.ReadSlice('\n')
		if err == bufio.ErrBufferFull {
			return bufio.ErrBufferFull
		}
		if err != nil {
			// Check for marker at EOF without final \n.
			if err == io.EOF && isMGoDirective(buf) {
				err = io.ErrUnexpectedEOF
				return err
			}
			break
		}
		if isMGoDirective(buf) {
			fmt.Println(string(buf))
		}
	}

	return nil
}

func isMGoDirective(in []byte) bool {
	return bytes.HasPrefix(in, []byte("//mgo:model"))
}
