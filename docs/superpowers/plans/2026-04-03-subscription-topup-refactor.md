# Subscription & Topup Tier Management Refactor — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refactor subscription management into a plan-group model (group → billing-cycle variants), and extract pay-as-you-go topup tiers from JSON config into a dedicated database table with admin CRUD.

**Architecture:** Two new database tables (`subscription_plan_group`, `topup_tier`), modification of the existing `subscription_plan` table (add `group_id`, remove `title`/`subtitle`). New admin APIs and pages for both modules. Homepage refactored to consume group-structured data. Old `AmountOptions`/`AmountDiscount` JSON config removed.

**Tech Stack:** Go 1.22+, Gin, GORM v2, React 18, Semi Design UI, Bun

**Spec:** `docs/superpowers/specs/2026-04-03-subscription-topup-refactor-design.md`

---

## File Structure

### Backend — New Files
| File | Responsibility |
|------|---------------|
| `model/subscription_plan_group.go` | `SubscriptionPlanGroup` model, CRUD methods, cache |
| `model/topup_tier.go` | `TopupTier` model, CRUD methods, cache |
| `controller/topup_tier.go` | TopupTier admin + public API handlers |

### Backend — Modified Files
| File | Changes |
|------|---------|
| `model/subscription.go` | `SubscriptionPlan`: add `GroupID`, remove `Title`/`Subtitle`; update `SubscriptionPlanInfo` |
| `model/main.go` | Register new tables in `migrateDB()` |
| `controller/subscription.go` | Add group CRUD handlers; refactor `GetPublicSubscriptionPlans` to return group-aggregated data; refactor admin plan handlers |
| `controller/topup.go` | `getPayMoney()` reads discount from `TopupTier` table; `GetTopUpInfo()` removes `amount_options`/`amount_discount` |
| `router/api-router.go` | Register new group + topup tier routes |
| `setting/operation_setting/payment_setting.go` | Remove `AmountOptions` and `AmountDiscount` fields |

### Frontend — New Files
| File | Responsibility |
|------|---------------|
| `web/src/components/common/FeatureListEditor.jsx` | Reusable structured feature list editor (add/remove/reorder items with icon+style) |
| `web/src/components/table/topup-tiers/index.jsx` | TopupTier admin table page container |
| `web/src/components/table/topup-tiers/TopupTiersTable.jsx` | Table component |
| `web/src/components/table/topup-tiers/TopupTiersColumnDefs.jsx` | Column definitions |
| `web/src/components/table/topup-tiers/TopupTiersActions.jsx` | Action buttons |
| `web/src/components/table/topup-tiers/modals/AddEditTopupTierModal.jsx` | Create/edit modal |
| `web/src/hooks/topup-tiers/useTopupTiersData.jsx` | Data fetching hook |

### Frontend — Modified Files
| File | Changes |
|------|---------|
| `web/src/components/table/subscriptions/modals/AddEditSubscriptionModal.jsx` | Split into group modal + plan variant modal |
| `web/src/components/table/subscriptions/index.jsx` | Group-first list view |
| `web/src/components/table/subscriptions/SubscriptionsTable.jsx` | Show groups with expandable variants |
| `web/src/components/table/subscriptions/SubscriptionsColumnDefs.jsx` | Group columns + variant columns |
| `web/src/components/table/subscriptions/SubscriptionsActions.jsx` | Group + variant actions |
| `web/src/hooks/subscriptions/useSubscriptionsData.jsx` | Fetch groups with nested plans |
| `web/src/pages/Home/index.jsx` | `SubscriptionSection`: consume group API; `PayAsYouGoSection`: consume tier API |
| `web/src/components/topup/RechargeCard.jsx` | Read tiers from new API |
| `web/src/components/topup/SubscriptionPlansCard.jsx` | Consume group-structured data |
| `web/src/components/layout/SiderBar.jsx` | Add topup-tiers menu item |
| `web/src/pages/Setting/Payment/SettingsPaymentGateway.jsx` | Remove `AmountOptions`/`AmountDiscount` fields |
| `web/src/App.jsx` (or routing config) | Add route for `/topup-tiers` page |

---

## Task 1: Backend — SubscriptionPlanGroup Model

**Files:**
- Create: `model/subscription_plan_group.go`
- Modify: `model/main.go:258-296`

- [ ] **Step 1: Create the SubscriptionPlanGroup model file**

Create `model/subscription_plan_group.go`:

```go
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
		ttl := subscriptionPlanCacheTTL() // reuse same TTL config
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

func GetSubscriptionPlanGroupById(id int) (*SubscriptionPlanGroup, error) {
	cache := getSubscriptionPlanGroupCache()
	key := strconv.Itoa(id)
	group, ok := cache.Get(key)
	if ok {
		return &group, nil
	}
	var g SubscriptionPlanGroup
	if err := DB.First(&g, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSubscriptionPlanGroupNotFound
		}
		return nil, err
	}
	cache.Set(key, g)
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
	// Check no plans reference this group
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
```

- [ ] **Step 2: Register the new table in migrateDB()**

In `model/main.go`, add `&SubscriptionPlanGroup{}` to the `DB.AutoMigrate(...)` call at line 258. Add it before `&SubscriptionOrder{}` (line 278):

