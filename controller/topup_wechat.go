package controller

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/thanhpk/randstr"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/downloader"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
)

const PaymentMethodWechat = "wechat"

type WechatPayRequest struct {
	Amount        int64  `json:"amount"`
	PaymentMethod string `json:"payment_method"`
}

// getWechatPayClient 初始化微信支付 V3 SDK Client
func getWechatPayClient(ctx context.Context) (*core.Client, error) {
	privateKey, err := utils.LoadPrivateKey(setting.WechatPayMchPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("加载商户私钥失败: %v", err)
	}

	client, err := core.NewClient(
		ctx,
		option.WithWechatPayAutoAuthCipher(
			setting.WechatPayMchId,
			setting.WechatPayMchSerialNo,
			privateKey,
			setting.WechatPayMchApiV3Key,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("初始化微信支付客户端失败: %v", err)
	}
	return client, nil
}

// getWechatPayMoney 计算微信支付金额（CNY）
func getWechatPayMoney(amount int64, group string) float64 {
	dAmount := decimal.NewFromInt(amount)
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
		dAmount = dAmount.Div(dQuotaPerUnit)
	}

	topupGroupRatio := common.GetTopupGroupRatio(group)
	if topupGroupRatio == 0 {
		topupGroupRatio = 1
	}

	// 使用微信专用单价，若为 0 则使用全局 CNY 单价
	unitPrice := setting.WechatPayUnitPrice
	if unitPrice <= 0 {
		unitPrice = operation_setting.Price
	}

	discount := 1.0
	if tier, err := model.GetTopupTierByAmount(amount); err == nil && tier != nil {
		if tier.Discount > 0 {
			discount = tier.Discount
		}
	}

	dTopupGroupRatio := decimal.NewFromFloat(topupGroupRatio)
	dPrice := decimal.NewFromFloat(unitPrice)
	dDiscount := decimal.NewFromFloat(discount)

	payMoney := dAmount.Mul(dPrice).Mul(dTopupGroupRatio).Mul(dDiscount)
	return payMoney.InexactFloat64()
}

func getWechatPayMinTopup() int64 {
	minTopup := setting.WechatPayMinTopUp
	// 全局 MinTopUp 作为下限，确保各网关不低于全局设定
	if operation_setting.MinTopUp > minTopup {
		minTopup = operation_setting.MinTopUp
	}
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		dMinTopup := decimal.NewFromInt(int64(minTopup))
		dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
		minTopup = int(dMinTopup.Mul(dQuotaPerUnit).IntPart())
	}
	return int64(minTopup)
}

