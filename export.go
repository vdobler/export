package export

import (
	"fmt"
	"io"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"encoding"
	"encoding/csv"
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
type Field interface {
	Name() string  // The name of the field
	Type()   FieldType // The type of the field
	Value(i int) interface{} // The value, maybe nil
	HasValue(i int) bool // Check if a value is present
	String(i int) string // A generic string representation
	Int(i int) int64
	Float(i int) float64
	Time(i int) time.Time
}

// -------------------------------------------------------------------------
// Dumper

type Dumper interface {
	Dump(e Extractor) error
}

type CSVDumper struct {
	Writer csv.Writer
	ShowHeader bool
	MissingValue string
}

func (d CSVDumper) Dump(e Extractor) error {
	row := make([]string, e.N)
	if d.ShowHeader {
		for i, field := range e.Fields {
			row[i] = field.Name()
		}
		d.Writer.Write(row)
	}
	for row := 0; row < e.N; row++ {
		for col, field := range e.Fields {
			if value := field.Value(row); value == nil {
				row[col] = d.MissingValue
			} else {
				row[col] = field.String(row)
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
		return newSOMExtractor(data, fieldnames ...string)
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

type intField struct {
	
}

func newSOMExtractor(data interface{}, fieldnames ...string) (Extractor, error) {
	t := reflect.TypeOf(data).Elem()
	v := reflect.ValueOf(data)
	n := v.Len()
	ex := Extractor{}
	ex.N = n

	for _, fieldname := range fieldnames {
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.Name != fieldname {
				continue
			}
			if !canHandle(f.Type) {
				return ex, fmt.Errorf("Cannot extract field %q of type %s", fieldname, f.Type.String())
			}

			switch f.Type.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				field.Type = Int
				field.Origin = 0
				for j := 0; j < n; j++ {
					field.Data[j] = float64(v.Index(j).FieldByName(f.Name).Int())
				}
			case reflect.String:
				field.Type = String
				field.Origin = 0
				for j := 0; j < n; j++ {
					s := v.Index(j).FieldByName(f.Name).String()
					field.Data[j] = float64(pool.Add(s))
				}
			case reflect.Float32, reflect.Float64:
				field.Type = Float
				field.Origin = 0
				for j := 0; j < n; j++ {
					field.Data[j] = v.Index(j).FieldByName(f.Name).Float()
				}
			case reflect.Struct: // Checked above for beeing time.Time
				field.Type = Time
				if n > 0 {
					field.Origin = v.Index(0).FieldByName(f.Name).Interface().(time.Time).Unix()
				}
				for j := 0; j < n; j++ {
					t := v.Index(j).FieldByName(f.Name).Interface().(time.Time).Unix()
					field.Data[j] = float64(t - field.Origin)
				}
			}
			df.Columns[f.Name] = field
			
		}
	}

	// Fields first.
	for i := 0; i < t.NumField(); i++ {

		switch f.Type.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			field.Type = Int
			field.Origin = 0
			for j := 0; j < n; j++ {
				field.Data[j] = float64(v.Index(j).FieldByName(f.Name).Int())
			}
		case reflect.String:
			field.Type = String
			field.Origin = 0
			for j := 0; j < n; j++ {
				s := v.Index(j).FieldByName(f.Name).String()
				field.Data[j] = float64(pool.Add(s))
			}
		case reflect.Float32, reflect.Float64:
			field.Type = Float
			field.Origin = 0
			for j := 0; j < n; j++ {
				field.Data[j] = v.Index(j).FieldByName(f.Name).Float()
			}
		case reflect.Struct: // Checked above for beeing time.Time
			field.Type = Time
			if n > 0 {
				field.Origin = v.Index(0).FieldByName(f.Name).Interface().(time.Time).Unix()
			}
			for j := 0; j < n; j++ {
				t := v.Index(j).FieldByName(f.Name).Interface().(time.Time).Unix()
				field.Data[j] = float64(t - field.Origin)
			}
		}
		df.Columns[f.Name] = field

		// println("newSOMDataFrame: added field Name =", f.Name, "   type =", f.Type.String())

	}

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

	// TODO: Maybe pointer methods too?
	// v.Addr().MethodByName()

	return df, nil
}

// Filter extracts all rows from df where field==value.
// TODO: allow ranges
func Filter(df *DataFrame, field string, value interface{}) *DataFrame {
	if df == nil {
		return nil
	}

	dfft, ok := df.Columns[field]
	if !ok {
		// TODO: warn somhow...
		return df.Copy()
	}

	// Convert generic value into the float used for comparison.
	var floatVal float64
	switch reflect.TypeOf(value).Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		floatVal = float64(reflect.ValueOf(value).Int())
	case reflect.String:
		sidx := df.Pool.Find(value.(string))
		if sidx == -1 {
			return nil // TODO: is this sensible?
		}
		floatVal = float64(sidx)
	case reflect.Float32, reflect.Float64:
		floatVal = float64(reflect.ValueOf(value).Float())
	case reflect.Struct:
		if !isTime(reflect.TypeOf(value)) {
			panic("Bad type of value" + reflect.TypeOf(value).String())
		}
		floatVal = float64(value.(time.Time).Unix() - dfft.Origin)
	default:
		panic("Bad type of value" + reflect.TypeOf(value).String())
	}

	result := NewDataFrame(fmt.Sprintf("%s|%s=%v", df.Name, field, value), df.Pool)

	// How many rows will be in the result data frame?
	col := df.Columns[field].Data
	result.N = 0
	for i := 0; i < df.N; i++ {
		if col[i] != floatVal {
			continue
		}
		result.N++
	}

	// Actual filtering.
	for name, field := range df.Columns {
		f := field.CopyMeta()
		f.Data = make([]float64, result.N)
		n := 0
		for i := 0; i < df.N; i++ {
			if col[i] != floatVal {
				continue
			}
			f.Data[n] = field.Data[i]
			n++
		}
		result.Columns[name] = f
	}

	return result
}

// Sorting of int64 slices.
type IntSlice []int64

func (p IntSlice) Len() int           { return len(p) }
func (p IntSlice) Less(i, j int) bool { return p[i] < p[j] }
func (p IntSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func SortInts(a []int64)              { sort.Sort(IntSlice(a)) }

// Levels returns the levels of field.
func Levels(df *DataFrame, field string) FloatSet {
	if df == nil {
		return NewFloatSet()
	}
	t, ok := df.Columns[field]
	if !ok {
		panic(fmt.Sprintf("No such field %q in data frame %q.", field, df.Name))
	}
	if !t.Discrete() {
		panic(fmt.Sprintf("Field %q (%s) in data frame %q is not discrete.",
			field, t.Type, df.Name))
	}

	return df.Columns[field].Levels()
}

func (f Field) Levels() FloatSet {
	if !f.Discrete() {
		panic("Called Levels on non-discrete Field")
	}
	levels := NewFloatSet()
	for _, v := range f.Data {
		levels.Add(v)
	}

	return levels
}

// MinMax returns the minimum and maximum element and their indixes.
func MinMax(df *DataFrame, field string) (minval, maxval float64, minidx, maxidx int) {
	if df == nil {
		return math.NaN(), math.NaN(), -1, -1
	}
	_, ok := df.Columns[field]
	if !ok {
		return math.NaN(), math.NaN(), -1, -1
	}

	return df.Columns[field].MinMax()
}

func (f Field) MinMax() (minval, maxval float64, minidx, maxidx int) {
	if len(f.Data) == 0 {
		println("MinMax", f.Type.String(), ": no data -> NaN")
		return math.NaN(), math.NaN(), -1, -1
	}

	column := f.Data
	minval, maxval = column[0], column[0]
	// println("min/max start", minval, maxval)
	minidx, maxidx = 0, 0
	for i, v := range column {
		// println("  ", v)
		if v < minval {
			minval, minidx = v, i
			// println("    lower")
		} else if v > maxval {
			maxval, maxidx = v, i
			// println("    higher")
		}
	}

	return minval, maxval, minidx, maxidx
}

func (df *DataFrame) Print(out io.Writer) {
	names := df.FieldNames()

	fmt.Fprintf(out, "Data Frame %q:\n", df.Name)

	w := new(tabwriter.Writer)
	w.Init(out, 0, 8, 2, ' ', 0)
	for _, name := range names {
		fmt.Fprintf(w, "\t%s", name)
	}
	fmt.Fprintln(w)
	for i := 0; i < df.N; i++ {
		fmt.Fprintf(w, "%d", i)
		for _, name := range names {
			field := df.Columns[name]
			var s string
			s = field.String(field.Data[i])
			fmt.Fprintf(w, "\t%s", s)
		}
		fmt.Fprintln(w)
	}
	w.Flush()

}

// GroupingField constructs a new Field of type String with the same length
// as data. The values are the concationation of the named columns.
// The named columns in data must be discrete.
func GroupingField(data *DataFrame, names []string) Field {
	// Check names
	for _, n := range names {
		if f, ok := data.Columns[n]; !ok {
			panic(fmt.Sprintf("Data frame %q has no column %q to group by.",
				data.Name, n))
		} else if !f.Discrete() {
			panic(fmt.Sprintf("Column %q in data frame %q is of type %s and cannot be used for grouping",
				n, data.Name, f.Type))
		}
	}

	field := NewField(data.N, String, data.Pool)
	for i := 0; i < data.N; i++ {
		group := ""
		for _, name := range names {
			f := data.Columns[name]
			val := f.Data[i]
			if group != "" {
				group += " | " // TODO: ist this clever? No. Maybe int-Type?
			}
			group += f.String(val)
		}
		field.Data[i] = float64(data.Pool.Add(group))
	}
	return field
}

func (f Field) Resolution() float64 {
	resolution := math.Inf(+1)
	d := f.Data
	for i := 0; i < len(f.Data)-1; i++ {
		r := math.Abs(d[i] - d[i+1])
		if r < resolution {
			resolution = r
		}
	}
	return resolution
}

// Partition df.
func Partition(df *DataFrame, field string, levels []float64) []*DataFrame {
	part := make([]*DataFrame, len(levels))
	idx := make(map[float64]int)
	for i, level := range levels {
		part[i] = df.CopyMeta()
		part[i].Delete(field)
		idx[level] = i
	}

	fc := df.Columns[field].Data
	for j := 0; j < df.N; j++ {
		level := fc[j]
		i := idx[level]
		for name, f := range df.Columns {
			if name == field {
				continue
			}
			t := part[i].Columns[name]
			t.Data = append(t.Data, f.Data[j])
			part[i].N = len(t.Data)
			part[i].Columns[name] = t
		}
	}

	return part
}
