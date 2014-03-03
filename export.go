package export

import (
	"encoding"
	"encoding/csv"
	"fmt"
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
	Int
	Float
	String
	Time
)

// String representation of ft.
func (ft FieldType) String() string {
	return []string{"NA", "Int", "Float", "String", "Time"}[ft]
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

// A generic string representation
func (f Field) String(i int) string {
	val := f.Value(i)
	if val == nil {
		return "<nil>"
	}
	switch f.Type {
	case Int:
		return fmt.Sprintf("%d", val.(int64))
	case Float:
		return fmt.Sprintf("%g", val.(float64))
	case String:
		return val.(string)
	case Time:
		return val.(time.Time).Format(time.RFC3339)
	}

	if tm, ok := val.(encoding.TextMarshaler); ok {
		if text, err := tm.MarshalText(); err == nil {
			return string(text)
		} else {
			return fmt.Sprintf("Ooops %s", err)
		}
	}

	return fmt.Sprintf("Ooops: %s %v", f.Type.String(), val)
}

// -------------------------------------------------------------------------
// Dumper

type Dumper interface {
	Dump(e Extractor) error
}

type CSVDumper struct {
	Writer       *csv.Writer
	ShowHeader   bool
	MissingValue string
}

func (d CSVDumper) Dump(e Extractor) error {
	row := make([]string, len(e.Fields))
	if d.ShowHeader {
		for i, field := range e.Fields {
			row[i] = field.Name
		}
		d.Writer.Write(row)
	}
	for r := 0; r < e.N; r++ {
		for col, field := range e.Fields {
			if value := field.Value(r); value == nil {
				row[col] = d.MissingValue
			} else {
				row[col] = field.String(r)
			}
		}
		err := d.Writer.Write(row)
		if err != nil {
			return err
		}
	}
	d.Writer.Flush()
	return d.Writer.Error()
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
				return ex, fmt.Errorf("export: cannot extract field %q of type %s",
					name, f.Type.String())
			}

			switch f.Type.Kind() {
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
			if mt.NumIn() != 1 || (numOut != 1 && numOut != 2) {
				continue
			}
			mayFail := false
			if numOut == 2 && mt.Out(1).Kind() == reflect.Interface {
				if mt.Out(1).Implements(errorInterfaceType) {
					mayFail = true
					println("We may fail on ", name)
				}
			}
			if !canHandle(mt.Out(0)) {
				return ex, fmt.Errorf("export: cannot use method %q of type %s",
					name, mt.Out(0).String())
			}

			// TODO: Move mayFail code out of function closure.
			switch mt.Out(0).Kind() {
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

			if field.Type != NA {
				ex.Fields = append(ex.Fields, field)
				continue
			}
		}

		// TODO: Maybe pointer methods too?
		// v.Addr().MethodByName()
	}

	return ex, nil
}
