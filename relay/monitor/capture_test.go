package monitor

import (
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
)

func TestSanitizeLongStrings_NonJSON(t *testing.T) {
	// SSE streams should pass through unchanged
	input := []byte("data: {\"type\":\"content_block_delta\"}\n\n")
	result := SanitizeLongStrings(input, 512)
	if string(result) != string(input) {
		t.Errorf("non-JSON input should be returned unchanged, got: %s", result)
	}
}

func TestSanitizeLongStrings_EmptyInput(t *testing.T) {
	result := SanitizeLongStrings(nil, 512)
	if result != nil {
		t.Errorf("nil input should return nil")
	}
	result = SanitizeLongStrings([]byte{}, 512)
	if len(result) != 0 {
		t.Errorf("empty input should return empty")
	}
}

func TestSanitizeLongStrings_ShortStringsUnchanged(t *testing.T) {
	input := map[string]any{
		"model": "claude-opus-4-6",
		"messages": []any{
			map[string]any{
				"role":    "user",
				"content": "Hello, how are you?",
			},
		},
	}
	data, _ := common.Marshal(input)
	result := SanitizeLongStrings(data, 512)

	var parsed map[string]any
	if err := common.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("result should be valid JSON: %v", err)
	}
	msgs := parsed["messages"].([]any)
	msg := msgs[0].(map[string]any)
	if msg["content"] != "Hello, how are you?" {
		t.Errorf("short string should not be truncated, got: %v", msg["content"])
	}
}

func TestSanitizeLongStrings_PreservesCacheControl(t *testing.T) {
	longText := strings.Repeat("这是一段很长的系统提示词。", 200) // ~2400 CJK chars
	input := map[string]any{
		"model": "claude-opus-4-6",
		"system": []any{
			map[string]any{
				"type":          "text",
				"text":          longText,
				"cache_control": map[string]any{"type": "ephemeral"},
			},
		},
		"messages": []any{
			map[string]any{
				"role": "user",
				"content": []any{
					map[string]any{
						"type": "text",
						"text": "Hello",
					},
				},
			},
		},
	}
	data, _ := common.Marshal(input)
	result := SanitizeLongStrings(data, 512)

	var parsed map[string]any
	if err := common.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("result should be valid JSON: %v", err)
	}

	// Verify cache_control is preserved
	system := parsed["system"].([]any)
	sysBlock := system[0].(map[string]any)
	cc, ok := sysBlock["cache_control"]
	if !ok || cc == nil {
		t.Fatal("cache_control should be preserved")
	}
	ccMap := cc.(map[string]any)
	if ccMap["type"] != "ephemeral" {
		t.Errorf("cache_control type should be ephemeral, got: %v", ccMap["type"])
	}

	// Verify type field preserved
	if sysBlock["type"] != "text" {
		t.Errorf("type should be preserved, got: %v", sysBlock["type"])
	}

	// Verify text was truncated
	text := sysBlock["text"].(string)
	if !strings.Contains(text, "...[+") {
		t.Errorf("long text should be truncated with placeholder, got length: %d", len(text))
	}

	// Verify model preserved
	if parsed["model"] != "claude-opus-4-6" {
		t.Errorf("model should be preserved, got: %v", parsed["model"])
	}

	// Verify short user message not truncated
	msgs := parsed["messages"].([]any)
	msg := msgs[0].(map[string]any)
	content := msg["content"].([]any)
	textBlock := content[0].(map[string]any)
	if textBlock["text"] != "Hello" {
		t.Errorf("short text should not be truncated, got: %v", textBlock["text"])
	}
}

func TestSanitizeLongStrings_Unicode(t *testing.T) {
	// 600 CJK characters - each is one rune but 3 bytes
	longCJK := strings.Repeat("中", 600)
	input := map[string]any{"text": longCJK}
	data, _ := common.Marshal(input)
	result := SanitizeLongStrings(data, 512)

	var parsed map[string]any
	if err := common.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("result should be valid JSON: %v", err)
	}

	text := parsed["text"].(string)
	if !strings.Contains(text, "...[+400 chars]") {
		t.Errorf("should truncate with correct char count, got: %s", text[len(text)-30:])
	}
	// Prefix should be 200 runes of "中"
	prefix := strings.Repeat("中", 200)
	if !strings.HasPrefix(text, prefix) {
		t.Error("prefix should contain first 200 runes intact")
	}
}

