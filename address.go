package main

import (
	"fmt"
	"regexp"
	"strings"
)

// Address is the intermediate representation both supported formats parse
// into and render out of. It's also what --json prints, so the two
// directions (to-usps, to-line) agree on one shape.
type Address struct {
	Name    string `json:"name,omitempty"`
	Street1 string `json:"street1"`
	Street2 string `json:"street2,omitempty"`
	City    string `json:"city"`
	State   string `json:"state"`
	Zip     string `json:"zip"`
}

var (
	zipPattern   = regexp.MustCompile(`^\d{5}(-\d{4})?$`)
	statePattern = regexp.MustCompile(`^[A-Z]{2}$`)
)

func (a Address) Validate() error {
	var problems []string
	if strings.TrimSpace(a.Street1) == "" {
		problems = append(problems, "missing street")
	}
	if strings.TrimSpace(a.City) == "" {
		problems = append(problems, "missing city")
	}
	if !statePattern.MatchString(a.State) {
		problems = append(problems, fmt.Sprintf("state must be a two-letter code, got %q", a.State))
	}
	if !zipPattern.MatchString(a.Zip) {
		problems = append(problems, fmt.Sprintf("zip must be 5 digits or ZIP+4, got %q", a.Zip))
	}
	if len(problems) > 0 {
		return fmt.Errorf("invalid address: %s", strings.Join(problems, "; "))
	}
	return nil
}

// USPSBlock renders the mailing label block described in USPS Publication
// 28: all uppercase, unit info folded onto the street line, with the
// recipient name (if set) on its own line at the top.
func (a Address) USPSBlock() []string {
	var lines []string
	if a.Name != "" {
		lines = append(lines, strings.ToUpper(a.Name))
	}
	line1 := a.Street1
	if a.Street2 != "" {
		line1 = line1 + " " + a.Street2
	}
	lines = append(lines, strings.ToUpper(line1))
	lines = append(lines, strings.ToUpper(fmt.Sprintf("%s %s %s", a.City, a.State, a.Zip)))
	return lines
}

// Line renders the address as a single comma-separated line, the shape
// most web forms and CSV exports use.
func (a Address) Line() string {
	parts := []string{a.Street1}
	if a.Street2 != "" {
		parts = append(parts, a.Street2)
	}
	parts = append(parts, a.City, a.State+" "+a.Zip)
	return strings.Join(parts, ", ")
}
