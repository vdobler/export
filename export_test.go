package export

import (
	"encoding/csv"
	"errors"
	"fmt"
	"math"
	"os"
	"reflect"
	"testing"
	"time"
)

type Obs struct {
	Age     int
	Origin  string
	Weight  float64
	Height  float64
	Special []byte
}

func (o Obs) BMI() float64 {
	return o.Weight / (o.Height * o.Height)
}

func (o Obs) Group() int {
	return 10*(o.Age/10) + 5
}

func (o Obs) Fancy() (int, error) {
	if o.Height < 1.65 {
		return 0, fmt.Errorf("too small (was %.2f)", o.Height)
	}
	return int(100 * math.Sqrt(o.Height-1.65)), nil
}

func (o Obs) Country() string {
	o2c := map[string]string{
		"ch": "Schweiz",
		"de": "Deutschland",
		"uk": "England",
	}
	return o2c[o.Origin]
}

func (o Obs) IsEU() bool {
	return o.Origin != "ch"
}

var measurement = []Obs{
	Obs{Age: 20, Origin: "de", Weight: 80, Height: 1.88},
	Obs{Age: 22, Origin: "de", Weight: 85, Height: 1.85},
	Obs{Age: 20, Origin: "de", Weight: 90, Height: 1.95},
	Obs{Age: 25, Origin: "de", Weight: 90, Height: 1.72},

	Obs{Age: 20, Origin: "ch", Weight: 77, Height: 1.78},
	Obs{Age: 20, Origin: "ch", Weight: 82, Height: 1.75},
	Obs{Age: 28, Origin: "ch", Weight: 85, Height: 1.80},
	Obs{Age: 20, Origin: "ch", Weight: 84, Height: 1.62},

	Obs{Age: 31, Origin: "de", Weight: 85, Height: 1.88},
	Obs{Age: 30, Origin: "de", Weight: 90, Height: 1.85},
	Obs{Age: 30, Origin: "de", Weight: 99, Height: 1.95},
	Obs{Age: 42, Origin: "de", Weight: 95, Height: 1.72},

	Obs{Age: 30, Origin: "ch", Weight: 80, Height: 1.78},
	Obs{Age: 30, Origin: "ch", Weight: 85, Height: 1.75},
	Obs{Age: 37, Origin: "ch", Weight: 87, Height: 1.80},
	Obs{Age: 47, Origin: "ch", Weight: 90, Height: 1.62},

	Obs{Age: 42, Origin: "uk", Weight: 60, Height: 1.68},
	Obs{Age: 42, Origin: "uk", Weight: 65, Height: 1.65},
	Obs{Age: 44, Origin: "uk", Weight: 55, Height: 1.52},
	Obs{Age: 44, Origin: "uk", Weight: 70, Height: 1.72},
}

func TestCSVExtractor(t *testing.T) {
	extractor, err := NewExtractor(measurement, "Age", "Origin", "Weight", "BMI", "Fancy", "IsEU")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	d := CSVDumper{
		Writer:     csv.NewWriter(os.Stdout),
		OmitHeader: false,
	}

	d.Dump(extractor, DefaultFormat)
	TabDumper{Writer: os.Stdout}.Dump(extractor, RFormat)
}

func TestRVecExtractor(t *testing.T) {
	extractor, err := NewExtractor(measurement, "Age", "Origin", "BMI", "Fancy", "IsEU")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	d := RVecDumper{
		Writer: os.Stdout,
		Name:   "body.data",
	}

	d.Dump(extractor, RFormat)
}

var someError = errors.New("some error")

type S struct {
	B bool
	I int
	F float64
	S string
	T time.Time
	E error
}

func (s S) BM() bool      { return s.B }
func (s S) IM() int       { return s.I }
func (s S) FM() float64   { return s.F }
func (s S) SM() string    { return s.S }
func (s S) TM() time.Time { return s.T }
func (s S) EM() error     { return s.E }