```go
// Inside migrateDB(), in the DB.AutoMigrate() call, add:
&SubscriptionPlanGroup{},
```

- [ ] **Step 3: Verify it compiles**

Run: `cd /Users/wzp/Projects/nebula-claw/nebula-api && go build ./model/...`
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add model/subscription_plan_group.go model/main.go
git commit -m "feat: add SubscriptionPlanGroup model with CRUD and cache"
```

---

## Task 2: Backend — Modify SubscriptionPlan Model

**Files:**
- Modify: `model/subscription.go:144-180`

- [ ] **Step 1: Add GroupID field, remove Title and Subtitle from SubscriptionPlan**

In `model/subscription.go`, replace the `SubscriptionPlan` struct (lines 144-180):

Remove these two fields:
```go
Title    string `json:"title" gorm:"type:varchar(128);not null"`
Subtitle string `json:"subtitle" gorm:"type:varchar(255);default:''"`
```

Add this field after `Id`:
```go
GroupID int `json:"group_id" gorm:"not null;index"`
```

The struct should now start like:
```go
type SubscriptionPlan struct {
	Id      int `json:"id"`
	GroupID int `json:"group_id" gorm:"not null;index"`

	// Display money amount
	PriceAmount float64 `json:"price_amount" gorm:"type:decimal(10,6);not null;default:0"`
	Currency    string  `json:"currency" gorm:"type:varchar(8);not null;default:'USD'"`
	// ... rest unchanged
}
```

- [ ] **Step 2: Update SubscriptionPlanInfo to include group title**

Find `SubscriptionPlanInfo` struct (around line 1138) and update it:

```go
type SubscriptionPlanInfo struct {
	PlanId    int
	PlanTitle string // This now comes from the group
}
```

Update `GetSubscriptionPlanInfoByUserSubscriptionId` (around line 1143) to join with the group table:

```go
func GetSubscriptionPlanInfoByUserSubscriptionId(userSubscriptionId int) (*SubscriptionPlanInfo, error) {
	cache := getSubscriptionPlanInfoCache()
	key := strconv.Itoa(userSubscriptionId)
	info, ok := cache.Get(key)
	if ok {
		return &info, nil
	}

	var result SubscriptionPlanInfo
	err := DB.Table("user_subscriptions").
		Select("subscription_plans.id as plan_id, subscription_plan_groups.title as plan_title").
		Joins("JOIN subscription_plans ON subscription_plans.id = user_subscriptions.plan_id").
		Joins("JOIN subscription_plan_groups ON subscription_plan_groups.id = subscription_plans.group_id").
		Where("user_subscriptions.id = ?", userSubscriptionId).
		Scan(&result).Error
	if err != nil {
		return nil, err
	}
	if result.PlanId == 0 {
		return nil, errors.New("subscription plan info not found")
	}
	cache.Set(key, result)
	return &result, nil
}
```

- [ ] **Step 3: Add helper to get plans by group ID**

Add to `model/subscription.go`:

```go
func GetSubscriptionPlansByGroupID(groupID int) ([]SubscriptionPlan, error) {
	var plans []SubscriptionPlan
	if err := DB.Where("group_id = ?", groupID).Order("sort_order desc, id desc").Find(&plans).Error; err != nil {
		return nil, err
	}
	return plans, nil
}

func GetEnabledSubscriptionPlansByGroupID(groupID int) ([]SubscriptionPlan, error) {
	var plans []SubscriptionPlan
	if err := DB.Where("group_id = ? AND enabled = ?", groupID, true).Order("sort_order desc, id desc").Find(&plans).Error; err != nil {
		return nil, err
	}
	return plans, nil
}
```

- [ ] **Step 4: Handle SQLite migration for the new column**

Check `ensureSubscriptionPlanTableSQLite()` in `model/subscription.go` — this function handles SQLite-specific migration. It will need to handle the `group_id` column. Since there are no existing subscriptions on production, the simplest approach is to let GORM AutoMigrate handle it for all databases. Update `migrateDB()` in `model/main.go`:

Move `&SubscriptionPlan{}` from the conditional SQLite block into the main `DB.AutoMigrate()` call (since there's no production data to worry about for SQLite column issues). If `ensureSubscriptionPlanTableSQLite()` still exists and handles other concerns, keep it but make sure the `group_id` column is included.

- [ ] **Step 5: Verify it compiles**

Run: `cd /Users/wzp/Projects/nebula-claw/nebula-api && go build ./...`
Expected: Compilation errors in `controller/subscription.go` are expected at this stage (references to removed Title/Subtitle fields). These will be fixed in Task 4.

- [ ] **Step 6: Commit**

```bash
git add model/subscription.go model/main.go
git commit -m "feat: add GroupID to SubscriptionPlan, remove Title/Subtitle"
```

---

## Task 3: Backend — TopupTier Model

**Files:**
- Create: `model/topup_tier.go`
- Modify: `model/main.go`

- [ ] **Step 1: Create the TopupTier model file**

Create `model/topup_tier.go`:

```go
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
	Features   string  `json:"features" gorm:"type:text"` // JSON array of FeatureItem (shared type from subscription_plan_group.go)
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
	tier, ok := cache.Get(key)
	if ok {
		return &tier, nil
	}
	var t TopupTier
	if err := DB.First(&t, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTopupTierNotFound
		}
		return nil, err
	}
	cache.Set(key, t)
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
```

- [ ] **Step 2: Register TopupTier in migrateDB()**

In `model/main.go`, add `&TopupTier{}` to the `DB.AutoMigrate()` call.

- [ ] **Step 3: Verify it compiles**

Run: `cd /Users/wzp/Projects/nebula-claw/nebula-api && go build ./model/...`
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add model/topup_tier.go model/main.go
git commit -m "feat: add TopupTier model with CRUD and cache"
```