// RequestWechatPay 创建微信支付 Native 订单，返回 code_url 供前端生成二维码
func RequestWechatPay(c *gin.Context) {
	if !setting.WechatPayEnabled {
		c.JSON(200, gin.H{"message": "error", "data": "微信支付未启用"})
		return
	}

	var req WechatPayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "参数错误"})
		return
	}

	if req.Amount < getWechatPayMinTopup() {
		c.JSON(200, gin.H{"message": "error", "data": fmt.Sprintf("充值数量不能小于 %d", getWechatPayMinTopup())})
		return
	}
	if req.Amount > 10000 {
		c.JSON(200, gin.H{"message": "error", "data": "充值数量不能大于 10000"})
		return
	}

	id := c.GetInt("id")

	// 限制每个用户的待支付订单数，防止滥用
	const maxPendingOrders = 5
	pendingCount, err := model.CountUserPendingTopUps(id, PaymentMethodWechat)
	if err != nil {
		log.Printf("微信支付查询待支付订单失败: %v", err)
		c.JSON(200, gin.H{"message": "error", "data": "系统繁忙，请稍后重试"})
		return
	}
	if pendingCount >= maxPendingOrders {
		c.JSON(200, gin.H{"message": "error", "data": "您有太多未完成的支付订单，请先完成或等待已有订单过期后再试"})
		return
	}

	group, err := model.GetUserGroup(id, true)
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "获取用户分组失败"})
		return
	}

	payMoney := getWechatPayMoney(req.Amount, group)
	if payMoney < 0.01 {
		c.JSON(200, gin.H{"message": "error", "data": "充值金额过低"})
		return
	}

	// 生成唯一订单号
	tradeNo := fmt.Sprintf("WXPAY-%d-%d-%s", id, time.Now().UnixMilli(), randstr.String(6))

	// Token 模式下归一化 Amount（存等价数量，避免 RechargeWechat 双重放大）
	amount := req.Amount
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		amount = int64(float64(req.Amount) / common.QuotaPerUnit)
		if amount < 1 {
			amount = 1
		}
	}

	// 创建本地订单
	topUp := &model.TopUp{
		UserId:          id,
		Amount:          amount,
		Money:           payMoney,
		TradeNo:         tradeNo,
		PaymentMethod:   PaymentMethodWechat,
		PaymentProvider: model.PaymentProviderWechat,
		CreateTime:      time.Now().Unix(),
		Status:          common.TopUpStatusPending,
	}
	if err := topUp.Insert(); err != nil {
		log.Printf("微信支付创建本地订单失败: %v", err)
		c.JSON(200, gin.H{"message": "error", "data": "创建订单失败"})
		return
	}

	// 初始化微信支付客户端
	ctx := c.Request.Context()
	client, err := getWechatPayClient(ctx)
	if err != nil {
		log.Printf("微信支付客户端初始化失败: %v", err)
		topUp.Status = common.TopUpStatusFailed
		_ = topUp.Update()
		c.JSON(200, gin.H{"message": "error", "data": "支付配置错误"})
		return
	}

	// 确定回调地址
	notifyUrl := setting.WechatPayNotifyUrl
	if notifyUrl == "" {
		notifyUrl = service.GetCallbackAddress() + "/api/pay/notify/wechat"
	}

	// 金额转换为分（微信支付要求整数分）
	totalFen := int64(payMoney * 100)
	if totalFen <= 0 {
		topUp.Status = common.TopUpStatusFailed
		_ = topUp.Update()
		c.JSON(200, gin.H{"message": "error", "data": "充值金额过低"})
		return
	}

	// 调用微信支付 Native 预下单
	svc := native.NativeApiService{Client: client}
	description := fmt.Sprintf("充值 %d 额度", req.Amount)
	currency := "CNY"

	// 订单 15 分钟后过期，防止大量未支付订单积压
	expireTime := time.Now().Add(15 * time.Minute)

	resp, _, err := svc.Prepay(ctx, native.PrepayRequest{
		Appid:       core.String(setting.WechatPayAppId),
		Mchid:       core.String(setting.WechatPayMchId),
		Description: core.String(description),
		OutTradeNo:  core.String(tradeNo),
		NotifyUrl:   core.String(notifyUrl),
		TimeExpire:  &expireTime,
		Amount: &native.Amount{
			Total:    core.Int64(totalFen),
			Currency: core.String(currency),
		},
	})
	if err != nil {
		log.Printf("微信支付预下单失败: %v, 订单: %s", err, tradeNo)
		topUp.Status = common.TopUpStatusFailed
		_ = topUp.Update()
		c.JSON(200, gin.H{"message": "error", "data": "拉起支付失败"})
		return
	}

	if resp.CodeUrl == nil || *resp.CodeUrl == "" {
		log.Printf("微信支付预下单返回空 code_url, 订单: %s", tradeNo)
		topUp.Status = common.TopUpStatusFailed
		_ = topUp.Update()
		c.JSON(200, gin.H{"message": "error", "data": "拉起支付失败"})
		return
	}

	// 保存 code_url 到订单，供用户刷新页面后重新展示二维码
	topUp.CodeUrl = *resp.CodeUrl
	_ = topUp.Update()

	log.Printf("微信支付订单创建成功 - 用户: %d, 订单: %s, 金额: %.2f CNY", id, tradeNo, payMoney)

	c.JSON(200, gin.H{
		"message": "success",
		"data": gin.H{
			"code_url":  *resp.CodeUrl,
			"order_id":  tradeNo,
			"pay_money": payMoney,
		},
	})
}

