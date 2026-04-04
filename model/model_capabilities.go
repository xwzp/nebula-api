package model

import (
	"strings"

	"github.com/QuantumNous/new-api/common"
)

// Reasoning field values (tri-state).
const (
	ReasoningNotSet       = 0
	ReasoningSupported    = 1
	ReasoningNotSupported = 2
)

// Capability source indicators (returned in API responses).
const (
	SourceOverride = "override"
	SourceFallback = "fallback"
	SourceDefault  = "default"
)

// Global defaults (tier 3) — used when neither DB override nor fallback table has a value.
const (
	DefaultContextWindow   = 128000
	DefaultMaxOutputTokens = 4096
	DefaultReasoning       = ReasoningNotSupported
)

var DefaultInputModalities = []string{"text"}

// ModelCapability holds fallback capability data for a known model or model family.
// Pattern supports exact match or prefix match (ending with "*").
type ModelCapability struct {
	Pattern         string
	ContextWindow   int      // 0 = not specified
	MaxOutputTokens int      // 0 = not specified
	Reasoning       int      // 0 = not specified, 1 = true, 2 = false
	InputModalities []string // nil = not specified
}

// fallbackCapabilities is the tier-2 known model table.
// Exact entries come first; prefix ("*"-suffix) entries follow.
// When multiple prefix entries match, the longest prefix wins.
var fallbackCapabilities = []ModelCapability{
	// ── Anthropic Claude 4.6 ──
	{Pattern: "claude-opus-4-6*", ContextWindow: 1000000, MaxOutputTokens: 128000, Reasoning: 1, InputModalities: []string{"text", "image"}},
	{Pattern: "claude-sonnet-4-6*", ContextWindow: 1000000, MaxOutputTokens: 128000, Reasoning: 1, InputModalities: []string{"text", "image"}},

	// ── Anthropic Claude 4.5 ──
	{Pattern: "claude-opus-4-5*", ContextWindow: 200000, MaxOutputTokens: 64000, Reasoning: 1, InputModalities: []string{"text", "image"}},
	{Pattern: "claude-sonnet-4-5*", ContextWindow: 200000, MaxOutputTokens: 64000, Reasoning: 1, InputModalities: []string{"text", "image"}},

	// ── Anthropic Claude 4 ──
	{Pattern: "claude-sonnet-4*", ContextWindow: 200000, MaxOutputTokens: 16000, Reasoning: 1, InputModalities: []string{"text", "image"}},

	// ── Anthropic Claude 3.5 ──
	{Pattern: "claude-3-5-sonnet*", ContextWindow: 200000, MaxOutputTokens: 8192, Reasoning: 2, InputModalities: []string{"text", "image"}},
	{Pattern: "claude-3-5-haiku*", ContextWindow: 200000, MaxOutputTokens: 8192, Reasoning: 2, InputModalities: []string{"text", "image"}},

	// ── Anthropic Claude 3 ──
	{Pattern: "claude-3-opus*", ContextWindow: 200000, MaxOutputTokens: 4096, Reasoning: 2, InputModalities: []string{"text", "image"}},
	{Pattern: "claude-3-haiku*", ContextWindow: 200000, MaxOutputTokens: 4096, Reasoning: 2, InputModalities: []string{"text", "image"}},

	// ── OpenAI GPT-4.1 ──
	{Pattern: "gpt-4.1", ContextWindow: 1047576, MaxOutputTokens: 32768, Reasoning: 2, InputModalities: []string{"text", "image"}},
	{Pattern: "gpt-4.1-mini", ContextWindow: 1047576, MaxOutputTokens: 32768, Reasoning: 2, InputModalities: []string{"text", "image"}},
	{Pattern: "gpt-4.1-nano", ContextWindow: 1047576, MaxOutputTokens: 32768, Reasoning: 2, InputModalities: []string{"text", "image"}},

	// ── OpenAI GPT-4o ──
	{Pattern: "gpt-4o", ContextWindow: 128000, MaxOutputTokens: 16384, Reasoning: 2, InputModalities: []string{"text", "image"}},
	{Pattern: "gpt-4o-mini*", ContextWindow: 128000, MaxOutputTokens: 16384, Reasoning: 2, InputModalities: []string{"text", "image"}},
	{Pattern: "gpt-4o-2*", ContextWindow: 128000, MaxOutputTokens: 16384, Reasoning: 2, InputModalities: []string{"text", "image"}},
	{Pattern: "chatgpt-4o*", ContextWindow: 128000, MaxOutputTokens: 16384, Reasoning: 2, InputModalities: []string{"text", "image"}},

	// ── OpenAI GPT-4 Turbo / GPT-4 ──
	{Pattern: "gpt-4-turbo*", ContextWindow: 128000, MaxOutputTokens: 4096, Reasoning: 2, InputModalities: []string{"text", "image"}},
	{Pattern: "gpt-4-1106*", ContextWindow: 128000, MaxOutputTokens: 4096, Reasoning: 2, InputModalities: []string{"text"}},
	{Pattern: "gpt-4", ContextWindow: 8192, MaxOutputTokens: 8192, Reasoning: 2, InputModalities: []string{"text"}},

	// ── OpenAI o-series (reasoning) ──
	{Pattern: "o1*", ContextWindow: 200000, MaxOutputTokens: 100000, Reasoning: 1, InputModalities: []string{"text", "image"}},
	{Pattern: "o3*", ContextWindow: 200000, MaxOutputTokens: 100000, Reasoning: 1, InputModalities: []string{"text", "image"}},
	{Pattern: "o4-mini*", ContextWindow: 200000, MaxOutputTokens: 100000, Reasoning: 1, InputModalities: []string{"text", "image"}},

	// ── OpenAI GPT-3.5 ──
	{Pattern: "gpt-3.5-turbo*", ContextWindow: 16385, MaxOutputTokens: 4096, Reasoning: 2, InputModalities: []string{"text"}},

	// ── Google Gemini 2.5 ──
	{Pattern: "gemini-2.5-pro*", ContextWindow: 1048576, MaxOutputTokens: 65536, Reasoning: 1, InputModalities: []string{"text", "image", "audio", "video"}},
	{Pattern: "gemini-2.5-flash*", ContextWindow: 1048576, MaxOutputTokens: 65536, Reasoning: 1, InputModalities: []string{"text", "image", "audio", "video"}},

	// ── Google Gemini 2.0 ──
	{Pattern: "gemini-2.0-flash*", ContextWindow: 1048576, MaxOutputTokens: 8192, Reasoning: 2, InputModalities: []string{"text", "image", "audio", "video"}},

	// ── Google Gemini 1.5 ──
	{Pattern: "gemini-1.5-pro*", ContextWindow: 2097152, MaxOutputTokens: 8192, Reasoning: 2, InputModalities: []string{"text", "image", "audio", "video"}},
	{Pattern: "gemini-1.5-flash*", ContextWindow: 1048576, MaxOutputTokens: 8192, Reasoning: 2, InputModalities: []string{"text", "image", "audio", "video"}},

	// ── DeepSeek ──
	{Pattern: "deepseek-r1*", ContextWindow: 65536, MaxOutputTokens: 8192, Reasoning: 1, InputModalities: []string{"text"}},
	{Pattern: "deepseek-reasoner*", ContextWindow: 65536, MaxOutputTokens: 8192, Reasoning: 1, InputModalities: []string{"text"}},
	{Pattern: "deepseek-chat*", ContextWindow: 65536, MaxOutputTokens: 8192, Reasoning: 2, InputModalities: []string{"text"}},
	{Pattern: "deepseek-v3*", ContextWindow: 65536, MaxOutputTokens: 8192, Reasoning: 2, InputModalities: []string{"text"}},

	// ── xAI Grok ──
	{Pattern: "grok-3*", ContextWindow: 131072, MaxOutputTokens: 16384, Reasoning: 2, InputModalities: []string{"text", "image"}},
	{Pattern: "grok-2*", ContextWindow: 131072, MaxOutputTokens: 8192, Reasoning: 2, InputModalities: []string{"text", "image"}},

	// ── Mistral ──
	{Pattern: "mistral-large*", ContextWindow: 131072, MaxOutputTokens: 8192, Reasoning: 2, InputModalities: []string{"text"}},
	{Pattern: "mistral-medium*", ContextWindow: 32768, MaxOutputTokens: 8192, Reasoning: 2, InputModalities: []string{"text"}},
	{Pattern: "mistral-small*", ContextWindow: 32768, MaxOutputTokens: 8192, Reasoning: 2, InputModalities: []string{"text"}},

	// ── Qwen ──
	{Pattern: "qwen-max*", ContextWindow: 131072, MaxOutputTokens: 8192, Reasoning: 2, InputModalities: []string{"text"}},
	{Pattern: "qwen-plus*", ContextWindow: 131072, MaxOutputTokens: 8192, Reasoning: 2, InputModalities: []string{"text"}},
	{Pattern: "qwen-turbo*", ContextWindow: 131072, MaxOutputTokens: 8192, Reasoning: 2, InputModalities: []string{"text"}},
	{Pattern: "qwq*", ContextWindow: 131072, MaxOutputTokens: 8192, Reasoning: 1, InputModalities: []string{"text"}},
}

