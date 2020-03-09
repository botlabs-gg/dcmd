package dcmd

import (
	"strconv"
	"strings"

	"github.com/jonas747/discordgo"
	"github.com/jonas747/dstate"
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

func (p *ParsedArg) MemberState() *dstate.MemberState {
	if p.Value == nil {
		return nil
	}

	switch t := p.Value.(type) {
	case *dstate.MemberState:
		return t
	case *AdvUserMatch:
		return t.Member
	}

	return nil
}

func (p *ParsedArg) User() *discordgo.User {
	if p.Value == nil {
		return nil
	}

	switch t := p.Value.(type) {
	case *dstate.MemberState:
		return t.DGoUser()
	case *AdvUserMatch:
		return t.User
	}

	return nil
}

func (p *ParsedArg) AdvUser() *AdvUserMatch {
	if p.Value == nil {
		return nil
	}

	switch t := p.Value.(type) {
	case *AdvUserMatch:
		return t
	}

	return nil
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
	Int             = &IntArg{}
	Float           = &FloatArg{}
	String          = &StringArg{}
	User            = &UserArg{}
	UserReqMention  = &UserArg{RequireMention: true}
	UserID          = &UserIDArg{}
	Channel         = &ChannelArg{}
	AdvUser         = &AdvUserArg{EnableUserID: true, EnableUsernameSearch: true, RequireMembership: true}
	AdvUserNoMember = &AdvUserArg{EnableUserID: true, EnableUsernameSearch: true}
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

func (s *StringArg) Matches(def *ArgDef, part string) bool { return true }
func (s *StringArg) Parse(def *ArgDef, part string, data *Data) (interface{}, error) {
	return part, nil
}
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
	if strings.HasPrefix(part, "<@") && len(part) > 3 {
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
		m, err := FindDiscordMemberByName(data.GS, part)
		if m != nil {
			return m.DGoUser(), nil
		}
		return nil, err
	}

	return nil, &ImproperMention{part}
}

func (u *UserArg) HelpName() string {
	if u.RequireMention {
		return "User Mention"
	}
	return "User"
}

func FindDiscordMemberByName(gs *dstate.GuildState, str string) (*dstate.MemberState, error) {
	gs.RLock()
	defer gs.RUnlock()

	lowerIn := strings.ToLower(str)

	partialMatches := make([]*dstate.MemberState, 0, 5)
	fullMatches := make([]*dstate.MemberState, 0, 5)

	for _, v := range gs.Members {
		if v == nil {
			continue
		}

		if v.Username == "" {
			continue
		}

		if strings.EqualFold(str, v.Username) || strings.EqualFold(str, v.Nick) {
			fullMatches = append(fullMatches, v.Copy())
			if len(fullMatches) >= 5 {
				break
			}
		} else if len(partialMatches) < 5 {
			if strings.Contains(strings.ToLower(v.Username), lowerIn) {
				partialMatches = append(partialMatches, v)
			}
		}
	}

	if len(fullMatches) == 1 {
		return fullMatches[0].Copy(), nil
	}

	if len(fullMatches) == 0 && len(partialMatches) == 0 {
		return nil, &UserNotFound{str}
	}

	out := ""
	for _, v := range fullMatches {
		if out != "" {
			out += ", "
		}

		out += "`" + v.Username + "`"
	}

	for _, v := range partialMatches {
		if out != "" {
			out += ", "
		}

		out += "`" + v.Username + "`"
	}

	if len(fullMatches) > 1 {
		return nil, NewSimpleUserError("Too many users with that name, " + out + ". Please re-run the command with a narrower search, mention or ID.")
	}

	return nil, NewSimpleUserError("Did you mean one of these? " + out + ". Please re-run the command with a narrower search, mention or ID")
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
	if strings.HasPrefix(part, "<@") && len(part) > 3 {
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
	if strings.HasPrefix(part, "<#") && len(part) > 3 {
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

type AdvUserMatch struct {
	// Member may not be present if "RequireMembership" is false
	Member *dstate.MemberState

	// User is always present
	User *discordgo.User
}

func (a *AdvUserMatch) UsernameOrNickname() string {
	if a.Member != nil {
		if a.Member.Nick != "" {
			return a.Member.Nick
		}
	}

	return a.User.Username
}

// AdvUserArg is a more advanced version of UserArg and UserIDArg, it will return a AdvUserMatch
type AdvUserArg struct {
	EnableUserID         bool // Whether to check for user IDS
	EnableUsernameSearch bool // Whether to search for usernames
	RequireMembership    bool // Whether this requires a membership of the server, if set then Member will always be populated
}

func (u *AdvUserArg) Matches(def *ArgDef, part string) bool {
	if strings.HasPrefix(part, "<@") && strings.HasSuffix(part, ">") {
		return true
	}

	if u.EnableUserID {
		_, err := strconv.ParseInt(part, 10, 64)
		if err == nil {
			return true
		}
	}

	if u.EnableUsernameSearch {
		// username search
		return true
	}

	return false
}

func (u *AdvUserArg) Parse(def *ArgDef, part string, data *Data) (interface{}, error) {

	var user *discordgo.User
	var ms *dstate.MemberState

	// check mention
	if strings.HasPrefix(part, "<@") && len(part) > 3 {
		user = u.ParseMention(def, part, data)
	}

	msFailed := false
	if user == nil && u.EnableUserID {
		// didn't find a match in the previous step
		// try userID search
		if parsed, err := strconv.ParseInt(part, 10, 64); err == nil {
			ms, user = u.SearchID(parsed, data)
			if ms == nil {
				msFailed = true
			}
		}
	}

	if u.EnableUsernameSearch && data.GS != nil && ms == nil && user == nil {
		// Search for username
		var err error
		ms, err = FindDiscordMemberByName(data.GS, part)
		if err != nil {
			return nil, err
		}
	}

	if ms == nil && user == nil {
		return nil, NewSimpleUserError("User/Member not found")
	}

	if ms != nil && user == nil {
		user = ms.DGoUser()
	} else if ms == nil && user != nil && !msFailed {
		ms, user = u.SearchID(user.ID, data)
	}

	return &AdvUserMatch{
		Member: ms,
		User:   user,
	}, nil
}

func (u *AdvUserArg) SearchID(parsed int64, data *Data) (member *dstate.MemberState, user *discordgo.User) {

	if data.GS != nil {
		// attempt to fetch member
		member = data.GS.MemberCopy(true, parsed)
		if member != nil {
			return member, member.DGoUser()
		}

		m, err := data.Session.GuildMember(data.GS.ID, parsed)
		if err == nil {
			member = dstate.MSFromDGoMember(data.GS, m)
			return member, m.User
		}
	}

	if u.RequireMembership {
		return nil, nil
	}

	// fallback to standard user
	user, _ = data.Session.User(parsed)
	return
}

func (u *AdvUserArg) ParseMention(def *ArgDef, part string, data *Data) (user *discordgo.User) {
	// Direct mention
	id := part[2 : len(part)-1]
	if id[0] == '!' {
		// Nickname mention
		id = id[1:]
	}

	parsed, _ := strconv.ParseInt(id, 10, 64)
	for _, v := range data.Msg.Mentions {
		if parsed == v.ID {
			return v
		}
	}

	return nil
}

func (u *AdvUserArg) HelpName() string {
	out := "User mention"
	if u.EnableUsernameSearch {
		out += "/Name"
	}
	if u.EnableUserID {
		out += "/ID"
	}

	return out
}