func TestSanitizeLongStrings_ExactThreshold(t *testing.T) {
	exact := strings.Repeat("a", 512)
	input := map[string]any{"text": exact}
	data, _ := common.Marshal(input)
	result := SanitizeLongStrings(data, 512)

	var parsed map[string]any
	common.Unmarshal(result, &parsed)
	if parsed["text"] != exact {
		t.Error("string at exact threshold should not be truncated")
	}

	// One over
	over := strings.Repeat("a", 513)
	input["text"] = over
	data, _ = common.Marshal(input)
	result = SanitizeLongStrings(data, 512)
	common.Unmarshal(result, &parsed)
	text := parsed["text"].(string)
	if !strings.Contains(text, "...[+") {
		t.Error("string one rune over threshold should be truncated")
	}
}

func TestSanitizeLongStrings_NestedObjects(t *testing.T) {
	longStr := strings.Repeat("x", 1000)
	input := map[string]any{
		"level1": map[string]any{
			"level2": map[string]any{
				"deep_text": longStr,
				"short":     "ok",
			},
		},
	}
	data, _ := common.Marshal(input)
	result := SanitizeLongStrings(data, 512)

	var parsed map[string]any
	common.Unmarshal(result, &parsed)
	l1 := parsed["level1"].(map[string]any)
	l2 := l1["level2"].(map[string]any)
	if l2["short"] != "ok" {
		t.Error("short nested string should be preserved")
	}
	deep := l2["deep_text"].(string)
	if !strings.Contains(deep, "...[+") {
		t.Error("long nested string should be truncated")
	}
}

func TestSanitizeLongStrings_ArrayStrings(t *testing.T) {
	longStr := strings.Repeat("y", 1000)
	input := []any{"short", longStr, "also short"}
	data, _ := common.Marshal(input)
	result := SanitizeLongStrings(data, 512)

	var parsed []any
	common.Unmarshal(result, &parsed)
	if parsed[0] != "short" {
		t.Error("short array string should be preserved")
	}
	if !strings.Contains(parsed[1].(string), "...[+") {
		t.Error("long array string should be truncated")
	}
	if parsed[2] != "also short" {
		t.Error("short array string should be preserved")
	}
}

func TestCaptureBodyFromBytes_FullPipeline(t *testing.T) {
	longSystem := strings.Repeat("System instructions here. ", 500) // ~13KB
	input := map[string]any{
		"model": "claude-opus-4-6",
		"system": []any{
			map[string]any{
				"type":          "text",
				"text":          longSystem,
				"cache_control": map[string]any{"type": "ephemeral"},
			},
		},
		"messages": []any{
			map[string]any{
				"role":    "user",
				"content": "Hello",
			},
		},
	}
	data, _ := common.Marshal(input)

	captured := CaptureBodyFromBytes(data, DefaultMaxBodyBytes)

	// Should not be truncated (after sanitization, body is well under 64KB)
	if captured.Truncated {
		t.Errorf("sanitized body should fit within 64KB, body len: %d", len(captured.Body))
	}

	// BodyLen should reflect original size
	if captured.BodyLen != len(data) {
		t.Errorf("BodyLen should be original data length, got %d, want %d", captured.BodyLen, len(data))
	}

	// Body should be valid JSON with cache_control visible
	if !strings.Contains(captured.Body, "cache_control") {
		t.Error("cache_control should be visible in captured body")
	}
	if !strings.Contains(captured.Body, "ephemeral") {
		t.Error("cache_control value should be visible")
	}
	if !strings.Contains(captured.Body, "claude-opus-4-6") {
		t.Error("model should be visible")
	}
}