---

## Task 4: Backend — Subscription Group Admin API + Refactor Plan API

**Files:**
- Modify: `controller/subscription.go`
- Modify: `router/api-router.go`

This is the largest backend task. It refactors the subscription controller to work with groups.

- [ ] **Step 1: Add group admin CRUD handlers**

Add to `controller/subscription.go` (new handler functions):

```go
// --- Group Admin Handlers ---

func AdminListSubscriptionPlanGroups(c *gin.Context) {
	groups, err := model.GetAllSubscriptionPlanGroups()
	if err != nil {
		common.ApiError(c, err)
		return
	}

	// For each group, load its plans
	type GroupWithPlans struct {
		model.SubscriptionPlanGroup
		Plans []model.SubscriptionPlan `json:"plans"`
	}
	result := make([]GroupWithPlans, 0, len(groups))
	for _, g := range groups {
		plans, err := model.GetSubscriptionPlansByGroupID(g.Id)
		if err != nil {
			common.ApiError(c, err)
			return
		}
		result = append(result, GroupWithPlans{
			SubscriptionPlanGroup: g,
			Plans:                 plans,
		})
	}
	common.ApiSuccess(c, result)
}

func AdminCreateSubscriptionPlanGroup(c *gin.Context) {
	var group model.SubscriptionPlanGroup
	if err := common.DecodeJson(c.Request.Body, &group); err != nil {
		common.ApiErrorMsg(c, "invalid request body")
		return
	}
	if group.Title == "" {
		common.ApiErrorMsg(c, "title is required")
		return
	}
	if err := model.CreateSubscriptionPlanGroup(&group); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, group)
}

func AdminUpdateSubscriptionPlanGroup(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiErrorMsg(c, "invalid id")
		return
	}
	var req map[string]interface{}
	if err := common.DecodeJson(c.Request.Body, &req); err != nil {
		common.ApiErrorMsg(c, "invalid request body")
		return
	}

	// Build update map with allowed fields
	updates := make(map[string]interface{})
	allowedFields := []string{"title", "subtitle", "tag", "features", "sort_order", "enabled"}
	for _, f := range allowedFields {
		if v, ok := req[f]; ok {
			updates[f] = v
		}
	}
	if len(updates) == 0 {
		common.ApiErrorMsg(c, "no fields to update")
		return
	}
	if err := model.UpdateSubscriptionPlanGroup(id, updates); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

func AdminDeleteSubscriptionPlanGroup(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiErrorMsg(c, "invalid id")
		return
	}
	if err := model.DeleteSubscriptionPlanGroup(id); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

func AdminUpdateSubscriptionPlanGroupStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiErrorMsg(c, "invalid id")
		return
	}
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := common.DecodeJson(c.Request.Body, &req); err != nil {
		common.ApiErrorMsg(c, "invalid request body")
		return
	}
	if err := model.UpdateSubscriptionPlanGroup(id, map[string]interface{}{"enabled": req.Enabled}); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}
```

- [ ] **Step 2: Refactor admin plan create/update to require group_id**

Update `AdminCreateSubscriptionPlan` in `controller/subscription.go`:
- Remove `title` validation (title is now on the group)
- Add `group_id` validation — verify the group exists
- Remove `Title`/`Subtitle` from the plan fields

Update `AdminUpdateSubscriptionPlan`:
- Remove `title`/`subtitle` from the updatable fields map
- Ensure `group_id` cannot be changed after creation (or allow re-parenting if needed)

- [ ] **Step 3: Refactor GetPublicSubscriptionPlans to return group-aggregated data**

Replace the `PublicSubscriptionPlanDTO` and `GetPublicSubscriptionPlans` in `controller/subscription.go`:

