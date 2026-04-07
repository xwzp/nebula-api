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
// the provided reverse mapping. Handles both unescaped and escaped forms.
func ReverseMapToolNamesInJSON(data string, reverseMap map[string]string) string {
	for from, to := range reverseMap {
		// Unescaped form (top-level SSE JSON)
		data = strings.ReplaceAll(data, `"name":"`+from+`"`, `"name":"`+to+`"`)
		data = strings.ReplaceAll(data, `"name": "`+from+`"`, `"name": "`+to+`"`)
		// Escaped form (inside JSON string values)
		data = strings.ReplaceAll(data, `\"name\":\"`+from+`\"`, `\"name\":\"`+to+`\"`)
		data = strings.ReplaceAll(data, `\"name\": \"`+from+`\"`, `\"name\": \"`+to+`\"`)
	}
	return data
}

// ReverseMapParamNamesInJSON replaces parameter names (JSON object keys) in
// a raw JSON string.  Handles both unescaped ("key":) and escaped (\"key\":)
// forms because in SSE streaming the partial_json value is a JSON-encoded
// string where inner quotes are backslash-escaped.
func ReverseMapParamNamesInJSON(data string, reverseMap map[string]string) string {
	for from, to := range reverseMap {
		// Unescaped form: "key": or "key":  (in non-streaming / parsed JSON)
		data = strings.ReplaceAll(data, `"`+from+`":`, `"`+to+`":`)
		data = strings.ReplaceAll(data, `"`+from+`": `, `"`+to+`": `)
		// Escaped form: \"key\": or \"key\":  (inside SSE partial_json strings)
		data = strings.ReplaceAll(data, `\"`+from+`\":`, `\"`+to+`\":`)
		data = strings.ReplaceAll(data, `\"`+from+`\": `, `\"`+to+`\": `)
	}
	return data
}
