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
)

// ---- Shared types ----

type BillingPreferenceRequest struct {
	BillingPreference string `json:"billing_preference"`
}

type SubscriptionPayRequest struct {
	PlanId     int    `json:"plan_id"`
	PeriodType string `json:"period_type"`
}

// validateAndCreateSubscriptionOrder validates the plan + period, checks purchase limits,
// and creates a pending SubscriptionOrder. Returns the plan, order, and true on success.
func validateAndCreateSubscriptionOrder(c *gin.Context, tradePrefix string, paymentMethod string) (*model.SubscriptionPlan, *model.SubscriptionOrder, bool) {
	var req SubscriptionPayRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.PlanId <= 0 {
		common.ApiErrorMsg(c, "参数错误")
		return nil, nil, false
	}
	if !model.ValidatePeriodType(req.PeriodType) {
		common.ApiErrorMsg(c, "无效的付款周期")
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
	if !plan.IsPeriodEnabled(req.PeriodType) {
		common.ApiErrorMsg(c, "该付款周期未启用")
		return nil, nil, false
	}

	price := plan.CalcPeriodPrice(req.PeriodType)
	if price < 0.01 {
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
		PeriodType:    req.PeriodType,
		Money:         price,
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

type PublicPeriodDTO struct {
	Enabled  bool    `json:"enabled"`
	Price    float64 `json:"price"`
	Discount int     `json:"discount"`
	Features string  `json:"features,omitempty"`
}

type PublicSubscriptionPlanDTO struct {
	Id           int                        `json:"id"`
	Title        string                     `json:"title"`
	Subtitle     string                     `json:"subtitle"`
	Tag          string                     `json:"tag"`
	Features     string                     `json:"features"`
	PriceMonthly float64                    `json:"price_monthly"`
	Currency     string                     `json:"currency"`
	Periods      map[string]PublicPeriodDTO  `json:"periods"`
	TotalAmount  int64                      `json:"total_amount"`
	UpgradeGroup string                     `json:"upgrade_group"`
}

func GetPublicSubscriptionPlans(c *gin.Context) {
	plans, err := model.GetEnabledSubscriptionPlans()
	if err != nil {
		common.ApiError(c, err)
		return
	}

	result := make([]PublicSubscriptionPlanDTO, 0, len(plans))
	for _, p := range plans {
		// Skip plans with no enabled period
		if !p.MonthlyEnabled && !p.QuarterlyEnabled && !p.YearlyEnabled {
			continue
		}
		dto := PublicSubscriptionPlanDTO{
			Id:           p.Id,
			Title:        p.Title,
			Subtitle:     p.Subtitle,
			Tag:          p.Tag,
			Features:     p.Features,
			PriceMonthly: p.PriceMonthly,
			Currency:     p.Currency,
			TotalAmount:  p.TotalAmount,
			UpgradeGroup: p.UpgradeGroup,
			Periods: map[string]PublicPeriodDTO{
				model.PeriodMonthly: {
					Enabled:  p.MonthlyEnabled,
					Price:    p.CalcPeriodPrice(model.PeriodMonthly),
					Discount: 0,
					Features: p.MonthlyFeatures,
				},
				model.PeriodQuarterly: {
					Enabled:  p.QuarterlyEnabled,
					Price:    p.CalcPeriodPrice(model.PeriodQuarterly),
					Discount: p.QuarterlyDiscount,
					Features: p.QuarterlyFeatures,
				},
				model.PeriodYearly: {
					Enabled:  p.YearlyEnabled,
					Price:    p.CalcPeriodPrice(model.PeriodYearly),
					Discount: p.YearlyDiscount,
					Features: p.YearlyFeatures,
				},
			},
		}
		result = append(result, dto)
	}

	c.Header("Cache-Control", "public, max-age=60")
	common.ApiSuccess(c, result)
}

// ---- User APIs ----

func GetSubscriptionPlans(c *gin.Context) {
	plans, err := model.GetEnabledSubscriptionPlans()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, plans)
}

func GetSubscriptionSelf(c *gin.Context) {
	userId := c.GetInt("id")
	settingMap, _ := model.GetUserSetting(userId, false)
	pref := common.NormalizeBillingPreference(settingMap.BillingPreference)

	allSubscriptions, err := model.GetAllUserSubscriptions(userId)
	if err != nil {
		allSubscriptions = []model.SubscriptionSummary{}
	}

	activeSubscriptions, err := model.GetAllActiveUserSubscriptions(userId)
	if err != nil {
		activeSubscriptions = []model.SubscriptionSummary{}
	}

	common.ApiSuccess(c, gin.H{
		"billing_preference": pref,
		"subscriptions":      activeSubscriptions,
		"all_subscriptions":  allSubscriptions,
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

// ---- Admin APIs: Plan Management ----

func AdminListSubscriptionPlans(c *gin.Context) {
	plans, err := model.GetAllSubscriptionPlans()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, plans)
}

func AdminCreateSubscriptionPlan(c *gin.Context) {
	var plan model.SubscriptionPlan
	if err := common.DecodeJson(c.Request.Body, &plan); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	plan.Id = 0
	if strings.TrimSpace(plan.Title) == "" {
		common.ApiErrorMsg(c, "订阅标题不能为空")
		return
	}
	if plan.PriceMonthly < 0 {
		common.ApiErrorMsg(c, "价格不能为负数")
		return
	}
	if plan.PriceMonthly > 99999 {
		common.ApiErrorMsg(c, "价格不能超过99999")
		return
	}
	if plan.Currency == "" {
		plan.Currency = "USD"
	}
	if plan.MaxPurchasePerUser < 0 {
		common.ApiErrorMsg(c, "购买上限不能为负数")
		return
	}
	if plan.TotalAmount < 0 {
		common.ApiErrorMsg(c, "总额度不能为负数")
		return
	}
	plan.UpgradeGroup = strings.TrimSpace(plan.UpgradeGroup)
	if plan.UpgradeGroup != "" {
		if _, ok := ratio_setting.GetGroupRatioCopy()[plan.UpgradeGroup]; !ok {
			common.ApiErrorMsg(c, "升级分组不存在")
			return
		}
	}
	plan.QuotaResetPeriod = model.NormalizeResetPeriod(plan.QuotaResetPeriod)
	if plan.QuotaResetPeriod == model.SubscriptionResetCustom && plan.QuotaResetCustomSeconds <= 0 {
		common.ApiErrorMsg(c, "自定义重置周期需大于0秒")
		return
	}
	// Clamp discount values
	if plan.QuarterlyDiscount < 0 {
		plan.QuarterlyDiscount = 0
	}
	if plan.QuarterlyDiscount > 100 {
		plan.QuarterlyDiscount = 100
	}
	if plan.YearlyDiscount < 0 {
		plan.YearlyDiscount = 0
	}
	if plan.YearlyDiscount > 100 {
		plan.YearlyDiscount = 100
	}

	if err := model.CreateSubscriptionPlan(&plan); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, plan)
}

func AdminUpdateSubscriptionPlan(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		common.ApiErrorMsg(c, "无效的ID")
		return
	}
	var plan model.SubscriptionPlan
	if err := common.DecodeJson(c.Request.Body, &plan); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	if plan.PriceMonthly < 0 {
		common.ApiErrorMsg(c, "价格不能为负数")
		return
	}
	if plan.PriceMonthly > 99999 {
		common.ApiErrorMsg(c, "价格不能超过99999")
		return
	}
	if plan.Currency == "" {
		plan.Currency = "USD"
	}
	if plan.MaxPurchasePerUser < 0 {
		common.ApiErrorMsg(c, "购买上限不能为负数")
		return
	}
	if plan.TotalAmount < 0 {
		common.ApiErrorMsg(c, "总额度不能为负数")
		return
	}
	plan.UpgradeGroup = strings.TrimSpace(plan.UpgradeGroup)
	if plan.UpgradeGroup != "" {
		if _, ok := ratio_setting.GetGroupRatioCopy()[plan.UpgradeGroup]; !ok {
			common.ApiErrorMsg(c, "升级分组不存在")
			return
		}
	}
	plan.QuotaResetPeriod = model.NormalizeResetPeriod(plan.QuotaResetPeriod)
	if plan.QuotaResetPeriod == model.SubscriptionResetCustom && plan.QuotaResetCustomSeconds <= 0 {
		common.ApiErrorMsg(c, "自定义重置周期需大于0秒")
		return
	}
	// Clamp discount values
	if plan.QuarterlyDiscount < 0 {
		plan.QuarterlyDiscount = 0
	}
	if plan.QuarterlyDiscount > 100 {
		plan.QuarterlyDiscount = 100
	}
	if plan.YearlyDiscount < 0 {
		plan.YearlyDiscount = 0
	}
	if plan.YearlyDiscount > 100 {
		plan.YearlyDiscount = 100
	}

	updateMap := map[string]interface{}{
		"title":                        plan.Title,
		"subtitle":                     plan.Subtitle,
		"tag":                          plan.Tag,
		"features":                     plan.Features,
		"price_monthly":                plan.PriceMonthly,
		"currency":                     plan.Currency,
		"monthly_enabled":              plan.MonthlyEnabled,
		"monthly_features":             plan.MonthlyFeatures,
		"monthly_stripe_price_id":      plan.MonthlyStripePriceId,
		"monthly_creem_product_id":     plan.MonthlyCreemProductId,
		"quarterly_enabled":            plan.QuarterlyEnabled,
		"quarterly_discount":           plan.QuarterlyDiscount,
		"quarterly_features":           plan.QuarterlyFeatures,
		"quarterly_stripe_price_id":    plan.QuarterlyStripePriceId,
		"quarterly_creem_product_id":   plan.QuarterlyCreemProductId,
		"yearly_enabled":               plan.YearlyEnabled,
		"yearly_discount":              plan.YearlyDiscount,
		"yearly_features":              plan.YearlyFeatures,
		"yearly_stripe_price_id":       plan.YearlyStripePriceId,
		"yearly_creem_product_id":      plan.YearlyCreemProductId,
		"total_amount":                 plan.TotalAmount,
		"quota_reset_period":           plan.QuotaResetPeriod,
		"quota_reset_custom_seconds":   plan.QuotaResetCustomSeconds,
		"upgrade_group":                plan.UpgradeGroup,
		"max_purchase_per_user":        plan.MaxPurchasePerUser,
		"sort_order":                   plan.SortOrder,
		"enabled":                      plan.Enabled,
		"updated_at":                   common.GetTimestamp(),
	}
	if err := model.UpdateSubscriptionPlan(id, updateMap); err != nil {
		common.ApiError(c, err)
		return
	}
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
	if err := model.UpdateSubscriptionPlan(id, map[string]interface{}{"enabled": *req.Enabled}); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

func AdminDeleteSubscriptionPlan(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		common.ApiErrorMsg(c, "无效的ID")
		return
	}
	if err := model.DeleteSubscriptionPlan(id); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

// ---- Admin: Bind Subscription ----

type AdminBindSubscriptionRequest struct {
	UserId     int    `json:"user_id"`
	PlanId     int    `json:"plan_id"`
	PeriodType string `json:"period_type"`
}

func AdminBindSubscription(c *gin.Context) {
	var req AdminBindSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.UserId <= 0 || req.PlanId <= 0 {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	if req.PeriodType == "" {
		req.PeriodType = model.PeriodMonthly
	}
	msg, err := model.AdminBindSubscription(req.UserId, req.PlanId, req.PeriodType, "")
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

// ---- Admin: User Subscription Management ----

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
	PlanId     int    `json:"plan_id"`
	PeriodType string `json:"period_type"`
}

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
	if req.PeriodType == "" {
		req.PeriodType = model.PeriodMonthly
	}
	msg, err := model.AdminBindSubscription(userId, req.PlanId, req.PeriodType, "")
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
