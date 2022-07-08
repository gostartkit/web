package web

import (
	"errors"
	"strings"
)

// SqlOrderBy create sql for order by
func SqlOrderBy(orderBy string, str *strings.Builder, prefix string, fn func(variableName string) (string, string, string, error)) error {
	vals := orderByParse(orderBy)

	l := len(vals)

	if l == 0 {
		return errors.New("orderBy invalid")
	}

	for i := 0; i < l; i++ {

		val := vals[i]

		switch val {
		case ",":
			str.WriteString(", ")
		case "asc":
			str.WriteString(" ASC")
		case "desc":
			str.WriteString(" DESC")
		default:
			n, _, _, err := fn(val)

			if err != nil {
				return err
			}

			str.WriteString(prefix)
			str.WriteByte('`')
			str.WriteString(n)
			str.WriteByte('`')
		}
	}

	return nil
}

// orderByParse parse orderBy to vals
func orderByParse(orderBy string) []string {
	l := len(orderBy)

	prev := 0

	vals := []string{}

	for pos := 0; pos < l; pos++ {
		r := orderBy[pos]

		switch r {
		case ',':

			if pos > prev {
				vals = append(vals, orderBy[prev:pos])
			}

			prev = pos + 1

			vals = append(vals, string(r))

		case ' ', '\t':

			if pos > prev {
				vals = append(vals, orderBy[prev:pos])
			}

			prev = pos + 1
		}
	}

	if prev < l {
		vals = append(vals, orderBy[prev:])
	}

	return vals
}
