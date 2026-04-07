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

// ContextKeyParamReverseMap re-exports the shared context key for convenience.
const ContextKeyParamReverseMap = relaycommon.ContextKeyParamReverseMap

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
	// ---- Intersection ----
	"read":       "Kx7",
	"write":      "Mv3",
	"edit":       "Rq9",
	"exec":       "Tn4",
	"web_search": "Jb6",
	"web_fetch":  "Wp2",

	// ---- Non-intersection ----
	"process":          "Uf8",
	"canvas":           "Zd1",
	"nodes":            "Hy5",
	"cron":             "Oa0",
	"message":          "Lc3",
	"tts":              "Bg7",
	"gateway":          "Xe9",
	"agents_list":      "Fs2",
	"sessions_list":    "Qi4",
	"sessions_history": "Nv6",
	"sessions_send":    "Pw1",
	"sessions_yield":   "Ek8",
	"sessions_spawn":   "Gm5",
	"subagents":        "Dr0",
	"session_status":   "Aj7",
	"browser":          "Vc3",
	"memory_search":    "Yl6",
	"memory_get":       "Sh9",
}

// descriptionOverrideMap maps OpenClaw tool names → rewritten descriptions.
// Descriptions are rewritten to remove OpenClaw-specific phrasing and avoid
// fingerprinting by upstream description matching.
var descriptionOverrideMap = map[string]string{
	"read": "Accepts a path argument and produces the corresponding blob. " +
		"Binary payloads such as raster graphics are attached as supplementary parts. " +
		"Long results may be windowed; callers should iterate with positional arguments until exhausted.",
	"edit": "Performs surgical mutations on a single resource. " +
		"Supply one or more pairs of before/after fragments; each pair must reference a distinct, " +
		"non-overlapping slice of the original payload. Adjacent mutations should be coalesced into one pair.",
	"write": "Persists the supplied payload to the designated location, replacing any prior version. " +
		"Intermediate path segments are provisioned on demand.",
	"exec": "Delegates an instruction string to the host interpreter and collects the resulting output. " +
		"May be deferred to a background context via timing or flag parameters. " +
		"An optional terminal-emulation flag is available for programs that require raw I/O.",
	"process": "Provides lifecycle control over previously deferred interpreter contexts: " +
		"enumerate, inspect logs, inject input, relay key sequences, or halt.",
	"canvas": "Governs an embedded rendering surface — toggling visibility, " +
		"evaluating scripts within it, and capturing rasterized snapshots of the current frame.",
	"message": "Routes a payload to one or more external notification endpoints. " +
		"Accepts single-target and fan-out modes.",
	"tts": "Converts a text argument into an audio waveform delivered as an inline attachment.",
	"agents_list": "Returns the set of identifiers eligible for delegation to isolated workers.",
	"sessions_list": "Enumerates open contexts, optionally narrowed by age, category, or preview depth.",
	"sessions_history": "Materializes the conversation transcript of a given context reference.",
	"sessions_send": "Injects a payload into a peer context specified by reference or alias.",
	"sessions_yield": "Signals that the current turn is complete and control should pass to pending peer results.",
	"sessions_spawn": "Provisions a new isolated context. Can operate in fire-and-forget " +
		"or persistent mode; inherits the caller's working root.",
	"subagents": "Administers child contexts: inventory, signal termination, or redirect.",
	"session_status": "Emits a diagnostic card for the active context — resource counters, wall-clock elapsed, " +
		"and optional cost projection.",
	"web_search": "Issues a keyword query against public indices and returns a ranked " +
		"collection of titles, locations, and preview fragments.",
	"web_fetch": "Retrieves a remote document by URI and reduces it to a structured text representation.",
	"browser": "Drives an automated viewport: load locations, " +
		"capture visual or structural representations, and replay interaction sequences.",
	"memory_search": "Executes a relevance-ranked lookup across persisted notes and " +
		"prior transcripts, surfacing the closest matching fragments with provenance metadata.",
	"memory_get": "Extracts a bounded slice from a specific persisted note by positional coordinates.",
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
	// Default: prefix with "Zx" + capitalize first letter for unmapped tools
	capitalized := "Zx" + capitalizeFirst(name)
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
				// Override description if we have a rewritten version
				if newDesc, ok := descriptionOverrideMap[name]; ok {
					tool["description"] = newDesc
				}
				// Rewrite parameter names and descriptions
				sanitizeToolParams(name, tool)
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

// renameToolUseInMessages renames tool names AND input parameter keys in
// assistant message content blocks of type "tool_use".
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
			originalName, _ := block["name"].(string)
			// Rename tool name
			if newName, exists := fwd[originalName]; exists {
				block["name"] = newName
			}
			// Rename input parameter keys
			if inputAny, ok := block["input"]; ok {
				if inputMap, ok := inputAny.(map[string]interface{}); ok {
					renameInputKeys(originalName, inputMap)
				}
			}
		}
	}
}

