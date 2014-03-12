// Copyright 2014 Volker DObler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package export provides tools to dump tabulated data.
//
// Export allows to dump tabular data in different output formats.
// The main type is Exporter which determines which data is output and in
// which order. An Exporter is constructed from (almost) any slice type
// and may access nested fields.
//
// Export can export the following Go types:
//   - bool
//   - uint8, uint16, ...,  int64
//   - float32 and float64
//   - string
//   - time.Time
// This package handles floats and int as 64bit values. Thus an uint64
// may overflow without notice.
//
package export

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"reflect"
	"strings"
	"text/tabwriter"
	"time"
)

func isTime(x reflect.Type) bool {
	return x.PkgPath() == "time" && x.Kind() == reflect.Struct && x.Name() == "Time"
}

// -------------------------------------------------------------------------
// Type and Column

// Type represents the basic type of a column.
type Type uint

const (
	NA Type = iota
	Bool
	Int
	Float
	String
	Time
)

// String returns the name of ft.
func (ft Type) String() string {
	return []string{"NA", "Bool", "Int", "Float", "String", "Time"}[ft]
}

// Column represents one column in the export.
type Column struct {
	Name string // The name of the column.
	Type Type   // The type of the column.

	// Value returns the i'th value in this column.
	// For errors or nil pointers nil is returned.
	Value func(i int) interface{}

	mayFail bool   // Pointer fields or erroring methods.
	access  []step // The steps needed to access the result.
}

// Print the i'th entry of column c given format.
func (c Column) Print(format Format, i int) string {
	val := c.Value(i)
	if val == nil {
		return format.NA
	}
	switch c.Type {
	case Bool:
		if val.(bool) {
			return format.True
		}
		return format.False
	case Int:
		return fmt.Sprintf(format.IntFmt, val.(int64))
	case Float:
		x := val.(float64)
		if math.IsNaN(x) {
			return format.NaN
		}
		return fmt.Sprintf(format.FloatFmt, val.(float64))
	case String:
		return fmt.Sprintf(format.StringFmt, val.(string))
	case Time:
		t := val.(time.Time)
		if format.TimeLoc != nil {
			t = t.In(format.TimeLoc)
		}
		return t.Format(format.TimeFmt)
	}

	return fmt.Sprintf("Ooops: %v", val)
}

// -------------------------------------------------------------------------
// Dumper

// CSVDumper dumps values in CSV format
type CSVDumper struct {
	Writer     *csv.Writer // The csv.Writer to output the data.
	OmitHeader bool        // OmitHeader suppresses the header line in the generated CSV.
}

// Dump dumps the fields from e to d.
func (d CSVDumper) Dump(e *Extractor, format Format) error {
	row := make([]string, len(e.Columns))
	if !d.OmitHeader {
		for i, field := range e.Columns {
			row[i] = field.Name
		}
		d.Writer.Write(row)
	}
	for r := 0; r < e.N; r++ {
		for col, field := range e.Columns {
			row[col] = field.Print(format, r)
		}
		err := d.Writer.Write(row)
		if err != nil {
			return err
		}
	}
	d.Writer.Flush()
	return d.Writer.Error()
}

type TabDumper struct {
	Writer io.Writer
	Flags  uint
}

// Dump dumps the fields from e to d.
func (d TabDumper) Dump(e *Extractor, format Format) error {
	w := new(tabwriter.Writer)
	w.Init(d.Writer, 1, 8, 1, ' ', d.Flags)
	for i, field := range e.Columns {
		if i == 0 {
			fmt.Fprintf(w, "%s", field.Name)
		} else {
			fmt.Fprintf(w, "\t%s", field.Name)
		}
	}
	fmt.Fprintln(w)
	for r := 0; r < e.N; r++ {
		for i, field := range e.Columns {
			if i == 0 {
				fmt.Fprintf(w, "%s", field.Print(format, r))
			} else {
				fmt.Fprintf(w, "\t%s", field.Print(format, r))
			}
		}
		fmt.Fprintln(w)
	}
	w.Flush()
	return nil
}

// RVecDumper dumps as a R vectors.
type RVecDumper struct {
	Writer io.Writer

	// If Name is nonempty a data frame named Name consisting of all
	// fields is constructed too.
	Name string
}

