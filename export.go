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
	Writer       csv.Writer
	ShowHeader   bool
	MissingValue string
}

func (d CSVDumper) Dump(e Extractor) error {
	row := make([]string, e.N)
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

func newSOMExtractor(data interface{}, fieldnames ...string) (Extractor, error) {
	t := reflect.TypeOf(data).Elem()
	v := reflect.ValueOf(data)
	n := v.Len()
	ex := Extractor{}
	ex.N = n

	// Direkct struct fields first.
	for index, name := range fieldnames {
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.Name != name {
				continue
			}
			if !canHandle(f.Type) {
				return ex, fmt.Errorf("export: cannot extract field %q of type %s",
					name, f.Type.String())
			}

			field := Field{Name: name}
			switch f.Type.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				field.Type = Int
				field.Value = func(i int) interface{} {
					return v.Index(i).Field(index).Int()
				}
			case reflect.String:
				field.Type = String
				field.Value = func(i int) interface{} {
					return v.Index(i).Field(index).String()
				}
			case reflect.Float32, reflect.Float64:
				field.Type = Float
				field.Value = func(i int) interface{} {
					return v.Index(i).Field(index).Float()
				}
			case reflect.Struct: // Checked above for beeing time.Time
				field.Type = Time
				field.Value = func(i int) interface{} {
					return v.Index(i).Field(index).Interface()
				}
			}
			ex.Fields = append(ex.Fields, field)
			break
		}
	}

	/*
		// The same for methods.
		for i := 0; i < t.NumMethod(); i++ {
			m := t.Method(i)

			// Look for methods with signatures like "func(elemtype) [int,string,float,time]"
			mt := m.Type
			if mt.NumIn() != 1 || mt.NumOut() != 1 {
				continue
			}
			switch mt.Out(0).Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			case reflect.String:
			case reflect.Float32, reflect.Float64:
			case reflect.Struct:
				if !isTime(mt.Out(0)) {
					continue
				}
			default:
				continue
			}

			field := Field{
				Data: make([]float64, n),
				Pool: pool,
			}

			switch mt.Out(0).Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				field.Type = Int
				for j := 0; j < n; j++ {
					field.Data[j] = float64(m.Func.Call([]reflect.Value{v.Index(j)})[0].Int())
				}
			case reflect.String:
				field.Type = String
				for j := 0; j < n; j++ {
					s := m.Func.Call([]reflect.Value{v.Index(j)})[0].String()
					field.Data[j] = float64(pool.Add(s))
				}
			case reflect.Float32, reflect.Float64:
				field.Type = Float
				for j := 0; j < n; j++ {
					field.Data[j] = m.Func.Call([]reflect.Value{v.Index(j)})[0].Float()
				}
			case reflect.Struct: // checked above for beeing time.Time
				field.Type = Float
				if n > 0 {
					field.Origin = m.Func.Call([]reflect.Value{v.Index(0)})[0].Interface().(time.Time).Unix()
				}
				for j := 0; j < n; j++ {
					t := m.Func.Call([]reflect.Value{v.Index(j)})[0].Interface().(time.Time).Unix()
					field.Data[j] = float64(t - field.Origin)
				}
			default:
				panic("Oooops")
			}

			df.Columns[m.Name] = field

			// println("newSOMDataFrame: added method Name =", m.Name, "   type =", df.Type[m.Name].String())
		}

	*/
	// TODO: Maybe pointer methods too?
	// v.Addr().MethodByName()

	return ex, nil
}
