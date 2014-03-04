package export

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"reflect"
	"time"
)

// Interplay Extractor-Accessor-Formatter-Dumper
// Extractor constructs a list of Accessors
// Dumper writes rows to output.
// The output format depends basically on the used Dumper, i.e. only the
// Dumper knows how to format the values. So no formatter is needed.
// But it would be nice if Accessors could help here.

var m = math.Floor

func isTime(x reflect.Type) bool {
	return x.PkgPath() == "time" && x.Kind() == reflect.Struct && x.Name() == "Time"
}

// -------------------------------------------------------------------------
// Field Types

// FieldType represents the basisc type of a field.
type FieldType uint

const (
	NA FieldType = iota
	Bool
	Int
	Float
	String
	Time
)

// String representation of ft.
func (ft FieldType) String() string {
	return []string{"NA", "Bool", "Int", "Float", "String", "Time"}[ft]
}

// -------------------------------------------------------------------------
// Formating Options

// Format describes how different fields types will be formated
type Format struct {
	True, False string // String values of boolean true and false.
	IntFmt      string // Package fmt style verb for int64 printing.
	FloatFmt    string // Package fmt style verb for float64 printing.
	StringFmt   string // Package fmt style verb for string printing.
	TimeFmt     string // A package time layout string.

	// TimeLoc is the location in which times are presented.
	// If a nil TimeLoc is used the times are presented in their
	// original location.
	TimeLoc *time.Location

	NA string // Representation of a missing value.
}

// DefaultFormat are the default formating options.
var DefaultFormat = Format{
	True:      "true",
	False:     "false",
	IntFmt:    "%d",
	FloatFmt:  "%.5g",
	StringFmt: "%s",
	TimeFmt:   "2006-01-02T15:04:05.999",
	TimeLoc:   time.Local,
	NA:        "",
}

// RFormat is a useful format for dumping stuff you want to read into R.
var RFormat = Format{
	True:      "TRUE",
	False:     "FALSE",
	IntFmt:    "%d",
	FloatFmt:  "%.6g",
	StringFmt: "%s",
	TimeFmt:   "2006-01-02 15:04:05",
	TimeLoc:   time.Local,
	NA:        "NA",
}

// -------------------------------------------------------------------------
// Field

// Field represents a column in a data frame.
type Field struct {
	Name  string                  // The name of the field
	Type  FieldType               // The type of the field
	Value func(i int) interface{} // The value, maybe nil
}

//	HasValue(i int) bool // Check if a value is present

