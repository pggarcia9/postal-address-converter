package main

import (
	"strings"
	"testing"
)

func TestParseLine(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  Address
	}{
		{
			"no unit",
			"123 Main St, Springfield, IL 62704",
			Address{Street1: "123 Main St", City: "Springfield", State: "IL", Zip: "62704"},
		},
		{
			"with unit",
			"123 Main St, Apt 4, Springfield, IL 62704",
			Address{Street1: "123 Main St", Street2: "Apt 4", City: "Springfield", State: "IL", Zip: "62704"},
		},
		{
			"extra whitespace",
			" 123 Main St ,  Springfield , IL 62704 ",
			Address{Street1: "123 Main St", City: "Springfield", State: "IL", Zip: "62704"},
		},
		{
			"zip+4",
			"123 Main St, Springfield, IL 62704-1234",
			Address{Street1: "123 Main St", City: "Springfield", State: "IL", Zip: "62704-1234"},
		},
		{
			"trailing comma",
			"123 Main St, Springfield, IL 62704,",
			Address{Street1: "123 Main St", City: "Springfield", State: "IL", Zip: "62704"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ParseLine(c.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != c.want {
				t.Fatalf("got %+v, want %+v", got, c.want)
			}
		})
	}
}

func TestParseLineErrors(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"too few parts", "123 Main St, Springfield", "line format needs 3 or 4"},
		{"too many parts", "123 Main St, Apt 4, Unit B, Springfield, IL 62704", "line format needs 3 or 4"},
		{"malformed state zip", "123 Main St, Springfield, Illinois", "expected \"ST ZIP\""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := ParseLine(c.input)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("expected error containing %q, got %v", c.want, err)
			}
		})
	}
}

func TestSplitUnit(t *testing.T) {
	cases := []struct {
		name       string
		input      string
		wantStreet string
		wantUnit   string
	}{
		{"no unit", "123 Main St", "123 Main St", ""},
		{"apt", "123 Main St APT 4", "123 Main St", "APT 4"},
		{"lowercase designator", "123 Main St apt 4", "123 Main St", "apt 4"},
		{"hash prefix", "123 Main St #4", "123 Main St", "#4"},
		{"hash glued to number", "123 Main St #4B", "123 Main St", "#4B"},
		{"ste", "123 Main St STE 200", "123 Main St", "STE 200"},
		{"designator as street name", "1 Fl Street", "1", "Fl Street"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			street, unit := splitUnit(c.input)
			if street != c.wantStreet || unit != c.wantUnit {
				t.Fatalf("got street=%q unit=%q, want street=%q unit=%q", street, unit, c.wantStreet, c.wantUnit)
			}
		})
	}
}

func TestParseUSPSBlock(t *testing.T) {
	got, err := ParseUSPSBlock("123 MAIN ST", "SPRINGFIELD IL 62704")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := Address{Street1: "123 MAIN ST", City: "SPRINGFIELD", State: "IL", Zip: "62704"}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}

	got, err = ParseUSPSBlock("JANE DOE", "123 MAIN ST APT 4", "SPRINGFIELD IL 62704")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want = Address{Name: "JANE DOE", Street1: "123 MAIN ST", Street2: "APT 4", City: "SPRINGFIELD", State: "IL", Zip: "62704"}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}

	got, err = ParseUSPSBlock("123 MAIN ST", "NEW YORK CITY NY 10001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.City != "NEW YORK CITY" {
		t.Fatalf("got city %q, want %q", got.City, "NEW YORK CITY")
	}
}

func TestParseUSPSBlockErrors(t *testing.T) {
	cases := []struct {
		name  string
		lines []string
		want  string
	}{
		{"one line", []string{"123 MAIN ST"}, "expected 2 lines"},
		{"four lines", []string{"A", "B", "C", "D"}, "expected 2 lines"},
		{"empty street", []string{"  ", "SPRINGFIELD IL 62704"}, "street line is empty"},
		{"short city line", []string{"123 MAIN ST", "SPRINGFIELD"}, "must be \"CITY ST ZIP\""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := ParseUSPSBlock(c.lines...)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("expected error containing %q, got %v", c.want, err)
			}
		})
	}
}

func TestSplitStateZip(t *testing.T) {
	state, zip, err := splitStateZip("IL 62704")
	if err != nil || state != "IL" || zip != "62704" {
		t.Fatalf("got state=%q zip=%q err=%v", state, zip, err)
	}

	state, zip, err = splitStateZip("il 62704")
	if err != nil || state != "IL" {
		t.Fatalf("expected state to be uppercased, got %q err=%v", state, err)
	}

	if _, _, err := splitStateZip("Illinois"); err == nil {
		t.Fatal("expected error for single-token input")
	}
	if _, _, err := splitStateZip("IL 62704 extra"); err == nil {
		t.Fatal("expected error for too many tokens")
	}
}

func TestRoundTrip(t *testing.T) {
	line := "123 Main St, Apt 4, Springfield, IL 62704"
	addr, err := ParseLine(line)
	if err != nil {
		t.Fatalf("ParseLine: %v", err)
	}
	block := addr.USPSBlock()
	if len(block) != 2 {
		t.Fatalf("expected 2-line block, got %d: %v", len(block), block)
	}

	back, err := ParseUSPSBlock(block...)
	if err != nil {
		t.Fatalf("ParseUSPSBlock: %v", err)
	}
	if back.Street1 != "123 MAIN ST" || back.Street2 != "APT 4" {
		t.Fatalf("unexpected round trip result: %+v", back)
	}
}
