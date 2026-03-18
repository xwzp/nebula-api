package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relay"
	"github.com/QuantumNous/new-api/relay/channel/brave_search"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relay/helper"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
)

// RelaySearch handles /v1/search requests by routing them to the Brave Search adaptor.
// It follows a simplified billing flow: pre-consume -> upstream request -> settle.
func RelaySearch(c *gin.Context) {
	requestId := c.GetString(common.RequestIdKey)

	// 1. Build RelayInfo from context (Distribute middleware has already set channel info).
	relayInfo := relaycommon.GenRelayInfoSearch(c)
	relayInfo.InitChannelMeta(c)

	// 2. Get the adaptor and assert it is Brave Search.
	adaptor := relay.GetAdaptor(relayInfo.ApiType)
	braveAdaptor, ok := adaptor.(*brave_search.Adaptor)
	if !ok {
		respondSearchError(c, http.StatusBadRequest, "search_adaptor_error",
			"channel is not configured as Brave Search", requestId)
		return
	}
	braveAdaptor.Init(relayInfo)

	// 3. Parse the search request body.
	if err := braveAdaptor.ParseSearchRequest(c); err != nil {
		respondSearchError(c, http.StatusBadRequest, "invalid_search_request",
			err.Error(), requestId)
		return
	}

	// 4. Estimate tokens (fixed 1 token for per-call search billing).
	relayInfo.SetEstimatePromptTokens(1)

	// 5. Get pricing.
	priceData, err := helper.ModelPriceHelper(c, relayInfo, 1, &types.TokenCountMeta{})
	if err != nil {
		respondSearchError(c, http.StatusInternalServerError, "model_price_error",
			err.Error(), requestId)
		return
	}

	// 6. Pre-consume billing (skip for free models).
	if priceData.FreeModel {
		logger.LogInfo(c, fmt.Sprintf("model %s is free, skipping pre-consume", relayInfo.OriginModelName))
	} else {
		if apiErr := service.PreConsumeBilling(c, priceData.QuotaToPreConsume, relayInfo); apiErr != nil {
			apiErr.SetMessage(common.MessageWithRequestId(apiErr.Error(), requestId))
			c.JSON(apiErr.StatusCode, gin.H{
				"error": apiErr.ToOpenAIError(),
			})
			return
		}
	}

	// 7. Execute the upstream search request.
	httpResp, doErr := braveAdaptor.DoSearchRequest(c, relayInfo)
	if doErr != nil {
		if relayInfo.Billing != nil {
			relayInfo.Billing.Refund(c)
		}
		respondSearchError(c, http.StatusBadGateway, "search_request_failed",
			doErr.Error(), requestId)
		return
	}

	// 8. Check upstream status code.
	if httpResp.StatusCode != http.StatusOK {
		if relayInfo.Billing != nil {
			relayInfo.Billing.Refund(c)
		}
		// Forward upstream error response as-is.
		apiErr := service.RelayErrorHandler(c.Request.Context(), httpResp, false)
		if apiErr != nil {
			apiErr.SetMessage(common.MessageWithRequestId(apiErr.Error(), requestId))
			c.JSON(apiErr.StatusCode, gin.H{
				"error": apiErr.ToOpenAIError(),
			})
		}
		return
	}

	// 9. Write upstream response to client and get usage.
	usageAny, apiErr := braveAdaptor.DoResponse(c, httpResp, relayInfo)
	if apiErr != nil {
		if relayInfo.Billing != nil {
			relayInfo.Billing.Refund(c)
		}
		apiErr.SetMessage(common.MessageWithRequestId(apiErr.Error(), requestId))
		c.JSON(apiErr.StatusCode, gin.H{
			"error": apiErr.ToOpenAIError(),
		})
		return
	}

	// 10. Settle billing and log consumption.
	usage, _ := usageAny.(*dto.Usage)
	settleSearchBilling(c, relayInfo, usage)
}

// settleSearchBilling settles the billing for a successful search request.
func settleSearchBilling(c *gin.Context, relayInfo *relaycommon.RelayInfo, usage *dto.Usage) {
	if usage == nil {
		usage = &dto.Usage{
			PromptTokens: 1,
			TotalTokens:  1,
		}
	}

	promptTokens := usage.PromptTokens
	if promptTokens < 1 {
		promptTokens = 1
	}

	// Calculate quota using the same pattern as the text handler.
	var quota int
	groupRatio := relayInfo.PriceData.GroupRatioInfo.GroupRatio
	if relayInfo.PriceData.UsePrice {
		quota = int(relayInfo.PriceData.ModelPrice * common.QuotaPerUnit * groupRatio)
	} else {
		ratio := relayInfo.PriceData.ModelRatio * groupRatio
		quota = int(float64(promptTokens) * ratio)
	}
	if quota == 0 && (relayInfo.PriceData.ModelRatio != 0 || relayInfo.PriceData.ModelPrice != 0) && groupRatio != 0 {
		quota = 1
	}

	// Update user/channel used quota.
	if quota > 0 {
		model.UpdateUserUsedQuotaAndRequestCount(relayInfo.UserId, quota)
		model.UpdateChannelUsedQuota(relayInfo.ChannelId, quota)
	}

	// Settle via BillingSession.
	if err := service.SettleBilling(c, relayInfo, quota); err != nil {
		logger.LogError(c, "error settling search billing: "+err.Error())
	}

	// Log consumption.
	tokenName := c.GetString("token_name")
	useTimeSeconds := int(time.Since(relayInfo.StartTime).Seconds())
	other := map[string]interface{}{
		"request_path": c.Request.URL.Path,
		"model_price":  relayInfo.PriceData.ModelPrice,
		"model_ratio":  relayInfo.PriceData.ModelRatio,
		"group_ratio":  groupRatio,
	}
	if relayInfo.IsModelMapped {
		other["is_model_mapped"] = true
		other["upstream_model_name"] = relayInfo.UpstreamModelName
	}
	model.RecordConsumeLog(c, relayInfo.UserId, model.RecordConsumeLogParams{
		ChannelId:      relayInfo.ChannelId,
		PromptTokens:   promptTokens,
		ModelName:      relayInfo.OriginModelName,
		TokenName:      tokenName,
		Quota:          quota,
		Content:        "Brave Search",
		TokenId:        relayInfo.TokenId,
		UseTimeSeconds: useTimeSeconds,
		Group:          relayInfo.UsingGroup,
		Other:          other,
	})
}

// respondSearchError returns a standard OpenAI-format error response.
func respondSearchError(c *gin.Context, statusCode int, code, message, requestId string) {
	c.JSON(statusCode, gin.H{
		"error": types.OpenAIError{
			Message: common.MessageWithRequestId(message, requestId),
			Type:    "search_error",
			Code:    code,
		},
	})
}
