package controller

import (
	"fmt"
	"io"
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/relay"
	"github.com/QuantumNous/new-api/relay/channel/brave_search"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relay/helper"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
)

// firecrawlSearchRequest matches the Firecrawl POST /v2/search request body.
type firecrawlSearchRequest struct {
	Query      string   `json:"query"`
	Limit      *int     `json:"limit,omitempty"`
	Sources    []string `json:"sources,omitempty"`
	Categories []string `json:"categories,omitempty"`
}

// firecrawlSearchItem is a single result in the Firecrawl response.
type firecrawlSearchItem struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// RelayFirecrawlSearch handles POST /v2/search with Firecrawl-compatible request/response format.
// Internally it routes to the Brave Search channel, translating between formats.
func RelayFirecrawlSearch(c *gin.Context) {
	requestId := c.GetString(common.RequestIdKey)

	// 1. Parse Firecrawl request body.
	var fcReq firecrawlSearchRequest
	if err := common.UnmarshalBodyReusable(c, &fcReq); err != nil {
		respondFirecrawlError(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}
	if fcReq.Query == "" {
		respondFirecrawlError(c, http.StatusBadRequest, "query is required")
		return
	}

	// 2. Build RelayInfo (Distribute middleware already set channel to brave-search).
	relayInfo := relaycommon.GenRelayInfoSearch(c)
	relayInfo.InitChannelMeta(c)

	// 3. Get the Brave Search adaptor.
	adaptor := relay.GetAdaptor(relayInfo.ApiType)
	braveAdaptor, ok := adaptor.(*brave_search.Adaptor)
	if !ok {
		respondFirecrawlError(c, http.StatusBadRequest, "channel is not configured as Brave Search")
		return
	}
	braveAdaptor.Init(relayInfo)

	// 4. Convert Firecrawl request to Brave Search request directly.
	braveSearchReq := &brave_search.SearchRequest{
		Model: "brave-search",
		Query: fcReq.Query,
		Count: fcReq.Limit,
	}
	braveAdaptor.SetSearchRequest(braveSearchReq)

	// 5. Billing: estimate, price, pre-consume.
	relayInfo.SetEstimatePromptTokens(1)
	priceData, err := helper.ModelPriceHelper(c, relayInfo, 1, &types.TokenCountMeta{})
	if err != nil {
		respondFirecrawlError(c, http.StatusInternalServerError, err.Error())
		return
	}
	if priceData.FreeModel {
		logger.LogInfo(c, fmt.Sprintf("model %s is free, skipping pre-consume", relayInfo.OriginModelName))
	} else {
		if apiErr := service.PreConsumeBilling(c, priceData.QuotaToPreConsume, relayInfo); apiErr != nil {
			respondFirecrawlError(c, apiErr.StatusCode, common.MessageWithRequestId(apiErr.Error(), requestId))
			return
		}
	}

	// 6. Execute upstream search request.
	httpResp, doErr := braveAdaptor.DoSearchRequest(c, relayInfo)
	if doErr != nil {
		if relayInfo.Billing != nil {
			relayInfo.Billing.Refund(c)
		}
		respondFirecrawlError(c, http.StatusBadGateway, doErr.Error())
		return
	}
	defer httpResp.Body.Close()

	// 7. Read upstream response.
	if httpResp.StatusCode != http.StatusOK {
		if relayInfo.Billing != nil {
			relayInfo.Billing.Refund(c)
		}
		body, _ := io.ReadAll(httpResp.Body)
		respondFirecrawlError(c, httpResp.StatusCode, fmt.Sprintf("upstream error: %s", string(body)))
		return
	}

	body, readErr := io.ReadAll(httpResp.Body)
	if readErr != nil {
		if relayInfo.Billing != nil {
			relayInfo.Billing.Refund(c)
		}
		respondFirecrawlError(c, http.StatusInternalServerError, "failed to read upstream response")
		return
	}

	// 8. Parse Brave response and convert to Firecrawl format.
	var braveResp struct {
		Web struct {
			Results []struct {
				Title       string `json:"title"`
				URL         string `json:"url"`
				Description string `json:"description"`
			} `json:"results"`
		} `json:"web"`
	}
	if err := common.Unmarshal(body, &braveResp); err != nil {
		// If we can't parse, return raw response wrapped in Firecrawl format.
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []any{},
		})
		settleSearchBilling(c, relayInfo, &dto.Usage{PromptTokens: 1, TotalTokens: 1})
		return
	}

	items := make([]firecrawlSearchItem, 0, len(braveResp.Web.Results))
	for _, r := range braveResp.Web.Results {
		items = append(items, firecrawlSearchItem{
			Title:       r.Title,
			URL:         r.URL,
			Description: r.Description,
		})
	}

	// 9. Return Firecrawl-compatible response.
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    items,
	})

	// 10. Settle billing.
	settleSearchBilling(c, relayInfo, &dto.Usage{PromptTokens: 1, TotalTokens: 1})
}

// respondFirecrawlError returns an error in Firecrawl format.
func respondFirecrawlError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{
		"success": false,
		"error":   message,
	})
}