func (s S) BME() (bool, error) {
	if s.B {
		return true, nil
	}
	return false, someError
}

func (s S) IME() (int, error) {
	if s.I > 10 {
		return s.I, nil
	}
	return 0, someError
}

func (s S) FME() (float64, error) {
	if s.F > 10 {
		return s.F, nil
	}
	return 0, someError
}

func (s S) SME() (string, error) {
	if len(s.S) > 10 {
		return s.S, nil
	}
	return "", someError
}

func (s S) TME() (time.Time, error) {
	if s.T.Hour() > 10 {
		return s.T, nil
	}
	return time.Time{}, someError
}

func (s S) EME() (error, error) {
	return s.E, someError
}

func (s S) ExtraArg(int) int {
	return 12
}

func (s S) WrongReturn() (int, int) {
	return 13, 14
}

var time1 = time.Date(2000, 1, 2, 15, 20, 30, 0, time.UTC)
var time2 = time.Date(2000, 1, 2, 3, 20, 30, 0, time.UTC)

var ss = []S{
	S{true, 23, 45.67, "Hello World!", time1, nil},
	S{false, 9, 8.76, "Short", time2, nil},
}

func TestExtractor(t *testing.T) {
	fieldNames := []string{"B", "I", "F", "S", "T",
		"BM", "IM", "FM", "SM", "TM",
		"BME", "IME", "FME", "SME", "TME"}
	extractor, err := NewExtractor(ss, fieldNames...)

	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	// Fields are in order and have proper type.
	for i, name := range fieldNames {
		field := extractor.Columns[i]
		if field.Name != name {
			t.Errorf("Column %d, got name %s, want %s", i, field.Name, name)
		}
		ft := field.Type.String()
		if ft[0] != name[0] {
			t.Errorf("Column %d, got type %s, want '%s'", i, ft, name)
		}
	}

	// Check for proper NA handling.
	for i := 10; i < 15; i++ {
		val := extractor.Columns[i].Value(0)
		na := extractor.Columns[i].Value(1)
		if val == nil {
			t.Errorf("Column %s unexpected nil", fieldNames[i])
		}
		if na != nil {
			t.Errorf("Column %s unexpected non nil, got %v", fieldNames[i], na)
		}
	}

	// Check values.
	for i, s := range ss {
		// Booleans
		bfv := extractor.Columns[0].Value(i).(bool)
		bmv := extractor.Columns[5].Value(i).(bool)
		bemv := s.B
		if i%2 == 0 {
			bemv = extractor.Columns[10].Value(i).(bool)
		}
		if bfv != s.B || bmv != s.B || bemv != s.B {
			t.Errorf("Bool %d: Got field=%t method=%t errmethod=%t, want %t",
				i, bfv, bmv, bemv, s.B)
		}

		// Integers
		ifv := extractor.Columns[1].Value(i).(int64)
		imv := extractor.Columns[6].Value(i).(int64)
		iemv := int64(s.I)
		if i%2 == 0 {
			iemv = extractor.Columns[11].Value(i).(int64)
		}
		if ifv != int64(s.I) || imv != int64(s.I) || iemv != int64(s.I) {
			t.Errorf("Int %d: Got field=%d method=%d errmethod=%d, want %d",
				i, ifv, imv, iemv, s.I)
		}

		// Floats
		ffv := extractor.Columns[2].Value(i).(float64)
		fmv := extractor.Columns[7].Value(i).(float64)
		femv := s.F
		if i%2 == 0 {
			femv = extractor.Columns[12].Value(i).(float64)
		}
		if ffv != s.F || fmv != s.F || femv != s.F {
			t.Errorf("Float %d: Got field=%g method=%g errmethod=%g, want %g",
				i, ffv, fmv, femv, s.F)
		}

		// Strings
		sfv := extractor.Columns[3].Value(i).(string)
		smv := extractor.Columns[8].Value(i).(string)
		semv := s.S
		if i%2 == 0 {
			semv = extractor.Columns[13].Value(i).(string)
		}
		if sfv != s.S || smv != s.S || semv != s.S {
			t.Errorf("String %d: Got field=%s method=%s errmethod=%s, want %s",
				i, sfv, smv, semv, s.S)
		}

		// Times
		tfv := extractor.Columns[4].Value(i).(time.Time)
		tmv := extractor.Columns[9].Value(i).(time.Time)
		temv := s.T
		if i%2 == 0 {
			temv = extractor.Columns[14].Value(i).(time.Time)
		}
		if !tfv.Equal(s.T) || !tmv.Equal(s.T) || !temv.Equal(s.T) {
			t.Errorf("Time %d: Got field=%s method=%s errmethod=%s, want %s",
				i, tfv, tmv, temv, s.T)
		}
	}

}

