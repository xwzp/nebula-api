package claude_oauth

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/QuantumNous/new-api/common"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
)

// senderMetadataRe matches the OpenClaw "Sender (untrusted metadata)" block
// followed by an optional timestamp like [Sun 2026-04-05 11:53 GMT+8].
// The entire match is stripped, leaving only the actual user message.
var senderMetadataRe = regexp.MustCompile("(?s)Sender \\(untrusted metadata\\):\\s*```json\\s*\\{[^}]*\\}\\s*```\\s*(?:\\[[^\\]]*\\]\\s*)?")

// openClawRe matches "OpenClaw" or "openclaw" (case-insensitive).
var openClawRe = regexp.MustCompile("(?i)openclaw")

// ContextKeyToolReverseMap re-exports the shared context key for convenience.
const ContextKeyToolReverseMap = relaycommon.ContextKeyToolReverseMap

// ---------------------------------------------------------------------------
// Tool name mapping configuration
//
// There are two categories:
//
// 1. Intersection tools — tools that have a direct counterpart in Claude Code
//    CLI with a DIFFERENT name. These are explicitly mapped.
//
// 2. All other tools — renamed by capitalizing the first letter (e.g.
//    "process" → "Process", "browser" → "Browser"). This makes them look
//    like Claude Code CLI naming convention without needing a manual entry.
//
// To DISABLE a specific tool (drop it entirely), add it to droppedTools.
// To add a custom mapping, add it to intersectionForwardMap.
// ---------------------------------------------------------------------------

// intersectionForwardMap maps OpenClaw tool names → upstream tool names.
//
// - Intersection tools (have a Claude Code CLI counterpart): mapped to the CLI name.
// - Non-intersection tools (OpenClaw-only): mapped to a PascalCase name to
//   match Claude Code CLI naming convention.
//
// To test which tools trigger detection, move entries to droppedTools.
var intersectionForwardMap = map[string]string{
	// ---- Intersection: tools that exist in both OpenClaw and Claude Code CLI ----
	"read":       "Read",
	"write":      "Write",
	"edit":       "Edit",
	"exec":       "Bash",
	"web_search": "WebSearch",
	"web_fetch":  "WebFetch",

	// ---- Non-intersection: OpenClaw-only tools, renamed to PascalCase ----
	"process":          "Process",
	"canvas":           "Canvas",
	"nodes":            "Nodes",
	"cron":             "Cron",
	"message":          "Message",
	"tts":              "Tts",
	"gateway":          "Gateway",
	"agents_list":      "AgentsList",
	"sessions_list":    "SessionsList",
	"sessions_history": "SessionsHistory",
	"sessions_send":    "SessionsSend",
	"sessions_yield":   "SessionsYield",
	"sessions_spawn":   "SessionsSpawn",
	"subagents":        "Subagents",
	"session_status":   "SessionStatus",
	"browser":          "Browser",
	"memory_search":    "MemorySearch",
	"memory_get":       "MemoryGet",
}

// droppedTools lists tool names that should be removed entirely.
// Add a tool name here to prevent it from being forwarded to upstream.
var droppedTools = map[string]bool{
	// Example: "canvas": true,
}

// buildReverseMaps constructs the forward (openclaw→upstream) and reverse
// (upstream→openclaw) mappings from the configuration above.
func buildReverseMaps() (forward map[string]string, reverse map[string]string) {
	// Start with explicit intersection mappings
	forward = make(map[string]string, len(intersectionForwardMap))
	reverse = make(map[string]string, len(intersectionForwardMap))
	for oc, cli := range intersectionForwardMap {
		forward[oc] = cli
		reverse[cli] = oc
	}
	return forward, reverse
}

// forwardMap and ToolNameReverseMap are computed once at init time from the
// configuration tables above.
var forwardMap map[string]string

// ToolNameReverseMap maps upstream tool names → OpenClaw tool names.
// Used in response path to convert tool names back before sending to client.
var ToolNameReverseMap map[string]string

func init() {
	forwardMap, ToolNameReverseMap = buildReverseMaps()
}

