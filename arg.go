package dcmd

import (
	"github.com/jonas747/discordgo"
	"github.com/jonas747/dstate"
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
	if p.Value == nil {
		return ""
	}

	switch t := p.Value.(type) {
	case string:
		return t
	case int, int32, int64, uint, uint32, uint64:
		return strconv.FormatInt(p.Int64(), 10)
	default:
		return ""
	}
}

// TODO: GO-Generate the number ones
func (p *ParsedArg) Int() int {
	if p.Value == nil {
		return 0
	}

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
	if p.Value == nil {
		return 0
	}

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

func (p *ParsedArg) Bool() bool {
	if p.Value == nil {
		return false
	}

	switch t := p.Value.(type) {
	case bool:
		return t
	case int, int32, int64, uint, uint32, uint64:
		return p.Int64() > 0
	case string:
		return t != ""
	}

	return false
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
	Matches(def *ArgDef, part string) bool

	// Attempt to parse it, returning any error if one occured.
	Parse(def *ArgDef, part string, data *Data) (val interface{}, err error)

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
	UserID         = &UserIDArg{}
	Channel        = &ChannelArg{}
)

// IntArg matches and parses integer arguments
// If min and max are not equal then the value has to be within min and max or else it will fail parsing
type IntArg struct {
	Min, Max int64
}

func (i *IntArg) Matches(def *ArgDef, part string) bool {
	_, err := strconv.ParseInt(part, 10, 64)
	return err == nil
}
func (i *IntArg) Parse(def *ArgDef, part string, data *Data) (interface{}, error) {
	v, err := strconv.ParseInt(part, 10, 64)
	if err != nil {
		return nil, &InvalidInt{part}
	}

	// A valid range has been specified
	if i.Max != i.Min {
		if i.Max < v || i.Min > v {
			return nil, &OutOfRangeError{ArgName: def.Name, Got: v, Min: i.Min, Max: i.Max}
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

func (f *FloatArg) Matches(def *ArgDef, part string) bool {
	_, err := strconv.ParseFloat(part, 64)
	return err == nil
}
func (f *FloatArg) Parse(def *ArgDef, part string, data *Data) (interface{}, error) {
	v, err := strconv.ParseFloat(part, 64)
	if err != nil {
		return nil, &InvalidFloat{part}
	}

	// A valid range has been specified
	if f.Max != f.Min {
		if f.Max < v || f.Min > v {
			return nil, &OutOfRangeError{ArgName: def.Name, Got: v, Min: f.Min, Max: f.Max, Float: true}
		}
	}

	return v, nil
}

func (f *FloatArg) HelpName() string {
	return "Decimal number"
}

// StringArg matches and parses float arguments
type StringArg struct{}

func (s *StringArg) Matches(def *ArgDef, part string) bool                           { return true }
func (s *StringArg) Parse(def *ArgDef, part string, data *Data) (interface{}, error) { return part, nil }
func (s *StringArg) HelpName() string {
	return "Text"
}

// UserArg matches and parses user argument, optionally searching for the member if RequireMention is false
type UserArg struct {
	RequireMention bool
}

func (u *UserArg) Matches(def *ArgDef, part string) bool {
	if u.RequireMention {
		return strings.HasPrefix(part, "<@") && strings.HasSuffix(part, ">")
	}

	// username searches are enabled, any string can be used
	return true
}

func (u *UserArg) Parse(def *ArgDef, part string, data *Data) (interface{}, error) {
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
	} else if !u.RequireMention && data.GS != nil {
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

		if v.Username == "" {
			continue
		}

		if strings.EqualFold(str, v.Username) {
			return v.DGoUser(), nil
		}
	}

	return nil, &UserNotFound{str}
}

// UserIDArg matches a mention or a plain id, the user does not have to be a part of the server
// The type of the ID is parsed into a int64
type UserIDArg struct{}

func (u *UserIDArg) Matches(def *ArgDef, part string) bool {
	// Check for mention
	if strings.HasPrefix(part, "<@") && strings.HasSuffix(part, ">") {
		return true
	}

	// Check for ID
	_, err := strconv.ParseInt(part, 10, 64)
	if err == nil {
		return true
	}

	return false
}

func (u *UserIDArg) Parse(def *ArgDef, part string, data *Data) (interface{}, error) {
	if strings.HasPrefix(part, "<@") {
		// Direct mention
		id := part[2 : len(part)-1]
		if id[0] == '!' {
			// Nickname mention
			id = id[1:]
		}

		parsed, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return nil, &ImproperMention{part}
		}

		return parsed, nil
	}

	id, err := strconv.ParseInt(part, 10, 64)
	if err == nil {
		return id, nil
	}

	return nil, &ImproperMention{part}
}

func (u *UserIDArg) HelpName() string {
	return "Mention/ID"
}

// UserIDArg matches a mention or a plain id, the user does not have to be a part of the server
// The type of the ID is parsed into a int64
type ChannelArg struct{}

func (ca *ChannelArg) Matches(def *ArgDef, part string) bool {
	// Check for mention
	if strings.HasPrefix(part, "<#") && strings.HasSuffix(part, ">") {
		return true
	}

	// Check for ID
	_, err := strconv.ParseInt(part, 10, 64)
	if err == nil {
		return true
	}

	return false
}

func (ca *ChannelArg) Parse(def *ArgDef, part string, data *Data) (interface{}, error) {
	if data.GS == nil {
		return nil, nil
	}

	var cID int64
	if strings.HasPrefix(part, "<#") {
		// Direct mention
		id := part[2 : len(part)-1]

		parsed, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return nil, &ImproperMention{part}
		}

		cID = parsed
	} else {
		id, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return nil, &ImproperMention{part}
		}
		cID = id
	}

	data.GS.RLock()
	if c, ok := data.GS.Channels[cID]; ok {
		data.GS.RUnlock()
		return c, nil
	}
	data.GS.RUnlock()

	return nil, &ImproperMention{part}
}

func (ca *ChannelArg) HelpName() string {
	return "Channel"
}