// renameInputKeys renames the keys in a tool_use input map according to
// toolParamOverrides.  Also handles nested objects (e.g. edits[].oldText).
func renameInputKeys(originalToolName string, input map[string]interface{}) {
	cfg, ok := toolParamOverrides[originalToolName]
	if !ok {
		return
	}

	// Collect renames to apply (can't mutate map while iterating)
	type rename struct {
		oldKey string
		newKey string
	}
	var renames []rename
	for oldKey := range input {
		if remap, found := cfg.Props[oldKey]; found && remap.Name != "" && remap.Name != oldKey {
			renames = append(renames, rename{oldKey, remap.Name})
		}
	}
	for _, r := range renames {
		input[r.newKey] = input[r.oldKey]
		delete(input, r.oldKey)
	}

	// Handle nested overrides (e.g. edits is an array of objects)
	for _, nested := range cfg.Nested {
		// Find the array under its NEW name (already renamed above)
		arrayKey := nested.ArrayPropName
		if remap, ok := cfg.Props[arrayKey]; ok && remap.Name != "" {
			arrayKey = remap.Name
		}
		arrAny, ok := input[arrayKey]
		if !ok {
			continue
		}
		arr, ok := arrAny.([]interface{})
		if !ok {
			continue
		}
		for _, itemAny := range arr {
			itemMap, ok := itemAny.(map[string]interface{})
			if !ok {
				continue
			}
			var nestedRenames []rename
			for oldKey := range itemMap {
				if remap, found := nested.Items[oldKey]; found && remap.Name != "" && remap.Name != oldKey {
					nestedRenames = append(nestedRenames, rename{oldKey, remap.Name})
				}
			}
			for _, r := range nestedRenames {
				itemMap[r.newKey] = itemMap[r.oldKey]
				delete(itemMap, r.oldKey)
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

// replaceOpenClaw replaces case-insensitive "openclaw" with "the client",
// but preserves path components like ".openclaw" or "/openclaw".
func replaceOpenClaw(text string) string {
	lower := strings.ToLower(text)
	if !strings.Contains(lower, "openclaw") {
		return text
	}
	// Protect path patterns before replacing
	text = strings.ReplaceAll(text, ".openclaw", "\x00PATHOC_DOT\x00")
	text = strings.ReplaceAll(text, ".OpenClaw", "\x00PATHOC_DOT\x00")
	text = strings.ReplaceAll(text, "/openclaw", "\x00PATHOC_SLASH\x00")
	text = strings.ReplaceAll(text, "/OpenClaw", "\x00PATHOC_SLASH\x00")
	// Do the replacement
	text = openClawRe.ReplaceAllString(text, "the client")
	// Restore protected paths
	text = strings.ReplaceAll(text, "\x00PATHOC_DOT\x00", ".openclaw")
	text = strings.ReplaceAll(text, "\x00PATHOC_SLASH\x00", "/openclaw")
	return text
}