// LookupFallbackCapability finds the best-matching fallback entry for a model name.
// Returns nil if no match is found.
func LookupFallbackCapability(modelName string) *ModelCapability {
	// 1. Exact match first
	for i := range fallbackCapabilities {
		cap := &fallbackCapabilities[i]
		if !strings.HasSuffix(cap.Pattern, "*") && cap.Pattern == modelName {
			return cap
		}
	}
	// 2. Prefix match — longest prefix wins
	var best *ModelCapability
	bestLen := 0
	for i := range fallbackCapabilities {
		cap := &fallbackCapabilities[i]
		if strings.HasSuffix(cap.Pattern, "*") {
			prefix := strings.TrimSuffix(cap.Pattern, "*")
			if strings.HasPrefix(modelName, prefix) && len(prefix) > bestLen {
				best = cap
				bestLen = len(prefix)
			}
		}
	}
	return best
}

// ResolveCapabilities populates the Effective* / *Source / Fallback* fields on a Model
// following the 3-tier priority: DB override > Fallback table > Global defaults.
func ResolveCapabilities(m *Model) {
	fallback := LookupFallbackCapability(m.ModelName)

	// Always populate fallback reference values
	if fallback != nil {
		m.FallbackContextWindow = fallback.ContextWindow
		m.FallbackMaxOutputTokens = fallback.MaxOutputTokens
		m.FallbackReasoning = fallback.Reasoning
		m.FallbackInputModalities = fallback.InputModalities
	}

	// ── Context Window ──
	if m.ContextWindow > 0 {
		m.EffectiveContextWindow = m.ContextWindow
		m.ContextWindowSource = SourceOverride
	} else if fallback != nil && fallback.ContextWindow > 0 {
		m.EffectiveContextWindow = fallback.ContextWindow
		m.ContextWindowSource = SourceFallback
	} else {
		m.EffectiveContextWindow = DefaultContextWindow
		m.ContextWindowSource = SourceDefault
	}

	// ── Max Output Tokens ──
	if m.MaxOutputTokens > 0 {
		m.EffectiveMaxOutputTokens = m.MaxOutputTokens
		m.MaxOutputTokensSource = SourceOverride
	} else if fallback != nil && fallback.MaxOutputTokens > 0 {
		m.EffectiveMaxOutputTokens = fallback.MaxOutputTokens
		m.MaxOutputTokensSource = SourceFallback
	} else {
		m.EffectiveMaxOutputTokens = DefaultMaxOutputTokens
		m.MaxOutputTokensSource = SourceDefault
	}

	// ── Reasoning ──
	if m.Reasoning > ReasoningNotSet {
		m.EffectiveReasoning = m.Reasoning
		m.ReasoningSource = SourceOverride
	} else if fallback != nil && fallback.Reasoning > ReasoningNotSet {
		m.EffectiveReasoning = fallback.Reasoning
		m.ReasoningSource = SourceFallback
	} else {
		m.EffectiveReasoning = DefaultReasoning
		m.ReasoningSource = SourceDefault
	}

	// ── Input Modalities ──
	if m.InputModalities != "" {
		var modalities []string
		if err := common.UnmarshalJsonStr(m.InputModalities, &modalities); err == nil && len(modalities) > 0 {
			m.EffectiveInputModalities = modalities
			m.InputModalitiesSource = SourceOverride
		}
	}
	if m.InputModalitiesSource == "" {
		if fallback != nil && len(fallback.InputModalities) > 0 {
			m.EffectiveInputModalities = fallback.InputModalities
			m.InputModalitiesSource = SourceFallback
		} else {
			m.EffectiveInputModalities = DefaultInputModalities
			m.InputModalitiesSource = SourceDefault
		}
	}
}

