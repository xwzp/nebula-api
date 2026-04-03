package controller

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/gin-gonic/gin"
	"github.com/thanhpk/randstr"
	"gorm.io/gorm"
)

// ---- Shared types ----

type SubscriptionPlanDTO struct {
	Plan          model.SubscriptionPlan `json:"plan"`
	GroupTitle    string                 `json:"group_title,omitempty"`
	GroupSubtitle string                 `json:"group_subtitle,omitempty"`
	GroupTag      string                 `json:"group_tag,omitempty"`
}

type BillingPreferenceRequest struct {
	BillingPreference string `json:"billing_preference"`
}

type SubscriptionPayRequest struct {
	PlanId int `json:"plan_id"`
}

// validateAndCreateSubscriptionOrder validates the plan, checks purchase limits,
// and creates a pending SubscriptionOrder. Returns the plan, order, and true on success.
// On failure it writes the error response and returns false.
func validateAndCreateSubscriptionOrder(c *gin.Context, tradePrefix string, paymentMethod string) (*model.SubscriptionPlan, *model.SubscriptionOrder, bool) {
	var req SubscriptionPayRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.PlanId <= 0 {
		common.ApiErrorMsg(c, "参数错误")
		return nil, nil, false
	}

	plan, err := model.GetSubscriptionPlanById(req.PlanId)
	if err != nil {
		common.ApiError(c, err)
		return nil, nil, false
	}
	if !plan.Enabled {
		common.ApiErrorMsg(c, "套餐未启用")
		return nil, nil, false
	}
	if plan.PriceAmount < 0.01 {
		common.ApiErrorMsg(c, "套餐金额过低")
		return nil, nil, false
	}

	userId := c.GetInt("id")
	if plan.MaxPurchasePerUser > 0 {
		count, err := model.CountUserSubscriptionsByPlan(userId, plan.Id)
		if err != nil {
			common.ApiError(c, err)
			return nil, nil, false
		}
		if count >= int64(plan.MaxPurchasePerUser) {
			common.ApiErrorMsg(c, "已达到该套餐购买上限")
			return nil, nil, false
		}
	}

	tradeNo := fmt.Sprintf("%s-%d-%d-%s", tradePrefix, userId, time.Now().UnixMilli(), randstr.String(6))

	order := &model.SubscriptionOrder{
		UserId:        userId,
		PlanId:        plan.Id,
		Money:         plan.PriceAmount,
		TradeNo:       tradeNo,
		PaymentMethod: paymentMethod,
		CreateTime:    time.Now().Unix(),
		Status:        common.TopUpStatusPending,
	}
	if err := order.Insert(); err != nil {
		common.ApiErrorMsg(c, "创建订单失败")
		return nil, nil, false
	}

	return plan, order, true
}

// ---- Public APIs ----

type PublicPlanVariantDTO struct {
	Id                      int     `json:"id"`
	PriceAmount             float64 `json:"price_amount"`
	Currency                string  `json:"currency"`
	DurationUnit            string  `json:"duration_unit"`
	DurationValue           int     `json:"duration_value"`
	CustomSeconds           int64   `json:"custom_seconds"`
	SortOrder               int     `json:"sort_order"`
	TotalAmount             int64   `json:"total_amount"`
	UpgradeGroup            string  `json:"upgrade_group"`
	QuotaResetPeriod        string  `json:"quota_reset_period"`
	QuotaResetCustomSeconds int64   `json:"quota_reset_custom_seconds"`
	MaxPurchasePerUser      int     `json:"max_purchase_per_user"`
}

type PublicPlanGroupDTO struct {
	Id       int                    `json:"id"`
	Title    string                 `json:"title"`
	Subtitle string                 `json:"subtitle"`
	Tag      string                 `json:"tag"`
	Features string                 `json:"features"`
	Plans    []PublicPlanVariantDTO  `json:"plans"`
}

