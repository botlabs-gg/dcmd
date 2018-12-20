package dcmd

import (
	"fmt"
	"github.com/pkg/errors"
)

type InvalidInt struct {
	Part string
}

func (i *InvalidInt) Error() string {
	return fmt.Sprintf("%q is not a whole number", i.Part)
}

func (i *InvalidInt) IsUserError() bool {
	return true
}

type InvalidFloat struct {
	Part string
}

func (i *InvalidFloat) Error() string {
	return fmt.Sprintf("%q is not a number", i.Part)
}

func (i *InvalidFloat) IsUserError() bool {
	return true
}

type ImproperMention struct {
	Part string
}

func (i *ImproperMention) Error() string {
	return fmt.Sprintf("Improper mention %q", i.Part)
}

func (i *ImproperMention) IsUserError() bool {
	return true
}

type NoMention struct {
	Part string
}

func (i *NoMention) Error() string {
	return fmt.Sprintf("No mention found in %q", i.Part)
}

func (i *NoMention) IsUserError() bool {
	return true
}

type UserNotFound struct {
	Part string
}

func (i *UserNotFound) Error() string {
	return fmt.Sprintf("User %q not found", i.Part)
}

func (i *UserNotFound) IsUserError() bool {
	return true
}

type ChannelNotFound struct {
	ID int64
}

func (c *ChannelNotFound) Error() string {
	return fmt.Sprintf("Channel %d not found", c.ID)
}

func (c *ChannelNotFound) IsUserError() bool {
	return true
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

func (o *OutOfRangeError) IsUserError() bool {
	return true
}

type UserError interface {
	IsUserError() bool
}

func IsUserError(err error) bool {
	v, ok := errors.Cause(err).(UserError)
	if ok && v.IsUserError() {
		return true
	}

	return false
}

type simpleUserError string

func (s simpleUserError) Error() string {
	return string(s)
}

func (s simpleUserError) IsUserError() bool {
	return true
}

func NewSimpleUserError(args ...interface{}) error {
	return simpleUserError(fmt.Sprint(args...))
}
