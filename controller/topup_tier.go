package controller

import (
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

func AdminListTopupTiers(c *gin.Context) {
	tiers, err := model.GetAllTopupTiers()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, tiers)
}

func AdminCreateTopupTier(c *gin.Context) {
	var tier model.TopupTier
	if err := common.DecodeJson(c.Request.Body, &tier); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	if tier.Title == "" {
		common.ApiErrorMsg(c, "标题不能为空")
		return
	}
	if tier.Amount <= 0 {
		common.ApiErrorMsg(c, "金额必须为正数")
		return
	}
	if tier.Discount <= 0 || tier.Discount > 1 {
		tier.Discount = 1.0
	}
	if err := model.CreateTopupTier(&tier); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, tier)
}

func AdminUpdateTopupTier(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiErrorMsg(c, "无效的ID")
		return
	}
	var req map[string]interface{}
	if err := common.DecodeJson(c.Request.Body, &req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	updates := make(map[string]interface{})
	allowedFields := []string{"title", "subtitle", "tag", "amount", "discount", "bonus_quota", "features", "sort_order", "enabled"}
	for _, f := range allowedFields {
		if v, ok := req[f]; ok {
			updates[f] = v
		}
	}
	if len(updates) == 0 {
		common.ApiErrorMsg(c, "没有需要更新的字段")
		return
	}
	if err := model.UpdateTopupTier(id, updates); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

func AdminUpdateTopupTierStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
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
	if err := model.UpdateTopupTier(id, map[string]interface{}{"enabled": req.Enabled}); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

func AdminDeleteTopupTier(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiErrorMsg(c, "无效的ID")
		return
	}
	if err := model.DeleteTopupTier(id); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

func GetPublicTopupTiers(c *gin.Context) {
	tiers, err := model.GetEnabledTopupTiers()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.Header("Cache-Control", "public, max-age=60")
	common.ApiSuccess(c, tiers)
}
