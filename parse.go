package dcmd

import (
	"github.com/pkg/errors"
	"strings"
)

var (
	ErrNoComboFound       = errors.New("No matching combo found")
	ErrNotEnoughArguments = errors.New("No enough arguments passed")
)

func ArgParserMW(inner RunFunc) RunFunc {
	return func(data *Data) (interface{}, error) {
		// Parse Args
		err := ParseCmdArgs(data)
		if err != nil {
			return nil, err
		}

		return inner(data)
	}
}

// ParseCmdArgs is the standard argument parser
// todo, more doc on the format
func ParseCmdArgs(data *Data) error {
	argDefsCommand, argDefsOk := data.Cmd.(CmdWithArgDefs)
	switchesCmd, switchesOk := data.Cmd.(CmdWithSwitches)

	if !argDefsOk && !switchesOk {
		// Command dosen't use the standard arg parsing
		return nil
	}

	// Split up the args
	split := SplitArgs(data.MsgStrippedPrefix)

	var err error
	if switchesOk {
		// Parse the switches first
		split, err = ParseSwitches(switchesCmd.Switches(), data, split)
		if err != nil {
			return err
		}
	}

	if argDefsOk {
		defs, req, combos := argDefsCommand.ArgDefs()
		err = ParseArgDefs(defs, req, combos, data, split)
		if err != nil {
			return err
		}
	}

	return nil
}

// ParseArgDefs parses ordered argument definition for a CmdWithArgDefs
func ParseArgDefs(defs []*ArgDef, required int, combos [][]int, data *Data, split []*RawArg) error {

	combo, ok := FindCombo(defs, combos, split)
	if !ok {
		return ErrNoComboFound
	}

	parsedArgs := NewParsedArgs(defs)

	for i, v := range combo {
		def := defs[v]
		if i >= len(split) {
			return ErrNotEnoughArguments
		}

		combined := ""
		if i == len(combo)-1 && len(split)-1 > i {
			// Last arg, but still more after, combine and rebuilt them
			for j := i; j < len(split); j++ {
				if j != i {
					combined += " "
				}

				temp := split[j]
				if temp.Container != 0 {
					combined += string(temp.Container) + temp.Str + string(temp.Container)
				} else {
					combined += temp.Str
				}
			}
		} else {
			combined = split[i].Str
		}

		v, err := def.Type.Parse(combined, data)
		if err != nil {
			return err
		}
		parsedArgs[i].Value = v
	}

	data.Args = parsedArgs

	return nil
}

// ParseSwitches parses all switches for a CmdWithSwitches, and also takes them out of the raw args
func ParseSwitches(switches []*ArgDef, data *Data, split []*RawArg) ([]*RawArg, error) {
	newRaws := make([]*RawArg, 0, len(split))

	// Initialise the parsed switches
	parsedSwitches := make(map[string]*ParsedArg)
	for _, v := range switches {
		parsedSwitches[v.Name] = &ParsedArg{
			Value: v.Default,
			Def:   v,
		}
	}

	for i := 0; i < len(split); i++ {
		raw := split[i]
		if raw.Container == 0 {
			newRaws = append(newRaws, raw)
			continue
		}

		if !strings.HasPrefix(raw.Str, "-") {
			newRaws = append(newRaws, raw)
			continue
		}

		rest := raw.Str[1:]
		var matchedArg *ArgDef
		for _, v := range switches {
			if v.Name == rest {
				matchedArg = v
				break
			}
		}

		if matchedArg == nil {
			newRaws = append(newRaws, raw)
			continue
		}

		if matchedArg.Type == nil {
			parsedSwitches[matchedArg.Name].Raw = raw
			parsedSwitches[matchedArg.Name].Value = true
			continue
		}

		if i >= len(split)-1 {
			// A switch with extra stuff requird, but no extra data provided
			// Can't handle this case...
			continue
		}

		// At this point, we have encountered a switch with data
		// so we need to skip the next RawArg
		i++

		val, err := matchedArg.Type.Parse(split[i].Str, data)
		if err != nil {
			// TODO: Use custom error type for helpfull errror
			return nil, err
		}

		parsedSwitches[matchedArg.Name].Raw = raw
		parsedSwitches[matchedArg.Name].Value = val
	}
	data.Switches = parsedSwitches
	return newRaws, nil
}

