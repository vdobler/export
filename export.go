// Copyright Volker Dobler 2014

// Package export provides tools to dump tabulated data.
//
// Export allows to dump tabular data in different output formats.
// The main type is Exporter which determines which data is output
// and in which order. An Exporter is constructed from (almost)
// any slice type.  Take a struct S with two fields A and B and two
// methods C and D and.
//
//     type S struct {
//         A float64
//         B string
//     }
//
//     func (s S) C() int {
//         return int(s.A+0.5)
//     }
//
//     func (s S) D() (bool, error) {
//         if s.B == "" {
//             return false, errors.New("empty")
//         }
//         return len(s.B) > 5, nil
//     }
//
//     var data = []S{S{3.14, "Hello"}, S{55.5, ""}}
//     exp, err := NewExtractor("C", "D", "A", B")
//     dumper := TabDumper{Writer: os.Stdout}
//     dumper.dump(exp)
//
// Would produce a text table and output the collumns C, D, A and B.
// Note that column D may contain NA for "not avialable".
//
package export


// Stuff is getting complicated once there are pointers involved.
//     data := []*S{...}
//     type S struct {
//         A *int
//         B struct { C bool; D string}
           E *T
//     }
//     type T struct {
//             F string
//         }
//     }
//     func (t T) M() float64 { return 3.14 }
//     type U struct { G int64 }
//     func (t T) N() (*U, error) { return &U{123}, nil }
// Might be interesting to extract from nested structs like
//     NewExtractor{"A", "B.C", "E.F", "E.M", "E.N.G"}
// So we would have:
//   - plain field: a leaf
//   - ptr to plain field: may fail, one to go for a leaf
//   - nested struct
//   - ptr to nested struct
//   - ptr to ptr to ... ? (realy needed)
//   - method call
// That means a columns spec like "E.N.G" above translates to
// an access rule for row n like:
//   - take element n
//   - deref ptr (to get a S) if non nil, else --> nil
//   - deref E (to get a T) if non nil, else --> nil
//   - call N (to get a *U) if err --> nil
//   - deref ret (to get a U) if non nil, else --> nil
//   - return N
// Each step can be summarized as working on cur
//   If cur is Method:
//       cur = call method
//       if mayFail && hasErr  -->  nil
//       restart
//   If a ptr:
//       If nil  -->  nil
//       Else cur = deref cur
//       restart
//   If struct:
//       cur = field from next part in colspec
//       restart
//   If field:
//       If known type --> return
//       If assignable to known type --> return
//       If encodingTextMarshal-able --> do it an return
//       Printf %v --> return.
// NewExtractor would build such a list of access rules and Column.Value
// would execute this rule list. Pre-Built closures are of limits.
// It would be basically what encoding/gob does, just without cycle
// detection and other fancyness.
//

import (
	"encoding/csv"
	"fmt"
	"io"
	"reflect"
	"text/tabwriter"
	"time"
)

func isTime(x reflect.Type) bool {
	return x.PkgPath() == "time" && x.Kind() == reflect.Struct && x.Name() == "Time"
}

// -------------------------------------------------------------------------
// Type

// Type represents the basisc type of a columne.
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

// Column
type Column struct {
	Name  string                  // The name of the field
	Type  Type                    // The type of the field
	Value func(i int) interface{} // The value, maybe nil

	MayFail bool

	fieldNo int           // >= 0 ==> a field; <0 ==> a method
	mfunc   reflect.Value // the function of the method if fieldNo < 0
	ptr     bool          // for pointer fields
}

// Print the i'th entry of f according to the given format.
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

type Dumper interface {
	Dump(e Extractor, format Format) error
}

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

	// som is true fro slice-of-measurement type data
	som bool

	// typ contains the go type this Extractor can work on.
	typ reflect.Type
}

// NewExtractor returns an extractor for the given column names of data.
func NewExtractor(data interface{}, columnnames ...string) (*Extractor, error) {
	t := reflect.TypeOf(data)
	switch t.Kind() {
	case reflect.Slice:
		ex, err := newSOMExtractor(data, columnnames...)
		if err != nil {
			return ex, err
		}
		ex.som = true
		ex.bindSOM(data)
		return ex, nil
	case reflect.Struct:
		panic("COS data frame not implemented")
	}
	return &Extractor{}, fmt.Errorf("Cannot build Extrator for %s", t.String())
}

