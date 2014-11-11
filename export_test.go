// Copyright 2014 Volker Dobler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package export

import (
	"encoding/csv"
	"errors"
	"fmt"
	"math"
	"os"
	"reflect"
	"strings"
	"testing"
	"text/tabwriter"
	"time"
)

var someError = errors.New("some error")

type S struct {
	B bool
	I int
	F float64
	S string
	T time.Time
	E error
	N Named
}

type Named uint16
type Named32 uint32

func (s S) BM() bool      { return s.B }
func (s S) IM() int       { return s.I }
func (s S) FM() float64   { return s.F }
func (s S) SM() string    { return s.S }
func (s S) TM() time.Time { return s.T }
func (s S) EM() error     { return s.E }
func (s S) NM() Named32   { return Named32(s.N * 2) }

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
var time3 = time.Date(2009, 12, 28, 9, 45, 0, 0, time.UTC)

var ss = []S{
	S{true, 23, 45.67, "Hello World!", time1, nil, 123},
	S{false, 9, 8.76, "Short", time2, nil, 456},
}

func TestExtractor(t *testing.T) {
	fieldNames := []string{"B", "I", "F", "S", "T",
		"BM()", "IM()", "FM()", "SM()", "TM()",
		"BME()", "IME()", "FME()", "SME()", "TME()",
	}
	extractor, err := NewExtractor(ss, fieldNames...)

	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	// Fields are in order and have proper type.
	for i, name := range fieldNames {
		field := extractor.Columns[i]
		if strings.HasSuffix(name, "()") {
			name = name[:len(name)-2]
		}
		if field.Name != name {
			t.Errorf("Column %d, got name %s, want %s", i, field.Name, name)
		}
		ft := field.typ.String()
		if ft[0] != name[0] {
			t.Errorf("Column %d, got type %s, want '%s'", i, ft, name)
		}
	}

	// Check for proper NA handling.
	for i := 10; i < 15; i++ {
		val := extractor.Columns[i].value(0)
		na := extractor.Columns[i].value(1)
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
		bfv := extractor.Columns[0].value(i).(bool)
		bmv := extractor.Columns[5].value(i).(bool)
		bemv := s.B
		if i%2 == 0 {
			bemv = extractor.Columns[10].value(i).(bool)
		}
		if bfv != s.B || bmv != s.B || bemv != s.B {
			t.Errorf("Bool %d: Got field=%t method=%t errmethod=%t, want %t",
				i, bfv, bmv, bemv, s.B)
		}

		// Integers
		ifv := extractor.Columns[1].value(i).(int64)
		imv := extractor.Columns[6].value(i).(int64)
		iemv := int64(s.I)
		if i%2 == 0 {
			iemv = extractor.Columns[11].value(i).(int64)
		}
		if ifv != int64(s.I) || imv != int64(s.I) || iemv != int64(s.I) {
			t.Errorf("Int %d: Got field=%d method=%d errmethod=%d, want %d",
				i, ifv, imv, iemv, s.I)
		}

		// Floats
		ffv := extractor.Columns[2].value(i).(float64)
		fmv := extractor.Columns[7].value(i).(float64)
		femv := s.F
		if i%2 == 0 {
			femv = extractor.Columns[12].value(i).(float64)
		}
		if ffv != s.F || fmv != s.F || femv != s.F {
			t.Errorf("Float %d: Got field=%g method=%g errmethod=%g, want %g",
				i, ffv, fmv, femv, s.F)
		}

		// Strings
		sfv := extractor.Columns[3].value(i).(string)
		smv := extractor.Columns[8].value(i).(string)
		semv := s.S
		if i%2 == 0 {
			semv = extractor.Columns[13].value(i).(string)
		}
		if sfv != s.S || smv != s.S || semv != s.S {
			t.Errorf("String %d: Got field=%s method=%s errmethod=%s, want %s",
				i, sfv, smv, semv, s.S)
		}

		// Times
		tfv := extractor.Columns[4].value(i).(time.Time)
		tmv := extractor.Columns[9].value(i).(time.Time)
		temv := s.T
		if i%2 == 0 {
			temv = extractor.Columns[14].value(i).(time.Time)
		}
		if !tfv.Equal(s.T) || !tmv.Equal(s.T) || !temv.Equal(s.T) {
			t.Errorf("Time %d: Got field=%s method=%s errmethod=%s, want %s",
				i, tfv, tmv, temv, s.T)
		}
	}

}

