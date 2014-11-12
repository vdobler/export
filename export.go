// Copyright 2014 Volker Dobler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package export provides tools to dump tabulated data.
//
// Export allows to dump tabular data in different output formats.
// The main type is Extractor which determines which data is output and in
// which order. An Extractor is constructed from (almost) any slice type
// and may access nested fields and/or methods of the slice elements.
//
// Example
//
// Given a struct type S with a method M and a slice of S data
//
//     type S struct {
//         A int
//         B string
//         C struct{T time.Time}
//     }
//
//     func (s S) M() float64 { return float64(s.A)/2 }
//
//     data := []S{
//         {4, "Hello"},
//         {5, "World!"},
//     }
//
// an Extractor ex for data could be constructed like
//
//     ex, _ := NewExtractor(data, "B", "M()", "A", "C.T", "C.T.Day()")
//
// This Extractor can be used to dump data in CSV format like this:
//
//     csvdumper := CSVDumper{Writer: csv.NewWriter(os.Stdout)}
//     csvdumper.Dump(ex, DefaultFormat)
//
// Column Specifiers
//
// A columns specifier during construction of an Extractor determines which
// field, method, nested field, method on nested field, and so on shall be
// exported:
//   - Only exported fields can be exported.
//   - Accessing a nested field (in the example T) inside a field (C in the
//     example) is written as T.C
//   - Methods require "()" in the columne specifier (here "M()").
//   - Methods may not take arguments.
//   - Only methods returnig one value or a (value, error) pair may
//     be used.
//   - Pointers are dereferenced automatically.
//   - Nil Pointers and method calls returning a non-nil error result in
//     a NA value for this field.
//
// The final field (or the type returned by a final method call) must be
// one of:
//   - bool
//   - uint8, uint16, ...,  int64
//   - float32 and float64
//   - complex64 and complex128
//   - string
//   - time.Time and time.Duration
//
// This package handles floats and int as 64bit values and complex values
// as complex128. Thus an uint64 may overflow without notice.
//
// Dumping
//
// Dumping the data bound to an Extractor is done via a Dumper. This package
// provides three types: CSVDumper, TabDumper and RVecDumper. It is the
// dumpers responsibility to iterate over the rows and columns of an Extractor
// and generating values via the the Columns Print method which takes a
// Formater which does the actual string generation.
package export

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

// -------------------------------------------------------------------------
// Extractor

// Extractor provides access to fields and methods of tabular data.
// An extractor must be constructed with NewExtractor and can be rebound
// to new data sets anytime by Bind.
type Extractor struct {
	// N is the numer of elements in the currently bound data.
	N int

	// Columns contains all the columns to extract. After
	// creation of an Extractor Columns may be manipulated, e.g.
	// setting a custom name for a column or rearanging or dropping
	// columns.
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

// -------------------------------------------------------------------------
// Type and Column

// Type represents the basic type of a column.
type Type uint

const (
	NA Type = iota
	Bool
	Int
	Float
	Complex
	String
	Time
	Duration
)

// String returns the name of t.
func (t Type) String() string {
	return []string{"NA", "Bool", "Int", "Float", "Complex", "String",
		"Time", "Duration"}[t]
}

// Column represents one column in the export. Columns are created
// during construction of an Extractor only.
type Column struct {
	// Name is the name of the column. It is created based on the
	// column spec during construction of a new Extractor and may
	// be changed afterwards.
	Name string

	typ Type // The type of the column.

	// value returns the i'th value in this column.
	// For errors or nil pointers nil is returned.
	value func(i int) interface{}

	access   []step // The steps needed to access the result.
	unsigned bool   // For Type == Int
}

// Type returns the type of the column c.
func (c Column) Type() Type { return c.typ }

// Print the i'th entry of column c with the given format.
func (c Column) Print(f Formater, i int) string {
	val := c.value(i)
	if val == nil {
		return f.NA()
	}
	switch c.typ {
	case Bool:
		return f.Bool(val.(bool))
	case Int:
		return f.Int(val.(int64))
	case Float:
		return f.Float(val.(float64))
	case Complex:
		return f.Complex(val.(complex128))
	case String:
		return f.String(val.(string))
	case Time:
		return f.Time(val.(time.Time))
	case Duration:
		return f.Duration(val.(time.Duration))
	}

	return fmt.Sprintf("%v", val)
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
		steps, rType, unsigned, err := buildSteps(typ, spec)
		if err != nil {
			return nil, err
		}
		name := ""
		for s := range steps {
			if s > 0 {
				name += "."
			}
			name += steps[s].name
		}

		field := Column{
			Name:     name,
			typ:      rType,
			access:   steps,
			unsigned: unsigned,
		}
		ex.Columns = append(ex.Columns, field)
	}

	return &ex, nil
}

// bindSOM is the slice-of-measurements version of Bind.
func (e *Extractor) bindSOM(data interface{}) {
	v := reflect.ValueOf(data)
	e.N = v.Len()
	for fn, field := range e.Columns {
		access := field.access
		typ := field.Type()
		unsigned := field.unsigned
		e.Columns[fn].value = func(i int) interface{} {
			return retrieve(v.Index(i), access, e.indir, typ, unsigned)
		}
	}
}

