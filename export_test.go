package export

import (
	"encoding/csv"
	"errors"
	"fmt"
	"math"
	"os"
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
		field := extractor.Fields[i]
		if field.Name != name {
			t.Errorf("Field %d, got name %s, want %s", i, field.Name, name)
		}
		ft := field.Type.String()
		if ft[0] != name[0] {
			t.Errorf("Field %d, got type %s, want '%s'", i, ft, name)
		}
	}

	// Check for proper NA handling.
	for i := 10; i < 15; i++ {
		val := extractor.Fields[i].Value(0)
		na := extractor.Fields[i].Value(1)
		if val == nil {
			t.Errorf("Field %s unexpected nil", fieldNames[i])
		}
		if na != nil {
			t.Errorf("Field %s unexpected non nil, got %v", fieldNames[i], na)
		}
	}

	// Check values.
	for i, s := range ss {
		// Booleans
		bfv := extractor.Fields[0].Value(i).(bool)
		bmv := extractor.Fields[5].Value(i).(bool)
		bemv := s.B
		if i%2 == 0 {
			bemv = extractor.Fields[10].Value(i).(bool)
		}
		if bfv != s.B || bmv != s.B || bemv != s.B {
			t.Errorf("Bool %d: Got field=%t method=%t errmethod=%t, want %t",
				i, bfv, bmv, bemv, s.B)
		}

		// Integers
		ifv := extractor.Fields[1].Value(i).(int64)
		imv := extractor.Fields[6].Value(i).(int64)
		iemv := int64(s.I)
		if i%2 == 0 {
			iemv = extractor.Fields[11].Value(i).(int64)
		}
		if ifv != int64(s.I) || imv != int64(s.I) || iemv != int64(s.I) {
			t.Errorf("Int %d: Got field=%d method=%d errmethod=%d, want %d",
				i, ifv, imv, iemv, s.I)
		}

		// Floats
		ffv := extractor.Fields[2].Value(i).(float64)
		fmv := extractor.Fields[7].Value(i).(float64)
		femv := s.F
		if i%2 == 0 {
			femv = extractor.Fields[12].Value(i).(float64)
		}
		if ffv != s.F || fmv != s.F || femv != s.F {
			t.Errorf("Float %d: Got field=%g method=%g errmethod=%g, want %g",
				i, ffv, fmv, femv, s.F)
		}

		// Strings
		sfv := extractor.Fields[3].Value(i).(string)
		smv := extractor.Fields[8].Value(i).(string)
		semv := s.S
		if i%2 == 0 {
			semv = extractor.Fields[13].Value(i).(string)
		}
		if sfv != s.S || smv != s.S || semv != s.S {
			t.Errorf("String %d: Got field=%s method=%s errmethod=%s, want %s",
				i, sfv, smv, semv, s.S)
		}

		// Times
		tfv := extractor.Fields[4].Value(i).(time.Time)
		tmv := extractor.Fields[9].Value(i).(time.Time)
		temv := s.T
		if i%2 == 0 {
			temv = extractor.Fields[14].Value(i).(time.Time)
		}
		if !tfv.Equal(s.T) || !tmv.Equal(s.T) || !temv.Equal(s.T) {
			t.Errorf("Time %d: Got field=%s method=%s errmethod=%s, want %s",
				i, tfv, tmv, temv, s.T)
		}
	}

}

func TestBadField(t *testing.T) {
	for i, name := range []string{"Unexisting", "E", "EM", "EME", "ExtraArg", "WrongReturn"} {
		_, err := NewExtractor(ss, name)
		if err == nil {
			t.Errorf("%d: Got nil error on field %s", i, name)
		}
	}
}
