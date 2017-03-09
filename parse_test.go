package dcmd

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseArgDefs(t *testing.T) {
	cases := []struct {
		name         string
		input        string
		defs         []*ArgDef
		expectedArgs []*ParsedArg
	}{
		{"simple int", "15", []*ArgDef{{Type: Int}}, []*ParsedArg{{Value: int64(15)}}},
		{"simple float", "15.5", []*ArgDef{{Type: Float}}, []*ParsedArg{{Value: float64(15.5)}}},
		{"simple string", "hello", []*ArgDef{{Type: String}}, []*ParsedArg{{Value: "hello"}}},
		{"int float", "15 30.5", []*ArgDef{{Type: Int}, {Type: Float}}, []*ParsedArg{{Value: int64(15)}, {Value: float64(30.5)}}},
		{"string int", "hey_man 30", []*ArgDef{{Type: String}, {Type: Int}}, []*ParsedArg{{Value: "hey_man"}, {Value: int64(30)}}},
		{"quoted strings", "first `middle quoted` last", []*ArgDef{{Type: String}, {Type: String}, {Type: String}}, []*ParsedArg{{Value: "first"}, {Value: "middle quoted"}, {Value: "last"}}},
	}

	for i, v := range cases {
		t.Run(fmt.Sprintf("#%d-%s", i, v.name), func(t *testing.T) {
			d := new(Data)
			err := ParseArgDefs(v.defs, 0, nil, d, SplitArgs(v.input))

			if err != nil {
				t.Fatal("ParseArgDefs returned a bad error", err)
			}

			// Check if we got the expected output
			for i, ea := range v.expectedArgs {
				if i >= len(d.Args) {
					t.Fatal("Unexpected end of parsed args")
				}

				assert.Equal(t, ea.Value, d.Args[i].Value, "Should be equal")
			}
		})
	}
}