// RequestWechatAmount 查询微信支付的实付金额（参照 RequestStripeAmount 模式）
func RequestWechatAmount(c *gin.Context) {
	var req WechatPayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "参数错误"})
		return
	}
	if req.Amount < getWechatPayMinTopup() {
		c.JSON(200, gin.H{"message": "error", "data": fmt.Sprintf("充值数量不能小于 %d", getWechatPayMinTopup())})
		return
	}
	id := c.GetInt("id")
	group, err := model.GetUserGroup(id, true)
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "获取用户分组失败"})
		return
	}
	payMoney := getWechatPayMoney(req.Amount, group)
	if payMoney <= 0.01 {
		c.JSON(200, gin.H{"message": "error", "data": "充值金额过低"})
		return
	}
	c.JSON(200, gin.H{"message": "success", "data": strconv.FormatFloat(payMoney, 'f', 2, 64)})
}

// WechatPayNotify 处理微信支付异步回调通知
func WechatPayNotify(c *gin.Context) {
	ctx := c.Request.Context()

	// 确保 downloader manager 已注册（getWechatPayClient 会触发注册）
	if _, err := getWechatPayClient(ctx); err != nil {
		log.Printf("微信支付回调：客户端初始化失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": "内部错误"})
		return
	}

	// 从证书下载管理器获取验证器，用于验签
	certVisitor := downloader.MgrInstance().GetCertificateVisitor(setting.WechatPayMchId)
	handler, err := notify.NewRSANotifyHandler(
		setting.WechatPayMchApiV3Key,
		verifiers.NewSHA256WithRSAVerifier(certVisitor),
	)
	if err != nil {
		log.Printf("微信支付回调：创建通知处理器失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": "内部错误"})
		return
	}

	var transaction payments.Transaction
	_, err = handler.ParseNotifyRequest(ctx, c.Request, &transaction)
	if err != nil {
		log.Printf("微信支付回调：验签/解密失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": "验签失败"})
		return
	}

	// 检查交易状态
	if transaction.TradeState == nil {
		log.Printf("微信支付回调：交易状态为空")
		c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "OK"})
		return
	}

	if *transaction.TradeState != "SUCCESS" {
		log.Printf("微信支付回调：交易状态非成功: %s, 订单: %s",
			*transaction.TradeState, safeDeref(transaction.OutTradeNo))
		c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "OK"})
		return
	}

	outTradeNo := safeDeref(transaction.OutTradeNo)
	if outTradeNo == "" {
		log.Printf("微信支付回调：缺少商户订单号")
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": "缺少订单号"})
		return
	}

	LockOrder(outTradeNo)
	defer UnlockOrder(outTradeNo)

	if err := model.RechargeWechat(outTradeNo); err != nil {
		log.Printf("微信支付回调：充值处理失败: %v, 订单: %s", err, outTradeNo)
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": err.Error()})
		return
	}

	log.Printf("微信支付回调：充值成功 - 订单: %s, 微信交易号: %s",
		outTradeNo, safeDeref(transaction.TransactionId))
	c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "OK"})
}