func GetPublicSubscriptionPlans(c *gin.Context) {
	groups, err := model.GetEnabledSubscriptionPlanGroups()
	if err != nil {
		common.ApiError(c, err)
		return
	}

	result := make([]PublicPlanGroupDTO, 0, len(groups))
	for _, g := range groups {
		plans, err := model.GetEnabledSubscriptionPlansByGroupID(g.Id)
		if err != nil {
			common.ApiError(c, err)
			return
		}
		if len(plans) == 0 {
			continue
		}
		variants := make([]PublicPlanVariantDTO, 0, len(plans))
		for _, p := range plans {
			variants = append(variants, PublicPlanVariantDTO{
				Id:                      p.Id,
				PriceAmount:             p.PriceAmount,
				Currency:                p.Currency,
				DurationUnit:            p.DurationUnit,
				DurationValue:           p.DurationValue,
				CustomSeconds:           p.CustomSeconds,
				SortOrder:               p.SortOrder,
				TotalAmount:             p.TotalAmount,
				UpgradeGroup:            p.UpgradeGroup,
				QuotaResetPeriod:        p.QuotaResetPeriod,
				QuotaResetCustomSeconds: p.QuotaResetCustomSeconds,
				MaxPurchasePerUser:      p.MaxPurchasePerUser,
			})
		}
		result = append(result, PublicPlanGroupDTO{
			Id:       g.Id,
			Title:    g.Title,
			Subtitle: g.Subtitle,
			Tag:      g.Tag,
			Features: g.Features,
			Plans:    variants,
		})
	}

	c.Header("Cache-Control", "public, max-age=60")
	common.ApiSuccess(c, result)
}

// ---- User APIs ----

func GetSubscriptionPlans(c *gin.Context) {
	var plans []model.SubscriptionPlan
	if err := model.DB.Where("enabled = ?", true).Order("sort_order desc, id desc").Find(&plans).Error; err != nil {
		common.ApiError(c, err)
		return
	}
	// Pre-load group titles for enrichment
	groupCache := make(map[int]*model.SubscriptionPlanGroup)
	result := make([]SubscriptionPlanDTO, 0, len(plans))
	for _, p := range plans {
		dto := SubscriptionPlanDTO{Plan: p}
		if p.GroupID > 0 {
			g, ok := groupCache[p.GroupID]
			if !ok {
				g, _ = model.GetSubscriptionPlanGroupById(p.GroupID)
				groupCache[p.GroupID] = g // may be nil
			}
			if g != nil {
				dto.GroupTitle = g.Title
				dto.GroupSubtitle = g.Subtitle
				dto.GroupTag = g.Tag
			}
		}
		result = append(result, dto)
	}
	common.ApiSuccess(c, result)
}

func GetSubscriptionSelf(c *gin.Context) {
	userId := c.GetInt("id")
	settingMap, _ := model.GetUserSetting(userId, false)
	pref := common.NormalizeBillingPreference(settingMap.BillingPreference)

	// Get all subscriptions (including expired)
	allSubscriptions, err := model.GetAllUserSubscriptions(userId)
	if err != nil {
		allSubscriptions = []model.SubscriptionSummary{}
	}

	// Get active subscriptions for backward compatibility
	activeSubscriptions, err := model.GetAllActiveUserSubscriptions(userId)
	if err != nil {
		activeSubscriptions = []model.SubscriptionSummary{}
	}

	common.ApiSuccess(c, gin.H{
		"billing_preference": pref,
		"subscriptions":      activeSubscriptions, // all active subscriptions
		"all_subscriptions":  allSubscriptions,    // all subscriptions including expired
	})
}

func UpdateSubscriptionPreference(c *gin.Context) {
	userId := c.GetInt("id")
	var req BillingPreferenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	pref := common.NormalizeBillingPreference(req.BillingPreference)

	user, err := model.GetUserById(userId, true)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	current := user.GetSetting()
	current.BillingPreference = pref
	user.SetSetting(current)
	if err := user.Update(false); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"billing_preference": pref})
}

// ---- Admin APIs ----

