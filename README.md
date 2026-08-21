# addrconv

Every address I deal with shows up in one of two shapes: a single
comma-separated line (what web forms and CSV exports give you), or the
two-line block the USPS actually wants on an envelope, per
[Publication 28](https://pe.usps.com/text/pub28/welcome.htm) (all caps, no
punctuation, unit folded onto the street line). Copying between the two by
hand is the kind of thing that's easy to get subtly wrong, so this is a
small converter that does it the same way every time.

## Formats

**line** - a single line, three or four comma-separated parts:

```
123 Main St, Springfield, IL 62704
123 Main St, Apt 4, Springfield, IL 62704
```

**usps** - the two-line mailing label block, or three lines with a
recipient name on top:

```
123 MAIN ST APT 4
SPRINGFIELD IL 62704

JANE DOE
123 MAIN ST APT 4
SPRINGFIELD IL 62704
```

## Usage

```
$ addrconv to-usps "123 Main St, Apt 4, Springfield, IL 62704"
123 MAIN ST APT 4
SPRINGFIELD IL 62704

$ addrconv to-usps --json "123 Main St, Springfield, IL 62704"
{
  "street1": "123 Main St",
  "city": "Springfield",
  "state": "IL",
  "zip": "62704"
}

$ addrconv to-line "123 MAIN ST APT 4" "SPRINGFIELD IL 62704"
123 MAIN ST APT 4, SPRINGFIELD, IL 62704

$ echo "123 Main St, Springfield, IL 62704" | addrconv to-usps
123 MAIN ST
SPRINGFIELD IL 62704

$ addrconv to-usps --name "Jane Doe" "123 Main St, Springfield, IL 62704"
JANE DOE
123 MAIN ST
SPRINGFIELD IL 62704

$ addrconv to-line --json "JANE DOE" "123 MAIN ST" "SPRINGFIELD IL 62704"
{
  "name": "JANE DOE",
  "street1": "123 MAIN ST",
  "city": "SPRINGFIELD",
  "state": "IL",
  "zip": "62704"
}
```

`--name` prints a recipient name as the first line of the usps block. Going
the other way, give `to-line` three lines instead of two (name, street,
city/state/zip) and it comes back out on the `Address` as `Name`. The plain
comma-separated `line` format has no slot for a name - it only round-trips
street/city/state/zip.

`--json` doesn't just reformat the target text as JSON - it prints the
parsed `Address` struct, which is the same shape either direction produces.
That's the easiest way to check what the parser actually understood before
trusting the formatted output.

Known limitation: parsing a `usps` block back out, the street line is only
split into street + unit if it ends in one of the designators from
Publication 28 appendix C (APT, UNIT, STE, #, and the like). An
unrecognized designator stays folded into the street, since Publication 28
doesn't fix where on the line it would go.

## Building

Standard library only, no dependencies to fetch:

```
go build ./...
```

## License

MIT, see [LICENSE](LICENSE).
