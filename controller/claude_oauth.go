package controller

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	claude_oauth "github.com/QuantumNous/new-api/relay/channel/claude_oauth"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
)

type importClaudeOAuthKeyRequest struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func ImportClaudeOAuthKey(c *gin.Context) {
	channelID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, fmt.Errorf("invalid channel id: %w", err))
		return
	}

	ch, err := model.GetChannelById(channelID, false)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if ch == nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "channel not found"})
		return
	}
	if ch.Type != constant.ChannelTypeClaudeOAuth {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "channel type is not Claude OAuth"})
		return
	}

	req := importClaudeOAuthKeyRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}
	if req.AccessToken == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "access_token is required"})
		return
	}

	now := time.Now().Format(time.RFC3339)
	key := claude_oauth.OAuthKey{
		AccessToken:  req.AccessToken,
		RefreshToken: req.RefreshToken,
		LastRefresh:  now,
		Type:         "claude_oauth",
	}
	encoded, err := common.Marshal(key)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	if err := model.DB.Model(&model.Channel{}).Where("id = ?", channelID).Update("key", string(encoded)).Error; err != nil {
		common.ApiError(c, err)
		return
	}

	model.InitChannelCache()
	service.ResetProxyClientCache()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "saved",
		"data": gin.H{
			"channel_id":   channelID,
			"expires_at":   key.ExpiresAt,
			"last_refresh": key.LastRefresh,
		},
	})
}

func RefreshClaudeOAuthChannelCredential(c *gin.Context) {
	channelID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, fmt.Errorf("invalid channel id: %w", err))
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	oauthKey, _, err := service.RefreshClaudeOAuthChannelCredential(ctx, channelID, service.ClaudeOAuthCredentialRefreshOptions{ResetCaches: true})
	if err != nil {
		common.SysError("failed to refresh claude oauth channel credential: " + err.Error())
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "刷新凭证失败，请稍后重试"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "refreshed",
		"data": gin.H{
			"channel_id":   channelID,
			"expires_at":   oauthKey.ExpiresAt,
			"last_refresh": oauthKey.LastRefresh,
		},
	})
}
