// +build ignore

package main

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/rogpeppe/line-protocol-corpus/lpcorpus"
)

func main() {
	corpus, err := lpcorpus.ReadCorpus("../..")
	if err != nil {
		log.Fatal(err)
	}
	inputs := make(map[string]bool)
	for _, input := range corpus.Decode {
		inputs[string(input.Text)] = true
	}
	if err := os.MkdirAll("corpus", 0777); err != nil {
		log.Fatal(err)
	}
	for data := range inputs {
		if err := ioutil.WriteFile(filepath.Join("corpus", fmt.Sprintf("%x", md5.Sum([]byte(data)))), []byte(data), 0666); err != nil {
			log.Fatal(err)
		}
	}
}
