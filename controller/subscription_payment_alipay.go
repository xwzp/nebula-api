package controller

import (
	"fmt"
	"log"
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting"
	"github.com/gin-gonic/gin"
	"github.com/smartwalle/alipay/v3"
)

func SubscriptionRequestAlipayPay(c *gin.Context) {
	if !setting.AlipayEnabled {
		common.ApiErrorMsg(c, "支付宝支付未启用")
		return
	}

	plan, order, ok := validateAndCreateSubscriptionOrder(c, "SUBAL", PaymentMethodAlipay)
	if !ok {
		return
	}

	client, err := getAlipayClient()
	if err != nil {
		log.Printf("订阅支付宝客户端初始化失败: %v", err)
		_ = model.ExpireSubscriptionOrder(order.TradeNo)
		common.ApiErrorMsg(c, "支付配置错误")
		return
	}

	// 订阅支付使用独立的回调路径，不复用钱包充值的 notifyUrl
	notifyUrl := service.GetCallbackAddress() + "/api/subscription/alipay/notify"
	totalAmount := fmt.Sprintf("%.2f", order.Money)

	var p = alipay.TradePreCreate{}
	p.NotifyURL = notifyUrl
	p.Subject = fmt.Sprintf("订阅套餐: %s", plan.Title)
	p.OutTradeNo = order.TradeNo
	p.TotalAmount = totalAmount
	p.TimeoutExpress = "15m"

	resp, err := client.TradePreCreate(c.Request.Context(), p)
	if err != nil {
		log.Printf("订阅支付宝预创建失败: %v, 订单: %s", err, order.TradeNo)
		_ = model.ExpireSubscriptionOrder(order.TradeNo)
		common.ApiErrorMsg(c, "拉起支付失败")
		return
	}

	if resp.QRCode == "" {
		log.Printf("订阅支付宝预创建返回空 qr_code, 订单: %s, resp: %+v", order.TradeNo, resp)
		_ = model.ExpireSubscriptionOrder(order.TradeNo)
		common.ApiErrorMsg(c, "拉起支付失败")
		return
	}

	log.Printf("订阅支付宝订单创建成功 - 用户: %d, 订单: %s, 金额: %s CNY", c.GetInt("id"), order.TradeNo, totalAmount)

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data": gin.H{
			"code_url":  resp.QRCode,
			"order_id":  order.TradeNo,
			"pay_money": order.Money,
		},
	})
}

func SubscriptionAlipayNotify(c *gin.Context) {
	client, err := getAlipayClient()
	if err != nil {
		log.Printf("订阅支付宝回调：客户端初始化失败: %v", err)
		c.String(http.StatusInternalServerError, "fail")
		return
	}

	if err := c.Request.ParseForm(); err != nil {
		log.Printf("订阅支付宝回调：解析 form 失败: %v", err)
		c.String(http.StatusBadRequest, "fail")
		return
	}

	notification, err := client.DecodeNotification(c.Request.Context(), c.Request.PostForm)
	if err != nil {
		log.Printf("订阅支付宝回调：验签失败: %v", err)
		c.String(http.StatusBadRequest, "fail")
		return
	}

	if notification.TradeStatus != alipay.TradeStatusSuccess && notification.TradeStatus != alipay.TradeStatusFinished {
		log.Printf("订阅支付宝回调：交易状态非成功: %s, 订单: %s", notification.TradeStatus, notification.OutTradeNo)
		c.String(http.StatusOK, "success")
		return
	}

	outTradeNo := notification.OutTradeNo
	if outTradeNo == "" {
		c.String(http.StatusBadRequest, "fail")
		return
	}

	LockOrder(outTradeNo)
	defer UnlockOrder(outTradeNo)

	if err := model.CompleteSubscriptionOrder(outTradeNo, common.GetJsonString(notification)); err != nil {
		log.Printf("订阅支付宝回调：处理失败: %v, 订单: %s", err, outTradeNo)
		c.String(http.StatusInternalServerError, "fail")
		return
	}

	log.Printf("订阅支付宝回调：成功 - 订单: %s, 支付宝交易号: %s", outTradeNo, notification.TradeNo)
	c.String(http.StatusOK, "success")
}
