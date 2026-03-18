package claude_oauth

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/relay/channel"
	"github.com/QuantumNous/new-api/relay/channel/claude"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
)

type Adaptor struct {
	inner claude.Adaptor
}

func (a *Adaptor) Init(info *relaycommon.RelayInfo) {
	a.inner.Init(info)
}

func (a *Adaptor) GetRequestURL(info *relaycommon.RelayInfo) (string, error) {
	return a.inner.GetRequestURL(info)
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Header, info *relaycommon.RelayInfo) error {
	channel.SetupApiRequestHeader(info, c, req)

	key := strings.TrimSpace(info.ApiKey)
	if key == "" {
		return errors.New("claude oauth channel: empty key")
	}

	var accessToken string
	if strings.HasPrefix(key, "{") {
		// JSON format: {"access_token": "...", "refresh_token": "..."}
		oauthKey, err := ParseOAuthKey(key)
		if err != nil {
			return err
		}
		accessToken = strings.TrimSpace(oauthKey.AccessToken)
	} else {
		// Plain token from `claude setup-token`: sk-ant-oat01-...
		accessToken = key
	}

	if accessToken == "" {
		return errors.New("claude oauth channel: access_token is required")
	}

	anthropicVersion := c.Request.Header.Get("anthropic-version")
	if anthropicVersion == "" {
		anthropicVersion = "2023-06-01"
	}
	req.Set("anthropic-version", anthropicVersion)

	// Call CommonClaudeHeadersOperation first — it may set anthropic-beta from client headers.
	// OAuth headers are set AFTER so they won't be overwritten.
	claude.CommonClaudeHeadersOperation(c, req, info)

	// Setup-tokens (sk-ant-oat01-*) use Bearer auth + oauth/claude-code beta headers
	// Regular API keys (sk-ant-api03-*) use x-api-key header
	if strings.Contains(accessToken, "sk-ant-oat") {
		req.Set("Authorization", "Bearer "+accessToken)
		// OAuth betas set last to ensure they are never overwritten by client headers
		req.Set("anthropic-beta", "claude-code-20250219,oauth-2025-04-20,fine-grained-tool-streaming-2025-05-14,interleaved-thinking-2025-05-14")
		req.Set("anthropic-dangerous-direct-browser-access", "true")
		req.Set("User-Agent", "claude-cli/2.1.75")
		req.Set("x-app", "cli")
	} else {
		req.Set("x-api-key", accessToken)
	}

	return nil
}

func (a *Adaptor) ConvertOpenAIRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.GeneralOpenAIRequest) (any, error) {
	converted, err := a.inner.ConvertOpenAIRequest(c, info, request)
	if err != nil {
		return nil, err
	}
	if a.isOAuth(info) {
		if claudeReq, ok := converted.(*dto.ClaudeRequest); ok {
			injectClaudeCodeSystemPrompt(claudeReq)
		}
	}
	return converted, nil
}

func (a *Adaptor) ConvertClaudeRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.ClaudeRequest) (any, error) {
	converted, err := a.inner.ConvertClaudeRequest(c, info, request)
	if err != nil {
		return nil, err
	}
	if a.isOAuth(info) {
		if claudeReq, ok := converted.(*dto.ClaudeRequest); ok {
			injectClaudeCodeSystemPrompt(claudeReq)
		}
	}
	return converted, nil
}

func (a *Adaptor) ConvertGeminiRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.GeminiChatRequest) (any, error) {
	return a.inner.ConvertGeminiRequest(c, info, request)
}

func (a *Adaptor) ConvertRerankRequest(c *gin.Context, relayMode int, request dto.RerankRequest) (any, error) {
	return a.inner.ConvertRerankRequest(c, relayMode, request)
}

func (a *Adaptor) ConvertEmbeddingRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.EmbeddingRequest) (any, error) {
	return a.inner.ConvertEmbeddingRequest(c, info, request)
}

func (a *Adaptor) ConvertAudioRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.AudioRequest) (io.Reader, error) {
	return a.inner.ConvertAudioRequest(c, info, request)
}

func (a *Adaptor) ConvertImageRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.ImageRequest) (any, error) {
	return a.inner.ConvertImageRequest(c, info, request)
}

func (a *Adaptor) ConvertOpenAIResponsesRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.OpenAIResponsesRequest) (any, error) {
	return a.inner.ConvertOpenAIResponsesRequest(c, info, request)
}

func (a *Adaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (any, error) {
	return channel.DoApiRequest(a, c, info, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (usage any, err *types.NewAPIError) {
	return a.inner.DoResponse(c, resp, info)
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return ChannelName
}

const claudeCodeSystemPrompt = "You are Claude Code, Anthropic's official CLI for Claude."

// isOAuth checks if the current channel key is an OAuth setup-token
func (a *Adaptor) isOAuth(info *relaycommon.RelayInfo) bool {
	return strings.Contains(info.ApiKey, "sk-ant-oat")
}

// injectClaudeCodeSystemPrompt prepends the Claude Code identity system prompt.
// Anthropic requires this for OAuth tokens accessing advanced models.
func injectClaudeCodeSystemPrompt(req *dto.ClaudeRequest) {
	identity := map[string]any{
		"type": "text",
		"text": claudeCodeSystemPrompt,
	}

	switch existing := req.System.(type) {
	case nil:
		req.System = []any{identity}
	case string:
		if existing == "" {
			req.System = []any{identity}
		} else {
			req.System = []any{identity, map[string]any{"type": "text", "text": existing}}
		}
	case []any:
		req.System = append([]any{identity}, existing...)
	default:
		req.System = []any{identity}
	}
}