// Bind (re)binds e to data which must be of the same type as the data used
// during the construction of e.
func (e *Extractor) Bind(data interface{}) {
	t := reflect.TypeOf(data).Elem()
	if t != e.typ {
		panic(fmt.Sprintf("Cannot bind extractor for %s to data of type %s", e.typ, t))
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
		n := field.fieldNo
		f := field.mfunc
		switch {
		case n >= 0 && !field.ptr:
			// Plain field access
			switch field.Type {
			case Bool:
				e.Columns[fn].Value = func(i int) interface{} {
					return v.Index(i).Field(n).Bool()
				}
			case Int:
				e.Columns[fn].Value = func(i int) interface{} {
					return v.Index(i).Field(n).Int()
				}
			case Float:
				e.Columns[fn].Value = func(i int) interface{} {
					return v.Index(i).Field(n).Float()
				}
			case String:
				e.Columns[fn].Value = func(i int) interface{} {
					return v.Index(i).Field(n).String()
				}
			case Time:
				e.Columns[fn].Value = func(i int) interface{} {
					return v.Index(i).Field(n).Interface()
				}
			}
		case n >= 0 && field.ptr:
			// Pointer field.
			switch field.Type {
			case Bool:
				e.Columns[fn].Value = func(i int) interface{} {
					z := v.Index(i).Field(n)
					if z.IsNil() {
						return nil
					}
					return reflect.Indirect(z).Bool()
				}
			case Int:
				e.Columns[fn].Value = func(i int) interface{} {
					z := v.Index(i).Field(n)
					if z.IsNil() {
						return nil
					}
					return reflect.Indirect(z).Int()
				}
			case Float:
				e.Columns[fn].Value = func(i int) interface{} {
					z := v.Index(i).Field(n)
					if z.IsNil() {
						return nil
					}
					return reflect.Indirect(z).Float()
				}
			case String:
				e.Columns[fn].Value = func(i int) interface{} {
					z := v.Index(i).Field(n)
					if z.IsNil() {
						return nil
					}
					return reflect.Indirect(z).String()
				}
			case Time:
				e.Columns[fn].Value = func(i int) interface{} {
					z := v.Index(i).Field(n)
					if z.IsNil() {
						return nil
					}
					return reflect.Indirect(z).Interface()
				}
			}
		case n < 0 && field.MayFail:
			// Method access with possible failure
			switch field.Type {
			case Bool:
				e.Columns[fn].Value = func(i int) interface{} {
					z := f.Call([]reflect.Value{v.Index(i)})
					if z[1].Interface() != nil {
						return nil
					}
					return z[0].Bool()
				}
			case Int:
				e.Columns[fn].Value = func(i int) interface{} {
					z := f.Call([]reflect.Value{v.Index(i)})
					if z[1].Interface() != nil {
						return nil
					}
					return z[0].Int()
				}
			case Float:
				e.Columns[fn].Value = func(i int) interface{} {
					z := f.Call([]reflect.Value{v.Index(i)})
					if z[1].Interface() != nil {
						return nil
					}
					return z[0].Float()
				}
			case String:
				e.Columns[fn].Value = func(i int) interface{} {
					z := f.Call([]reflect.Value{v.Index(i)})
					if z[1].Interface() != nil {
						return nil
					}
					return z[0].String()
				}
			case Time:
				e.Columns[fn].Value = func(i int) interface{} {
					z := f.Call([]reflect.Value{v.Index(i)})
					if z[1].Interface() != nil {
						return nil
					}
					return z[0].Interface()
				}
			}
		case n < 0 && !field.MayFail:
			// Method access without failure
			switch field.Type {
			case Bool:
				e.Columns[fn].Value = func(i int) interface{} {
					return f.Call([]reflect.Value{v.Index(i)})[0].Bool()
				}
			case Int:
				e.Columns[fn].Value = func(i int) interface{} {
					return f.Call([]reflect.Value{v.Index(i)})[0].Int()
				}
			case Float:
				e.Columns[fn].Value = func(i int) interface{} {
					return f.Call([]reflect.Value{v.Index(i)})[0].Float()
				}
			case String:
				e.Columns[fn].Value = func(i int) interface{} {
					return f.Call([]reflect.Value{v.Index(i)})[0].String()
				}
			case Time:
				e.Columns[fn].Value = func(i int) interface{} {
					return f.Call([]reflect.Value{v.Index(i)})[0].Interface()
				}
			}
		default:
			println(n, field.MayFail, field.ptr)
		}
	}
}

// superType returns our types which group Go's low level types.
// A Go type which cannot be handled will yield NA.
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

// newSOMExtractor sets up an unbound Extractor for a slice-of-measurements
// type data.
func newSOMExtractor(data interface{}, fieldnames ...string) (*Extractor, error) {
	t := reflect.TypeOf(data).Elem()
	ex := Extractor{}
	ex.typ = t

	for _, name := range fieldnames {
		field := Column{
			Type: NA,
			Name: name,
		}
		// Fields.
		for n := 0; n < t.NumField(); n++ {
			f := t.Field(n)
			if f.Name != name {
				continue
			}
			var st Type
			if f.Type.Kind() == reflect.Ptr {
				field.ptr = true
				st = superType(f.Type.Elem())
			} else {
				st = superType(f.Type)
			}
			if st == NA {
				return &ex, fmt.Errorf("export: cannot use field %q type is %#v", name, f.Type)
			}
			field.MayFail = false
			field.fieldNo = n
			field.Type = st
			break
		}
		if field.Type != NA {
			ex.Columns = append(ex.Columns, field)
			continue
		}

		// The same for methods.
		// TODO: As n is not used in the field value closures a simple
		// MethodByName would be much nicer here.
		for n := 0; n < t.NumMethod(); n++ {
			m := t.Method(n)
			if m.Name != name {
				continue
			}
			// Look for methods with signatures like
			//   func(elemtype) [int,string,float,time]
			// or
			//   func(elemtype) ([int,string,float,time], error)
			mt := m.Type
			numOut := mt.NumOut()
			st := superType(mt.Out(0))
			if mt.NumIn() != 1 || (numOut != 1 && numOut != 2) || st == NA {
				return &ex, fmt.Errorf("export: cannot use method %q", name)
			}
			mayFail := false
			if numOut == 2 {
				if mt.Out(1).Kind() == reflect.Interface &&
					mt.Out(1).Implements(errorInterfaceType) {
					mayFail = true
				} else {
					return &ex, fmt.Errorf("export: cannot use method %q", name)
				}
			}
			field.MayFail = mayFail
			field.fieldNo = -1
			field.mfunc = m.Func
			field.Type = st
			break
		}
		if field.Type != NA {
			ex.Columns = append(ex.Columns, field)
			continue
		}

		return &ex, fmt.Errorf("export: no such field or method %q", name)

		// TODO: Maybe pointer methods too?
		// v.Addr().MethodByName()
	}

	return &ex, nil
}
