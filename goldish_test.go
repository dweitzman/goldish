package goldish

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	CheckGoldens(t, []string{"keys", "text"}, []string{"ast"},
		func(args map[string]string) (map[string]string, error) {
			var keys []string
			err := json.Unmarshal([]byte(args["keys"]), &keys)
			if err != nil {
				return nil, err
			}

			parsedC, caseErr := parse(strings.NewReader(args["text"]))

			result := make(map[string]string)
			if caseErr != nil {
				result["ast"] = caseErr.Error()
			} else {
				gotAst, err := json.MarshalIndent(parsedC, "", "  ")
				if err != nil {
					return nil, err
				}
				result["ast"] = string(gotAst)
			}
			return result, nil
		},
	)
}

func TestParseAndSerialize(t *testing.T) {
	CheckGoldens(t, []string{"keys", "text"}, []string{"serialized"},
		func(args map[string]string) (map[string]string, error) {
			var keys []string
			err := json.Unmarshal([]byte(args["keys"]), &keys)
			if err != nil {
				return nil, err
			}

			parsedC, err := parse(strings.NewReader(args["text"]))
			if err != nil {
				return nil, err
			}

			buf := bytes.NewBuffer(nil)
			err = parsedC.Write(buf, keys)
			if err != nil {
				return nil, err
			}

			result := make(map[string]string)
			result["serialized"] = buf.String()
			return result, nil
		},
	)
}
