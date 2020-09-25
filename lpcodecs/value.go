package lpcodecs

import (
	"bytes"
	"fmt"
	"math"

	"gopkg.in/yaml.v3"
)

// Note: this hacky union type is total overkill for this use, but the code might come in
// useful later in some place where high performance is more
// important so let's keep it for now.

// Value holds one of the possible line-protocol field values.
type Value struct {
	// number covers:
	//	- signed integer
	//	- unsigned integer
	//	- bool
	//	- float
	number uint64
	// bytes holds the string bytes or a sentinel (see below)
	// when the value's not holding a string.
	bytes []byte
}

var (
	intSentinel   = [1]byte{'i'}
	uintSentinel  = [1]byte{'u'}
	floatSentinel = [1]byte{'f'}
	boolSentinel  = [1]byte{'b'}
)

func MustNewValue(x interface{}) Value {
	v, ok := NewValue(x)
	if !ok {
		panic(fmt.Errorf("invalid value for NewValue: %T (%#v)", x, x))
	}
	return v
}

func (v1 Value) Equal(v2 Value) bool {
	return v1.Kind() == v2.Kind() && v1.number == v2.number && bytes.Equal(v1.bytes, v2.bytes)
}

func NewValue(x interface{}) (Value, bool) {
	switch x := x.(type) {
	case int64:
		return Value{
			number: uint64(x),
			bytes:  intSentinel[:],
		}, true
	case uint64:
		return Value{
			number: uint64(x),
			bytes:  uintSentinel[:],
		}, true
	case float64:
		//if math.IsInf(x, 0) || math.IsNaN(x) {
		//	return Value{}, false
		//}
		return Value{
			number: math.Float64bits(x),
			bytes:  floatSentinel[:],
		}, true
	case bool:
		n := uint64(0)
		if x {
			n = 1
		}
		return Value{
			number: uint64(n),
			bytes:  boolSentinel[:],
		}, true
	case string:
		return Value{
			bytes: []byte(x),
		}, true
	case []byte:
		// Could retain a reference to the []byte, but that's dubious.
		return Value{
			bytes: append([]byte(nil), x...),
		}, true
	}
	return Value{}, false
}

func (v Value) IntV() int64 {
	v.mustBe(Int)
	return int64(v.number)
}

func (v Value) UintV() uint64 {
	v.mustBe(Uint)
	return v.number
}

func (v Value) FloatV() float64 {
	v.mustBe(Float)
	return math.Float64frombits(v.number)
}

func (v Value) StringV() string {
	v.mustBe(String)
	return string(v.bytes)
}

func (v Value) BytesV() []byte {
	v.mustBe(String)
	return v.bytes
}

func (v Value) BoolV() bool {
	v.mustBe(Bool)
	return v.number != 0
}

func (v Value) Interface() interface{} {
	switch v.Kind() {
	case Int:
		return v.IntV()
	case Uint:
		return v.UintV()
	case String:
		return v.StringV()
	case Bool:
		return v.BoolV()
	case Float:
		return v.FloatV()
	default:
		panic("unknown value kind")
	}
}

type encodedValue struct {
	Type  ValueKind `yaml:"type"`
	Value interface{}
}

func (v Value) MarshalYAML() (interface{}, error) {
	switch v.Kind() {
	case String:
		return encodedValue{
			Type:  String,
			Value: Bytes(v.BytesV()),
		}, nil
	case Float:
		var x interface{}
		f := v.FloatV()
		switch {
		case math.IsNaN(f):
			x = "NaN"
		case math.IsInf(f, 0):
			x = "Inf"
		default:
			x = f
		}
		return encodedValue{
			Type:  v.Kind(),
			Value: x,
		}, nil
	default:
		return encodedValue{
			Type:  v.Kind(),
			Value: v.Interface(),
		}, nil
	}
}

func (v *Value) UnmarshalYAML(n *yaml.Node) error {
	var e encodedValue
	if err := n.Decode(&e); err != nil {
		return err
	}
	switch e.Type {
	case Int:
		switch n := e.Value.(type) {
		case int:
			*v = MustNewValue(int64(n))
		case int64:
			*v = MustNewValue(n)
		case float64:
			*v = MustNewValue(int64(n))
		default:
			return fmt.Errorf("unknown type for int (%T)", e.Value)
		}
	case Uint:
		switch n := e.Value.(type) {
		case int:
			*v = MustNewValue(uint64(n))
		case float64:
			*v = MustNewValue(uint64(n))
		case uint32:
			*v = MustNewValue(uint64(n))
		case uint64:
			*v = MustNewValue(n)
		default:
			return fmt.Errorf("unknown type for uint (%T)", e.Value)
		}
	case Float:
		switch n := e.Value.(type) {
		case int:
			*v = MustNewValue(float64(n))
		case float64:
			*v = MustNewValue(n)
		case string:
			switch n {
			case "NaN":
				*v = MustNewValue(math.NaN())
			case "Inf":
				*v = MustNewValue(math.Inf(1))
			default:
				return fmt.Errorf("unknown string string value for float %q (need NaN or Inf)", n)
			}
		default:
			return fmt.Errorf("unknown type for float (%T)", e.Value)
		}
	case Bool:
		*v = MustNewValue(e.Value)
	case String:
		*v = MustNewValue(e.Value)
	default:
		return fmt.Errorf("unknown value kind")
	}
	return nil
}

func (v Value) mustBe(k ValueKind) {
	if v.Kind() != k {
		panic(fmt.Errorf("value has unexpected kind; got %v want %v", v.Kind(), k))
	}
}

func (v Value) Kind() ValueKind {
	if len(v.bytes) != 1 {
		return String
	}
	switch &v.bytes[0] {
	case &intSentinel[0]:
		return Int
	case &uintSentinel[0]:
		return Uint
	case &floatSentinel[0]:
		return Float
	case &boolSentinel[0]:
		return Bool
	}
	return String
}

func (v Value) String() string {
	switch v.Kind() {
	case Float:
		return fmt.Sprint(v.FloatV())
	case Int:
		return fmt.Sprintf("%di", v.IntV())
	case Uint:
		return fmt.Sprintf("%du", v.UintV())
	case Bool:
		return fmt.Sprint(v.BoolV())
	case String:
		return fmt.Sprintf("%q", v.StringV())
	default:
		panic("unknown kind")
	}
}

type ValueKind uint8

const (
	Unknown ValueKind = iota
	String
	Int
	Uint
	Float
	Bool
)

var kinds = []string{
	Unknown: "unknown",
	String:  "string",
	Int:     "int",
	Uint:    "uint",
	Float:   "float",
	Bool:    "bool",
}

func (k ValueKind) String() string {
	return kinds[k]
}

func (k ValueKind) MarshalText() ([]byte, error) {
	if k == Unknown {
		return nil, fmt.Errorf("cannot marshal 'unknown' value type")
	}
	return []byte(k.String()), nil
}

func (k *ValueKind) UnmarshalText(data []byte) error {
	s := string(data)
	for i, kstr := range kinds {
		if i > 0 && kstr == s {
			*k = ValueKind(i)
			return nil
		}
	}
	return fmt.Errorf("unknown Value kind %q", s)
}
