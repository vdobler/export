// Copyright 2014 Volker DObler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package export

import (
	"errors"
	"os"
	"text/tabwriter"
	"time"
)

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
	exp, err := NewExtractor(data,
		"Flt", "Str", "IntP", // Accessing of fields and pointer fields
		"Method1", "Method2", // Accessing results of methods.
		"Other.Start", "OtherP.Unix") // Accessing nested elements.
	if err != nil {
		panic(err.Error())
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 1, 8, 1, ' ', 0)
	tab := TabDumper{Writer: w}
	tab.Dump(exp, DefaultFormat)
	w.Flush()

	// Output:
	// Flt  Str   IntP Method1 Method2 Start               Unix
	// 3.14 Hello 8    3       false   2009-12-28T09:45:00 1418428799
	// 2.72 Go         3       false   2014-12-13T00:59:59 4070908860
	// 1.41       9    1               2099-01-01T01:01:00

}
