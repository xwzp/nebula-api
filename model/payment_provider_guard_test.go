package model

import (
	"errors"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/require"
)

func TestTopUpProviderGuardRejectsCrossGatewayStatusUpdate(t *testing.T) {
	truncateTables(t)

	topUp := &TopUp{
		UserId:          1,
		Amount:          100,
		Money:           100,
		TradeNo:         "WXPAY-1-1000-abcd",
		PaymentMethod:   "wechat",
		PaymentProvider: PaymentProviderWechat,
		Status:          common.TopUpStatusPending,
	}
	require.NoError(t, DB.Create(topUp).Error)

	err := UpdatePendingTopUpStatus(topUp.TradeNo, PaymentProviderStripe, common.TopUpStatusExpired)
	require.True(t, errors.Is(err, ErrPaymentProviderMismatch))
}

func TestTopUpProviderGuardAllowsOnlyMatchingLegacyOrders(t *testing.T) {
	require.True(t, IsTopUpPaymentProvider(&TopUp{
		TradeNo:       "WXPAY-1-1000-abcd",
		PaymentMethod: "wechat",
	}, PaymentProviderWechat))

	require.False(t, IsTopUpPaymentProvider(&TopUp{
		TradeNo:       "WXPAY-1-1000-abcd",
		PaymentMethod: "wechat",
	}, PaymentProviderStripe))

	require.True(t, IsTopUpPaymentProvider(&TopUp{
		TradeNo:       "USR1NOabcdef1000",
		PaymentMethod: "alipay",
	}, PaymentProviderEpay))
}

func TestCompleteSubscriptionOrderRejectsCrossGatewayCallback(t *testing.T) {
	truncateTables(t)

	order := &SubscriptionOrder{
		UserId:          2,
		PlanId:          1,
		Money:           10,
		PeriodType:      PeriodMonthly,
		TradeNo:         "sub_ref_guard",
		PaymentMethod:   "stripe",
		PaymentProvider: PaymentProviderStripe,
		Status:          common.TopUpStatusPending,
	}
	require.NoError(t, DB.Create(order).Error)

	err := CompleteSubscriptionOrder(order.TradeNo, `{"provider":"epay"}`, PaymentProviderEpay, "alipay")
	require.True(t, errors.Is(err, ErrPaymentProviderMismatch))
}

func TestSubscriptionProviderGuardAllowsOnlyMatchingLegacyOrders(t *testing.T) {
	require.True(t, subscriptionPaymentProviderMatches(&SubscriptionOrder{
		TradeNo:       "SUBWX-1-1000-abcd",
		PaymentMethod: "wechat",
	}, PaymentProviderWechat))

	require.False(t, subscriptionPaymentProviderMatches(&SubscriptionOrder{
		TradeNo:       "SUBWX-1-1000-abcd",
		PaymentMethod: "wechat",
	}, PaymentProviderStripe))

	require.True(t, subscriptionPaymentProviderMatches(&SubscriptionOrder{
		TradeNo:       "SUBUSR1NOabcdef1000",
		PaymentMethod: "alipay",
	}, PaymentProviderEpay))
}