```go
type PublicPlanVariantDTO struct {
	Id                      int     `json:"id"`
	PriceAmount             float64 `json:"price_amount"`
	Currency                string  `json:"currency"`
	DurationUnit            string  `json:"duration_unit"`
	DurationValue           int     `json:"duration_value"`
	CustomSeconds           int64   `json:"custom_seconds"`
	SortOrder               int     `json:"sort_order"`
	TotalAmount             int64   `json:"total_amount"`
	UpgradeGroup            string  `json:"upgrade_group"`
	QuotaResetPeriod        string  `json:"quota_reset_period"`
	QuotaResetCustomSeconds int64   `json:"quota_reset_custom_seconds"`
	MaxPurchasePerUser      int     `json:"max_purchase_per_user"`
}

type PublicPlanGroupDTO struct {
	Id       int                `json:"id"`
	Title    string             `json:"title"`
	Subtitle string             `json:"subtitle"`
	Tag      string             `json:"tag"`
	Features string             `json:"features"` // JSON string — frontend parses it
	Plans    []PublicPlanVariantDTO `json:"plans"`
}

func GetPublicSubscriptionPlans(c *gin.Context) {
	groups, err := model.GetEnabledSubscriptionPlanGroups()
	if err != nil {
		common.ApiError(c, err)
		return
	}

	result := make([]PublicPlanGroupDTO, 0, len(groups))
	for _, g := range groups {
		plans, err := model.GetEnabledSubscriptionPlansByGroupID(g.Id)
		if err != nil {
			common.ApiError(c, err)
			return
		}
		if len(plans) == 0 {
			continue // skip groups with no enabled plans
		}
		variants := make([]PublicPlanVariantDTO, 0, len(plans))
		for _, p := range plans {
			variants = append(variants, PublicPlanVariantDTO{
				Id:                      p.Id,
				PriceAmount:             p.PriceAmount,
				Currency:                p.Currency,
				DurationUnit:            p.DurationUnit,
				DurationValue:           p.DurationValue,
				CustomSeconds:           p.CustomSeconds,
				SortOrder:               p.SortOrder,
				TotalAmount:             p.TotalAmount,
				UpgradeGroup:            p.UpgradeGroup,
				QuotaResetPeriod:        p.QuotaResetPeriod,
				QuotaResetCustomSeconds: p.QuotaResetCustomSeconds,
				MaxPurchasePerUser:      p.MaxPurchasePerUser,
			})
		}
		result = append(result, PublicPlanGroupDTO{
			Id:       g.Id,
			Title:    g.Title,
			Subtitle: g.Subtitle,
			Tag:      g.Tag,
			Features: g.Features,
			Plans:    variants,
		})
	}

	c.Header("Cache-Control", "public, max-age=60")
	common.ApiSuccess(c, result)
}
```

- [ ] **Step 4: Register new routes**

In `router/api-router.go`, update the subscription admin routes section (around line 168-181):

```go
// Group management
subscriptionAdminRoute.GET("/groups", controller.AdminListSubscriptionPlanGroups)
subscriptionAdminRoute.POST("/groups", controller.AdminCreateSubscriptionPlanGroup)
subscriptionAdminRoute.PUT("/groups/:id", controller.AdminUpdateSubscriptionPlanGroup)
subscriptionAdminRoute.DELETE("/groups/:id", controller.AdminDeleteSubscriptionPlanGroup)
subscriptionAdminRoute.PATCH("/groups/:id", controller.AdminUpdateSubscriptionPlanGroupStatus)

// Plan variant management (under group)
subscriptionAdminRoute.POST("/groups/:id/plans", controller.AdminCreateSubscriptionPlan)
// Keep existing plan update/toggle routes as-is (they use plan ID directly)
```

- [ ] **Step 5: Verify it compiles**

Run: `cd /Users/wzp/Projects/nebula-claw/nebula-api && go build ./...`
Expected: no errors

- [ ] **Step 6: Commit**

```bash
git add controller/subscription.go router/api-router.go
git commit -m "feat: add subscription plan group admin API and refactor public plans endpoint"
```

---

## Task 5: Backend — TopupTier Admin API + Refactor Topup Controller

**Files:**
- Create: `controller/topup_tier.go`
- Modify: `controller/topup.go:25-181,206-239`
- Modify: `router/api-router.go`
- Modify: `setting/operation_setting/payment_setting.go`

- [ ] **Step 1: Create TopupTier controller**

Create `controller/topup_tier.go`:

```go
package controller

import (
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

func AdminListTopupTiers(c *gin.Context) {
	tiers, err := model.GetAllTopupTiers()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, tiers)
}

func AdminCreateTopupTier(c *gin.Context) {
	var tier model.TopupTier
	if err := common.DecodeJson(c.Request.Body, &tier); err != nil {
		common.ApiErrorMsg(c, "invalid request body")
		return
	}
	if tier.Title == "" {
		common.ApiErrorMsg(c, "title is required")
		return
	}
	if tier.Amount <= 0 {
		common.ApiErrorMsg(c, "amount must be positive")
		return
	}
	if tier.Discount <= 0 || tier.Discount > 1 {
		tier.Discount = 1.0
	}
	if err := model.CreateTopupTier(&tier); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, tier)
}

func AdminUpdateTopupTier(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiErrorMsg(c, "invalid id")
		return
	}
	var req map[string]interface{}
	if err := common.DecodeJson(c.Request.Body, &req); err != nil {
		common.ApiErrorMsg(c, "invalid request body")
		return
	}

	updates := make(map[string]interface{})
	allowedFields := []string{"title", "subtitle", "tag", "amount", "discount", "bonus_quota", "features", "sort_order", "enabled"}
	for _, f := range allowedFields {
		if v, ok := req[f]; ok {
			updates[f] = v
		}
	}
	if len(updates) == 0 {
		common.ApiErrorMsg(c, "no fields to update")
		return
	}
	if err := model.UpdateTopupTier(id, updates); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

func AdminUpdateTopupTierStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiErrorMsg(c, "invalid id")
		return
	}
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := common.DecodeJson(c.Request.Body, &req); err != nil {
		common.ApiErrorMsg(c, "invalid request body")
		return
	}
	if err := model.UpdateTopupTier(id, map[string]interface{}{"enabled": req.Enabled}); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

func AdminDeleteTopupTier(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiErrorMsg(c, "invalid id")
		return
	}
	if err := model.DeleteTopupTier(id); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

func GetPublicTopupTiers(c *gin.Context) {
	tiers, err := model.GetEnabledTopupTiers()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.Header("Cache-Control", "public, max-age=60")
	common.ApiSuccess(c, tiers)
}
```

