package setting

var (
	AlipayEnabled    bool
	AlipayAppId      string
	AlipayPrivateKey string  // 应用私钥（RSA2）
	AlipayNotifyUrl  string  // 异步通知地址（可选，留空自动拼接）
	AlipayUnitPrice  float64 // CNY per unit, 0 means use global Price
	AlipayMinTopUp   int     = 1

	// ── 普通公钥模式 ──
	AlipayPublicKey string // 支付宝公钥（纯文本字符串，非证书）

	// ── 公钥证书模式（三个证书内容均为 PEM 文本） ──
	AlipayAppCertPublicKey string // 应用公钥证书 appCertPublicKey_xxxx.crt
	AlipayCertPublicKey    string // 支付宝公钥证书 alipayCertPublicKey_RSA2.crt
	AlipayRootCert         string // 支付宝根证书 alipayRootCert.crt
)

// IsAlipayCertMode 判断是否使用公钥证书模式
func IsAlipayCertMode() bool {
	return AlipayAppCertPublicKey != "" && AlipayCertPublicKey != "" && AlipayRootCert != ""
}

// IsAlipayConfigured 判断支付宝是否已配置（任一模式均可）
func IsAlipayConfigured() bool {
	if AlipayAppId == "" || AlipayPrivateKey == "" {
		return false
	}
	// 证书模式 OR 普通公钥模式
	return IsAlipayCertMode() || AlipayPublicKey != ""
}