// mapToolName returns the upstream name for an OpenClaw tool.
// For intersection tools it uses the explicit mapping; for all others it
// capitalizes the first letter. Returns ("", false) if the tool is dropped.
func mapToolName(name string) (string, bool) {
	if droppedTools[name] {
		return "", false
	}
	if mapped, ok := forwardMap[name]; ok {
		return mapped, true
	}
	// Default: capitalize first letter
	capitalized := capitalizeFirst(name)
	return capitalized, true
}

// capitalizeFirst returns s with its first rune uppercased.
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return s
	}
	upper := unicode.ToUpper(r)
	if upper == r {
		return s
	}
	return string(upper) + s[size:]
}

// SanitizeClientFingerprints removes OpenClaw client fingerprints from a
// Claude API request JSON body before forwarding to upstream.
//
// It performs:
//  1. Strips "Sender (untrusted metadata)" blocks from user message text.
//  2. Replaces "OpenClaw"/"openclaw" with "the client" in system prompts.
//  3. Renames tools and tool_use references in messages.
func SanitizeClientFingerprints(jsonData []byte) ([]byte, error) {
	var data map[string]interface{}
	if err := common.Unmarshal(jsonData, &data); err != nil {
		common.SysError("SanitizeClientFingerprints Unmarshal error: " + err.Error())
		return jsonData, nil
	}

	sanitizeMessages(data)
	sanitizeSystem(data)
	sanitizeTools(data)

	result, err := common.Marshal(data)
	if err != nil {
		common.SysError("SanitizeClientFingerprints Marshal error: " + err.Error())
		return jsonData, nil
	}
	return result, nil
}

// sanitizeTools renames tool definitions, removes dropped tools, and renames
// tool_use blocks in assistant messages.
//
// For each tool:
//   - If in droppedTools → removed
//   - If in intersectionForwardMap → renamed to the explicit mapping
//   - Otherwise → first letter capitalized
//
// Also populates dynamicReverseEntries on the data map under a special key
// so the caller can retrieve the full reverse map for response processing.
func sanitizeTools(data map[string]interface{}) {
	// Collect the full reverse map (intersection + auto-capitalized)
	fullReverse := make(map[string]string)
	for k, v := range ToolNameReverseMap {
		fullReverse[k] = v
	}

	// Rename/filter tool definitions
	if toolsAny, ok := data["tools"]; ok {
		if tools, ok := toolsAny.([]interface{}); ok {
			filtered := make([]interface{}, 0, len(tools))
			for _, toolAny := range tools {
				tool, ok := toolAny.(map[string]interface{})
				if !ok {
					continue
				}
				name, _ := tool["name"].(string)
				newName, keep := mapToolName(name)
				if !keep {
					continue
				}
				if newName != name {
					tool["name"] = newName
					// Add to reverse map if not already present
					if _, exists := fullReverse[newName]; !exists {
						fullReverse[newName] = name
					}
				}
				filtered = append(filtered, tool)
			}
			data["tools"] = filtered
		}
	}

	// Store the full reverse map for the caller to retrieve
	data["__tool_reverse_map"] = fullReverse

	// Rename tool_choice if it references a specific tool name
	if tcAny, ok := data["tool_choice"]; ok {
		if tc, ok := tcAny.(map[string]interface{}); ok {
			if name, ok := tc["name"].(string); ok {
				if newName, keep := mapToolName(name); keep && newName != name {
					tc["name"] = newName
				}
			}
		}
	}

	// Rename tool_use blocks in assistant messages
	renameToolUseInMessages(data, fullReverse)
}