- [ ] **Step 2: Modify getPayMoney() to use TopupTier table**

In `controller/topup.go`, replace the discount lookup in `getPayMoney()` (lines 227-233):

**Old:**
```go
discount := 1.0
if ds, ok := operation_setting.GetPaymentSetting().AmountDiscount[int(amount)]; ok {
    if ds > 0 {
        discount = ds
    }
}
```

**New:**
```go
discount := 1.0
if tier, err := model.GetTopupTierByAmount(amount); err == nil && tier != nil {
    if tier.Discount > 0 {
        discount = tier.Discount
    }
}
```

Add `"github.com/QuantumNous/new-api/model"` to imports if not already present.

- [ ] **Step 3: Remove amount_options and discount from GetTopUpInfo()**

In `controller/topup.go`, in the `GetTopUpInfo()` function, remove these two lines from the `data` map (around lines 176-177):

```go
"amount_options":      operation_setting.GetPaymentSetting().AmountOptions,
"discount":            operation_setting.GetPaymentSetting().AmountDiscount,
```

- [ ] **Step 4: Remove AmountOptions and AmountDiscount from PaymentSetting**

In `setting/operation_setting/payment_setting.go`, remove the `AmountOptions` and `AmountDiscount` fields. If these are the only fields, consider whether the struct is still needed for other purposes. If it has no remaining fields, the struct and its registration can be removed entirely. Otherwise, just remove the two fields:

```go
type PaymentSetting struct {
	// Fields removed: AmountOptions, AmountDiscount
	// If other fields exist, keep them. Otherwise, the file can be emptied/removed.
}
```

Since currently these are the only fields, remove the entire file contents and replace with a minimal stub or delete the file. Check if `GetPaymentSetting()` is called anywhere else first — if the only callers were for `AmountOptions`/`AmountDiscount` and those are now removed, the entire file can be deleted.

- [ ] **Step 5: Register topup tier routes**

In `router/api-router.go`, add:

```go
// Topup tier - public (no auth required, like public-plans)
apiRouter.GET("/topup/tiers", controller.GetPublicTopupTiers)

// Topup tier - admin
topupTierAdminRoute := apiRouter.Group("/topup/admin")
topupTierAdminRoute.Use(middleware.AdminAuth())
topupTierAdminRoute.GET("/tiers", controller.AdminListTopupTiers)
topupTierAdminRoute.POST("/tiers", controller.AdminCreateTopupTier)
topupTierAdminRoute.PUT("/tiers/:id", controller.AdminUpdateTopupTier)
topupTierAdminRoute.PATCH("/tiers/:id", controller.AdminUpdateTopupTierStatus)
topupTierAdminRoute.DELETE("/tiers/:id", controller.AdminDeleteTopupTier)
```

- [ ] **Step 6: Handle bonus_quota in recharge flow**

In `model/topup.go`, update the `Recharge()` function to add bonus quota when a tier match is found. After the standard quota calculation, add bonus:

```go
// After calculating quota from amount, check for tier bonus
if tier, err := GetTopupTierByAmount(topUp.Amount); err == nil && tier != nil && tier.BonusQuota > 0 {
    quota += float64(tier.BonusQuota)
}
```

- [ ] **Step 7: Verify it compiles**

Run: `cd /Users/wzp/Projects/nebula-claw/nebula-api && go build ./...`
Expected: no errors

- [ ] **Step 8: Commit**

```bash
git add controller/topup_tier.go controller/topup.go model/topup.go router/api-router.go setting/operation_setting/payment_setting.go
git commit -m "feat: add topup tier admin API, refactor discount to use DB, remove JSON config"
```

---

## Task 6: Frontend — Shared FeatureListEditor Component

**Files:**
- Create: `web/src/components/common/FeatureListEditor.jsx`

- [ ] **Step 1: Create the FeatureListEditor component**

This is a reusable component for editing the structured feature list. Used by both subscription group and topup tier admin modals.

Create `web/src/components/common/FeatureListEditor.jsx`:

