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
	"github.com/smartwalle/alipay/v3"
	"github.com/thanhpk/randstr"
)

const PaymentMethodAlipay = "alipay"

type AlipayPayRequest struct {
	Amount        int64  `json:"amount"`
	PaymentMethod string `json:"payment_method"`
}

// getAlipayClient 初始化支付宝 SDK Client，自动检测公钥模式或证书模式
func getAlipayClient() (*alipay.Client, error) {
	client, err := alipay.New(setting.AlipayAppId, setting.AlipayPrivateKey, true)
	if err != nil {
		return nil, fmt.Errorf("初始化支付宝客户端失败: %v", err)
	}

	if setting.IsAlipayCertMode() {
		// 公钥证书模式：加载三个证书
		if err := client.LoadAppCertPublicKey(setting.AlipayAppCertPublicKey); err != nil {
			return nil, fmt.Errorf("加载应用公钥证书失败: %v", err)
		}
		if err := client.LoadAlipayCertPublicKey(setting.AlipayCertPublicKey); err != nil {
			return nil, fmt.Errorf("加载支付宝公钥证书失败: %v", err)
		}
		if err := client.LoadAliPayRootCert(setting.AlipayRootCert); err != nil {
			return nil, fmt.Errorf("加载支付宝根证书失败: %v", err)
		}
	} else {
		// 普通公钥模式：直接加载公钥字符串
		if err := client.LoadAliPayPublicKey(setting.AlipayPublicKey); err != nil {
			return nil, fmt.Errorf("加载支付宝公钥失败: %v", err)
		}
	}

	return client, nil
}

// getAlipayPayMoney 计算支付宝支付金额（CNY）
func getAlipayPayMoney(amount int64, group string) float64 {
	dAmount := decimal.NewFromInt(amount)
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
		dAmount = dAmount.Div(dQuotaPerUnit)
	}

	topupGroupRatio := common.GetTopupGroupRatio(group)
	if topupGroupRatio == 0 {
		topupGroupRatio = 1
	}

	// 使用支付宝专用单价，若为 0 则使用全局 CNY 单价
	unitPrice := setting.AlipayUnitPrice
	if unitPrice <= 0 {
		unitPrice = operation_setting.Price
	}

	discount := 1.0
	if ds, ok := operation_setting.GetPaymentSetting().AmountDiscount[int(amount)]; ok {
		if ds > 0 {
			discount = ds
		}
	}

	dTopupGroupRatio := decimal.NewFromFloat(topupGroupRatio)
	dPrice := decimal.NewFromFloat(unitPrice)
	dDiscount := decimal.NewFromFloat(discount)

	payMoney := dAmount.Mul(dPrice).Mul(dTopupGroupRatio).Mul(dDiscount)
	return payMoney.InexactFloat64()
}

func getAlipayMinTopup() int64 {
	minTopup := setting.AlipayMinTopUp
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

// RequestAlipayPay 创建支付宝当面付订单，返回 qr_code 供前端生成二维码
func RequestAlipayPay(c *gin.Context) {
	if !setting.AlipayEnabled {
		c.JSON(200, gin.H{"message": "error", "data": "支付宝支付未启用"})
		return
	}

	var req AlipayPayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "参数错误"})
		return
	}

	if req.Amount < getAlipayMinTopup() {
		c.JSON(200, gin.H{"message": "error", "data": fmt.Sprintf("充值数量不能小于 %d", getAlipayMinTopup())})
		return
	}
	if req.Amount > 10000 {
		c.JSON(200, gin.H{"message": "error", "data": "充值数量不能大于 10000"})
		return
	}

	id := c.GetInt("id")

	// 限制每个用户的待支付订单数，防止滥用
	const maxPendingOrders = 5
	pendingCount, err := model.CountUserPendingTopUps(id, PaymentMethodAlipay)
	if err != nil {
		log.Printf("支付宝查询待支付订单失败: %v", err)
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

	payMoney := getAlipayPayMoney(req.Amount, group)
	if payMoney < 0.01 {
		c.JSON(200, gin.H{"message": "error", "data": "充值金额过低"})
		return
	}

	// 生成唯一订单号
	tradeNo := fmt.Sprintf("ALIPAY-%d-%d-%s", id, time.Now().UnixMilli(), randstr.String(6))

	// Token 模式下归一化 Amount
	amount := req.Amount
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		amount = int64(float64(req.Amount) / common.QuotaPerUnit)
		if amount < 1 {
			amount = 1
		}
	}

	// 创建本地订单
	topUp := &model.TopUp{
		UserId:        id,
		Amount:        amount,
		Money:         payMoney,
		TradeNo:       tradeNo,
		PaymentMethod: PaymentMethodAlipay,
		CreateTime:    time.Now().Unix(),
		Status:        common.TopUpStatusPending,
	}
	if err := topUp.Insert(); err != nil {
		log.Printf("支付宝创建本地订单失败: %v", err)
		c.JSON(200, gin.H{"message": "error", "data": "创建订单失败"})
		return
	}

	// 初始化支付宝客户端
	client, err := getAlipayClient()
	if err != nil {
		log.Printf("支付宝客户端初始化失败: %v", err)
		topUp.Status = common.TopUpStatusFailed
		_ = topUp.Update()
		c.JSON(200, gin.H{"message": "error", "data": "支付配置错误"})
		return
	}

	// 确定回调地址
	notifyUrl := setting.AlipayNotifyUrl
	if notifyUrl == "" {
		notifyUrl = service.GetCallbackAddress() + "/api/pay/notify/alipay"
	}

	// 金额格式化为元（支付宝要求字符串，精确到分）
	totalAmount := fmt.Sprintf("%.2f", payMoney)

	// 调用支付宝当面付预创建
	var p = alipay.TradePreCreate{}
	p.NotifyURL = notifyUrl
	p.Subject = fmt.Sprintf("充值 %d 额度", req.Amount)
	p.OutTradeNo = tradeNo
	p.TotalAmount = totalAmount
	p.TimeoutExpress = "15m" // 15 分钟超时

	resp, err := client.TradePreCreate(c.Request.Context(), p)
	if err != nil {
		log.Printf("支付宝预创建失败: %v, 订单: %s", err, tradeNo)
		topUp.Status = common.TopUpStatusFailed
		_ = topUp.Update()
		c.JSON(200, gin.H{"message": "error", "data": "拉起支付失败"})
		return
	}

	if resp.QRCode == "" {
		log.Printf("支付宝预创建返回空 qr_code, 订单: %s, resp: %+v", tradeNo, resp)
		topUp.Status = common.TopUpStatusFailed
		_ = topUp.Update()
		c.JSON(200, gin.H{"message": "error", "data": "拉起支付失败"})
		return
	}

	// 保存 qr_code 到订单
	topUp.CodeUrl = resp.QRCode
	_ = topUp.Update()

	log.Printf("支付宝订单创建成功 - 用户: %d, 订单: %s, 金额: %s CNY", id, tradeNo, totalAmount)

	c.JSON(200, gin.H{
		"message": "success",
		"data": gin.H{
			"code_url":  resp.QRCode,
			"order_id":  tradeNo,
			"pay_money": payMoney,
		},
	})
}

