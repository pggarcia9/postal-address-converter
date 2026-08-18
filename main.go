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
  addrconv to-usps [--json] "<line address>"
  addrconv to-line [--json] "<usps line 1>" "<usps line 2>"

If the address isn't given as an argument, it's read from stdin
(one line for to-usps, two lines for to-line).

--json prints the parsed address as a JSON object instead of the
formatted text of the target format.
`)
}

func runToUSPS(args []string) {
	fs := flag.NewFlagSet("to-usps", flag.ExitOnError)
	jsonOut := fs.Bool("json", false, "print the parsed address as JSON")
	fs.Parse(args)

	line, err := readOneOrArg(fs.Args(), "to-usps")
	if err != nil {
		fail(err)
	}

	addr, err := ParseLine(line)
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
	l1, l2 := addr.USPSLines()
	fmt.Println(l1)
	fmt.Println(l2)
}

func runToLine(args []string) {
	fs := flag.NewFlagSet("to-line", flag.ExitOnError)
	jsonOut := fs.Bool("json", false, "print the parsed address as JSON")
	fs.Parse(args)

	rest := fs.Args()
	var l1, l2 string
	var err error
	if len(rest) >= 2 {
		l1, l2 = rest[0], rest[1]
	} else {
		l1, l2, err = readTwoFromStdin()
		if err != nil {
			fail(err)
		}
	}

	addr, err := ParseUSPSBlock(l1, l2)
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

func readTwoFromStdin() (string, string, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", "", fmt.Errorf("reading stdin: %w", err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) < 2 {
		return "", "", fmt.Errorf("to-line: need two lines of input (usps block), got %d", len(lines))
	}
	return strings.TrimSpace(lines[0]), strings.TrimSpace(lines[1]), nil
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
