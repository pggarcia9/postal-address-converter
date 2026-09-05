package main

import (
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	base := Address{Street1: "123 Main St", City: "Springfield", State: "IL", Zip: "62704"}

	if err := base.Validate(); err != nil {
		t.Fatalf("expected valid address to pass, got %v", err)
	}

	cases := []struct {
		name string
		mod  func(a Address) Address
		want string
	}{
		{"missing street", func(a Address) Address { a.Street1 = "  "; return a }, "missing street"},
		{"missing city", func(a Address) Address { a.City = ""; return a }, "missing city"},
		{"lowercase state", func(a Address) Address { a.State = "il"; return a }, "state must be"},
		{"long state", func(a Address) Address { a.State = "ILL"; return a }, "state must be"},
		{"not a real state", func(a Address) Address { a.State = "ZZ"; return a }, "not a USPS state"},
		{"territory is fine", func(a Address) Address { a.State = "PR"; return a }, ""},
		{"military state is fine", func(a Address) Address { a.State = "AE"; return a }, ""},
		{"short zip", func(a Address) Address { a.Zip = "6270"; return a }, "zip must be"},
		{"zip+4 missing suffix digits", func(a Address) Address { a.Zip = "62704-12"; return a }, "zip must be"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.mod(base).Validate()
			if c.want == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("expected error containing %q, got %v", c.want, err)
			}
		})
	}

	multi := base
	multi.State = "illinois"
	multi.Zip = "bad"
	err := multi.Validate()
	if err == nil {
		t.Fatal("expected error for multiple problems")
	}
	if !strings.Contains(err.Error(), "state must be") || !strings.Contains(err.Error(), "zip must be") {
		t.Fatalf("expected both problems listed, got %v", err)
	}

	zipPlus4 := base
	zipPlus4.Zip = "62704-1234"
	if err := zipPlus4.Validate(); err != nil {
		t.Fatalf("expected zip+4 to pass, got %v", err)
	}
}

func TestUSPSBlock(t *testing.T) {
	a := Address{Street1: "123 main st", City: "springfield", State: "il", Zip: "62704"}
	got := a.USPSBlock()
	want := []string{"123 MAIN ST", "SPRINGFIELD IL 62704"}
	if !equalLines(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}

	a.Name = "jane doe"
	got = a.USPSBlock()
	want = []string{"JANE DOE", "123 MAIN ST", "SPRINGFIELD IL 62704"}
	if !equalLines(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}

	a.Street2 = "apt 4"
	got = a.USPSBlock()
	want = []string{"JANE DOE", "123 MAIN ST APT 4", "SPRINGFIELD IL 62704"}
	if !equalLines(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestLine(t *testing.T) {
	a := Address{Street1: "123 Main St", City: "Springfield", State: "IL", Zip: "62704"}
	if got, want := a.Line(), "123 Main St, Springfield, IL 62704"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}

	a.Street2 = "Apt 4"
	if got, want := a.Line(), "123 Main St, Apt 4, Springfield, IL 62704"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func equalLines(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
