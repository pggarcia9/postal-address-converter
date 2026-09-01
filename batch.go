package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

// BatchRow pairs one parsed address with the CSV row it came from, or the
// error that row produced. A bad row further down the file shouldn't stop
// the good ones from converting, so ParseBatchCSV collects everything and
// leaves it to the caller to decide what to do with the errors.
type BatchRow struct {
	Line int // 1-based, counting the header as line 1
	Addr Address
	Err  error
}

// batchColumns are the CSV header names ParseBatchCSV understands. name and
// street2 are optional and may be blank or omitted from the header
// entirely; the rest are required.
var batchColumns = []string{"street1", "city", "state", "zip"}

// ParseBatchCSV reads a CSV of addresses, one per row, with a header row
// naming its columns: name, street1, street2, city, state, zip. Column
// order doesn't matter and name/street2 may be left out. Each data row
// becomes one BatchRow, in file order, whether or not it parsed cleanly, so
// callers can report a bad row by its original line number.
func ParseBatchCSV(r io.Reader) ([]BatchRow, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1

	header, err := cr.Read()
	if err != nil {
		return nil, fmt.Errorf("batch csv: reading header: %w", err)
	}
	col := make(map[string]int, len(header))
	for i, h := range header {
		col[strings.ToLower(strings.TrimSpace(h))] = i
	}
	for _, c := range batchColumns {
		if _, ok := col[c]; !ok {
			return nil, fmt.Errorf("batch csv: missing required column %q", c)
		}
	}

	field := func(record []string, name string) string {
		i, ok := col[name]
		if !ok || i >= len(record) {
			return ""
		}
		return strings.TrimSpace(record[i])
	}

	var rows []BatchRow
	line := 1
	for {
		record, err := cr.Read()
		if err == io.EOF {
			break
		}
		line++
		if err != nil {
			rows = append(rows, BatchRow{Line: line, Err: fmt.Errorf("reading row: %w", err)})
			continue
		}

		addr := Address{
			Name:    field(record, "name"),
			Street1: field(record, "street1"),
			Street2: field(record, "street2"),
			City:    field(record, "city"),
			State:   strings.ToUpper(field(record, "state")),
			Zip:     field(record, "zip"),
		}
		row := BatchRow{Line: line, Addr: addr}
		if err := addr.Validate(); err != nil {
			row.Err = err
		}
		rows = append(rows, row)
	}
	return rows, nil
}
