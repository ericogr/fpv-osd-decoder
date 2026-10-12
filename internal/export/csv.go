package export

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Format defines how the CSV is written: column separator and decimal
// separator. The default (, and .) suits English-locale spreadsheets;
// Portuguese-locale ones expect ; and , (otherwise "3.6" becomes a date). The
// zero Format is the default.
type Format struct {
	Sep     rune
	Decimal string
}

// NewFormat validates a column separator and a decimal separator.
func NewFormat(sep, decimal string) (Format, error) {
	r := []rune(sep)
	if len(r) != 1 || r[0] == '"' || r[0] == '\r' || r[0] == '\n' || r[0] == utf8.RuneError {
		return Format{}, fmt.Errorf("column separator must be a single character, such as , or ; (got %q)", sep)
	}
	if decimal != "." && decimal != "," {
		return Format{}, fmt.Errorf("decimal separator must be . or , (got %q)", decimal)
	}
	if string(r[0]) == decimal {
		return Format{}, fmt.Errorf("column and decimal separators cannot be the same character (%q)", decimal)
	}
	return Format{Sep: r[0], Decimal: decimal}, nil
}

// WriteCSV writes the table as CSV: a header with the column keys, then one
// row per frame. The decimal separator only changes numeric columns.
func (t *Table) WriteCSV(w io.Writer, f Format) error {
	if f.Sep == 0 {
		f.Sep = ','
	}
	if f.Decimal == "" {
		f.Decimal = "."
	}

	cw := csv.NewWriter(w)
	cw.Comma = f.Sep

	header := make([]string, 0, len(t.Columns)+1)
	header = append(header, timeColumn)
	for _, c := range t.Columns {
		header = append(header, c.Key)
	}
	if err := cw.Write(header); err != nil {
		return err
	}

	rec := make([]string, len(header))
	for i, row := range t.Rows {
		rec[0] = strconv.FormatUint(uint64(t.Times[i]), 10)
		for j, v := range row {
			if !t.Columns[j].Text && f.Decimal != "." {
				v = strings.Replace(v, ".", f.Decimal, 1)
			}
			rec[j+1] = v
		}
		if err := cw.Write(rec); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}