```jsx
import React from 'react';
import { Button, Input, Select, Space, Typography } from '@douyinfe/semi-ui';
import { IconPlus, IconDelete, IconArrowUp, IconArrowDown } from '@douyinfe/semi-icons';
import { useTranslation } from 'react-i18next';

const { Text } = Typography;

const iconOptions = [
  { value: 'check', label: '✓ Check' },
  { value: 'x', label: '✗ Cross' },
  { value: 'info', label: 'ℹ Info' },
];

const styleOptions = [
  { value: 'default', label: 'Default' },
  { value: 'highlight', label: 'Highlight' },
  { value: 'disabled', label: 'Disabled' },
];

export default function FeatureListEditor({ value = [], onChange }) {
  const { t } = useTranslation();

  const items = Array.isArray(value) ? value : [];

  const updateItem = (index, field, val) => {
    const newItems = [...items];
    newItems[index] = { ...newItems[index], [field]: val };
    onChange(newItems);
  };

  const addItem = () => {
    onChange([...items, { text: '', icon: 'check', style: 'default' }]);
  };

  const removeItem = (index) => {
    onChange(items.filter((_, i) => i !== index));
  };

  const moveItem = (index, direction) => {
    const newItems = [...items];
    const targetIndex = index + direction;
    if (targetIndex < 0 || targetIndex >= newItems.length) return;
    [newItems[index], newItems[targetIndex]] = [newItems[targetIndex], newItems[index]];
    onChange(newItems);
  };

  return (
    <div>
      <Text strong style={{ marginBottom: 8, display: 'block' }}>
        {t('优势列表')}
      </Text>
      {items.map((item, index) => (
        <Space key={index} style={{ marginBottom: 8, width: '100%' }} align='start'>
          <Input
            value={item.text}
            onChange={(val) => updateItem(index, 'text', val)}
            placeholder={t('功能描述')}
            style={{ width: 200 }}
          />
          <Select
            value={item.icon}
            onChange={(val) => updateItem(index, 'icon', val)}
            optionList={iconOptions}
            style={{ width: 100 }}
          />
          <Select
            value={item.style}
            onChange={(val) => updateItem(index, 'style', val)}
            optionList={styleOptions}
            style={{ width: 110 }}
          />
          <Button
            icon={<IconArrowUp />}
            theme='borderless'
            disabled={index === 0}
            onClick={() => moveItem(index, -1)}
          />
          <Button
            icon={<IconArrowDown />}
            theme='borderless'
            disabled={index === items.length - 1}
            onClick={() => moveItem(index, 1)}
          />
          <Button
            icon={<IconDelete />}
            theme='borderless'
            type='danger'
            onClick={() => removeItem(index)}
          />
        </Space>
      ))}
      <Button icon={<IconPlus />} theme='light' onClick={addItem} style={{ marginTop: 4 }}>
        {t('添加优势')}
      </Button>
    </div>
  );
}
```

- [ ] **Step 2: Commit**

```bash
cd /Users/wzp/Projects/nebula-claw/nebula-api/web && git add src/components/common/FeatureListEditor.jsx
git commit -m "feat: add reusable FeatureListEditor component for admin modals"
```

---

## Task 7: Frontend — Refactor Subscription Admin Management

**Files:**
- Modify: `web/src/components/table/subscriptions/index.jsx`
- Modify: `web/src/components/table/subscriptions/SubscriptionsTable.jsx`
- Modify: `web/src/components/table/subscriptions/SubscriptionsColumnDefs.jsx`
- Modify: `web/src/components/table/subscriptions/SubscriptionsActions.jsx`
- Modify: `web/src/components/table/subscriptions/modals/AddEditSubscriptionModal.jsx`
- Modify: `web/src/hooks/subscriptions/useSubscriptionsData.jsx`

This task refactors the subscription admin UI from a flat plan list to a group-first view.

- [ ] **Step 1: Update useSubscriptionsData hook to fetch groups**

Change the API endpoint from `/api/subscription/admin/plans` to `/api/subscription/admin/groups`. The response now contains groups with nested `plans` arrays.

- [ ] **Step 2: Refactor SubscriptionsTable to show groups**

The table should show groups as primary rows. Each group row is expandable to show its plan variants. Use Semi Design's `Table` `expandedRowRender` or a nested approach.

- [ ] **Step 3: Update SubscriptionsColumnDefs for group columns**

Group columns: Title, Tag, Sort Order, Enabled, Actions (edit/delete).
Variant columns (in expanded view): Duration, Price, Quota, Payment IDs, Enabled, Actions.

- [ ] **Step 4: Split AddEditSubscriptionModal into two modals**

Create `AddEditGroupModal.jsx` for group fields: title, subtitle, tag, features (using FeatureListEditor), sort_order, enabled.

Refactor `AddEditSubscriptionModal.jsx` for plan variant fields: group_id (pre-filled when creating from within a group), duration, price, quota, payment IDs, upgrade_group, enabled.

- [ ] **Step 5: Update SubscriptionsActions**

Add "Create Group" button. Within group rows, add "Add Plan" action.

- [ ] **Step 6: Verify the admin page works**

Run: `cd /Users/wzp/Projects/nebula-claw/nebula-api/web && bun run dev`
Navigate to admin → subscription management. Verify:
- Groups are listed
- Expanding a group shows its plan variants
- Creating a group works (with feature list editor)
- Adding a plan variant to a group works
- Editing/toggling/deleting works

- [ ] **Step 7: Commit**

```bash
cd /Users/wzp/Projects/nebula-claw/nebula-api && git add web/src/components/table/subscriptions/ web/src/hooks/subscriptions/
git commit -m "feat: refactor subscription admin UI to group-first management"
```

---

## Task 8: Frontend — TopupTier Admin Management Page

**Files:**
- Create: `web/src/components/table/topup-tiers/index.jsx`
- Create: `web/src/components/table/topup-tiers/TopupTiersTable.jsx`
- Create: `web/src/components/table/topup-tiers/TopupTiersColumnDefs.jsx`
- Create: `web/src/components/table/topup-tiers/TopupTiersActions.jsx`
- Create: `web/src/components/table/topup-tiers/modals/AddEditTopupTierModal.jsx`
- Create: `web/src/hooks/topup-tiers/useTopupTiersData.jsx`
- Modify: `web/src/components/layout/SiderBar.jsx`
- Modify: App routing config

