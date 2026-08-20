package main

import (
	"fmt"
	"strings"
)

// ParseLine parses the "line" format: "Street, City, ST ZIP" or, with a
// unit, "Street, Unit, City, ST ZIP".
func ParseLine(s string) (Address, error) {
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}

	var addr Address
	switch len(parts) {
	case 3:
		addr.Street1 = parts[0]
		addr.City = parts[1]
	case 4:
		addr.Street1 = parts[0]
		addr.Street2 = parts[1]
		addr.City = parts[2]
	default:
		return Address{}, fmt.Errorf(
			"line format needs 3 or 4 comma-separated parts (street[, unit], city, state zip), got %d in %q",
			len(parts), s)
	}

	state, zip, err := splitStateZip(parts[len(parts)-1])
	if err != nil {
		return Address{}, err
	}
	addr.State = state
	addr.Zip = zip
	return addr, nil
}

// unitDesignators are the secondary-address abbreviations listed in USPS
// Publication 28 appendix C. "#" is handled separately below since it's a
// prefix glued to the unit number ("#4") rather than a standalone word.
var unitDesignators = []string{
	"APT", "BLDG", "DEPT", "FL", "HNGR", "KEY", "LOT", "PIER",
	"RM", "SLIP", "SPC", "STE", "TRLR", "UNIT",
}

// splitUnit looks for the first token in a street line that's a known unit
// designator and splits the line there. Everything from that token on
// becomes the unit; everything before it stays the street. If no
// designator is found, the whole line is returned as the street.
func splitUnit(line string) (street, unit string) {
	fields := strings.Fields(line)
	for i, f := range fields {
		upper := strings.ToUpper(f)
		if strings.HasPrefix(upper, "#") {
			return strings.Join(fields[:i], " "), strings.Join(fields[i:], " ")
		}
		for _, d := range unitDesignators {
			if upper == d {
				return strings.Join(fields[:i], " "), strings.Join(fields[i:], " ")
			}
		}
	}
	return line, ""
}

// ParseUSPSBlock parses the two-line USPS mailing label block. If the
// street line ends in a recognized unit designator (APT, UNIT, STE, #,
// etc.), it's split back out into Street2; anything else stays folded
// into Street1, since Publication 28 doesn't fix where on the line an
// unrecognized designator would go.
func ParseUSPSBlock(line1, line2 string) (Address, error) {
	line1 = strings.TrimSpace(line1)
	if line1 == "" {
		return Address{}, fmt.Errorf("usps block: line 1 (street) is empty")
	}

	fields := strings.Fields(line2)
	if len(fields) < 3 {
		return Address{}, fmt.Errorf("usps block: line 2 must be \"CITY ST ZIP\", got %q", line2)
	}
	zip := fields[len(fields)-1]
	state := fields[len(fields)-2]
	city := strings.Join(fields[:len(fields)-2], " ")

	street1, street2 := splitUnit(line1)

	return Address{
		Street1: street1,
		Street2: street2,
		City:    city,
		State:   strings.ToUpper(state),
		Zip:     zip,
	}, nil
}

func splitStateZip(s string) (state, zip string, err error) {
	fields := strings.Fields(s)
	if len(fields) != 2 {
		return "", "", fmt.Errorf("expected \"ST ZIP\" (e.g. \"IL 62704\"), got %q", s)
	}
	return strings.ToUpper(fields[0]), fields[1], nil
}
