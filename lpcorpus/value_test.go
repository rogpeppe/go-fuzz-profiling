package lpcorpus

import (
	"math"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gopkg.in/yaml.v3"
)

var valueMarshalTests = []struct {
	testName string
	val      Value
	expect   string
}{{
	testName: "int",
	val:      MustNewValue(int64(1234)),
	expect: `
type: int
value: 1234
`[1:],
}, {
	testName: "int",
	val:      MustNewValue(uint64(1234)),
	expect: `
type: uint
value: 1234
`[1:],
}, {
	testName: "float",
	val:      MustNewValue(4.5),
	expect: `
type: float
value: 4.5
`[1:],
}, {
	testName: "bool",
	val:      MustNewValue(false),
	expect: `
type: bool
value: false
`[1:],
}, {
	testName: "string",
	val:      MustNewValue("hello"),
	expect: `
type: string
value: hello
`[1:],
}, {
	testName: "binary",
	val:      MustNewValue("hello\xff"),
	expect: `
type: string
value: !!binary 'aGVsbG//'
`[1:],
}, {
	testName: "infinity",
	val:      MustNewValue(math.Inf(1)),
	expect: `
type: float
value: Inf
`[1:],
}, {
	testName: "NaN",
	val:      MustNewValue(math.NaN()),
	expect: `
type: float
value: NaN
`[1:],
}}

func TestValueMarshal(t *testing.T) {
	c := qt.New(t)
	for _, test := range valueMarshalTests {
		c.Run(test.testName, func(c *qt.C) {
			data, err := yaml.Marshal(test.val)
			c.Assert(err, qt.IsNil)
			c.Assert(string(data), qt.Equals, test.expect)
			var v Value
			err = yaml.Unmarshal(data, &v)
			c.Assert(err, qt.IsNil)
			c.Assert(v, qt.CmpEquals(cmpopts.EquateNaNs()), test.val)
		})
	}
}
