package lpcorpus

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"

	"gopkg.in/yaml.v3"
)

type Corpus struct {
	Encode map[string]*EncodeInput `yaml:"encode"`
	Decode map[string]*DecodeInput `yaml:"decode"`
}

func ReadCorpus(dir string) (*Corpus, error) {
	var c Corpus
	if err := exportCUE(dir, "corpus", &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func ReadResults(dir string) (*Results, error) {
	var r Results
	if err := exportCUE(dir, "results", &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func exportCUE(dir string, expr string, dest interface{}) error {
	var buf bytes.Buffer
	c := exec.Command("cue", "export", "-e", expr, "--out=yaml")
	c.Dir = dir
	c.Stdout = &buf
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("cue export failed: %v", err)
	}
	if err := yaml.Unmarshal(buf.Bytes(), dest); err != nil {
		return fmt.Errorf("cannot unmarshal YAML CUE output: %v", err)
	}
	return nil
}

func WriteResults(dir string, r *Results) error {
	if err := writeCorpusData(dir, "results-decode", wrap(r.Decode, "results", "decode")); err != nil {
		return err
	}
	if err := writeCorpusData(dir, "results-encode", wrap(r.Encode, "results", "encode")); err != nil {
		return err
	}
	return nil
}

// writeCorpusData writes data into a CUE file into the given corpus directory
// with filenamePrefix as the filename prefix.
func writeCorpusData(corpusDir string, filenamePrefix string, dataVal interface{}) error {
	data, err := yaml.Marshal(dataVal)
	if err != nil {
		return fmt.Errorf("cannot marshal test cases: %v", err)
	}
	outfile := filepath.Join(corpusDir, filenamePrefix+"-generated.cue")
	c := exec.Command(
		"cue",
		"import",
		"-f",
		"-o="+outfile,
		"-p=lpcorpus",
		"yaml:",
		"-",
	)
	c.Stdin = bytes.NewReader(data)
	c.Stdout = os.Stderr
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return err
	}
	return nil
}

func wrap(x interface{}, elem ...string) interface{} {
	for i := len(elem) - 1; i >= 0; i-- {
		x = map[string]interface{}{
			elem[i]: x,
		}
	}
	return x
}

type Results struct {
	Encode map[string]*EncodeResults `yaml:"encode,omitempty"`
	Decode map[string]*DecodeResults `yaml:"decode,omitempty"`
}

type DecodeResults struct {
	Input  *DecodeInput             `yaml:"input"`
	Output map[string]*DecodeOutput `yaml:"output"`
}

type EncodeResults struct {
	Input  *EncodeInput             `yaml:"input"`
	Output map[string]*EncodeOutput `yaml:"output"`
}

type DecodeOutput struct {
	Result *Metric `yaml:"result,omitempty"`
	Error  string  `yaml:"error,omitempty"`
}

// Two decode outputs compare equal if they both succeed
// with the same decoded metric or they both failed.
func (o1 *DecodeOutput) Equal(o2 *DecodeOutput) bool {
	if o1.Result != nil && o2.Result != nil {
		return reflect.DeepEqual(o1.Result, o2.Result)
	}
	return o1.Error != "" && o2.Error != ""
}

type EncodeOutput struct {
	Result Bytes  `yaml:"result,omitempty"`
	Error  string `yaml:"error,omitempty"`
}

// Two encode outputs compare equal if they both succeed
// with the same decoded metric or they both failed.
func (o1 *EncodeOutput) Equal(o2 *EncodeOutput) bool {
	if o1.Result != nil && o2.Result != nil {
		return reflect.DeepEqual(o1.Result, o2.Result)
	}
	return o1.Error != "" && o2.Error != ""
}
