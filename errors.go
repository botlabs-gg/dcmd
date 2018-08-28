package dcmd

import (
	"fmt"
)

type InvalidInt struct {
	Part string
}

func (i *InvalidInt) Error() string {
	return fmt.Sprintf("%q is not a whole number", i.Part)
}

type InvalidFloat struct {
	Part string
}

func (i *InvalidFloat) Error() string {
	return fmt.Sprintf("%q is not a number", i.Part)
}

type ImproperMention struct {
	Part string
}

func (i *ImproperMention) Error() string {
	return fmt.Sprintf("Improper mention %q", i.Part)
}

type NoMention struct {
	Part string
}

func (i *NoMention) Error() string {
	return fmt.Sprintf("No mention found in %q", i.Part)
}

type UserNotFound struct {
	Part string
}

func (i *UserNotFound) Error() string {
	return fmt.Sprintf("User %q not found", i.Part)
}

type OutOfRangeError struct {
	Min, Max interface{}
	Got      interface{}
	Float    bool
	ArgName  string
}

func (o *OutOfRangeError) Error() string {
	preStr := "too big"

	switch o.Got.(type) {
	case int64:
		if o.Got.(int64) < o.Min.(int64) {
			preStr = "too small"
		}
	case float64:
		if o.Got.(float64) < o.Min.(float64) {
			preStr = "too small"
		}
	}

	const floatFormat = "%s is %s (has to be within %f - %f)"
	const intFormat = "%s is %s (has to be within %d - %d)"

	if o.Float {
		return fmt.Sprintf(floatFormat, o.ArgName, preStr, o.Min, o.Max)
	}

	return fmt.Sprintf(intFormat, o.ArgName, preStr, o.Min, o.Max)
}