var (
	ArgContainers = []rune{
		'"',
		'`',
	}
)

type RawArg struct {
	Str       string
	Container rune
}

// SplitArgs splits the string into fields
func SplitArgs(in string) []*RawArg {
	rawArgs := make([]*RawArg, 0)

	curBuf := ""
	escape := false
	var container rune
	for _, r := range in {
		// Apply or remove escape mode
		if r == '\\' {
			if escape {
				escape = false
				curBuf += "\\"
			} else {
				escape = true
			}

			continue
		}

		// Check for other special tokens
		isSpecialToken := false
		if !escape {
			isSpecialToken = true

			if r == ' ' {
				// Maybe seperate by space
				if curBuf != "" && container == 0 {
					rawArgs = append(rawArgs, &RawArg{curBuf, 0})
					curBuf = ""
				} else if container != 0 { // If it is quoted proceed as it was a normal rune
					isSpecialToken = false
				}
			} else if r == container && container != 0 && !escape {
				// Split arg here
				rawArgs = append(rawArgs, &RawArg{curBuf, container})
				curBuf = ""
				container = 0
			} else if container == 0 {
				// Check if we should start containing a arg
				for _, v := range ArgContainers {
					if v == r {
						container = v
						break
					}
				}

				if container == 0 {
					isSpecialToken = false
				}
			} else {
				isSpecialToken = false
			}
		}

		if !isSpecialToken {
			curBuf += string(r)
		}

		// Reset escape mode
		escape = false
	}

	// Something was left in the buffer just add it to the end
	if curBuf != "" {
		if container != 0 {
			curBuf = string(container) + curBuf
		}
		rawArgs = append(rawArgs, &RawArg{curBuf, 0})
	}

	return rawArgs
}

// Finds a proper argument combo from the provided args
func FindCombo(defs []*ArgDef, combos [][]int, args []*RawArg) (combo []int, ok bool) {

	if len(combos) < 1 {
		out := make([]int, len(defs))
		for k, _ := range out {
			out[k] = k
		}
		return out, true
	}

	var selectedCombo []int

	// Find a possible match
OUTER:
	for _, combo := range combos {
		if len(combo) > len(args) {
			// No match
			continue
		}

		// See if this combos arguments matches that of the parsed command
		for i, comboArg := range combo {
			def := defs[comboArg]

			if !def.Type.Matches(args[i].Str) {
				continue OUTER
			}
		}

		// We got a match, if this match is stronger than the last one set it as selected
		if len(combo) > len(selectedCombo) || !ok {
			selectedCombo = combo
			ok = true
		}
	}

	return selectedCombo, ok
}

// Parses a command into a ParsedCommand
// Arguments are split at space or you can put arguments inside quotes
// You can escape both space and quotes using '\"' or '\ ' ('\\' to escape the escaping)
// Quotes in the middle of an argument is trated as a normal character and not a seperator
// func (sc *Command) ParseCommand(raw string, triggerData *TriggerData) (*ExecData, error) {

// 	data := &ExecData{
// 		Session: triggerData.Session,
// 		Message: triggerData.Message,
// 		Command: sc,
// 		State:   triggerData.DState,
// 	}

// 	// Retrieve guild and channel if possible (session not provided in testing)
// 	var channel *dstate.ChannelState
// 	var guild *dstate.GuildState

// 	if triggerData.DState != nil {
// 		channel = triggerData.DState.Channel(true, triggerData.Message.ChannelID)
// 		data.Channel = channel

// 		guild = channel.Guild
// 		data.Guild = guild
// 		if guild != nil {
// 			guild.RLock()
// 			defer guild.RUnlock()
// 		}
// 	}

// 	// No arguments needed
// 	if len(sc.Arguments) < 1 {
// 		return data, nil
// 	}

// 	// Strip away the command name (or alias if that was what triggered it)
// 	buf := ""
// 	if sc.Name != "" {
// 		split := strings.SplitN(raw, " ", 2)
// 		if len(split) < 1 {
// 			return nil, errors.New("Command not specified")
// 		}

