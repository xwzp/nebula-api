package operation_setting

import "github.com/QuantumNous/new-api/setting/config"

// PaymentSetting holds payment-related configuration.
// AmountOptions and AmountDiscount have been moved to the topup_tier database table.
type PaymentSetting struct{}

var paymentSetting = PaymentSetting{}

func init() {
	config.GlobalConfig.Register("payment_setting", &paymentSetting)
}

func GetPaymentSetting() *PaymentSetting {
	return &paymentSetting
}
