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

**usps** - the two-line mailing label block:

```
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
```

`--json` doesn't just reformat the target text as JSON - it prints the
parsed `Address` struct, which is the same shape either direction produces.
That's the easiest way to check what the parser actually understood before
trusting the formatted output.

Known limitation: parsing a `usps` block back out, the street line is kept
whole rather than split into street + unit, since Publication 28 doesn't
fix where on the line a unit designator goes. Round-tripping `line -> usps
-> line` will fold a unit into the street rather than keeping it separate.

## Building

Standard library only, no dependencies to fetch:

```
go build ./...
```

## License

MIT, see [LICENSE](LICENSE).
