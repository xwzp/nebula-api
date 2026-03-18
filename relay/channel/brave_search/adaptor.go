package brave_search

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/relay/channel"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
)

type Adaptor struct {
	searchReq *SearchRequest
}

func (a *Adaptor) Init(info *relaycommon.RelayInfo) {
}

func (a *Adaptor) ParseSearchRequest(c *gin.Context) error {
	var req SearchRequest
	if err := common.UnmarshalBodyReusable(c, &req); err != nil {
		return fmt.Errorf("brave search: failed to parse request: %w", err)
	}
	if strings.TrimSpace(req.Query) == "" {
		return errors.New("brave search: query (q) is required and must not be empty")
	}
	a.searchReq = &req
	return nil
}

func (a *Adaptor) GetRequestURL(info *relaycommon.RelayInfo) (string, error) {
	if a.searchReq == nil {
		return "", errors.New("brave search: request not parsed yet")
	}

	baseUrl := info.ChannelBaseUrl
	if baseUrl == "" {
		baseUrl = "https://api.search.brave.com"
	}
	baseUrl = strings.TrimRight(baseUrl, "/")

	params := url.Values{}
	params.Set("q", a.searchReq.Query)
	if a.searchReq.Count != nil {
		params.Set("count", strconv.Itoa(*a.searchReq.Count))
	}
	if a.searchReq.Offset != nil {
		params.Set("offset", strconv.Itoa(*a.searchReq.Offset))
	}
	if a.searchReq.Country != "" {
		params.Set("country", a.searchReq.Country)
	}
	if a.searchReq.SearchLang != "" {
		params.Set("search_lang", a.searchReq.SearchLang)
	}
	if a.searchReq.Freshness != "" {
		params.Set("freshness", a.searchReq.Freshness)
	}
	if a.searchReq.TextFormat != "" {
		params.Set("text_format", a.searchReq.TextFormat)
	}
	if a.searchReq.ResultFilter != "" {
		params.Set("result_filter", a.searchReq.ResultFilter)
	}

	return fmt.Sprintf("%s/res/v1/web/search?%s", baseUrl, params.Encode()), nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Header, info *relaycommon.RelayInfo) error {
	req.Set("Accept", "application/json")
	req.Set("X-Subscription-Token", info.ApiKey)
	return nil
}

func (a *Adaptor) DoSearchRequest(c *gin.Context, info *relaycommon.RelayInfo) (*http.Response, error) {
	fullRequestURL, err := a.GetRequestURL(info)
	if err != nil {
		return nil, fmt.Errorf("get request url failed: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, fullRequestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("new request failed: %w", err)
	}

	headers := req.Header
	if err := a.SetupRequestHeader(c, &headers, info); err != nil {
		return nil, fmt.Errorf("setup request header failed: %w", err)
	}

	headerOverride, err := channel.ResolveHeaderOverride(info, c)
	if err != nil {
		return nil, err
	}
	for key, value := range headerOverride {
		req.Header.Set(key, value)
		if strings.EqualFold(key, "Host") {
			req.Host = value
		}
	}

	var client *http.Client
	if info.ChannelSetting.Proxy != "" {
		client, err = service.NewProxyHttpClient(info.ChannelSetting.Proxy)
		if err != nil {
			return nil, fmt.Errorf("new proxy http client failed: %w", err)
		}
	} else {
		client = service.GetHttpClient()
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request failed: %w", err)
	}
	if resp == nil {
		return nil, errors.New("resp is nil")
	}
	return resp, nil
}

func (a *Adaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (any, error) {
	return a.DoSearchRequest(c, info)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (usage any, err *types.NewAPIError) {
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, types.NewError(readErr, types.ErrorCodeReadResponseBodyFailed)
	}

	// Write the upstream response as-is to the client.
	c.Data(resp.StatusCode, "application/json", body)

	promptTokens := info.GetEstimatePromptTokens()
	if promptTokens < 1 {
		promptTokens = 1
	}

	return &dto.Usage{
		PromptTokens: promptTokens,
		TotalTokens:  promptTokens,
	}, nil
}

// LLM conversion methods — Brave Search is not an LLM endpoint.

func (a *Adaptor) ConvertOpenAIRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.GeneralOpenAIRequest) (any, error) {
	return nil, errors.New("brave search: not an LLM endpoint")
}

func (a *Adaptor) ConvertClaudeRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.ClaudeRequest) (any, error) {
	return nil, errors.New("brave search: not an LLM endpoint")
}

func (a *Adaptor) ConvertGeminiRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.GeminiChatRequest) (any, error) {
	return nil, errors.New("brave search: not an LLM endpoint")
}

func (a *Adaptor) ConvertRerankRequest(c *gin.Context, relayMode int, request dto.RerankRequest) (any, error) {
	return nil, errors.New("brave search: not an LLM endpoint")
}

func (a *Adaptor) ConvertEmbeddingRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.EmbeddingRequest) (any, error) {
	return nil, errors.New("brave search: not an LLM endpoint")
}

func (a *Adaptor) ConvertAudioRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.AudioRequest) (io.Reader, error) {
	return nil, errors.New("brave search: not an LLM endpoint")
}

func (a *Adaptor) ConvertImageRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.ImageRequest) (any, error) {
	return nil, errors.New("brave search: not an LLM endpoint")
}

func (a *Adaptor) ConvertOpenAIResponsesRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.OpenAIResponsesRequest) (any, error) {
	return nil, errors.New("brave search: not an LLM endpoint")
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return ChannelName
}