// 		if strings.EqualFold(split[0], strings.ToLower(sc.Name)) {
// 			buf = raw[len(strings.ToLower(sc.Name)):]
// 		} else {
// 			for _, alias := range sc.Aliases {
// 				if strings.EqualFold(alias, split[0]) {
// 					buf = raw[len(strings.ToLower(alias)):]
// 					break
// 				}
// 			}
// 		}
// 	}

// 	buf = strings.TrimSpace(buf)
// 	parsedArgs := make([]*ParsedArgument, len(sc.Arguments))
// 	for i, v := range sc.Arguments {
// 		if v.Default != nil {
// 			parsedArgs[i] = &ParsedArgument{Parsed: v.Default}
// 		}
// 	}

// 	data.Args = parsedArgs

// 	// No parameters provided, and none required, just handle the mofo
// 	if buf == "" {
// 		if sc.RequiredArgs == 0 && len(sc.ArgumentCombos) < 1 {
// 			return data, nil
// 		} else {
// 			if len(sc.ArgumentCombos) < 1 {
// 				err := sc.ErrMissingArgs(0)
// 				return nil, err
// 			} else {
// 				// Check if one of the combox accepts zero arguments
// 				for _, combo := range sc.ArgumentCombos {
// 					if len(combo) < 1 {
// 						return data, nil
// 					}
// 				}
// 			}

// 			return nil, ErrInvalidParameters
// 		}
// 	}

// 	rawArgs := ReadArgs(buf)
// 	selectedCombo, ok := sc.findCombo(rawArgs)
// 	if !ok {
// 		if len(sc.ArgumentCombos) < 1 {
// 			err := sc.ErrMissingArgs(len(rawArgs))
// 			return nil, err
// 		}
// 		return nil, ErrInvalidParameters
// 	}

// 	// Parse the arguments and fill up the PArsedArgs slice
// 	for k, comboArg := range selectedCombo {
// 		var val interface{}
// 		var err error

// 		buf := rawArgs[k].Raw.Str
// 		// If last arg att all the remaning rawargs, building up the
// 		if k == len(selectedCombo)-1 {
// 			for i := k + 1; i < len(rawArgs); i++ {
// 				switch rawArgs[i].Raw.Seperator {
// 				case ArgSeperatorSpace:
// 					buf += " " + rawArgs[i].Raw.Str
// 				case ArgSeperatorQuote:
// 					buf += " \"" + rawArgs[i].Raw.Str + "\""
// 				}
// 			}
// 		}

// 		switch sc.Arguments[comboArg].Type {
// 		case ArgumentString:
// 			val = buf
// 		case ArgumentNumber:
// 			val, err = ParseNumber(buf)
// 		case ArgumentUser:
// 			if channel == nil || channel.Channel.IsPrivate {
// 				continue // can't provide users in direct messages
// 			}
// 			val, err = ParseUser(buf, triggerData.Message, guild, sc.UserArgRequireMention)
// 		}

// 		if err != nil {
// 			return nil, errors.New("Failed parsing arguments: " + err.Error())
// 		}

// 		parsedArgs[comboArg] = &ParsedArgument{
// 			Raw:    buf,
// 			Parsed: val,
// 		}
// 	}

// 	return data, nil
// }

// func (sc *Command) ErrMissingArgs(provided int) error {
// 	names := ""
// 	for i, v := range sc.Arguments {
// 		if i < provided {
// 			continue
// 		}

// 		if i != provided {
// 			names += ", "
// 		}

// 		if i > sc.RequiredArgs {
// 			names += "(optional)"
// 		}
// 		names += v.Name

// 	}

// 	return fmt.Errorf("Missing arguments: %s.", names)
// }

// func (sc *Command) checkArgumentMatch(raw *MatchedArg, definition ArgumentType) bool {
// 	switch definition {
// 	case ArgumentNumber:
// 		return raw.Type == ArgumentNumber
// 	case ArgumentUser:
// 		// Check if a user mention is required
// 		// Otherwise it can be of any type
// 		if sc.UserArgRequireMention {
// 			return raw.Type == ArgumentUser
// 		} else {
// 			return true
// 		}
// 	case ArgumentString:
// 		// Both number and user can be a string
// 		// So it willl always match string no matter what
// 		return true
// 	}

// 	return false
// }