// Print the i'th entry of f according to the given format.
func (f Field) Print(format Format, i int) string {
	val := f.Value(i)
	if val == nil {
		return format.NA
	}
	switch f.Type {
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
func (d CSVDumper) Dump(e Extractor, format Format) error {
	row := make([]string, len(e.Fields))
	if !d.OmitHeader {
		for i, field := range e.Fields {
			row[i] = field.Name
		}
		d.Writer.Write(row)
	}
	for r := 0; r < e.N; r++ {
		for col, field := range e.Fields {
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

// RVecDumper dumps as a R vectors.
type RVecDumper struct {
	Writer   io.Writer
	FloatFmt string // FloatFmt may be used to change the formating of floats.
	TimeFmt  string // TimeFmt may be used to change the formating of times.

	// If Name is nonempty a data frame named Name consisting of all
	// fields is constructed too.
	Name string
}

// Dump dumps the fields from e to d.
func (d RVecDumper) Dump(e Extractor, format Format) error {
	all := ""
	for f, field := range e.Fields {
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

type Extractor struct {
	// N is the number of elements/rows/measurements.
	N int

	// Field contains all the fields, i.e. the columns to extract.
	Fields []Field
}

// NewExtractor returns an extractor for the given fieldnames of data.
// The order of fieldnames determines the columns....
func NewExtractor(data interface{}, fieldnames ...string) (Extractor, error) {
	t := reflect.TypeOf(data)
	switch t.Kind() {
	case reflect.Slice:
		return newSOMExtractor(data, fieldnames...)
	case reflect.Struct:
		panic("COS data frame not implemented")
	}
	return Extractor{}, fmt.Errorf("Cannot build Extrator for %s", t.String())
}

func canHandle(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Bool:
		return true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.String:
		return true
	case reflect.Float32, reflect.Float64:
		return true
	case reflect.Struct:
		return isTime(t)
	}
	return false
}

var (
	errorInterfaceType = reflect.TypeOf((*error)(nil)).Elem()
)

func newSOMExtractor(data interface{}, fieldnames ...string) (Extractor, error) {
	t := reflect.TypeOf(data).Elem()
	v := reflect.ValueOf(data)
	ex := Extractor{}
	ex.N = v.Len()

	for _, name := range fieldnames {
		field := Field{
			Type: NA,
			Name: name,
		}

		// Fields.
		for n := 0; n < t.NumField(); n++ {
			f := t.Field(n)
			if f.Name != name {
				continue
			}
			if !canHandle(f.Type) {
				return ex, fmt.Errorf("export: cannot usefield %q", name)
			}

			switch f.Type.Kind() {
			case reflect.Bool:
				field.Type = Bool
				field.Value = func(i int) interface{} {
					return v.Index(i).Field(n).Bool()
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				field.Type = Int
				field.Value = func(i int) interface{} {
					return v.Index(i).Field(n).Int()
				}
			case reflect.String:
				field.Type = String
				field.Value = func(i int) interface{} {
					return v.Index(i).Field(n).String()
				}
			case reflect.Float32, reflect.Float64:
				field.Type = Float
				field.Value = func(i int) interface{} {
					return v.Index(i).Field(n).Float()
				}
			case reflect.Struct: // Checked above for beeing time.Time
				field.Type = Time
				field.Value = func(i int) interface{} {
					return v.Index(i).Field(n).Interface()
				}
			}
			break
		}
		if field.Type != NA {
			ex.Fields = append(ex.Fields, field)
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
			if mt.NumIn() != 1 || (numOut != 1 && numOut != 2) ||
				!canHandle(mt.Out(0)) {
				return ex, fmt.Errorf("export: cannot use method %q", name)
			}
			mayFail := false
			if numOut == 2 {
				if mt.Out(1).Kind() == reflect.Interface &&
					mt.Out(1).Implements(errorInterfaceType) {
					mayFail = true
				} else {
					return ex, fmt.Errorf("export: cannot use method %q", name)
				}
			}

			// TODO: Move mayFail code out of function closure.
			switch mt.Out(0).Kind() {
			case reflect.Bool:
				field.Type = Bool
				field.Value = func(i int) interface{} {
					z := m.Func.Call([]reflect.Value{v.Index(i)})
					if mayFail && z[1].Interface() != nil {
						return nil
					}
					return z[0].Bool()
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				field.Type = Int
				field.Value = func(i int) interface{} {
					z := m.Func.Call([]reflect.Value{v.Index(i)})
					if mayFail && z[1].Interface() != nil {
						return nil
					}
					return z[0].Int()
				}
			case reflect.String:
				field.Type = String
				field.Value = func(i int) interface{} {
					z := m.Func.Call([]reflect.Value{v.Index(i)})
					if mayFail && z[1].Interface() != nil {
						return nil
					}
					return z[0].String()
				}
			case reflect.Float32, reflect.Float64:
				field.Type = Float
				field.Value = func(i int) interface{} {
					z := m.Func.Call([]reflect.Value{v.Index(i)})
					if mayFail && z[1].Interface() != nil {
						return nil
					}
					return z[0].Float()
				}
			case reflect.Struct: // checked above for beeing time.Time
				field.Type = Time
				field.Value = func(i int) interface{} {
					z := m.Func.Call([]reflect.Value{v.Index(i)})
					if mayFail && z[1].Interface() != nil {
						return nil
					}
					return z[0].Interface()
				}
			default:
				panic("Oooops")
			}

			break
		}
		if field.Type != NA {
			ex.Fields = append(ex.Fields, field)
			continue
		}

		return ex, fmt.Errorf("export: no such field or method %q", name)

		// TODO: Maybe pointer methods too?
		// v.Addr().MethodByName()
	}

	return ex, nil
}