func TestAlias(t *testing.T) {
	extractor, err := NewExtractor(ss, "N", "NM()")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if g := extractor.Columns[0].value(0); g.(int64) != 123 {
		t.Errorf("N:0, got %v, want 123", g)
	}

	if g := extractor.Columns[1].value(0); g.(int64) != 246 {
		t.Errorf("N:0, got %v, want 246", g)
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
	data := []struct{ A int }{
		{0}, {2}, {4}, {6}, {8}, {10},
	}
	extractor, err := NewExtractor(data, "A")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	short := data[1:4]
	extractor.Bind(short)
	if extractor.N != 3 {
		t.Errorf("Got %d after rebinding, want 3", extractor.N)
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

	if extractor.N != 3 {
		t.Fatalf("Got %d elements, want 3", extractor.N)
	}

	if v := extractor.Columns[0].value(0); v == nil {
		t.Errorf("0: Unexpected nil")
	} else {
		g, ok := v.(int64)
		if !ok {
			t.Errorf("0: Got %v, want int", v)
		} else if g != 1 {
			t.Errorf("0: Got %d, want 1", g)
		}
	}

	if v := extractor.Columns[0].value(1); v != nil {
		t.Errorf("1: Got %v, want nil", v)
	}

	if v := extractor.Columns[0].value(2); v == nil {
		t.Errorf("2: Unexpected nil")
	} else {
		g, ok := v.(int64)
		if !ok {
			t.Errorf("2: Got %v, want int", v)
		} else if g != 2 {
			t.Errorf("2: Got %d, want 2", g)
		}
	}

}

func TestSliceOfPointers(t *testing.T) {
	data := []*S{
		&S{true, 23, 45.67, "Hello World!", time1, nil, 123},
		&S{false, 9, 8.76, "Short", time2, nil, 456},
		nil,
	}

	extractor, err := NewExtractor(data, "B", "I", "F", "S", "T")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 1, 8, 1, ' ', 0)
	TabDumper{Writer: w}.Dump(extractor, RFormat)
	w.Flush()
}

type T struct {
	A   int
	AP  *int
	APP **int
	B   TT
}

type TT struct {
	C  float64
	CP *float64
}

type TTT struct {
	E string
}

func (_ TT) D() int            { return 123 }
func (_ TT) F() TTT            { return TTT{E: "Hello"} }
func (_ TT) FE() (TTT, error)  { return TTT{}, fmt.Errorf("some err") }
func (_ TT) Fxyz() (TTT, bool) { return TTT{}, false }
func (t TTT) G() int           { return len(t.E) }
func (t TTT) GTT() TT          { return TT{} }

func TestBuildSteps(t *testing.T) {
	typ := reflect.TypeOf(T{})
	steps, _, _, err := buildSteps(typ, "B.F().E")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if steps[0].method.IsValid() {
		t.Errorf("B should be field, got method")
	}
	if !steps[1].method.IsValid() {
		t.Errorf("F should be method, got field")
	}
	if steps[2].method.IsValid() {
		t.Errorf("E should be field, got method")
	}

	steps, _, _, err = buildSteps(typ, "APP")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if steps[0].method.IsValid() {
		t.Errorf("APP should be field, got method")
	}
	if steps[0].indir != 2 {
		t.Errorf("Indir of APP = %, want 2", steps[0].indir)
	}
}

func TestBuildStepsErrors(t *testing.T) {
	typ := reflect.TypeOf(T{})

	_, _, _, err := buildSteps(typ, "X")
	if err == nil {
		t.Errorf("Expected no such field or method X.")
	}

	_, _, _, err = buildSteps(typ, "B")
	if err == nil {
		t.Errorf("Expected B to be of unusable typ for final element.")
	}

	_, _, _, err = buildSteps(typ, "B.X")
	if err == nil {
		t.Errorf("Expected no such field or method X.")
	}

	_, _, _, err = buildSteps(typ, "B.Fxyz.E")
	if err == nil {
		t.Errorf("Expected wrong method signature for Fxyz")
	}

	_, _, _, err = buildSteps(typ, "B.FE.GTT")
	if err == nil {
		t.Errorf("Expected wrong return type method GTT for last element.")
	}
}

func TestAccessRetrieve(t *testing.T) {
	i1, i2 := 11, 13
	pi2 := &i2
	fl := float64(17)
	data := T{
		A: i1, AP: nil, APP: &pi2,
		B: TT{C: 19, CP: &fl},
	}

	v := reflect.ValueOf(data)

	// Check access to a, ap and app.
	a := step{name: "A", field: 0}
	ap := step{name: "AP", field: 1, indir: 1}
	app := step{name: "APP", field: 2, indir: 2}

	if w := retrieve(v, []step{a}, 0, Int, false); w == nil {
		t.Fatalf("Unexpected nil")
	} else {
		if g := w.(int64); g != 11 {
			t.Errorf("got %d", g)
		}
	}
	if _, err := access(v, []step{ap}); err == nil {
		t.Fatalf("Missing error")
	}
	if w, err := access(v, []step{app}); err != nil {
		t.Fatalf("Unexpected error %s", err)
	} else {
		if g := w.Int(); g != 13 {
			t.Errorf("got %d", g)
		}
	}

	// Check deep down access.
	b := step{name: "B", field: 3}
	c := step{name: "C", field: 0}
	cp := step{name: "CP", field: 1, indir: 1}

	if w, err := access(v, []step{b, c}); err != nil {
		t.Fatalf("Unexpected error %s", err)
	} else {
		if g := w.Float(); g != 19 {
			t.Errorf("got %g", g)
		}
	}
	if w, err := access(v, []step{b, cp}); err != nil {
		t.Fatalf("Unexpected error %s", err)
	} else {
		if g := w.Float(); g != 17 {
			t.Errorf("got %g", g)
		}
	}

	// Check method access.
	m := reflect.TypeOf(TT{}).Method(0).Func
	d := step{name: "D", method: m}
	if w, err := access(v, []step{b, d}); err != nil {
		t.Fatalf("Unexpected error %s", err)
	} else {
		if g := w.Int(); g != 123 {
			t.Errorf("got %d", g)
		}
	}

	// Going even further.
	m = reflect.TypeOf(TT{}).Method(1).Func
	f := step{name: "f", method: m}
	e := step{name: "E", field: 0}
	if w := retrieve(v, []step{b, f, e}, 0, String, false); w == nil {
		t.Fatalf("Unexpected nil")
	} else {
		if g := w.(string); g != "Hello" {
			t.Errorf("got %q", g)
		}
	}
	m = reflect.TypeOf(TTT{}).Method(0).Func
	g := step{name: "G", method: m}
	if w, err := access(v, []step{b, f, g}); err != nil {
		t.Fatalf("Unexpected error %s", err)
	} else {
		if g := w.Int(); g != int64(len("Hello")) {
			t.Errorf("got %d", g)
		}
	}
}

// -------------------------------------------------------------------------

var table = []S{
	S{true, 12, 3.14149, "Hello", time1, nil, 123},
	S{true, 14, 2.71828, "World", time2, nil, 456},
	S{false, 14, math.NaN(), "Go", time1, nil, 789},
	S{false, 16, 6.02214e23, "A Lot", time3, nil, 246},
}

func TestCSVExtractor(t *testing.T) {
	extractor, err := NewExtractor(table, "B", "I", "F", "S", "T")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	d := CSVDumper{
		Writer:     csv.NewWriter(os.Stdout),
		OmitHeader: false,
	}

	d.Dump(extractor, DefaultFormat)
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 1, 8, 1, ' ', 0)
	TabDumper{Writer: w}.Dump(extractor, RFormat)
	w.Flush()
}

func TestRVecExtractor(t *testing.T) {
	extractor, err := NewExtractor(table, "B", "I", "F", "S", "T")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	d := RVecDumper{
		Writer:    os.Stdout,
		DataFrame: "body.data",
	}

	d.Dump(extractor, RFormat)
}
