// Copyright 2014 Volker Dobler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package export

import (
	"encoding/csv"
	"fmt"
	"io"
	"text/tabwriter"
)

// Dumper is the interface which wrapps the Dump methods
type Dumper interface {
	// Dump the data defined in e in the given format.
	Dump(e *Extractor, format Format) error
}

// CSVDumper dumps values to a csv writer.
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

// TabDumper dumps the value to a tabwriter.
type TabDumper struct {
	// Writer is the tabwriter to be used.
	Writer     *tabwriter.Writer
	OmitHeader bool // OmitHeader suppresses the header line in the generated CSV.
}

// Dump dumps the fields from e to d. Dump does not call Flush on the
// underlying tabwriter.
func (d TabDumper) Dump(e *Extractor, format Format) error {
	if !d.OmitHeader {
		ff := "%s"
		for _, field := range e.Columns {
			fmt.Fprintf(d.Writer, ff, field.Name)
			ff = "\t%s"
		}
	}
	fmt.Fprintln(d.Writer)
	for r := 0; r < e.N; r++ {
		ff := "%s"
		for _, field := range e.Columns {
			fmt.Fprintf(d.Writer, ff, field.Print(format, r))
			ff = "\t%s"
		}
		fmt.Fprintln(d.Writer)
	}

	return nil
}

// RVecDumper dumps as a R vectors.
type RVecDumper struct {
	Writer io.Writer

	// DataFrame is the name of a R data frame to construct from the
	// individual column vectors. A empty value suppresses the generation
	// of this combining data frame.
	DataFrame string
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

	if d.DataFrame != "" {
		if _, err := fmt.Fprintf(d.Writer, "%s <- data.frame(%s)\n", d.DataFrame, all); err != nil {
			return err
		}
	}
	return nil
}
