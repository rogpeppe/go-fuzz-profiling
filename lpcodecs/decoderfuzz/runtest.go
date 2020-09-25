// +build ignore

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	"github.com/rogpeppe/line-protocol-corpus/lpcodecs"
	"gopkg.in/yaml.v3"
)

var implementations = flag.String("impl", "", "comma-separated implementations to run (default all)")

var defaultTime = time.Date(2000, 1, 2, 12, 13, 14, 0, time.UTC).UnixNano()

func main() {
	var names []string
	for name := range lpcodecs.Implementations {
		names = append(names, name)
	}
	sort.Strings(names)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "runtest crashhash|-\n")
		fmt.Fprintf(os.Stderr, "- means use standard input as input\n")
		fmt.Fprintf(os.Stderr, "implementations: %v\n", strings.Join(names, ","))
		flag.PrintDefaults()
		os.Exit(2)
	}
	flag.Parse()
	if len(flag.Args()) != 1 {
		flag.Usage()
	}
	hash := flag.Arg(0)
	useImpl := func(string) bool {
		return true
	}
	if *implementations != "" {
		impls := strings.Split(*implementations, ",")
		useImpl = func(s string) bool {
			for _, impl := range impls {
				if impl == s {
					return true
				}
			}
			return false
		}
	}

	var data []byte
	if hash == "-" {
		var err error
		data, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal("error reading stdin: ", err)
		}
	} else {
		var err error
		data, err = ioutil.ReadFile("crashers/" + hash)
		if err != nil {
			log.Fatal(err)
		}
	}
	input := &lpcorpus.DecodeInput{
		Text:        data,
		DefaultTime: defaultTime,
		Precision: lpcorpus.Precision{
			Duration: time.Nanosecond,
		},
	}
	outputs := make(map[string]*lpcorpus.DecodeOutput)
	for _, name := range names {
		impl := lpcodecs.Implementations[name]
		if !useImpl(name) {
			continue
		}
		var m *lpcorpus.Metric
		var err error
		func() {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("panic: %v\n%s", r, debug.Stack())
				}
			}()
			m, err = impl.Decoder.Decode(input)
		}()
		if err != nil {
			if _, ok := err.(*lpcodecs.SkipError); ok {
				continue
			}
			outputs[name] = &lpcorpus.DecodeOutput{
				Error: err.Error(),
			}
		} else {
			outputs[name] = &lpcorpus.DecodeOutput{
				Result: m,
			}
		}
	}
	data, _ = yaml.Marshal(outputs)
	os.Stdout.Write(data)
}