// CancelTopUp 取消待支付订单（用户取消自己的，管理员取消任意）
func CancelTopUp(c *gin.Context) {
	var req struct {
		TradeNo string `json:"trade_no"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.TradeNo == "" {
		c.JSON(200, gin.H{"message": "error", "data": "参数错误"})
		return
	}

	userId := c.GetInt("id")

	LockOrder(req.TradeNo)
	defer UnlockOrder(req.TradeNo)

	// 查询订单以判断是否为微信支付订单
	topUp := model.GetTopUpByTradeNo(req.TradeNo)
	if topUp == nil {
		c.JSON(200, gin.H{"message": "error", "data": "订单不存在"})
		return
	}

	isAdminUser := c.GetInt("role") >= common.RoleAdminUser
	if err := model.CancelTopUp(req.TradeNo, userId, isAdminUser); err != nil {
		c.JSON(200, gin.H{"message": "error", "data": err.Error()})
		return
	}

	// 如果是微信支付订单，同步关闭微信侧订单
	if topUp.PaymentMethod == PaymentMethodWechat {
		go closeWechatOrder(req.TradeNo)
	}
	// 如果是支付宝订单，同步关闭支付宝侧订单
	if topUp.PaymentMethod == PaymentMethodAlipay {
		go closeAlipayOrder(req.TradeNo)
	}

	log.Printf("订单取消成功 - 用户: %d, 订单: %s, 管理员: %v", userId, req.TradeNo, isAdminUser)
	c.JSON(200, gin.H{"message": "success", "data": "订单已取消"})
}

// closeWechatOrder 调用微信支付关闭订单 API
func closeWechatOrder(tradeNo string) {
	ctx := context.Background()
	client, err := getWechatPayClient(ctx)
	if err != nil {
		log.Printf("关闭微信订单失败（客户端初始化）: %v, 订单: %s", err, tradeNo)
		return
	}
	svc := native.NativeApiService{Client: client}
	_, err = svc.CloseOrder(ctx, native.CloseOrderRequest{
		OutTradeNo: core.String(tradeNo),
		Mchid:      core.String(setting.WechatPayMchId),
	})
	if err != nil {
		log.Printf("关闭微信订单失败: %v, 订单: %s", err, tradeNo)
		return
	}
	log.Printf("微信订单已关闭: %s", tradeNo)
}

// GetTopUpQrCode 获取待支付微信订单的二维码（用户刷新页面后重新展示）
func GetTopUpQrCode(c *gin.Context) {
	tradeNo := c.Query("trade_no")
	if tradeNo == "" {
		c.JSON(200, gin.H{"message": "error", "data": "参数错误"})
		return
	}

	userId := c.GetInt("id")
	topUp := model.GetTopUpByTradeNo(tradeNo)
	if topUp == nil {
		c.JSON(200, gin.H{"message": "error", "data": "订单不存在"})
		return
	}
	if topUp.UserId != userId {
		c.JSON(200, gin.H{"message": "error", "data": "无权操作"})
		return
	}
	if topUp.Status != common.TopUpStatusPending {
		c.JSON(200, gin.H{"message": "error", "data": "订单非待支付状态"})
		return
	}
	if topUp.PaymentMethod != PaymentMethodWechat && topUp.PaymentMethod != PaymentMethodAlipay {
		c.JSON(200, gin.H{"message": "error", "data": "仅支持微信/支付宝支付订单"})
		return
	}
	if topUp.CodeUrl == "" {
		c.JSON(200, gin.H{"message": "error", "data": "二维码已失效，请重新创建订单"})
		return
	}

	// 检查订单是否已过期（创建后 15 分钟）
	if time.Now().Unix()-topUp.CreateTime > 15*60 {
		c.JSON(200, gin.H{"message": "error", "data": "订单已过期，请重新创建"})
		return
	}

	c.JSON(200, gin.H{
		"message": "success",
		"data": gin.H{
			"code_url":  topUp.CodeUrl,
			"pay_money": topUp.Money,
			"trade_no":  topUp.TradeNo,
		},
	})
}

// StartTopUpExpireCleanup 启动后台定时任务，每 15 分钟清理过期的 pending 订单
func StartTopUpExpireCleanup() {
	for {
		time.Sleep(15 * time.Minute)
		expireBefore := time.Now().Add(-15 * time.Minute).Unix()
		wechatTradeNos, alipayTradeNos, err := model.ExpirePendingTopUps(expireBefore)
		if err != nil {
			log.Printf("清理过期订单失败: %v", err)
			continue
		}
		if len(wechatTradeNos) > 0 {
			log.Printf("已将 %d 笔过期微信订单标记为 expired，正在关闭微信侧订单", len(wechatTradeNos))
			for _, tradeNo := range wechatTradeNos {
				closeWechatOrder(tradeNo)
			}
		}
		if len(alipayTradeNos) > 0 {
			log.Printf("已将 %d 笔过期支付宝订单标记为 expired，正在关闭支付宝侧订单", len(alipayTradeNos))
			for _, tradeNo := range alipayTradeNos {
				closeAlipayOrder(tradeNo)
			}
		}
	}
}

func safeDeref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
