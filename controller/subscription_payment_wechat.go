package controller

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting"
	"github.com/gin-gonic/gin"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/downloader"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
)

func SubscriptionRequestWechatPay(c *gin.Context) {
	if !setting.WechatPayEnabled {
		common.ApiErrorMsg(c, "微信支付未启用")
		return
	}

	plan, order, ok := validateAndCreateSubscriptionOrder(c, "SUBWX", PaymentMethodWechat)
	if !ok {
		return
	}

	ctx := c.Request.Context()
	client, err := getWechatPayClient(ctx)
	if err != nil {
		log.Printf("订阅微信支付客户端初始化失败: %v", err)
		_ = model.ExpireSubscriptionOrder(order.TradeNo)
		common.ApiErrorMsg(c, "支付配置错误")
		return
	}

	// 订阅支付使用独立的回调路径，不复用钱包充值的 notifyUrl
	notifyUrl := service.GetCallbackAddress() + "/api/subscription/wechat/notify"

	totalFen := int64(order.Money * 100)
	if totalFen <= 0 {
		_ = model.ExpireSubscriptionOrder(order.TradeNo)
		common.ApiErrorMsg(c, "支付金额过低")
		return
	}

	expireTime := time.Now().Add(15 * time.Minute)

	svc := native.NativeApiService{Client: client}
	resp, _, err := svc.Prepay(ctx, native.PrepayRequest{
		Appid:       core.String(setting.WechatPayAppId),
		Mchid:       core.String(setting.WechatPayMchId),
		Description: core.String(fmt.Sprintf("订阅套餐: %s", model.GetPlanGroupTitle(plan))),
		OutTradeNo:  core.String(order.TradeNo),
		NotifyUrl:   core.String(notifyUrl),
		TimeExpire:  &expireTime,
		Amount: &native.Amount{
			Total:    core.Int64(totalFen),
			Currency: core.String("CNY"),
		},
	})
	if err != nil {
		log.Printf("订阅微信支付预下单失败: %v, 订单: %s", err, order.TradeNo)
		_ = model.ExpireSubscriptionOrder(order.TradeNo)
		common.ApiErrorMsg(c, "拉起支付失败")
		return
	}

	if resp.CodeUrl == nil || *resp.CodeUrl == "" {
		log.Printf("订阅微信支付预下单返回空 code_url, 订单: %s", order.TradeNo)
		_ = model.ExpireSubscriptionOrder(order.TradeNo)
		common.ApiErrorMsg(c, "拉起支付失败")
		return
	}

	log.Printf("订阅微信支付订单创建成功 - 用户: %d, 订单: %s, 金额: %.2f", c.GetInt("id"), order.TradeNo, order.Money)

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data": gin.H{
			"code_url":  *resp.CodeUrl,
			"order_id":  order.TradeNo,
			"pay_money": order.Money,
		},
	})
}

func SubscriptionWechatNotify(c *gin.Context) {
	ctx := c.Request.Context()

	if _, err := getWechatPayClient(ctx); err != nil {
		log.Printf("订阅微信支付回调：客户端初始化失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": "内部错误"})
		return
	}

	certVisitor := downloader.MgrInstance().GetCertificateVisitor(setting.WechatPayMchId)
	handler, err := notify.NewRSANotifyHandler(
		setting.WechatPayMchApiV3Key,
		verifiers.NewSHA256WithRSAVerifier(certVisitor),
	)
	if err != nil {
		log.Printf("订阅微信支付回调：创建通知处理器失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": "内部错误"})
		return
	}

	var transaction payments.Transaction
	_, err = handler.ParseNotifyRequest(ctx, c.Request, &transaction)
	if err != nil {
		log.Printf("订阅微信支付回调：验签/解密失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": "验签失败"})
		return
	}

	if transaction.TradeState == nil || *transaction.TradeState != "SUCCESS" {
		if transaction.TradeState != nil {
			log.Printf("订阅微信支付回调：交易状态非成功: %s, 订单: %s",
				*transaction.TradeState, safeDeref(transaction.OutTradeNo))
		}
		c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "OK"})
		return
	}

	outTradeNo := safeDeref(transaction.OutTradeNo)
	if outTradeNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": "缺少订单号"})
		return
	}

	LockOrder(outTradeNo)
	defer UnlockOrder(outTradeNo)

	if err := model.CompleteSubscriptionOrder(outTradeNo, common.GetJsonString(transaction)); err != nil {
		log.Printf("订阅微信支付回调：处理失败: %v, 订单: %s", err, outTradeNo)
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": err.Error()})
		return
	}

	log.Printf("订阅微信支付回调：成功 - 订单: %s, 微信交易号: %s",
		outTradeNo, safeDeref(transaction.TransactionId))
	c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "OK"})
}
