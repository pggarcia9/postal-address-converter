package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "to-usps":
		runToUSPS(os.Args[2:])
	case "to-line":
		runToLine(os.Args[2:])
	case "batch":
		runBatch(os.Args[2:])
	case "-h", "--help", "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "addrconv: unknown command %q\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `addrconv converts a US postal address between two formats:

  line  a single comma-separated line, as used in web forms and CSV exports
        "123 Main St, Springfield, IL 62704"
        "123 Main St, Apt 4, Springfield, IL 62704"

  usps  the two-line mailing label block described in USPS Publication 28
        123 MAIN ST APT 4
        SPRINGFIELD IL 62704

Usage:
  addrconv to-usps [--json] [--name "<recipient>"] "<line address>"
  addrconv to-line [--json] "<usps line 1>" "<usps line 2>" ["<usps line 3>"]
  addrconv batch [--json] <file.csv>

If the address isn't given as an argument, it's read from stdin
(one line for to-usps, two or three lines for to-line).

--name adds a recipient name as the first line of the usps block. When
converting back with to-line, give three lines (name, street, city/state/zip)
instead of two and it comes back out through --json as the address's Name
field.

--json prints the parsed address as a JSON object instead of the
formatted text of the target format.

batch reads a CSV file with a header row naming its columns (name, street1,
street2, city, state, zip - name and street2 are optional) and prints a
usps block for each row, separated by blank lines. A row that fails to
parse or validate is reported on stderr by its line number and skipped;
addrconv exits 1 if any row was skipped. With --json, each row prints as a
JSON object instead.
`)
}

func runToUSPS(args []string) {
	fs := flag.NewFlagSet("to-usps", flag.ExitOnError)
	jsonOut := fs.Bool("json", false, "print the parsed address as JSON")
	name := fs.String("name", "", "recipient name, printed as the first line of the block")
	fs.Parse(args)

	line, err := readOneOrArg(fs.Args(), "to-usps")
	if err != nil {
		fail(err)
	}

	addr, err := ParseLine(line)
	if err != nil {
		fail(err)
	}
	addr.Name = strings.TrimSpace(*name)
	if err := addr.Validate(); err != nil {
		fail(err)
	}

	if *jsonOut {
		printJSON(addr)
		return
	}
	for _, l := range addr.USPSBlock() {
		fmt.Println(l)
	}
}

func runToLine(args []string) {
	fs := flag.NewFlagSet("to-line", flag.ExitOnError)
	jsonOut := fs.Bool("json", false, "print the parsed address as JSON")
	fs.Parse(args)

	rest := fs.Args()
	var lines []string
	var err error
	if len(rest) >= 2 {
		lines = rest
	} else {
		lines, err = readUSPSLinesFromStdin()
		if err != nil {
			fail(err)
		}
	}

	addr, err := ParseUSPSBlock(lines...)
	if err != nil {
		fail(err)
	}
	if err := addr.Validate(); err != nil {
		fail(err)
	}

	if *jsonOut {
		printJSON(addr)
		return
	}
	fmt.Println(addr.Line())
}

func runBatch(args []string) {
	fs := flag.NewFlagSet("batch", flag.ExitOnError)
	jsonOut := fs.Bool("json", false, "print each parsed address as JSON instead of a usps block")
	fs.Parse(args)

	rest := fs.Args()
	if len(rest) != 1 {
		fail(fmt.Errorf("batch: expected a single CSV file path, got %d arguments", len(rest)))
	}

	f, err := os.Open(rest[0])
	if err != nil {
		fail(fmt.Errorf("batch: %w", err))
	}
	defer f.Close()

	rows, err := ParseBatchCSV(f)
	if err != nil {
		fail(err)
	}

	ok := true
	first := true
	for _, row := range rows {
		if row.Err != nil {
			fmt.Fprintf(os.Stderr, "addrconv: batch line %d: %v\n", row.Line, row.Err)
			ok = false
			continue
		}
		if *jsonOut {
			printJSON(row.Addr)
			continue
		}
		if !first {
			fmt.Println()
		}
		first = false
		for _, l := range row.Addr.USPSBlock() {
			fmt.Println(l)
		}
	}
	if !ok {
		os.Exit(1)
	}
}

func readOneOrArg(args []string, cmd string) (string, error) {
	if len(args) >= 1 {
		return strings.Join(args, " "), nil
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("reading stdin: %w", err)
	}
	line := strings.TrimSpace(string(data))
	if line == "" {
		return "", fmt.Errorf("%s: no address given (pass it as an argument or via stdin)", cmd)
	}
	return line, nil
}

func readUSPSLinesFromStdin() ([]string, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("reading stdin: %w", err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) < 2 || len(lines) > 3 {
		return nil, fmt.Errorf(
			"to-line: need two lines of input (usps block) or three (with a recipient name), got %d", len(lines))
	}
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
	return lines, nil
}

func printJSON(addr Address) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(addr)
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "addrconv: %v\n", err)
	os.Exit(1)
}