func AdminListSubscriptionPlans(c *gin.Context) {
	var plans []model.SubscriptionPlan
	if err := model.DB.Order("sort_order desc, id desc").Find(&plans).Error; err != nil {
		common.ApiError(c, err)
		return
	}
	result := make([]SubscriptionPlanDTO, 0, len(plans))
	for _, p := range plans {
		result = append(result, SubscriptionPlanDTO{
			Plan: p,
		})
	}
	common.ApiSuccess(c, result)
}

type AdminUpsertSubscriptionPlanRequest struct {
	Plan model.SubscriptionPlan `json:"plan"`
}

func AdminCreateSubscriptionPlan(c *gin.Context) {
	var req AdminUpsertSubscriptionPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	req.Plan.Id = 0
	// Accept group_id from URL path parameter or request body
	if groupIdFromPath, err := strconv.Atoi(c.Param("id")); err == nil && groupIdFromPath > 0 {
		req.Plan.GroupID = groupIdFromPath
	}
	if req.Plan.GroupID <= 0 {
		common.ApiErrorMsg(c, "套餐组ID不能为空")
		return
	}
	// Verify group exists
	if _, err := model.GetSubscriptionPlanGroupById(req.Plan.GroupID); err != nil {
		common.ApiErrorMsg(c, "套餐组不存在")
		return
	}
	if req.Plan.PriceAmount < 0 {
		common.ApiErrorMsg(c, "价格不能为负数")
		return
	}
	if req.Plan.PriceAmount > 9999 {
		common.ApiErrorMsg(c, "价格不能超过9999")
		return
	}
	if req.Plan.Currency == "" {
		req.Plan.Currency = "USD"
	}
	req.Plan.Currency = "USD"
	if req.Plan.DurationUnit == "" {
		req.Plan.DurationUnit = model.SubscriptionDurationMonth
	}
	if req.Plan.DurationValue <= 0 && req.Plan.DurationUnit != model.SubscriptionDurationCustom {
		req.Plan.DurationValue = 1
	}
	if req.Plan.MaxPurchasePerUser < 0 {
		common.ApiErrorMsg(c, "购买上限不能为负数")
		return
	}
	if req.Plan.TotalAmount < 0 {
		common.ApiErrorMsg(c, "总额度不能为负数")
		return
	}
	req.Plan.UpgradeGroup = strings.TrimSpace(req.Plan.UpgradeGroup)
	if req.Plan.UpgradeGroup != "" {
		if _, ok := ratio_setting.GetGroupRatioCopy()[req.Plan.UpgradeGroup]; !ok {
			common.ApiErrorMsg(c, "升级分组不存在")
			return
		}
	}
	req.Plan.QuotaResetPeriod = model.NormalizeResetPeriod(req.Plan.QuotaResetPeriod)
	if req.Plan.QuotaResetPeriod == model.SubscriptionResetCustom && req.Plan.QuotaResetCustomSeconds <= 0 {
		common.ApiErrorMsg(c, "自定义重置周期需大于0秒")
		return
	}
	err := model.DB.Create(&req.Plan).Error
	if err != nil {
		common.ApiError(c, err)
		return
	}
	model.InvalidateSubscriptionPlanCache(req.Plan.Id)
	common.ApiSuccess(c, req.Plan)
}

