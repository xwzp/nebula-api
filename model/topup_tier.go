package model

import (
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/pkg/cachex"
	"github.com/samber/hot"
	"gorm.io/gorm"
)

type TopupTier struct {
	Id         int     `json:"id"`
	Title      string  `json:"title" gorm:"type:varchar(128);not null"`
	Subtitle   string  `json:"subtitle" gorm:"type:varchar(255);default:''"`
	Tag        string  `json:"tag" gorm:"type:varchar(64);default:''"`
	Amount     int64   `json:"amount" gorm:"type:bigint;not null"`
	Discount   float64 `json:"discount" gorm:"type:decimal(10,4);not null;default:1.0"`
	BonusQuota int64   `json:"bonus_quota" gorm:"type:bigint;not null;default:0"`
	Features   string  `json:"features" gorm:"type:text"` // JSON array of FeatureItem
	SortOrder  int     `json:"sort_order" gorm:"type:int;default:0"`
	Enabled    bool    `json:"enabled" gorm:"default:true"`
	CreatedAt  int64   `json:"created_at" gorm:"bigint"`
	UpdatedAt  int64   `json:"updated_at" gorm:"bigint"`
}

func (t *TopupTier) BeforeCreate(tx *gorm.DB) (err error) {
	t.CreatedAt = time.Now().Unix()
	t.UpdatedAt = t.CreatedAt
	return nil
}

func (t *TopupTier) BeforeUpdate(tx *gorm.DB) (err error) {
	t.UpdatedAt = time.Now().Unix()
	return nil
}

// --- Cache ---

const topupTierCacheNamespace = "new-api:topup_tier:v1"

var (
	topupTierCacheOnce sync.Once
	topupTierCache     *cachex.HybridCache[TopupTier]
)

func getTopupTierCache() *cachex.HybridCache[TopupTier] {
	topupTierCacheOnce.Do(func() {
		ttl := 5 * time.Minute
		topupTierCache = cachex.NewHybridCache[TopupTier](cachex.HybridCacheConfig[TopupTier]{
			Namespace: cachex.Namespace(topupTierCacheNamespace),
			Redis:     common.RDB,
			RedisEnabled: func() bool {
				return common.RedisEnabled && common.RDB != nil
			},
			RedisCodec: cachex.JSONCodec[TopupTier]{},
			Memory: func() *hot.HotCache[string, TopupTier] {
				return hot.NewHotCache[string, TopupTier](hot.LRU, 1000).
					WithTTL(ttl).
					WithJanitor().
					Build()
			},
		})
	})
	return topupTierCache
}

func InvalidateTopupTierCache(tierId int) {
	if tierId <= 0 {
		return
	}
	cache := getTopupTierCache()
	_, _ = cache.DeleteMany([]string{strconv.Itoa(tierId)})
}

// --- CRUD ---

var ErrTopupTierNotFound = errors.New("topup tier not found")

func GetTopupTierById(id int) (*TopupTier, error) {
	cache := getTopupTierCache()
	key := strconv.Itoa(id)
	if cached, found, err := cache.Get(key); err == nil && found {
		return &cached, nil
	}
	var t TopupTier
	if err := DB.First(&t, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTopupTierNotFound
		}
		return nil, err
	}
	_ = cache.SetWithTTL(key, t, 5*time.Minute)
	return &t, nil
}

func GetAllTopupTiers() ([]TopupTier, error) {
	var tiers []TopupTier
	if err := DB.Order("sort_order desc, id desc").Find(&tiers).Error; err != nil {
		return nil, err
	}
	return tiers, nil
}

func GetEnabledTopupTiers() ([]TopupTier, error) {
	var tiers []TopupTier
	if err := DB.Where("enabled = ?", true).Order("sort_order desc, id desc").Find(&tiers).Error; err != nil {
		return nil, err
	}
	return tiers, nil
}

func GetTopupTierByAmount(amount int64) (*TopupTier, error) {
	var t TopupTier
	if err := DB.Where("amount = ? AND enabled = ?", amount, true).First(&t).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // not found is not an error — means no tier-specific discount
		}
		return nil, err
	}
	return &t, nil
}

func CreateTopupTier(tier *TopupTier) error {
	return DB.Create(tier).Error
}

func UpdateTopupTier(id int, updates map[string]interface{}) error {
	result := DB.Model(&TopupTier{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	InvalidateTopupTierCache(id)
	return nil
}

func DeleteTopupTier(id int) error {
	result := DB.Delete(&TopupTier{}, id)
	if result.Error != nil {
		return result.Error
	}
	InvalidateTopupTierCache(id)
	return nil
}
