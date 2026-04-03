package model

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/pkg/cachex"
	"github.com/samber/hot"
	"gorm.io/gorm"
)

// FeatureItem represents a single item in the structured feature/advantage list
// used by both SubscriptionPlanGroup and TopupTier.
type FeatureItem struct {
	Text  string `json:"text"`
	Icon  string `json:"icon"`  // "check", "x", "info"
	Style string `json:"style"` // "default", "highlight", "disabled"
}

type SubscriptionPlanGroup struct {
	Id        int    `json:"id"`
	Title     string `json:"title" gorm:"type:varchar(128);not null"`
	Subtitle  string `json:"subtitle" gorm:"type:varchar(255);default:''"`
	Tag       string `json:"tag" gorm:"type:varchar(64);default:''"`
	Features  string `json:"features" gorm:"type:text"` // JSON array of FeatureItem
	SortOrder int    `json:"sort_order" gorm:"type:int;default:0"`
	Enabled   bool   `json:"enabled" gorm:"default:true"`
	CreatedAt int64  `json:"created_at" gorm:"bigint"`
	UpdatedAt int64  `json:"updated_at" gorm:"bigint"`
}

func (g *SubscriptionPlanGroup) BeforeCreate(tx *gorm.DB) (err error) {
	g.CreatedAt = time.Now().Unix()
	g.UpdatedAt = g.CreatedAt
	return nil
}

func (g *SubscriptionPlanGroup) BeforeUpdate(tx *gorm.DB) (err error) {
	g.UpdatedAt = time.Now().Unix()
	return nil
}

// --- Cache ---

const subscriptionPlanGroupCacheNamespace = "new-api:subscription_plan_group:v1"

var (
	subscriptionPlanGroupCacheOnce sync.Once
	subscriptionPlanGroupCache     *cachex.HybridCache[SubscriptionPlanGroup]
)

func getSubscriptionPlanGroupCache() *cachex.HybridCache[SubscriptionPlanGroup] {
	subscriptionPlanGroupCacheOnce.Do(func() {
		ttl := subscriptionPlanCacheTTL()
		subscriptionPlanGroupCache = cachex.NewHybridCache[SubscriptionPlanGroup](cachex.HybridCacheConfig[SubscriptionPlanGroup]{
			Namespace: cachex.Namespace(subscriptionPlanGroupCacheNamespace),
			Redis:     common.RDB,
			RedisEnabled: func() bool {
				return common.RedisEnabled && common.RDB != nil
			},
			RedisCodec: cachex.JSONCodec[SubscriptionPlanGroup]{},
			Memory: func() *hot.HotCache[string, SubscriptionPlanGroup] {
				return hot.NewHotCache[string, SubscriptionPlanGroup](hot.LRU, 1000).
					WithTTL(ttl).
					WithJanitor().
					Build()
			},
		})
	})
	return subscriptionPlanGroupCache
}

func InvalidateSubscriptionPlanGroupCache(groupId int) {
	if groupId <= 0 {
		return
	}
	cache := getSubscriptionPlanGroupCache()
	_, _ = cache.DeleteMany([]string{strconv.Itoa(groupId)})
}

// --- CRUD ---

var ErrSubscriptionPlanGroupNotFound = errors.New("subscription plan group not found")

// GetPlanGroupTitle returns the group title for a given plan. Used for display/logging.
func GetPlanGroupTitle(plan *SubscriptionPlan) string {
	if plan == nil {
		return ""
	}
	g, err := GetSubscriptionPlanGroupById(plan.GroupID)
	if err != nil {
		return fmt.Sprintf("Plan#%d", plan.Id)
	}
	return g.Title
}

func GetSubscriptionPlanGroupById(id int) (*SubscriptionPlanGroup, error) {
	cache := getSubscriptionPlanGroupCache()
	key := strconv.Itoa(id)
	if cached, found, err := cache.Get(key); err == nil && found {
		return &cached, nil
	}
	var g SubscriptionPlanGroup
	if err := DB.First(&g, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSubscriptionPlanGroupNotFound
		}
		return nil, err
	}
	_ = cache.SetWithTTL(key, g, subscriptionPlanCacheTTL())
	return &g, nil
}

func GetAllSubscriptionPlanGroups() ([]SubscriptionPlanGroup, error) {
	var groups []SubscriptionPlanGroup
	if err := DB.Order("sort_order desc, id desc").Find(&groups).Error; err != nil {
		return nil, err
	}
	return groups, nil
}

func GetEnabledSubscriptionPlanGroups() ([]SubscriptionPlanGroup, error) {
	var groups []SubscriptionPlanGroup
	if err := DB.Where("enabled = ?", true).Order("sort_order desc, id desc").Find(&groups).Error; err != nil {
		return nil, err
	}
	return groups, nil
}

func CreateSubscriptionPlanGroup(group *SubscriptionPlanGroup) error {
	return DB.Create(group).Error
}

func UpdateSubscriptionPlanGroup(id int, updates map[string]interface{}) error {
	result := DB.Model(&SubscriptionPlanGroup{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	InvalidateSubscriptionPlanGroupCache(id)
	return nil
}

func DeleteSubscriptionPlanGroup(id int) error {
	var count int64
	if err := DB.Model(&SubscriptionPlan{}).Where("group_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("cannot delete group with existing plans")
	}
	result := DB.Delete(&SubscriptionPlanGroup{}, id)
	if result.Error != nil {
		return result.Error
	}
	InvalidateSubscriptionPlanGroupCache(id)
	return nil
}
