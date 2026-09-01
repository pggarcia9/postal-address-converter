package main

import (
	"strings"
	"testing"
)

func TestParseBatchCSV(t *testing.T) {
	input := `name,street1,street2,city,state,zip
Jane Doe,123 Main St,Apt 4,Springfield,IL,62704
,456 Oak Ave,,Portland,OR,97201
`
	rows, err := ParseBatchCSV(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("got %d rows, want 2", len(rows))
	}

	if rows[0].Err != nil {
		t.Fatalf("row 0: unexpected error: %v", rows[0].Err)
	}
	want := Address{Name: "Jane Doe", Street1: "123 Main St", Street2: "Apt 4", City: "Springfield", State: "IL", Zip: "62704"}
	if rows[0].Addr != want {
		t.Fatalf("row 0: got %+v, want %+v", rows[0].Addr, want)
	}
	if rows[0].Line != 2 {
		t.Fatalf("row 0: got line %d, want 2", rows[0].Line)
	}

	if rows[1].Err != nil {
		t.Fatalf("row 1: unexpected error: %v", rows[1].Err)
	}
	want = Address{Street1: "456 Oak Ave", City: "Portland", State: "OR", Zip: "97201"}
	if rows[1].Addr != want {
		t.Fatalf("row 1: got %+v, want %+v", rows[1].Addr, want)
	}
	if rows[1].Line != 3 {
		t.Fatalf("row 1: got line %d, want 3", rows[1].Line)
	}
}

func TestParseBatchCSVColumnOrderAndOptionalColumns(t *testing.T) {
	input := `city,zip,street1,state
Springfield,62704,123 Main St,IL
`
	rows, err := ParseBatchCSV(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 || rows[0].Err != nil {
		t.Fatalf("got rows %+v", rows)
	}
	want := Address{Street1: "123 Main St", City: "Springfield", State: "IL", Zip: "62704"}
	if rows[0].Addr != want {
		t.Fatalf("got %+v, want %+v", rows[0].Addr, want)
	}
}

func TestParseBatchCSVMissingRequiredColumn(t *testing.T) {
	input := `street1,city,state
123 Main St,Springfield,IL
`
	_, err := ParseBatchCSV(strings.NewReader(input))
	if err == nil || !strings.Contains(err.Error(), `missing required column "zip"`) {
		t.Fatalf("expected missing column error, got %v", err)
	}
}

func TestParseBatchCSVBadRowDoesNotStopOthers(t *testing.T) {
	input := `street1,city,state,zip
123 Main St,Springfield,IL,62704
,Nowhere,IL,not-a-zip
456 Oak Ave,Portland,OR,97201
`
	rows, err := ParseBatchCSV(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("got %d rows, want 3", len(rows))
	}
	if rows[0].Err != nil {
		t.Fatalf("row 0: unexpected error: %v", rows[0].Err)
	}
	if rows[1].Err == nil || !strings.Contains(rows[1].Err.Error(), "invalid address") {
		t.Fatalf("row 1: expected validation error, got %v", rows[1].Err)
	}
	if rows[1].Line != 3 {
		t.Fatalf("row 1: got line %d, want 3", rows[1].Line)
	}
	if rows[2].Err != nil {
		t.Fatalf("row 2: unexpected error: %v", rows[2].Err)
	}
	if rows[2].Line != 4 {
		t.Fatalf("row 2: got line %d, want 4", rows[2].Line)
	}
}
