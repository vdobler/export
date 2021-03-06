// Copyright 2014 Volker Dobler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package export

import (
	"errors"
	"os"
	"text/tabwriter"
	"time"
)

// Some is some structure.
type Some struct {
	Flt    float64
	Str    string
	IntP   *int
	Other  Other
	OtherP *Other
}

func (s Some) Method1() int {
	return int(s.Flt + 0.5)
}

// Method2 may fail.
func (s Some) Method2() (bool, error) {
	if s.Str == "" {
		return false, errors.New("empty")
	}
	return len(s.Str) > 5, nil
}

type Other struct {
	Start time.Time
}

func (o Other) Unix() int64 {
	return o.Start.Unix()
}

func Example() {
	// Set up some values.
	eight, nine := 8, 9
	t0 := time.Date(2009, 12, 28, 8, 45, 0, 0, time.UTC)
	t1 := time.Date(2014, 12, 12, 23, 59, 59, 0, time.UTC)
	t2 := time.Date(2099, 1, 1, 0, 1, 0, 0, time.UTC)

	// Data is a slice of Some things.
	var data = []Some{
		Some{3.14, "Hello", &eight, Other{t0}, &Other{t1}},
		Some{2.72, "Go", nil, Other{t1}, &Other{t2}},
		Some{1.41, "", &nine, Other{t2}, nil},
	}
	extractor, err := NewExtractor(data,
		"Flt", "Str", "IntP", // Accessing fields and pointer fields
		"Method1()", "Method2()", // Accessing results of methods.
		"Other.Start", "OtherP.Unix()", // Accessing nested elements.
		"Other.Start.Day()") // Accessing methods on nested elements.
	if err != nil {
		panic(err.Error())
	}

	// Rename the last column which defaults to "Other.Start.Day".
	extractor.Columns[7].Name = "DayOfMonth"

	w := &tabwriter.Writer{}
	w.Init(os.Stdout, 1, 8, 1, ' ', 0)
	tab := TabDumper{Writer: w}
	format := DefaultFormat // A human readable format. Missing values are omited.
	format.TimeLoc = nil    // Clear location to output in original (UTC) location.
	tab.Dump(extractor, format)
	w.Flush()

	// Output:
	// Flt  Str   IntP Method1 Method2 Other.Start         OtherP.Unix DayOfMonth
	// 3.14 Hello 8    3       false   2009-12-28T08:45:00 1418428799  28
	// 2.72 Go         3       false   2014-12-12T23:59:59 4070908860  12
	// 1.41       9    1               2099-01-01T00:01:00             1

}
