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