// Dump dumps the fields from e to d.
func (d RVecDumper) Dump(e *Extractor, format Format) error {
	all := ""
	for f, field := range e.Columns {
		if _, err := fmt.Fprintf(d.Writer, "%s <- c(", field.Name); err != nil {
			return err
		}
		for r := 0; r < e.N; r++ {
			s := field.Print(format, r)
			if r < e.N-1 {
				if r%10 == 9 {
					s += ",\n"
				} else {
					s += ", "
				}
			}
			if _, err := fmt.Fprintf(d.Writer, "%s", s); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(d.Writer, ")\n"); err != nil {
			return err
		}
		if f > 0 {
			all += ", "
		}
		all += field.Name
	}

	if d.Name != "" {
		if _, err := fmt.Fprintf(d.Writer, "%s <- data.frame(%s)\n", d.Name, all); err != nil {
			return err
		}
	}
	return nil
}

// -------------------------------------------------------------------------
// Extractor

// Extractor provides access to fields and methods of tabular data.
type Extractor struct {
	// N is the numer of elements in the currently bound data.
	N int

	// Columns contains all the columns to extract.
	Columns []Column

	som   bool // som is true for slice-of-measurement type data.
	indir int  // number of primary som indirections; e.g. 2 for []**Data

	// typ contains the go type this Extractor
	// can work on i.e. can be bound to.
	typ reflect.Type
}

// NewExtractor returns an extractor for the given column specifications of data.
func NewExtractor(data interface{}, columnSpecs ...string) (*Extractor, error) {
	typ := reflect.TypeOf(data)
	switch typ.Kind() {
	case reflect.Slice:
		ex, err := newSOMExtractor(data, columnSpecs...)
		if err != nil {
			return ex, err
		}
		ex.som = true
		ex.typ = typ
		ex.bindSOM(data) // This sets up ex.N and ex.Columns[i].Value.
		return ex, nil
	case reflect.Struct:
		panic("COS data frame not implemented")
	}
	return &Extractor{}, fmt.Errorf("Cannot build Extrator for %s", typ.String())
}

// newSOMExtractor sets up an unbound Extractor for a slice-of-measurements
// type data.
func newSOMExtractor(data interface{}, colSpecs ...string) (*Extractor, error) {
	typ := reflect.TypeOf(data).Elem()
	indir := 0
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		indir++
	}
	ex := Extractor{
		indir: indir,
	}

	for _, spec := range colSpecs {
		steps, err := buildSteps(typ, spec)
		if err != nil {
			return nil, err
		}
		last := steps[len(steps)-1]
		mayFail := false
		if last.isMethodCall() {
			if last.mayFail {
				mayFail = true
			}
		} else {
			if last.indir > 0 {
				mayFail = true
			}
		}

		field := Column{
			Name:    last.name,
			Type:    superType(last.typ),
			mayFail: mayFail,
			access:  steps,
		}
		ex.Columns = append(ex.Columns, field)
	}

	return &ex, nil
}

// Bind (re)binds e to data which must be of the same type as the data used
// during the construction of e.
func (e *Extractor) Bind(data interface{}) {
	typ := reflect.TypeOf(data)
	if typ != e.typ {
		panic(fmt.Sprintf("Cannot bind extractor for %v to data of type %v",
			e.typ, typ))
	}
	if e.som {
		e.bindSOM(data)
	} else {
		panic("COS data frame not implemented")
	}
}

// bindSOM is the slice-of-measurements version of Bind.
func (e *Extractor) bindSOM(data interface{}) {
	v := reflect.ValueOf(data)
	e.N = v.Len()
	for fn, field := range e.Columns {
		access := field.access
		e.Columns[fn].Value = func(i int) interface{} {
			return retrieve(v.Index(i), access, e.indir)
		}
	}
}

// superType returns our types which group Go's low level types.
// A Go type which cannot be handled will yield NA.
// TODO: this might be the worst name possible for this function.
func superType(t reflect.Type) Type {
	switch t.Kind() {
	case reflect.Bool:
		return Bool
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return Int
	case reflect.String:
		return String
	case reflect.Float32, reflect.Float64:
		return Float
	case reflect.Struct:
		if isTime(t) {
			return Time
		}
	}
	return NA
}

