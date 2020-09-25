package lpcorpus

import (
	"fmt"
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M, pkg string) int {
	corpusDir := os.Getenv("CORPUS")
	if corpusDir == "" {
		fmt.Fprintf(os.Stderr, "warning: no $CORPUS set; not writing line-protocol corpus files\n")
	}
	code := m.Run()
	if corpusDir != "" {
		if err := writeCorpusData(corpusDir, "inputs-decode-"+pkg, wrap(allDecodeInputs(), "inputs", "decode", pkg)); err != nil {
			log.Printf("failed to write corpus decode inputs: %v")
		}
		if err := writeCorpusData(corpusDir, "inputs-encode-"+pkg, wrap(allEncodeInputs(), "inputs", "encode", pkg)); err != nil {
			log.Printf("failed to write corpus encode inputs: %v")
		}
	}
	return code
}

func allDecodeInputs() []DecodeInput {
	mu.Lock()
	defer mu.Unlock()
	return append([]DecodeInput(nil), decodeInputs...)
}

func allEncodeInputs() []EncodeInput {
	mu.Lock()
	defer mu.Unlock()
	return append([]EncodeInput(nil), encodeInputs...)
}