Follow the same patterns as the existing subscription admin table. The structure mirrors `web/src/components/table/subscriptions/`.

- [ ] **Step 1: Create useTopupTiersData hook**

Fetches from `GET /api/topup/admin/tiers`. Manages state, loading, pagination, CRUD operations via the admin API.

- [ ] **Step 2: Create TopupTiersColumnDefs**

Columns: Title, Amount, Discount (formatted as percentage), Tag, Bonus Quota, Sort Order, Enabled, Actions.

- [ ] **Step 3: Create AddEditTopupTierModal**

Form fields: title, subtitle, tag, amount, discount (0-1 slider or input), bonus_quota, features (FeatureListEditor), sort_order, enabled.

API: POST `/api/topup/admin/tiers` (create) or PUT `/api/topup/admin/tiers/:id` (update).

- [ ] **Step 4: Create TopupTiersActions, TopupTiersTable, and index.jsx**

Follow the same pattern as subscription management components.

- [ ] **Step 5: Add sidebar menu item and route**

In `web/src/components/layout/SiderBar.jsx`, add to the `adminItems` array (around line 161, after the subscription item):

```jsx
{
  text: t('充值档位'),
  itemKey: 'topup-tiers',
  to: '/topup-tiers',
  className: isAdmin() ? '' : 'tableHiddle',
},
```

Add the route mapping in `routerMap`:
```jsx
'topup-tiers': '/console/topup-tiers',
```

Add the route in the app routing config (likely `App.jsx` or a routes file) to render the TopupTiers component at `/console/topup-tiers`.

- [ ] **Step 6: Add to sidebar module visibility config**

In `web/src/pages/Setting/Operation/SettingsSidebarModulesAdmin.jsx`, add `topup-tiers` to the admin modules array so it can be toggled on/off.

- [ ] **Step 7: Verify the admin page works**

Run: `cd /Users/wzp/Projects/nebula-claw/nebula-api/web && bun run dev`
Navigate to admin → topup tier management. Verify CRUD operations.

- [ ] **Step 8: Commit**

```bash
cd /Users/wzp/Projects/nebula-claw/nebula-api && git add web/src/components/table/topup-tiers/ web/src/hooks/topup-tiers/ web/src/components/layout/SiderBar.jsx
git commit -m "feat: add topup tier admin management page"
```

---

## Task 9: Frontend — Refactor Homepage Subscription Section

**Files:**
- Modify: `web/src/pages/Home/index.jsx` (SubscriptionSection, around lines 250-477)

- [ ] **Step 1: Update data fetching**

The `SubscriptionSection` currently receives a flat `plans` array. Update it to receive the group-structured response from `GET /api/subscription/public-plans`.

- [ ] **Step 2: Refactor year/month toggle logic**

Currently (lines 259-268) it filters by `duration_unit` and groups by `title` using a Map. With the new structure:
- Each group has a `plans` array with variants
- The toggle switches which variant to display per group (monthly vs yearly)
- Savings calculation uses variants within the same group (no need to match by title)

```jsx
const SubscriptionSection = ({ t, groups = [], loading = false, navigate }) => {
  const [isYearly, setIsYearly] = useState(false);

  if (!loading && groups.length === 0) return null;

  const { symbol, rate } = getCurrencyConfig();

  // Check if any group has yearly variants
  const hasYearly = groups.some(g =>
    g.plans.some(p => p.duration_unit === 'year' && p.duration_value === 1)
  );

  // For each group, pick the variant matching the current billing toggle
  const displayGroups = groups.map(group => {
    const monthly = group.plans.find(p => p.duration_unit === 'month' && p.duration_value === 1);
    const yearly = group.plans.find(p => p.duration_unit === 'year' && p.duration_value === 1);
    const activePlan = isYearly && yearly ? yearly : monthly;
    const savingsPercent = monthly && yearly
      ? Math.round((1 - yearly.price_amount / (monthly.price_amount * 12)) * 100)
      : null;
    return { ...group, activePlan, monthly, yearly, savingsPercent };
  }).filter(g => g.activePlan); // Only show groups that have a plan for the current period
  // ...
};
```

- [ ] **Step 3: Render feature list from group data**

Replace the hardcoded `benefits` array (lines 386-391) with the group's `features` JSON:

```jsx
const features = group.features ? JSON.parse(group.features) : [];
// Render with appropriate icons (Check, X, Info from lucide-react)
```

- [ ] **Step 4: Render tag from group data**

Use `group.tag` for the card badge instead of the hardcoded `isPopular` logic.

- [ ] **Step 5: Verify the homepage**

Run dev server, navigate to homepage, verify:
- Subscription cards render with group title, subtitle, tag, features
- Year/month toggle works correctly
- Savings badge displays properly
- Subscribe button navigates correctly

- [ ] **Step 6: Commit**

```bash
cd /Users/wzp/Projects/nebula-claw/nebula-api && git add web/src/pages/Home/index.jsx
git commit -m "feat: refactor homepage subscription section to use plan groups"
```

