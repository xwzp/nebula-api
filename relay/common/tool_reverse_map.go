package common

import "strings"

// ContextKeyToolReverseMap is the gin context key for tool name reverse mapping.
// When set, response handlers should reverse-map Claude Code CLI tool names
// back to the original client tool names.
const ContextKeyToolReverseMap = "tool_name_reverse_map"

// ContextKeyParamReverseMap is the gin context key for parameter name reverse mapping.
// When set, response handlers should reverse-map remapped parameter names
// back to original client parameter names in tool_use input blocks.
const ContextKeyParamReverseMap = "param_name_reverse_map"

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

// ReverseMapParamNamesInJSON replaces parameter names (JSON object keys) in
// a raw JSON string.  Matches the pattern "KEY": which is how JSON keys appear.
// Only replaces keys that are in the reverse map (short random tokens like
// "h1", "g2", etc.) so false-positive risk is minimal.
func ReverseMapParamNamesInJSON(data string, reverseMap map[string]string) string {
	for from, to := range reverseMap {
		data = strings.ReplaceAll(data, `"`+from+`":`, `"`+to+`":`)
		data = strings.ReplaceAll(data, `"`+from+`": `, `"`+to+`": `)
	}
	return data
}