// OpenClawModelInfo is the per-model structure expected by OpenClaw's Custom Provider config.
type OpenClawModelInfo struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Reasoning     bool     `json:"reasoning"`
	Input         []string `json:"input"`
	ContextWindow int      `json:"contextWindow"`
	MaxTokens     int      `json:"maxTokens"`
	API           string   `json:"api"`
}

// ResolveOpenClawAPI returns the OpenClaw API format for a given model name.
func ResolveOpenClawAPI(modelName string) string {
	lower := strings.ToLower(modelName)
	switch {
	case strings.HasPrefix(lower, "claude-"):
		return "anthropic-messages"
	case strings.HasPrefix(lower, "gemini-"):
		return "google-generative-ai"
	case strings.HasPrefix(lower, "codex-"):
		return "openai-codex-responses"
	default:
		return "openai-completions"
	}
}

// openClawAllowedInputs is the set of input modalities that OpenClaw accepts.
var openClawAllowedInputs = map[string]bool{"text": true, "image": true}

// filterInputForOpenClaw keeps only modalities that OpenClaw supports.
func filterInputForOpenClaw(modalities []string) []string {
	filtered := make([]string, 0, len(modalities))
	for _, m := range modalities {
		if openClawAllowedInputs[m] {
			filtered = append(filtered, m)
		}
	}
	if len(filtered) == 0 {
		return DefaultInputModalities
	}
	return filtered
}

