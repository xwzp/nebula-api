package setting

var (
	WechatPayEnabled       bool
	WechatPayMchId         string
	WechatPayMchApiV3Key   string
	WechatPayMchSerialNo   string
	WechatPayMchPrivateKey string
	WechatPayAppId         string
	WechatPayNotifyUrl     string
	WechatPayUnitPrice     float64 // CNY per unit, 0 means use global Price
	WechatPayMinTopUp      int     = 1
)
