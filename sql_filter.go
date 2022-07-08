package web

import (
	"errors"
	"fmt"
	"strings"
)

// SqlFilter create sql for filter and args
func SqlFilter(filter string, str *strings.Builder, args *[]interface{}, prefix string, fn func(key string, val string) (string, interface{}, error)) error {

	vals := filterParse(filter)

	l := len(vals)

	if l == 0 {
		return errors.New("filter invalid")
	}

	var prev string
	var space bool

	for i := 0; i < l; i++ {

		val := vals[i]

		switch val {
		case "eq", "ne", "gt", "ge", "lt", "le":
			if space {
				str.WriteString(" ")
			} else {
				space = true
			}

			op := filterOp(val)

			i, val = filterNext(vals, i, l)

			n, v, err := fn(prev, val)

			if err != nil {
				return err
			}

			*args = append(*args, v)
			fmt.Fprintf(str, "%s`%s` %s ?", prefix, n, op)
		case "and", "or", "not":
			if space {
				str.WriteString(" ")
			} else {
				space = true
			}
			str.WriteString(strings.ToUpper(val))
		case "(", ")":
			str.WriteString(val)
		default:
			prev = val
		}
	}

	return nil
}

// filterParse parse filter to vals
func filterParse(filter string) []string {
	l := len(filter)

	prev := 0
	inSingle := false
	inQuotes := false

	vals := []string{}

	for pos := 0; pos < l; pos++ {
		r := filter[pos]

		switch r {
		case '(', ')':

			if inSingle || inQuotes {
				continue
			}

			if pos > prev {
				vals = append(vals, filter[prev:pos])
			}

			prev = pos + 1

			vals = append(vals, string(r))

		case ' ', '\t':

			if inSingle || inQuotes {
				continue
			}

			if pos > prev {
				vals = append(vals, filter[prev:pos])
			}

			prev = pos + 1
		case '\'':
			if inQuotes {
				continue
			}

			inSingle = !inSingle

			if pos > prev {
				vals = append(vals, filter[prev:pos])
			}

			prev = pos + 1
		case '"':
			if inSingle {
				continue
			}

			inQuotes = !inQuotes

			if pos > prev {
				vals = append(vals, filter[prev:pos])
			}

			prev = pos + 1
		}
	}

	if prev < l {
		vals = append(vals, filter[prev:])
	}

	return vals
}

// filterNext read next item
func filterNext(items []string, i int, l int) (int, string) {
	pos := i + 1

	if pos < l {
		return pos, items[pos]
	}

	return i, ""
}

// filterOp return op
func filterOp(key string) string {
	switch key {
	case "eq":
		return "="
	case "ne":
		return "<>"
	case "gt":
		return ">"
	case "ge":
		return ">="
	case "lt":
		return "<"
	case "le":
		return "<="
	default:
		return "="
	}
}
