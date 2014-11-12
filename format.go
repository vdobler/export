// Copyright 2014 Volker Dobler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package export

import (
	"fmt"
	"math"
	"math/cmplx"
	"time"
)

// A Formater can convert baisc types to strings.
type Formater interface {
	Bool(b bool) string
	Int(i int64) string
	Float(f float64) string
	Complex(c complex128) string
	String(s string) string
	Time(t time.Time) string
	Duration(d time.Duration) string

	// NA is used to produce missing values for nil pointers or
	// method invocations which returned an error.
	NA() string
}

// Format describes how different fields types will be formated,
// either by specifying a literal representation, a package fmt
// style verb or a package time time format string.
type Format struct {
	TrueRep, FalseRep string // String values of boolean true and false.
	IntFmt            string // Package fmt style verb for int printing.
	FloatFmt          string // Package fmt style verb for float and complex printing.
	StringFmt         string // Package fmt style verb for string printing.
	TimeFmt           string // A package time layout string.
	DurationFmt       string // Either %s (human redable) or %d (nanoseconds)

	// TimeLoc is the location in which times are presented.
	// If a nil TimeLoc is used the times are presented in their
	// original location.
	TimeLoc *time.Location

	NARep            string // Representation of a missing value.
	NaNRep           string // Representation of a floating point NaN.
	PInfRep, MInfRep string // Positiv and negativ infinite. Complex uses PInf only
}

var _ Formater = Format{} // Make sure Format satisfies Formater.

func (f Format) Bool(b bool) string {
	if b {
		return f.TrueRep
	}
	return f.FalseRep
}
func (f Format) Int(i int64) string {
	return fmt.Sprintf(f.IntFmt, i)
}
func (f Format) Float(x float64) string {
	switch {
	case math.IsNaN(x):
		return f.NaNRep
	case math.IsInf(x, -1):
		return f.MInfRep
	case math.IsInf(x, +1):
		return f.PInfRep
	default:
		return fmt.Sprintf(f.FloatFmt, x)
	}
}
func (f Format) String(s string) string {
	return fmt.Sprintf(f.StringFmt, s)
}
func (f Format) Time(t time.Time) string {
	if f.TimeLoc != nil {
		t = t.In(f.TimeLoc)
	}
	return t.Format(f.TimeFmt)
}
func (f Format) Duration(d time.Duration) string {
	return fmt.Sprintf(f.DurationFmt, d)
}
func (f Format) Complex(c complex128) string {
	switch {
	case cmplx.IsNaN(c):
		return f.NaNRep
	case cmplx.IsInf(c):
		return f.PInfRep
	default:
		return fmt.Sprintf(f.FloatFmt, c)
	}
}
func (f Format) NA() string {
	return f.NARep
}

// DefaultFormat contains default formating options which produce
// pleasant human readable output.
var DefaultFormat = Format{
	TrueRep:     "true",
	FalseRep:    "false",
	IntFmt:      "%d",
	FloatFmt:    "%.4g",
	StringFmt:   "%s",
	TimeFmt:     "2006-01-02T15:04:05",
	TimeLoc:     time.Local,
	DurationFmt: "%s",
	NARep:       "",
	NaNRep:      "",
	PInfRep:     "+\u221e",
	MInfRep:     "-\u221e",
}

// PreciseFormat contains formatin options which tries to preserve
// the original data pretty well.
var PreciseFormat = Format{
	TrueRep:     "true",
	FalseRep:    "false",
	IntFmt:      "%d",
	FloatFmt:    "%g",
	StringFmt:   "%q",
	TimeFmt:     time.RFC3339Nano,
	DurationFmt: "%s",
	TimeLoc:     nil,
	NARep:       "",
	NaNRep:      "NaN",
	PInfRep:     "+\u221e",
	MInfRep:     "-\u221e",
}

// RFormat contains formating options usefull if you want to
// read the generated dumps into R.
var RFormat = Format{
	TrueRep:     "TRUE",
	FalseRep:    "FALSE",
	IntFmt:      "%d",
	FloatFmt:    "%.9g",
	StringFmt:   "%q",
	TimeFmt:     `as.POSIXct("2006-01-02 15:04:05")`,
	DurationFmt: "%d",
	TimeLoc:     time.Local, // I have no idea how timezones work in R. Sorry.
	NARep:       "NA",
	NaNRep:      "NA",
	PInfRep:     "Inf",
	MInfRep:     "-Inf",
}