var (
	errorInterfaceType = reflect.TypeOf((*error)(nil)).Elem()
)

// -------------------------------------------------------------------------
// Steps and accessing fields/methods

// step describes one step during the way down the type hierarchy.
type step struct {
	name    string        // the name of this element
	indir   int           // number of ptr-indirections to take before a type is reached
	method  reflect.Value // the function to call, if zero: not a fn call but a field access
	field   int           // field number if method is zero
	mayFail bool          // for methods which return (result, error)
	typ     reflect.Type
}

func (s step) isMethodCall() bool { return s.method.IsValid() }

// buildSteps constructs a slice of steps to access the given elem in typ.
func buildSteps(typ reflect.Type, elem string) ([]step, error) {
	var steps []step
	elements := strings.Split(elem, ".")
	for e, cur := range elements {
		found := false
		last := e == len(elements)-1

		// Fields first.
		for f := 0; f < typ.NumField(); f++ {
			if typ.Field(f).Name == cur {
				typ = typ.Field(f).Type
				indir := 0
				for typ.Kind() == reflect.Ptr {
					typ = typ.Elem()
					indir++
				}
				t := superType(typ)
				if last && t == NA {
					return steps, fmt.Errorf("export: cannot use field %s of type %v as final element", cur, typ)
				}
				s := step{name: cur, field: f, indir: indir, typ: typ}
				steps = append(steps, s)
				found = true
				break
			}
		}
		if found {
			continue
		}

		// Methods next
		m, ok := typ.MethodByName(cur)
		if !ok {
			return steps, fmt.Errorf("export: no field or method %s in %T", cur, typ)
		}
		// Look for methods with signatures like
		//   func(elemtype) [int,string,float,time]
		// or
		//   func(elemtype) ([int,string,float,time], error)
		mt := m.Type
		numOut := mt.NumOut()
		if mt.NumIn() != 1 || (numOut != 1 && numOut != 2) {
			return steps, fmt.Errorf("export: cannot use method %s of %T", cur, typ)
		}
		if last && superType(mt.Out(0)) == NA {
			return steps, fmt.Errorf("export: cannot use methods %s return of type %v as final element", cur, mt.Out(0))
		}
		mayFail := false
		if numOut == 2 {
			if mt.Out(1).Kind() == reflect.Interface &&
				mt.Out(1).Implements(errorInterfaceType) {
				mayFail = true
			} else {
				return steps, fmt.Errorf("export: cannot use method %s of %T", cur, typ)
			}
		}
		typ = mt.Out(0)
		s := step{name: cur, method: m.Func, mayFail: mayFail, typ: typ}
		steps = append(steps, s)
	}

	return steps, nil
}

// access drills down in v according to the given steps.
// Any nil pointer dereferenceing and method calls resulting in an non nil
// error result in an error beeing returned.
func access(v reflect.Value, steps []step) (reflect.Value, error) {
	for _, s := range steps {
		// Step down in field or method.
		if s.method.IsValid() {
			// TODO: methods on pointers?
			z := s.method.Call([]reflect.Value{v})
			if s.mayFail && z[1].Interface() != nil {
				return v, fmt.Errorf("method call failed on %s", s.name)
			}
			v = z[0]
		} else {
			v = v.Field(s.field)
		}

		// Follow all pointer indirections.
		for i := 0; i < s.indir; i++ {
			if v.IsNil() {
				return v, fmt.Errorf("nil pointer on %s", s.name)
			}
			v = reflect.Indirect(v)
		}

	}

	return v, nil
}

// retrieve decends v according to steps and returns the last value
// either as bool, int64, float64, string or time.Time.
// indir is the primary number of indirections to take.
// If no value was found due to nil pointers or method failures
// nil is returned.
func retrieve(v reflect.Value, steps []step, indir int) interface{} {
	for i := 0; i < indir; i++ {
		if v.IsNil() {
			return nil
		}
		v = reflect.Indirect(v)
	}

	res, err := access(v, steps)
	if err != nil {
		return nil
	}
	switch superType(res.Type()) {
	case Bool:
		return res.Bool()
	case Int:
		return res.Int()
	case Float:
		return res.Float()
	case String:
		return res.String()
	case Time:
		return res.Interface()
	}
	return nil
}
