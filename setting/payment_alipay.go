package setting

var (
	AlipayEnabled    bool
	AlipayAppId      string
	AlipayPrivateKey string  // 应用私钥（RSA2）
	AlipayPublicKey  string  // 支付宝公钥（用于验签）
	AlipayNotifyUrl  string  // 异步通知地址（可选，留空自动拼接）
	AlipayUnitPrice  float64 // CNY per unit, 0 means use global Price
	AlipayMinTopUp   int     = 1
)
