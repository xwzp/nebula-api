package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"

	"github.com/bytedance/gopkg/util/gopool"
)

const (
	claudeOAuthCredentialRefreshTickInterval = 10 * time.Minute
	claudeOAuthCredentialRefreshThreshold    = 24 * time.Hour
	claudeOAuthCredentialRefreshBatchSize    = 200
	claudeOAuthCredentialRefreshTimeout      = 15 * time.Second
)

var (
	claudeOAuthCredentialRefreshOnce    sync.Once
	claudeOAuthCredentialRefreshRunning atomic.Bool
)

func StartClaudeOAuthCredentialAutoRefreshTask() {
	claudeOAuthCredentialRefreshOnce.Do(func() {
		if !common.IsMasterNode {
			return
		}

		gopool.Go(func() {
			logger.LogInfo(context.Background(), fmt.Sprintf("claude oauth credential auto-refresh task started: tick=%s threshold=%s", claudeOAuthCredentialRefreshTickInterval, claudeOAuthCredentialRefreshThreshold))

			ticker := time.NewTicker(claudeOAuthCredentialRefreshTickInterval)
			defer ticker.Stop()

			runClaudeOAuthCredentialAutoRefreshOnce()
			for range ticker.C {
				runClaudeOAuthCredentialAutoRefreshOnce()
			}
		})
	})
}

func runClaudeOAuthCredentialAutoRefreshOnce() {
	if !claudeOAuthCredentialRefreshRunning.CompareAndSwap(false, true) {
		return
	}
	defer claudeOAuthCredentialRefreshRunning.Store(false)

	ctx := context.Background()
	now := time.Now()

	var refreshed int
	var scanned int

	offset := 0
	for {
		var channels []*model.Channel
		err := model.DB.
			Select("id", "name", "key", "status", "channel_info").
			Where("type = ? AND status = 1", constant.ChannelTypeClaudeOAuth).
			Order("id asc").
			Limit(claudeOAuthCredentialRefreshBatchSize).
			Offset(offset).
			Find(&channels).Error
		if err != nil {
			logger.LogError(ctx, fmt.Sprintf("claude oauth credential auto-refresh: query channels failed: %v", err))
			return
		}
		if len(channels) == 0 {
			break
		}
		offset += claudeOAuthCredentialRefreshBatchSize

		for _, ch := range channels {
			if ch == nil {
				continue
			}
			scanned++
			if ch.ChannelInfo.IsMultiKey {
				continue
			}

			rawKey := strings.TrimSpace(ch.Key)
			if rawKey == "" {
				continue
			}

			oauthKey, err := parseClaudeOAuthKey(rawKey)
			if err != nil {
				continue
			}

			refreshToken := strings.TrimSpace(oauthKey.RefreshToken)
			if refreshToken == "" {
				continue
			}

			expiresAtRaw := strings.TrimSpace(oauthKey.ExpiresAt)
			expiresAt, err := time.Parse(time.RFC3339, expiresAtRaw)
			if err == nil && !expiresAt.IsZero() && expiresAt.Sub(now) > claudeOAuthCredentialRefreshThreshold {
				continue
			}

			refreshCtx, cancel := context.WithTimeout(ctx, claudeOAuthCredentialRefreshTimeout)
			newKey, _, err := RefreshClaudeOAuthChannelCredential(refreshCtx, ch.Id, ClaudeOAuthCredentialRefreshOptions{ResetCaches: false})
			cancel()
			if err != nil {
				logger.LogWarn(ctx, fmt.Sprintf("claude oauth credential auto-refresh: channel_id=%d name=%s refresh failed: %v", ch.Id, ch.Name, err))
				continue
			}

			refreshed++
			logger.LogInfo(ctx, fmt.Sprintf("claude oauth credential auto-refresh: channel_id=%d name=%s refreshed, expires_at=%s", ch.Id, ch.Name, newKey.ExpiresAt))
		}
	}

	if refreshed > 0 {
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.LogWarn(ctx, fmt.Sprintf("claude oauth credential auto-refresh: InitChannelCache panic: %v", r))
				}
			}()
			model.InitChannelCache()
		}()
		ResetProxyClientCache()
	}

	if common.DebugEnabled {
		logger.LogDebug(ctx, "claude oauth credential auto-refresh: scanned=%d refreshed=%d", scanned, refreshed)
	}
}
