package vertex

import (
	"encoding/json"

	"github.com/QuantumNous/new-api/dto"
)

type VertexAIClaudeRequest struct {
	AnthropicVersion string              `json:"anthropic_version"`
	Messages         []dto.ClaudeMessage `json:"messages"`
	System           any                 `json:"system,omitempty"`
	MaxTokens        *uint               `json:"max_tokens,omitempty"`
	StopSequences    []string            `json:"stop_sequences,omitempty"`
	Stream           *bool               `json:"stream,omitempty"`
	Temperature      *float64            `json:"temperature,omitempty"`
	TopP             *float64            `json:"top_p,omitempty"`
	TopK             *int                `json:"top_k,omitempty"`
	Tools            any                 `json:"tools,omitempty"`
	ToolChoice       any                 `json:"tool_choice,omitempty"`
	Thinking         *dto.Thinking       `json:"thinking,omitempty"`
	OutputConfig     json.RawMessage     `json:"output_config,omitempty"`
	//Metadata         json.RawMessage     `json:"metadata,omitempty"`
}

// Vertex AI Embedding types — used with the :predict endpoint
// https://docs.cloud.google.com/vertex-ai/generative-ai/docs/model-reference/text-embeddings-api

type VertexEmbeddingInstance struct {
	Content  string `json:"content"`
	TaskType string `json:"task_type,omitempty"`
}

type VertexEmbeddingParameters struct {
	AutoTruncate         *bool `json:"autoTruncate,omitempty"`
	OutputDimensionality *int  `json:"outputDimensionality,omitempty"`
}

type VertexEmbeddingRequest struct {
	Instances  []VertexEmbeddingInstance  `json:"instances"`
	Parameters *VertexEmbeddingParameters `json:"parameters,omitempty"`
}

type VertexEmbeddingStatistics struct {
	Truncated  bool `json:"truncated"`
	TokenCount int  `json:"token_count"`
}

type VertexEmbeddingValues struct {
	Values     []float64                 `json:"values"`
	Statistics VertexEmbeddingStatistics `json:"statistics"`
}

type VertexEmbeddingPrediction struct {
	Embeddings VertexEmbeddingValues `json:"embeddings"`
}

type VertexEmbeddingResponse struct {
	Predictions []VertexEmbeddingPrediction `json:"predictions"`
}

func copyRequest(req *dto.ClaudeRequest, version string) *VertexAIClaudeRequest {
	return &VertexAIClaudeRequest{
		AnthropicVersion: version,
		System:           req.System,
		Messages:         req.Messages,
		MaxTokens:        req.MaxTokens,
		Stream:           req.Stream,
		Temperature:      req.Temperature,
		TopP:             req.TopP,
		TopK:             req.TopK,
		StopSequences:    req.StopSequences,
		Tools:            req.Tools,
		ToolChoice:       req.ToolChoice,
		Thinking:         req.Thinking,
		OutputConfig:     req.OutputConfig,
	}
}
