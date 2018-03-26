package dcmd

import (
	"github.com/jonas747/discordgo"
	"github.com/jonas747/dutil/dstate"
	"strconv"
	"strings"
)

// ArgDef represents a argument definition, either a switch or plain arg
type ArgDef struct {
	Name    string
	Switch  string
	Type    ArgType
	Help    string
	Default interface{}
}

type ParsedArg struct {
	Def   *ArgDef
	Value interface{}
	Raw   *RawArg
}

func (p *ParsedArg) Str() string {
	switch t := p.Value.(type) {
	case string:
		return t
	default:
		return ""
	}
}

// TODO: GO-Generate the number ones
func (p *ParsedArg) Int() int {
	switch t := p.Value.(type) {
	case int:
		return t
	case uint:
		return int(t)
	case int32:
		return int(t)
	case int64:
		return int(t)
	case uint32:
		return int(t)
	case uint64:
		return int(t)
	default:
		return 0
	}
}

func (p *ParsedArg) Int64() int64 {
	switch t := p.Value.(type) {
	case int:
		return int64(t)
	case uint:
		return int64(t)
	case int32:
		return int64(t)
	case int64:
		return t
	case uint32:
		return int64(t)
	case uint64:
		return int64(t)
	default:
		return 0
	}
}

// NewParsedArgs creates a new ParsedArg slice from defs passed, also filling default values
func NewParsedArgs(defs []*ArgDef) []*ParsedArg {
	out := make([]*ParsedArg, len(defs))

	for k, _ := range out {
		out[k] = &ParsedArg{
			Def:   defs[k],
			Value: defs[k].Default,
		}
	}

	return out
}

// ArgType is the interface argument types has to implement,
type ArgType interface {
	// Return true if this argument part matches this type
	Matches(part string) bool

	// Attempt to parse it, returning any error if one occured.
	Parse(part string, data *Data) (val interface{}, err error)

	// Name as shown in help
	HelpName() string
}

var (
	// Create some convenience instances
	Int            = &IntArg{}
	Float          = &FloatArg{}
	String         = &StringArg{}
	User           = &UserArg{}
	UserReqMention = &UserArg{RequireMention: true}
)

// IntArg matches and parses integer arguments
// If min and max are not equal then the value has to be within min and max or else it will fail parsing
type IntArg struct {
	Min, Max int64
}

func (i *IntArg) Matches(part string) bool {
	_, err := strconv.ParseInt(part, 10, 64)
	return err == nil
}
func (i *IntArg) Parse(part string, data *Data) (interface{}, error) {
	v, err := strconv.ParseInt(part, 10, 64)
	if err != nil {
		return nil, &InvalidInt{part}
	}

	// A valid range has been specified
	if i.Max != i.Min {
		if i.Max < v || i.Min > v {
			return nil, &OutOfRangeError{Got: v, Min: i.Min, Max: i.Max}
		}
	}

	return v, nil
}

func (i *IntArg) HelpName() string {
	return "Whole number"
}

// FloatArg matches and parses float arguments
// If min and max are not equal then the value has to be within min and max or else it will fail parsing
type FloatArg struct {
	Min, Max float64
}

func (f *FloatArg) Matches(part string) bool {
	_, err := strconv.ParseFloat(part, 64)
	return err == nil
}
func (f *FloatArg) Parse(part string, data *Data) (interface{}, error) {
	v, err := strconv.ParseFloat(part, 64)
	if err != nil {
		return nil, &InvalidFloat{part}
	}

	// A valid range has been specified
	if f.Max != f.Min {
		if f.Max < v || f.Min > v {
			return nil, &OutOfRangeError{Got: v, Min: f.Min, Max: f.Max, Float: true}
		}
	}

	return v, nil
}

func (f *FloatArg) HelpName() string {
	return "Decimal number"
}

// StringArg matches and parses float arguments
type StringArg struct{}

func (s *StringArg) Matches(part string) bool                           { return true }
func (s *StringArg) Parse(part string, data *Data) (interface{}, error) { return part, nil }
func (s *StringArg) HelpName() string {
	return "Text/String"
}

// UserArg matches and parses user argument, optionally searching for the member if RequireMention is false
type UserArg struct {
	RequireMention bool
}

func (u *UserArg) Matches(part string) bool {
	if u.RequireMention {
		return strings.HasPrefix(part, "<@") && strings.HasSuffix(part, ">")
	}

	// username searches are enabled, any string can be used
	return true
}

func (u *UserArg) Parse(part string, data *Data) (interface{}, error) {
	if strings.HasPrefix(part, "<@") {
		// Direct mention
		id := part[2 : len(part)-1]
		if id[0] == '!' {
			// Nickname mention
			id = id[1:]
		}

		parsed, _ := strconv.ParseInt(id, 10, 64)
		for _, v := range data.Msg.Mentions {
			if parsed == v.ID {
				return v, nil
			}
		}
		return nil, &ImproperMention{part}
	} else if !u.RequireMention {
		// Search for username
		data.GS.RLock()
		u, err := FindDiscordUserByName(part, data.GS.Members)
		data.GS.RUnlock()
		return u, err
	}

	return nil, &ImproperMention{part}
}

func (u *UserArg) HelpName() string {
	if u.RequireMention {
		return "User Mention"
	}
	return "User"
}

func FindDiscordUserByName(str string, members map[int64]*dstate.MemberState) (*discordgo.User, error) {
	for _, v := range members {
		if v == nil {
			continue
		}

		var user *discordgo.User
		if v.Member != nil {
			user = v.Member.User
		} else if v.Presence != nil {
			user = v.Presence.User
		}

		if user == nil || user.Username == "" {
			continue
		}

		if strings.EqualFold(str, user.Username) {
			cop := new(discordgo.User)
			*cop = *user
			return cop, nil
		}
	}

	return nil, &UserNotFound{str}
}
