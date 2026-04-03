package model

import (
	"strings"

	"github.com/QuantumNous/new-api/constant"
)

// 简化的供应商映射规则
var defaultVendorRules = map[string]string{
	"gpt":      "OpenAI",
	"dall-e":   "OpenAI",
	"whisper":  "OpenAI",
	"o1":       "OpenAI",
	"o3":       "OpenAI",
	"claude":   "Anthropic",
	"gemini":   "Google",
	"moonshot": "Moonshot",
	"kimi":     "Moonshot",
	"chatglm":  "智谱",
	"glm-":     "智谱",
	"qwen":     "阿里巴巴",
	"deepseek": "DeepSeek",
	"abab":     "MiniMax",
	"minimax":  "MiniMax",
	"ernie":    "百度",
	"spark":    "讯飞",
	"hunyuan":  "腾讯",
	"command":  "Cohere",
	"@cf/":     "Cloudflare",
	"360":      "360",
	"yi":       "零一万物",
	"jina":     "Jina",
	"mistral":  "Mistral",
	"grok":     "xAI",
	"llama":    "Meta",
	"doubao":   "字节跳动",
	"kling":    "快手",
	"jimeng":   "即梦",
	"vidu":     "Vidu",
}

// 供应商默认图标映射
var defaultVendorIcons = map[string]string{
	"OpenAI":     "OpenAI",
	"Anthropic":  "Claude.Color",
	"Google":     "Gemini.Color",
	"Moonshot":   "Moonshot",
	"智谱":         "Zhipu.Color",
	"阿里巴巴":       "Qwen.Color",
	"DeepSeek":   "DeepSeek.Color",
	"MiniMax":    "Minimax.Color",
	"百度":         "Wenxin.Color",
	"讯飞":         "Spark.Color",
	"腾讯":         "Hunyuan.Color",
	"Cohere":     "Cohere.Color",
	"Cloudflare": "Cloudflare.Color",
	"360":        "Ai360.Color",
	"零一万物":       "Yi.Color",
	"Jina":       "Jina",
	"Mistral":    "Mistral.Color",
	"xAI":        "XAI",
	"Meta":       "Ollama",
	"字节跳动":       "Doubao.Color",
	"快手":         "Kling.Color",
	"即梦":         "Jimeng.Color",
	"Vidu":       "Vidu",
	"微软":         "AzureAI",
	"Microsoft":  "AzureAI",
	"Azure":      "AzureAI",
}

// channelTypeVendorMap 渠道类型到供应商名称的映射（仅供应商专属渠道，不含聚合器/代理类型）
var channelTypeVendorMap = map[int]string{
	constant.ChannelTypeAnthropic:   "Anthropic",
	constant.ChannelTypeClaudeOAuth: "Anthropic",
	constant.ChannelTypeBaidu:       "百度",
	constant.ChannelTypeBaiduV2:     "百度",
	constant.ChannelTypeZhipu:       "智谱",
	constant.ChannelTypeZhipu_v4:    "智谱",
	constant.ChannelTypeAli:         "阿里巴巴",
	constant.ChannelTypeXunfei:      "讯飞",
	constant.ChannelType360:         "360",
	constant.ChannelTypeTencent:     "腾讯",
	constant.ChannelTypeGemini:      "Google",
	constant.ChannelTypePaLM:        "Google",
	constant.ChannelTypeMoonshot:    "Moonshot",
	constant.ChannelTypeLingYiWanWu: "零一万物",
	constant.ChannelTypeCohere:      "Cohere",
	constant.ChannelTypeMiniMax:     "MiniMax",
	constant.ChannelTypeJina:        "Jina",
	constant.ChannelCloudflare:      "Cloudflare",
	constant.ChannelTypeMistral:     "Mistral",
	constant.ChannelTypeDeepSeek:    "DeepSeek",
	constant.ChannelTypeVolcEngine:  "字节跳动",
	constant.ChannelTypeDoubaoVideo: "字节跳动",
	constant.ChannelTypeXai:         "xAI",
	constant.ChannelTypeKling:       "快手",
	constant.ChannelTypeJimeng:      "即梦",
	constant.ChannelTypeVidu:        "Vidu",
	constant.ChannelTypeSora:        "OpenAI",
	constant.ChannelTypeCodex:       "OpenAI",
}

// initDefaultVendorMapping 默认供应商映射：先按模型名匹配，再按渠道类型回退
func initDefaultVendorMapping(metaMap map[string]*Model, vendorMap map[int]*Vendor, enableAbilities []AbilityWithChannel) {
	// 阶段一：基于模型名称的模式匹配
	for _, ability := range enableAbilities {
		modelName := ability.Model
		if _, exists := metaMap[modelName]; exists {
			continue
		}

		vendorID := 0
		modelLower := strings.ToLower(modelName)
		for pattern, vendorName := range defaultVendorRules {
			if strings.Contains(modelLower, pattern) {
				vendorID = getOrCreateVendor(vendorName, vendorMap)
				break
			}
		}

		metaMap[modelName] = &Model{
			ModelName: modelName,
			VendorID:  vendorID,
			Status:    1,
			NameRule:  NameRuleExact,
		}
	}

	// 阶段二：对仍无供应商的模型，根据渠道类型回退分配
	for _, ability := range enableAbilities {
		meta, exists := metaMap[ability.Model]
		if !exists || meta.VendorID != 0 {
			continue
		}
		if vendorName, ok := channelTypeVendorMap[ability.ChannelType]; ok {
			meta.VendorID = getOrCreateVendor(vendorName, vendorMap)
		}
	}
}

// 查找或创建供应商
func getOrCreateVendor(vendorName string, vendorMap map[int]*Vendor) int {
	// 查找现有供应商
	for id, vendor := range vendorMap {
		if vendor.Name == vendorName {
			return id
		}
	}

	// 创建新供应商
	newVendor := &Vendor{
		Name:   vendorName,
		Status: 1,
		Icon:   getDefaultVendorIcon(vendorName),
	}

	if err := newVendor.Insert(); err != nil {
		return 0
	}

	vendorMap[newVendor.Id] = newVendor
	return newVendor.Id
}

// 获取供应商默认图标
func getDefaultVendorIcon(vendorName string) string {
	if icon, exists := defaultVendorIcons[vendorName]; exists {
		return icon
	}
	return ""
}
