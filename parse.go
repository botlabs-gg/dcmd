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
	argDefsCommand, argDefsOk := data.Cmd.Command.(CmdWithArgDefs)
	switchesCmd, switchesOk := data.Cmd.Command.(CmdWithSwitches)

	if !argDefsOk && !switchesOk {
		// Command dosen't use the standard arg parsing
		return nil
	}

	// Split up the args
	split := SplitArgs(data.MsgStrippedPrefix)

	var err error
	if switchesOk {
		switches := switchesCmd.Switches()
		if len(switches) > 0 {
			// Parse the switches first
			split, err = ParseSwitches(switchesCmd.Switches(), data, split)
			if err != nil {
				return err
			}
		}
	}

	if argDefsOk {
		defs, req, combos := argDefsCommand.ArgDefs(data)
		if len(defs) > 0 {
			err = ParseArgDefs(defs, req, combos, data, split)
			if err != nil {
				return err
			}
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
			if i >= required && len(combos) < 1 {
				break
			}
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

		val, err := def.Type.Parse(combined, data)
		if err != nil {
			return err
		}
		parsedArgs[v].Value = val
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
		parsedSwitches[v.Switch] = &ParsedArg{
			Value: v.Default,
			Def:   v,
		}
	}

	for i := 0; i < len(split); i++ {
		raw := split[i]
		if raw.Container != 0 {
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
			if v.Switch == rest {
				matchedArg = v
				break
			}
		}

		if matchedArg == nil {
			newRaws = append(newRaws, raw)
			continue
		}

		if matchedArg.Type == nil {
			parsedSwitches[matchedArg.Switch].Raw = raw
			parsedSwitches[matchedArg.Switch].Value = true
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

		parsedSwitches[matchedArg.Switch].Raw = raw
		parsedSwitches[matchedArg.Switch].Value = val
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
