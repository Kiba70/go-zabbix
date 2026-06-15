package zabbix

import (
	"fmt"
	"strings"
)

type ZBXBoolean bool

func (bit *ZBXBoolean) UnmarshalJSON(data []byte) error {
	// Zabbix API encodes booleans as quoted strings: "1", "0", "true", "false".
	// Also accept unquoted forms for robustness.
	asString := strings.Trim(string(data), `"`)
	switch asString {
	case "1", "true":
		*bit = true
	case "0", "false":
		*bit = false
	default:
		return fmt.Errorf("Boolean unmarshal error: invalid input %s", asString)
	}
	return nil
}

func (bit *ZBXBoolean) MarshalJSON() ([]byte, error) {
	if *bit {
		return []byte("\"1\""), nil
	}
	return []byte("\"0\""), nil
}