func TestBadColumn(t *testing.T) {
	for i, name := range []string{"Unexisting", "E", "EM", "EME", "ExtraArg", "WrongReturn"} {
		_, err := NewExtractor(ss, name)
		if err == nil {
			t.Errorf("%d: Got nil error on field %s", i, name)
		}
	}
}

func TestBind(t *testing.T) {
	extractor, err := NewExtractor(measurement, "Age", "Origin")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	short := measurement[0:5]
	extractor.Bind(short)
	if extractor.N != 5 {
		t.Errorf("Expected length 5 after rebinding, got %d", extractor.N)
	}
}

func TestPointerFields(t *testing.T) {
	type P struct{ A *int }
	i, j := 1, 2
	data := []P{
		P{A: &i}, P{A: nil}, P{A: &j},
	}
	extractor, err := NewExtractor(data, "A")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	TabDumper{Writer: os.Stdout}.Dump(extractor, RFormat)
}

func TestSliceOfPointers(t *testing.T) {
	data := []*S{
		&S{true, 23, 45.67, "Hello World!", time1, nil},
		&S{false, 9, 8.76, "Short", time2, nil},
	}

	extractor, err := NewExtractor(data, "B", "I", "F", "S", "T")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	TabDumper{Writer: os.Stdout}.Dump(extractor, RFormat)
}

type T struct {
	A   int
	AP  *int
	APP **int
	B   TT
	BPP **TT
}

type TT struct {
	C  float64
	CP *float64
}

type TTT struct {
	E string
}

func (_ TT) D() int  { return 123 }
func (_ TT) F() TTT  { return TTT{E: "Hello"} }
func (t TTT) G() int { return len(t.E) }

func TestAccessor(t *testing.T) {
	i1, i2, i3 := 7, 11, 13
	pi3 := &i3
	f := 3.141
	data := T{
		A: i1, AP: &i2, APP: &pi3,
		B: TT{C: 2.7, CP: &f},
	}

	a0 := step{name: "A-0", typ: Struct, num: 0}
	a := step{name: "A", typ: Int}
	app0 := step{name: "APP-0", typ: Struct, num: 2}
	app := step{name: "APP", typ: Int, indir: 2}
	fmt.Printf("A=%v\n", access(v, []step{a0, a}))
	fmt.Printf("APP=%v\n", access(v, []step{app0, app}))

	b := step{name: "B", typ: Struct, num: 3}
	// bpp := step{name: "BPP",typ: Struct, num: 3, indir: 2}

	c0 := step{name: "C0", typ: Struct, num: 0}
	c := step{name: "C", typ: Float}
	cp0 := step{name: "CP0", typ: Struct, num: 1}
	cp := step{name: "CP", typ: Float, indir: 1}

	v := reflect.ValueOf(data)
	fmt.Printf("B.C=%v\n", access(v, []step{b, c0, c}))
	fmt.Printf("B.CP=%v\n", access(v, []step{b, cp0, cp}))

	//	c := step{name: "C",typ: Struct, num: 0}

}
