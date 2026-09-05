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

// validStateCodes are the two-letter codes USPS Publication 28 Appendix B
// actually assigns: the 50 states, DC, the inhabited territories, and the
// three military "state" codes (AA/AE/AP) used for APO/FPO/DPO mail. A
// well-formed two-letter code that isn't on this list (e.g. "ZZ") is still
// rejected - matching the pattern isn't the same as being a real place.
var validStateCodes = map[string]bool{
	"AL": true, "AK": true, "AZ": true, "AR": true, "CA": true, "CO": true,
	"CT": true, "DE": true, "FL": true, "GA": true, "HI": true, "ID": true,
	"IL": true, "IN": true, "IA": true, "KS": true, "KY": true, "LA": true,
	"ME": true, "MD": true, "MA": true, "MI": true, "MN": true, "MS": true,
	"MO": true, "MT": true, "NE": true, "NV": true, "NH": true, "NJ": true,
	"NM": true, "NY": true, "NC": true, "ND": true, "OH": true, "OK": true,
	"OR": true, "PA": true, "RI": true, "SC": true, "SD": true, "TN": true,
	"TX": true, "UT": true, "VT": true, "VA": true, "WA": true, "WV": true,
	"WI": true, "WY": true,
	"DC": true,
	"AS": true, "GU": true, "MP": true, "PR": true, "VI": true,
	"AA": true, "AE": true, "AP": true,
}

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
	} else if !validStateCodes[a.State] {
		problems = append(problems, fmt.Sprintf("%q is not a USPS state, DC, or territory code", a.State))
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