// ExtractAndCleanReverseMap extracts the dynamic reverse map from the
// sanitized data and removes the internal key before marshaling.
// Call this between SanitizeClientFingerprints marshal and sending.
func ExtractAndCleanReverseMap(jsonData []byte) (cleaned []byte, reverseMap map[string]string, err error) {
	var data map[string]interface{}
	if err := common.Unmarshal(jsonData, &data); err != nil {
		return jsonData, nil, err
	}
	if rm, ok := data["__tool_reverse_map"]; ok {
		if rmMap, ok := rm.(map[string]interface{}); ok {
			reverseMap = make(map[string]string, len(rmMap))
			for k, v := range rmMap {
				if vs, ok := v.(string); ok {
					reverseMap[k] = vs
				}
			}
		}
		delete(data, "__tool_reverse_map")
		cleaned, err = common.Marshal(data)
		if err != nil {
			return jsonData, reverseMap, err
		}
		return cleaned, reverseMap, nil
	}
	return jsonData, nil, nil
}

// renameToolUseInMessages renames tool names in assistant message content blocks
// of type "tool_use". Uses the full reverse map to find the forward mapping.
func renameToolUseInMessages(data map[string]interface{}, fullReverse map[string]string) {
	// Build a forward lookup from the reverse map
	fwd := make(map[string]string, len(fullReverse))
	for upstream, openclaw := range fullReverse {
		fwd[openclaw] = upstream
	}

	messagesAny, ok := data["messages"]
	if !ok {
		return
	}
	messages, ok := messagesAny.([]interface{})
	if !ok {
		return
	}
	for _, msgAny := range messages {
		msg, ok := msgAny.(map[string]interface{})
		if !ok {
			continue
		}
		role, _ := msg["role"].(string)
		if role != "assistant" {
			continue
		}
		contentAny, ok := msg["content"]
		if !ok {
			continue
		}
		contentArr, ok := contentAny.([]interface{})
		if !ok {
			continue
		}
		for _, blockAny := range contentArr {
			block, ok := blockAny.(map[string]interface{})
			if !ok {
				continue
			}
			typ, _ := block["type"].(string)
			if typ != "tool_use" {
				continue
			}
			if name, ok := block["name"].(string); ok {
				if newName, exists := fwd[name]; exists {
					block["name"] = newName
				}
			}
		}
	}
}

// sanitizeMessages strips Sender metadata blocks from user message text.
func sanitizeMessages(data map[string]interface{}) {
	messagesAny, ok := data["messages"]
	if !ok {
		return
	}
	messages, ok := messagesAny.([]interface{})
	if !ok {
		return
	}

	for _, msgAny := range messages {
		msg, ok := msgAny.(map[string]interface{})
		if !ok {
			continue
		}
		role, _ := msg["role"].(string)
		if role != "user" {
			continue
		}
		sanitizeMessageContent(msg)
	}
}

// sanitizeMessageContent handles both string content and array content.
func sanitizeMessageContent(msg map[string]interface{}) {
	content := msg["content"]
	switch c := content.(type) {
	case string:
		msg["content"] = stripSenderMetadata(c)
	case []interface{}:
		for _, itemAny := range c {
			item, ok := itemAny.(map[string]interface{})
			if !ok {
				continue
			}
			typ, _ := item["type"].(string)
			if typ != "text" {
				continue
			}
			if text, ok := item["text"].(string); ok {
				item["text"] = stripSenderMetadata(text)
			}
		}
	}
}

// stripSenderMetadata removes the Sender metadata prefix from a text string.
func stripSenderMetadata(text string) string {
	return senderMetadataRe.ReplaceAllString(text, "")
}

// sanitizeSystem replaces OpenClaw mentions in system prompts.
func sanitizeSystem(data map[string]interface{}) {
	systemAny, ok := data["system"]
	if !ok {
		return
	}

	switch s := systemAny.(type) {
	case string:
		data["system"] = replaceOpenClaw(s)
	case []interface{}:
		for _, itemAny := range s {
			item, ok := itemAny.(map[string]interface{})
			if !ok {
				continue
			}
			if text, ok := item["text"].(string); ok {
				item["text"] = replaceOpenClaw(text)
			}
		}
	}
}

// replaceOpenClaw replaces all case-insensitive occurrences of "openclaw"
// with "the client".
func replaceOpenClaw(text string) string {
	if !strings.Contains(strings.ToLower(text), "openclaw") {
		return text
	}
	return openClawRe.ReplaceAllString(text, "the client")
}
