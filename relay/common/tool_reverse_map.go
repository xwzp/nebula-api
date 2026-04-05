package common

import "strings"

// ContextKeyToolReverseMap is the gin context key for tool name reverse mapping.
// When set, response handlers should reverse-map Claude Code CLI tool names
// back to the original client tool names.
const ContextKeyToolReverseMap = "tool_name_reverse_map"

// ReverseMapToolNamesInJSON replaces tool names in a raw JSON string using
// the provided reverse mapping. Used for Claude-format responses where the
// raw JSON is forwarded directly.
func ReverseMapToolNamesInJSON(data string, reverseMap map[string]string) string {
	for from, to := range reverseMap {
		data = strings.ReplaceAll(data, `"name":"`+from+`"`, `"name":"`+to+`"`)
		data = strings.ReplaceAll(data, `"name": "`+from+`"`, `"name": "`+to+`"`)
	}
	return data
}