// GetModelsWithCapabilities resolves capability metadata for a list of model names
// and returns them in OpenClaw-compatible format.
func GetModelsWithCapabilities(modelNames []string) []OpenClawModelInfo {
	if len(modelNames) == 0 {
		return nil
	}

	// Batch-load Model records that exist in DB
	var dbModels []*Model
	if err := DB.Where("model_name IN ?", modelNames).Find(&dbModels).Error; err != nil {
		return nil
	}
	dbMap := make(map[string]*Model, len(dbModels))
	for _, m := range dbModels {
		dbMap[m.ModelName] = m
	}

	result := make([]OpenClawModelInfo, 0, len(modelNames))
	for _, name := range modelNames {
		m := dbMap[name]
		if m == nil {
			m = &Model{ModelName: name}
		}
		ResolveCapabilities(m)

		displayName := name
		if m.Description != "" {
			displayName = m.Description
		}

		result = append(result, OpenClawModelInfo{
			ID:            name,
			Name:          displayName,
			Reasoning:     m.EffectiveReasoning == ReasoningSupported,
			Input:         filterInputForOpenClaw(m.EffectiveInputModalities),
			ContextWindow: m.EffectiveContextWindow,
			MaxTokens:     m.EffectiveMaxOutputTokens,
			API:           ResolveOpenClawAPI(name),
		})
	}
	return result
}