---

## Task 10: Frontend — Refactor Homepage Pay-As-You-Go Section

**Files:**
- Modify: `web/src/pages/Home/index.jsx` (PayAsYouGoSection, around lines 480-540+)

- [ ] **Step 1: Fetch tiers from new API**

Replace reading from `statusState.status.topup_amount_options` / `topup_amount_discount` with a direct API call to `GET /api/topup/tiers`. This can be done in the parent component's data fetching or via a `useEffect` in `PayAsYouGoSection`.

- [ ] **Step 2: Render tier cards with full data**

Each tier now has title, subtitle, tag, features, discount, bonus_quota. Update the card rendering:

```jsx
const PayAsYouGoSection = ({ t, tiers = [], navigate }) => {
  if (tiers.length === 0) return null;

  return (
    <section id='pricing-payg' className='...'>
      {/* ... header unchanged ... */}
      <div className={cn('grid gap-4 md:gap-6 px-2', gridCols)}>
        {tiers.map((tier) => {
          const features = tier.features ? JSON.parse(tier.features) : [];
          const actualPrice = tier.discount < 1
            ? tier.amount * tier.discount
            : tier.amount;
          // Render: title, subtitle, tag badge, price, discount label, features list, bonus_quota
        })}
      </div>
    </section>
  );
};
```

- [ ] **Step 3: Verify the homepage**

Run dev server, verify topup section renders tiers with titles, tags, features, discounts.

- [ ] **Step 4: Commit**

```bash
cd /Users/wzp/Projects/nebula-claw/nebula-api && git add web/src/pages/Home/index.jsx
git commit -m "feat: refactor homepage pay-as-you-go section to use topup tiers API"
```

---

## Task 11: Frontend — Refactor Topup Page Components

**Files:**
- Modify: `web/src/components/topup/RechargeCard.jsx`
- Modify: `web/src/components/topup/SubscriptionPlansCard.jsx`
- Modify: `web/src/components/topup/index.jsx`

- [ ] **Step 1: Update RechargeCard to fetch from topup tiers API**

Replace `presetAmounts` (sourced from topup info's `amount_options`) with tiers from `GET /api/topup/tiers`. The tier cards should show title, tag, and discount information.

- [ ] **Step 2: Update SubscriptionPlansCard to use group-structured data**

Update the component to consume `GET /api/subscription/public-plans` (group-structured) instead of the flat plan list. The year/month toggle should work within groups.

- [ ] **Step 3: Update topup/index.jsx data fetching**

Adjust the parent component to fetch tiers data and pass it to children.

- [ ] **Step 4: Verify the topup page**

Run dev server, navigate to `/console/topup`. Verify:
- Recharge card shows tiers correctly
- Subscription plans card shows groups with billing toggle
- Payment flow still works (plan_id reference unchanged)

- [ ] **Step 5: Commit**

```bash
cd /Users/wzp/Projects/nebula-claw/nebula-api && git add web/src/components/topup/
git commit -m "feat: refactor topup page to use tiers API and group-structured plans"
```

---

## Task 12: Frontend — Remove AmountOptions/AmountDiscount from Settings

**Files:**
- Modify: `web/src/pages/Setting/Payment/SettingsPaymentGateway.jsx`

- [ ] **Step 1: Remove the AmountOptions and AmountDiscount form fields**

In `SettingsPaymentGateway.jsx`, find and remove:
- The `AmountOptions` JSON textarea field (around lines 219-229)
- The `AmountDiscount` JSON textarea field (around lines 237-248)
- Their corresponding state/save logic

Keep all other payment settings (Price, MinTopUp, TopupGroupRatio, gateway configs).

- [ ] **Step 2: Verify settings page**

Run dev server, navigate to admin → settings → payment. Verify:
- AmountOptions and AmountDiscount fields are gone
- Other settings still save correctly

- [ ] **Step 3: Commit**

```bash
cd /Users/wzp/Projects/nebula-claw/nebula-api && git add web/src/pages/Setting/Payment/SettingsPaymentGateway.jsx
git commit -m "chore: remove AmountOptions/AmountDiscount from payment settings UI"
```

---

## Task 13: Final Verification & Cleanup

- [ ] **Step 1: Full backend build**

Run: `cd /Users/wzp/Projects/nebula-claw/nebula-api && go build ./...`
Expected: no errors

- [ ] **Step 2: Frontend build**

Run: `cd /Users/wzp/Projects/nebula-claw/nebula-api/web && bun run build`
Expected: no errors

- [ ] **Step 3: Run backend tests**

Run: `cd /Users/wzp/Projects/nebula-claw/nebula-api && go test ./...`
Expected: all existing tests pass

- [ ] **Step 4: Manual smoke test**

Start the full stack with `make all` or `go run main.go` + `bun run dev`:
1. Admin: create a subscription plan group with features → add monthly and yearly variants
2. Admin: create topup tiers with tags, discounts, features
3. Homepage: verify subscription section shows groups with year/month toggle
4. Homepage: verify topup section shows tiers with all info
5. Topup page: verify payment flow works end-to-end

- [ ] **Step 5: Final commit if any cleanup needed**

```bash
git add -A && git commit -m "chore: final cleanup for subscription & topup tier refactor"
```