func AdminUpdateSubscriptionPlan(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		common.ApiErrorMsg(c, "无效的ID")
		return
	}
	var req AdminUpsertSubscriptionPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	if req.Plan.PriceAmount < 0 {
		common.ApiErrorMsg(c, "价格不能为负数")
		return
	}
	if req.Plan.PriceAmount > 9999 {
		common.ApiErrorMsg(c, "价格不能超过9999")
		return
	}
	req.Plan.Id = id
	if req.Plan.Currency == "" {
		req.Plan.Currency = "USD"
	}
	req.Plan.Currency = "USD"
	if req.Plan.DurationUnit == "" {
		req.Plan.DurationUnit = model.SubscriptionDurationMonth
	}
	if req.Plan.DurationValue <= 0 && req.Plan.DurationUnit != model.SubscriptionDurationCustom {
		req.Plan.DurationValue = 1
	}
	if req.Plan.MaxPurchasePerUser < 0 {
		common.ApiErrorMsg(c, "购买上限不能为负数")
		return
	}
	if req.Plan.TotalAmount < 0 {
		common.ApiErrorMsg(c, "总额度不能为负数")
		return
	}
	req.Plan.UpgradeGroup = strings.TrimSpace(req.Plan.UpgradeGroup)
	if req.Plan.UpgradeGroup != "" {
		if _, ok := ratio_setting.GetGroupRatioCopy()[req.Plan.UpgradeGroup]; !ok {
			common.ApiErrorMsg(c, "升级分组不存在")
			return
		}
	}
	req.Plan.QuotaResetPeriod = model.NormalizeResetPeriod(req.Plan.QuotaResetPeriod)
	if req.Plan.QuotaResetPeriod == model.SubscriptionResetCustom && req.Plan.QuotaResetCustomSeconds <= 0 {
		common.ApiErrorMsg(c, "自定义重置周期需大于0秒")
		return
	}

	err := model.DB.Transaction(func(tx *gorm.DB) error {
		// update plan (allow zero values updates with map)
		updateMap := map[string]interface{}{
			"group_id":                   req.Plan.GroupID,
			"price_amount":               req.Plan.PriceAmount,
			"currency":                   req.Plan.Currency,
			"duration_unit":              req.Plan.DurationUnit,
			"duration_value":             req.Plan.DurationValue,
			"custom_seconds":             req.Plan.CustomSeconds,
			"enabled":                    req.Plan.Enabled,
			"sort_order":                 req.Plan.SortOrder,
			"stripe_price_id":            req.Plan.StripePriceId,
			"creem_product_id":           req.Plan.CreemProductId,
			"max_purchase_per_user":      req.Plan.MaxPurchasePerUser,
			"total_amount":               req.Plan.TotalAmount,
			"upgrade_group":              req.Plan.UpgradeGroup,
			"quota_reset_period":         req.Plan.QuotaResetPeriod,
			"quota_reset_custom_seconds": req.Plan.QuotaResetCustomSeconds,
			"updated_at":                 common.GetTimestamp(),
		}
		if err := tx.Model(&model.SubscriptionPlan{}).Where("id = ?", id).Updates(updateMap).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		common.ApiError(c, err)
		return
	}
	model.InvalidateSubscriptionPlanCache(id)
	common.ApiSuccess(c, nil)
}

type AdminUpdateSubscriptionPlanStatusRequest struct {
	Enabled *bool `json:"enabled"`
}

func AdminUpdateSubscriptionPlanStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		common.ApiErrorMsg(c, "无效的ID")
		return
	}
	var req AdminUpdateSubscriptionPlanStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Enabled == nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	if err := model.DB.Model(&model.SubscriptionPlan{}).Where("id = ?", id).Update("enabled", *req.Enabled).Error; err != nil {
		common.ApiError(c, err)
		return
	}
	model.InvalidateSubscriptionPlanCache(id)
	common.ApiSuccess(c, nil)
}

type AdminBindSubscriptionRequest struct {
	UserId int `json:"user_id"`
	PlanId int `json:"plan_id"`
}

func AdminBindSubscription(c *gin.Context) {
	var req AdminBindSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.UserId <= 0 || req.PlanId <= 0 {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	msg, err := model.AdminBindSubscription(req.UserId, req.PlanId, "")
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if msg != "" {
		common.ApiSuccess(c, gin.H{"message": msg})
		return
	}
	common.ApiSuccess(c, nil)
}

// ---- Admin: user subscription management ----

func AdminListUserSubscriptions(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Param("id"))
	if userId <= 0 {
		common.ApiErrorMsg(c, "无效的用户ID")
		return
	}
	subs, err := model.GetAllUserSubscriptions(userId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, subs)
}

