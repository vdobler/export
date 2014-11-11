// Copyright 2014 Volker Dobler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package export

import (
	"time"
)

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

	NA  string // Representation of a missing value.
	NaN string // Representation of a floating point NaN.
}

// DefaultFormat contains default formating options.
var DefaultFormat = Format{
	True:      "true",
	False:     "false",
	IntFmt:    "%d",
	FloatFmt:  "%.4g",
	StringFmt: "%s",
	TimeFmt:   "2006-01-02T15:04:05",
	TimeLoc:   time.Local,
	NA:        "",
	NaN:       "",
}

// PreciseFormat contains formatin options which tries to preserve
// the original data very well.
var PreciseFormat = Format{
	True:      "true",
	False:     "false",
	IntFmt:    "%d",
	FloatFmt:  "%g",
	StringFmt: "%q",
	TimeFmt:   time.RFC3339Nano,
	TimeLoc:   nil,
	NA:        "",
	NaN:       "NaN",
}

// RFormat contains formating options usefull if you want to
// read the generated dumps into R.
var RFormat = Format{
	True:      "TRUE",
	False:     "FALSE",
	IntFmt:    "%d",
	FloatFmt:  "%.9g",
	StringFmt: "%q",
	TimeFmt:   "2006-01-02 15:04:05",
	TimeLoc:   time.Local,
	NA:        "NA",
	NaN:       "NA",
}