// RequestAlipayAmount 查询支付宝实付金额
func RequestAlipayAmount(c *gin.Context) {
	var req AlipayPayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "参数错误"})
		return
	}
	if req.Amount < getAlipayMinTopup() {
		c.JSON(200, gin.H{"message": "error", "data": fmt.Sprintf("充值数量不能小于 %d", getAlipayMinTopup())})
		return
	}
	id := c.GetInt("id")
	group, err := model.GetUserGroup(id, true)
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "获取用户分组失败"})
		return
	}
	payMoney := getAlipayPayMoney(req.Amount, group)
	if payMoney <= 0.01 {
		c.JSON(200, gin.H{"message": "error", "data": "充值金额过低"})
		return
	}
	c.JSON(200, gin.H{"message": "success", "data": strconv.FormatFloat(payMoney, 'f', 2, 64)})
}

// AlipayNotify 处理支付宝异步回调通知
func AlipayNotify(c *gin.Context) {
	client, err := getAlipayClient()
	if err != nil {
		log.Printf("支付宝回调：客户端初始化失败: %v", err)
		c.String(http.StatusInternalServerError, "fail")
		return
	}

	// 解析 POST form 参数
	if err := c.Request.ParseForm(); err != nil {
		log.Printf("支付宝回调：解析 form 失败: %v", err)
		c.String(http.StatusBadRequest, "fail")
		return
	}

	// 验签通知
	notification, err := client.DecodeNotification(c.Request.Context(), c.Request.PostForm)
	if err != nil {
		log.Printf("支付宝回调：验签失败: %v", err)
		c.String(http.StatusBadRequest, "fail")
		return
	}

	// 检查交易状态
	if notification.TradeStatus != alipay.TradeStatusSuccess && notification.TradeStatus != alipay.TradeStatusFinished {
		log.Printf("支付宝回调：交易状态非成功: %s, 订单: %s", notification.TradeStatus, notification.OutTradeNo)
		// 即使不是成功状态也返回 success，避免支付宝重复通知
		c.String(http.StatusOK, "success")
		return
	}

	outTradeNo := notification.OutTradeNo
	if outTradeNo == "" {
		log.Printf("支付宝回调：缺少商户订单号")
		c.String(http.StatusBadRequest, "fail")
		return
	}

	LockOrder(outTradeNo)
	defer UnlockOrder(outTradeNo)

	if err := model.RechargeAlipay(outTradeNo); err != nil {
		log.Printf("支付宝回调：充值处理失败: %v, 订单: %s", err, outTradeNo)
		c.String(http.StatusInternalServerError, "fail")
		return
	}

	log.Printf("支付宝回调：充值成功 - 订单: %s, 支付宝交易号: %s", outTradeNo, notification.TradeNo)
	// 支付宝要求返回纯文本 "success"
	c.String(http.StatusOK, "success")
}

// closeAlipayOrder 调用支付宝关闭订单 API
func closeAlipayOrder(tradeNo string) {
	client, err := getAlipayClient()
	if err != nil {
		log.Printf("关闭支付宝订单失败（客户端初始化）: %v, 订单: %s", err, tradeNo)
		return
	}

	var p = alipay.TradeClose{}
	p.OutTradeNo = tradeNo

	_, err = client.TradeClose(context.Background(), p)
	if err != nil {
		log.Printf("关闭支付宝订单失败: %v, 订单: %s", err, tradeNo)
		return
	}
	log.Printf("支付宝订单已关闭: %s", tradeNo)
}