type AdminCreateUserSubscriptionRequest struct {
	PlanId int `json:"plan_id"`
}

// AdminCreateUserSubscription creates a new user subscription from a plan (no payment).
func AdminCreateUserSubscription(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Param("id"))
	if userId <= 0 {
		common.ApiErrorMsg(c, "无效的用户ID")
		return
	}
	var req AdminCreateUserSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.PlanId <= 0 {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	msg, err := model.AdminBindSubscription(userId, req.PlanId, "")
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if msg != "" {
		common.ApiSuccess(c, gin.H{"message": msg})
		return
	}
	common.ApiSuccess(c, nil)
}

// AdminInvalidateUserSubscription cancels a user subscription immediately.
func AdminInvalidateUserSubscription(c *gin.Context) {
	subId, _ := strconv.Atoi(c.Param("id"))
	if subId <= 0 {
		common.ApiErrorMsg(c, "无效的订阅ID")
		return
	}
	msg, err := model.AdminInvalidateUserSubscription(subId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if msg != "" {
		common.ApiSuccess(c, gin.H{"message": msg})
		return
	}
	common.ApiSuccess(c, nil)
}

// AdminDeleteUserSubscription hard-deletes a user subscription.
func AdminDeleteUserSubscription(c *gin.Context) {
	subId, _ := strconv.Atoi(c.Param("id"))
	if subId <= 0 {
		common.ApiErrorMsg(c, "无效的订阅ID")
		return
	}
	msg, err := model.AdminDeleteUserSubscription(subId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if msg != "" {
		common.ApiSuccess(c, gin.H{"message": msg})
		return
	}
	common.ApiSuccess(c, nil)
}

// ---- Plan Group Admin APIs ----

type GroupWithPlans struct {
	model.SubscriptionPlanGroup
	Plans []model.SubscriptionPlan `json:"plans"`
}

func AdminListSubscriptionPlanGroups(c *gin.Context) {
	groups, err := model.GetAllSubscriptionPlanGroups()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	result := make([]GroupWithPlans, 0, len(groups))
	for _, g := range groups {
		plans, err := model.GetSubscriptionPlansByGroupID(g.Id)
		if err != nil {
			common.ApiError(c, err)
			return
		}
		result = append(result, GroupWithPlans{
			SubscriptionPlanGroup: g,
			Plans:                 plans,
		})
	}
	common.ApiSuccess(c, result)
}

func AdminCreateSubscriptionPlanGroup(c *gin.Context) {
	var group model.SubscriptionPlanGroup
	if err := common.DecodeJson(c.Request.Body, &group); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	if strings.TrimSpace(group.Title) == "" {
		common.ApiErrorMsg(c, "套餐组标题不能为空")
		return
	}
	if err := model.CreateSubscriptionPlanGroup(&group); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, group)
}

func AdminUpdateSubscriptionPlanGroup(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		common.ApiErrorMsg(c, "无效的ID")
		return
	}
	var req map[string]interface{}
	if err := common.DecodeJson(c.Request.Body, &req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	updates := make(map[string]interface{})
	allowedFields := []string{"title", "subtitle", "tag", "features", "sort_order", "enabled"}
	for _, f := range allowedFields {
		if v, ok := req[f]; ok {
			updates[f] = v
		}
	}
	if len(updates) == 0 {
		common.ApiErrorMsg(c, "没有需要更新的字段")
		return
	}
	if err := model.UpdateSubscriptionPlanGroup(id, updates); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

func AdminDeleteSubscriptionPlanGroup(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		common.ApiErrorMsg(c, "无效的ID")
		return
	}
	if err := model.DeleteSubscriptionPlanGroup(id); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

func AdminUpdateSubscriptionPlanGroupStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		common.ApiErrorMsg(c, "无效的ID")
		return
	}
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := common.DecodeJson(c.Request.Body, &req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	if err := model.UpdateSubscriptionPlanGroup(id, map[string]interface{}{"enabled": req.Enabled}); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}