// superType returns our types which group Go's low level types.
// A Go type which cannot be handled will yield a Type of NA.
// TODO: this might be the worst name possible for this function.
func superType(t reflect.Type) Type {
	switch t.Kind() {
	case reflect.Bool:
		return Bool
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if isDuration(t) {
			return Duration
		}
		return Int
	case reflect.String:
		return String
	case reflect.Float32, reflect.Float64:
		return Float
	case reflect.Complex64, reflect.Complex128:
		return Complex
	case reflect.Struct:
		if isTime(t) {
			return Time
		}
	}
	return NA
}

func isTime(x reflect.Type) bool {
	return x.PkgPath() == "time" && x.Kind() == reflect.Struct && x.Name() == "Time"
}
func isDuration(x reflect.Type) bool {
	return x.PkgPath() == "time" && x.Kind() == reflect.Int64 && x.Name() == "Duration"
}

var (
	errorInterface    = reflect.TypeOf((*error)(nil)).Elem()
	stringerInterface = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()
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
	// typ     reflect.Type
}

func (s step) isMethodCall() bool { return s.method.IsValid() }

// buildSteps constructs a slice of steps to access the given elem in typ.
// The Type of the final element is returend and whether the final element
// has to be converted first.
func buildSteps(typ reflect.Type, elem string) ([]step, Type, bool, error) {
	var steps []step
	elements := strings.Split(elem, ".")
	for _, cur := range elements {
		var s step
		var err error
		if strings.HasSuffix(cur, "()") {
			cur = cur[:len(cur)-2]
			s, typ, err = methodStep(cur, typ)
			if err != nil {
				return nil, NA, false, err
			}
		} else {
			s, typ, err = fieldStep(cur, typ)
			if err != nil {
				return nil, NA, false, err
			}
		}
		steps = append(steps, s)
	}

	finalType := superType(typ)
	unsigned := false

	if finalType == NA {
		// Maybe typ implements fmt.Stringer in which case we
		// append an extra String method step.
		if typ.Implements(stringerInterface) {
			m, _ := typ.MethodByName("String")
			s := step{
				name:   "String",
				method: m.Func,
			}
			steps = append(steps, s)
		} else {
			return steps, NA, false,
				fmt.Errorf("export: cannot use type %T", typ)
		}
	} else if finalType == Int {
		switch typ.Kind() {
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			unsigned = true
		}
	}

	return steps, finalType, unsigned, nil
}

// fieldStep tries to construct step on typ with the given field.
func fieldStep(fieldName string, typ reflect.Type) (step, reflect.Type, error) {
	if typ.Kind() != reflect.Struct {
		return step{}, typ, fmt.Errorf("export: type %s is not a struct", typ)
	}

	var fn int = -1
	var field reflect.StructField
	for i := 0; i < typ.NumField(); i++ {
		if typ.Field(i).Name == fieldName {
			fn = i
			field = typ.Field(i)
			break
		}
	}
	if fn == -1 {
		return step{}, typ, fmt.Errorf("export: type %s has no field %s",
			typ, fieldName)
	}

	typ = field.Type
	indir := 0
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		indir++
	}
	s := step{
		name:  fieldName,
		field: fn,
		indir: indir,
	}
	return s, typ, nil
}

// methodStep tries to construct step on typ with the given methodName.
// It looks for methods with signatures like
//   func(elemtype) [bool,int,string,float,time]
// or
//   func(elemtype) ([bool,int,string,float,time], error)
func methodStep(methodName string, typ reflect.Type) (step, reflect.Type, error) {
	m, ok := typ.MethodByName(methodName)
	if !ok {
		return step{}, typ,
			fmt.Errorf("export: no method %s in %s", methodName, typ)
	}

	mt := m.Type
	numOut := mt.NumOut()
	if mt.NumIn() != 1 || (numOut != 1 && numOut != 2) {
		return step{}, typ, fmt.Errorf("export: cannot use method %s of %s",
			methodName, typ)
	}
	mayFail := false
	if numOut == 2 {
		if mt.Out(1).Kind() == reflect.Interface &&
			mt.Out(1).Implements(errorInterface) {
			mayFail = true
		} else {
			return step{}, typ, fmt.Errorf("export: cannot use method %s of %s",
				methodName, typ)
		}
	}
	typ = mt.Out(0)
	s := step{
		name:    methodName,
		method:  m.Func,
		mayFail: mayFail,
	}
	return s, typ, nil
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
// either as bool, int64, float64, complex128, string, time.Time or time.Duration
// indir is the primary number of indirections to take.
// If no value was found due to nil pointers or method failures
// nil is returned.
func retrieve(v reflect.Value, steps []step, indir int, typ Type, unsigned bool) interface{} {
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
	switch typ {
	case Bool:
		return res.Bool()
	case Int:
		if unsigned {
			return int64(res.Uint())
		} else {
			return res.Int()
		}
	case Float:
		return res.Float()
	case Complex:
		return res.Complex()
	case String:
		return res.String()
	case Time:
		return res.Interface()
	case Duration:
		return time.Duration(res.Int())
	}
	return nil
}
