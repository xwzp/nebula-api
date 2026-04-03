package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type apiResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func setupSubscriptionTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	gin.SetMode(gin.TestMode)
	common.UsingSQLite = true
	common.UsingMySQL = false
	common.UsingPostgreSQL = false
	common.RedisEnabled = false

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	model.DB = db
	model.LOG_DB = db

	if err := db.AutoMigrate(
		&model.SubscriptionPlanGroup{},
		&model.SubscriptionPlan{},
		&model.TopupTier{},
	); err != nil {
		t.Fatalf("failed to migrate tables: %v", err)
	}

	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	return db
}

func seedGroup(t *testing.T, db *gorm.DB, title string) *model.SubscriptionPlanGroup {
	t.Helper()
	group := &model.SubscriptionPlanGroup{
		Title:    title,
		Subtitle: title + " subtitle",
		Tag:      "popular",
		Features: `[{"text":"Feature 1","icon":"check","style":"default"}]`,
		Enabled:  true,
	}
	if err := db.Create(group).Error; err != nil {
		t.Fatalf("failed to create group: %v", err)
	}
	return group
}

func seedPlan(t *testing.T, db *gorm.DB, groupID int, priceAmount float64, durationUnit string) *model.SubscriptionPlan {
	t.Helper()
	plan := &model.SubscriptionPlan{
		GroupID:       groupID,
		PriceAmount:   priceAmount,
		Currency:      "USD",
		DurationUnit:  durationUnit,
		DurationValue: 1,
		TotalAmount:   500000,
		Enabled:       true,
	}
	if err := db.Create(plan).Error; err != nil {
		t.Fatalf("failed to create plan: %v", err)
	}
	return plan
}

func seedTopupTier(t *testing.T, db *gorm.DB, title string, amount int64, discount float64) *model.TopupTier {
	t.Helper()
	tier := &model.TopupTier{
		Title:    title,
		Amount:   amount,
		Discount: discount,
		Features: `[{"text":"Bonus credits","icon":"check","style":"highlight"}]`,
		Enabled:  true,
	}
	if err := db.Create(tier).Error; err != nil {
		t.Fatalf("failed to create topup tier: %v", err)
	}
	return tier
}

func newCtx(t *testing.T, method, target string, body any) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	var bodyReader *strings.Reader
	if body != nil {
		payload, err := common.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal body: %v", err)
		}
		bodyReader = strings.NewReader(string(payload))
	} else {
		bodyReader = strings.NewReader("")
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(method, target, bodyReader)
	if body != nil {
		ctx.Request.Header.Set("Content-Type", "application/json")
	}
	ctx.Set("id", 1)
	ctx.Set("role", common.RoleAdminUser)
	return ctx, recorder
}

func decodeResp(t *testing.T, recorder *httptest.ResponseRecorder) apiResponse {
	t.Helper()
	var resp apiResponse
	if err := common.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v\nbody: %s", err, recorder.Body.String())
	}
	return resp
}

// ---- Subscription Plan Group Tests ----

func TestAdminCreateSubscriptionPlanGroup(t *testing.T) {
	setupSubscriptionTestDB(t)

	body := map[string]interface{}{
		"title":    "Pro Plan",
		"subtitle": "Best value",
		"tag":      "推荐",
		"features": `[{"text":"Unlimited access","icon":"check","style":"highlight"}]`,
	}
	ctx, recorder := newCtx(t, http.MethodPost, "/api/subscription/admin/groups", body)
	AdminCreateSubscriptionPlanGroup(ctx)

	resp := decodeResp(t, recorder)
	if !resp.Success {
		t.Fatalf("expected success, got: %s", resp.Message)
	}

	var created model.SubscriptionPlanGroup
	if err := common.Unmarshal(resp.Data, &created); err != nil {
		t.Fatalf("failed to decode created group: %v", err)
	}
	if created.Title != "Pro Plan" {
		t.Errorf("expected title 'Pro Plan', got %q", created.Title)
	}
	if created.Id <= 0 {
		t.Errorf("expected positive ID, got %d", created.Id)
	}
}

func TestAdminCreateGroupEmptyTitleFails(t *testing.T) {
	setupSubscriptionTestDB(t)

	body := map[string]interface{}{
		"title": "  ",
	}
	ctx, recorder := newCtx(t, http.MethodPost, "/api/subscription/admin/groups", body)
	AdminCreateSubscriptionPlanGroup(ctx)

	resp := decodeResp(t, recorder)
	if resp.Success {
		t.Fatalf("expected failure for empty title")
	}
}

func TestAdminListGroupsWithPlans(t *testing.T) {
	db := setupSubscriptionTestDB(t)
	g := seedGroup(t, db, "Basic")
	seedPlan(t, db, g.Id, 9.99, "month")
	seedPlan(t, db, g.Id, 99.99, "year")

	ctx, recorder := newCtx(t, http.MethodGet, "/api/subscription/admin/groups", nil)
	AdminListSubscriptionPlanGroups(ctx)

	resp := decodeResp(t, recorder)
	if !resp.Success {
		t.Fatalf("expected success: %s", resp.Message)
	}

	var groups []GroupWithPlans
	if err := common.Unmarshal(resp.Data, &groups); err != nil {
		t.Fatalf("failed to decode groups: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if len(groups[0].Plans) != 2 {
		t.Errorf("expected 2 plans, got %d", len(groups[0].Plans))
	}
}

func TestAdminUpdateSubscriptionPlanGroup(t *testing.T) {
	db := setupSubscriptionTestDB(t)
	g := seedGroup(t, db, "Old Title")

	body := map[string]interface{}{
		"title":    "New Title",
		"subtitle": "Updated subtitle",
	}
	ctx, recorder := newCtx(t, http.MethodPut, "/api/subscription/admin/groups/"+strconv.Itoa(g.Id), body)
	ctx.Params = gin.Params{{Key: "id", Value: strconv.Itoa(g.Id)}}
	AdminUpdateSubscriptionPlanGroup(ctx)

	resp := decodeResp(t, recorder)
	if !resp.Success {
		t.Fatalf("expected success: %s", resp.Message)
	}

	// Verify in DB
	updated, err := model.GetSubscriptionPlanGroupById(g.Id)
	if err != nil {
		t.Fatalf("failed to get updated group: %v", err)
	}
	if updated.Title != "New Title" {
		t.Errorf("expected title 'New Title', got %q", updated.Title)
	}
}

func TestAdminDeleteGroupWithPlansFails(t *testing.T) {
	db := setupSubscriptionTestDB(t)
	g := seedGroup(t, db, "Has Plans")
	seedPlan(t, db, g.Id, 9.99, "month")

	ctx, recorder := newCtx(t, http.MethodDelete, "/api/subscription/admin/groups/"+strconv.Itoa(g.Id), nil)
	ctx.Params = gin.Params{{Key: "id", Value: strconv.Itoa(g.Id)}}
	AdminDeleteSubscriptionPlanGroup(ctx)

	resp := decodeResp(t, recorder)
	if resp.Success {
		t.Fatalf("expected failure when deleting group with plans")
	}
}

func TestAdminDeleteEmptyGroupSucceeds(t *testing.T) {
	db := setupSubscriptionTestDB(t)
	g := seedGroup(t, db, "Empty Group")

	ctx, recorder := newCtx(t, http.MethodDelete, "/api/subscription/admin/groups/"+strconv.Itoa(g.Id), nil)
	ctx.Params = gin.Params{{Key: "id", Value: strconv.Itoa(g.Id)}}
	AdminDeleteSubscriptionPlanGroup(ctx)

	resp := decodeResp(t, recorder)
	if !resp.Success {
		t.Fatalf("expected success: %s", resp.Message)
	}
}

func TestAdminGroupStatusToggle(t *testing.T) {
	db := setupSubscriptionTestDB(t)
	g := seedGroup(t, db, "Toggle Me")

	body := map[string]interface{}{"enabled": false}
	ctx, recorder := newCtx(t, http.MethodPatch, "/api/subscription/admin/groups/"+strconv.Itoa(g.Id), body)
	ctx.Params = gin.Params{{Key: "id", Value: strconv.Itoa(g.Id)}}
	AdminUpdateSubscriptionPlanGroupStatus(ctx)

	resp := decodeResp(t, recorder)
	if !resp.Success {
		t.Fatalf("expected success: %s", resp.Message)
	}

	updated, _ := model.GetSubscriptionPlanGroupById(g.Id)
	if updated.Enabled {
		t.Errorf("expected group to be disabled")
	}
}

// ---- Subscription Plan Under Group Tests ----

func TestAdminCreatePlanUnderGroup(t *testing.T) {
	db := setupSubscriptionTestDB(t)
	g := seedGroup(t, db, "Pro")

	body := map[string]interface{}{
		"plan": map[string]interface{}{
			"price_amount":   19.99,
			"duration_unit":  "month",
			"duration_value": 1,
			"total_amount":   500000,
		},
	}
	ctx, recorder := newCtx(t, http.MethodPost, "/api/subscription/admin/groups/"+strconv.Itoa(g.Id)+"/plans", body)
	ctx.Params = gin.Params{{Key: "id", Value: strconv.Itoa(g.Id)}}
	AdminCreateSubscriptionPlan(ctx)

	resp := decodeResp(t, recorder)
	if !resp.Success {
		t.Fatalf("expected success: %s", resp.Message)
	}

	var plan model.SubscriptionPlan
	if err := common.Unmarshal(resp.Data, &plan); err != nil {
		t.Fatalf("failed to decode plan: %v", err)
	}
	if plan.GroupID != g.Id {
		t.Errorf("expected group_id %d, got %d", g.Id, plan.GroupID)
	}
}

func TestAdminCreatePlanWithoutGroupFails(t *testing.T) {
	setupSubscriptionTestDB(t)

	body := map[string]interface{}{
		"plan": map[string]interface{}{
			"price_amount":  19.99,
			"duration_unit": "month",
			"total_amount":  500000,
		},
	}
	ctx, recorder := newCtx(t, http.MethodPost, "/api/subscription/admin/groups/0/plans", body)
	ctx.Params = gin.Params{{Key: "id", Value: "0"}}
	AdminCreateSubscriptionPlan(ctx)

	resp := decodeResp(t, recorder)
	if resp.Success {
		t.Fatalf("expected failure for plan without group")
	}
}

// ---- Public API Tests ----

func TestGetPublicSubscriptionPlansGroupedResponse(t *testing.T) {
	db := setupSubscriptionTestDB(t)
	g := seedGroup(t, db, "Starter")
	seedPlan(t, db, g.Id, 5.99, "month")
	seedPlan(t, db, g.Id, 59.99, "year")

	// Also create a disabled group - should not appear
	disabled := seedGroup(t, db, "Disabled Group")
	db.Model(&model.SubscriptionPlanGroup{}).Where("id = ?", disabled.Id).Update("enabled", false)

	ctx, recorder := newCtx(t, http.MethodGet, "/api/subscription/public-plans", nil)
	GetPublicSubscriptionPlans(ctx)

	resp := decodeResp(t, recorder)
	if !resp.Success {
		t.Fatalf("expected success: %s", resp.Message)
	}

	var groups []PublicPlanGroupDTO
	if err := common.Unmarshal(resp.Data, &groups); err != nil {
		t.Fatalf("failed to decode public groups: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected 1 enabled group, got %d", len(groups))
	}
	if groups[0].Title != "Starter" {
		t.Errorf("expected title 'Starter', got %q", groups[0].Title)
	}
	if len(groups[0].Plans) != 2 {
		t.Errorf("expected 2 plan variants, got %d", len(groups[0].Plans))
	}
}

func TestGetPublicPlansSkipsGroupsWithNoPlans(t *testing.T) {
	db := setupSubscriptionTestDB(t)
	seedGroup(t, db, "Empty Group") // no plans

	ctx, recorder := newCtx(t, http.MethodGet, "/api/subscription/public-plans", nil)
	GetPublicSubscriptionPlans(ctx)

	resp := decodeResp(t, recorder)
	if !resp.Success {
		t.Fatalf("expected success: %s", resp.Message)
	}

	var groups []PublicPlanGroupDTO
	if err := common.Unmarshal(resp.Data, &groups); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(groups) != 0 {
		t.Errorf("expected 0 groups (no plans), got %d", len(groups))
	}
}

func TestGetSubscriptionPlansEnrichesGroupFields(t *testing.T) {
	db := setupSubscriptionTestDB(t)
	g := seedGroup(t, db, "Premium")
	// Invalidate any stale cache entry from previous tests (cache is global)
	model.InvalidateSubscriptionPlanGroupCache(g.Id)
	seedPlan(t, db, g.Id, 29.99, "month")

	ctx, recorder := newCtx(t, http.MethodGet, "/api/subscription/plans", nil)
	GetSubscriptionPlans(ctx)

	resp := decodeResp(t, recorder)
	if !resp.Success {
		t.Fatalf("expected success: %s", resp.Message)
	}

	var plans []SubscriptionPlanDTO
	if err := common.Unmarshal(resp.Data, &plans); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(plans) != 1 {
		t.Fatalf("expected 1 plan, got %d", len(plans))
	}
	if plans[0].GroupTitle != "Premium" {
		t.Errorf("expected group_title 'Premium', got %q", plans[0].GroupTitle)
	}
	if plans[0].GroupSubtitle == "" {
		t.Errorf("expected non-empty group_subtitle")
	}
}

// ---- Topup Tier Tests ----

func TestAdminCreateTopupTier(t *testing.T) {
	setupSubscriptionTestDB(t)

	body := map[string]interface{}{
		"title":    "Starter Pack",
		"amount":   10,
		"discount": 1.0,
		"features": `[{"text":"Basic support","icon":"check","style":"default"}]`,
	}
	ctx, recorder := newCtx(t, http.MethodPost, "/api/topup/admin/tiers", body)
	AdminCreateTopupTier(ctx)

	resp := decodeResp(t, recorder)
	if !resp.Success {
		t.Fatalf("expected success: %s", resp.Message)
	}

	var tier model.TopupTier
	if err := common.Unmarshal(resp.Data, &tier); err != nil {
		t.Fatalf("failed to decode tier: %v", err)
	}
	if tier.Title != "Starter Pack" {
		t.Errorf("expected title 'Starter Pack', got %q", tier.Title)
	}
	if tier.Amount != 10 {
		t.Errorf("expected amount 10, got %d", tier.Amount)
	}
}

func TestAdminCreateTopupTierEmptyTitleFails(t *testing.T) {
	setupSubscriptionTestDB(t)

	body := map[string]interface{}{
		"title":  "",
		"amount": 10,
	}
	ctx, recorder := newCtx(t, http.MethodPost, "/api/topup/admin/tiers", body)
	AdminCreateTopupTier(ctx)

	resp := decodeResp(t, recorder)
	if resp.Success {
		t.Fatalf("expected failure for empty title")
	}
}

func TestAdminCreateTopupTierZeroAmountFails(t *testing.T) {
	setupSubscriptionTestDB(t)

	body := map[string]interface{}{
		"title":  "Invalid",
		"amount": 0,
	}
	ctx, recorder := newCtx(t, http.MethodPost, "/api/topup/admin/tiers", body)
	AdminCreateTopupTier(ctx)

	resp := decodeResp(t, recorder)
	if resp.Success {
		t.Fatalf("expected failure for zero amount")
	}
}

func TestAdminListTopupTiers(t *testing.T) {
	db := setupSubscriptionTestDB(t)
	seedTopupTier(t, db, "Small", 10, 1.0)
	seedTopupTier(t, db, "Medium", 50, 0.95)
	seedTopupTier(t, db, "Large", 100, 0.9)

	ctx, recorder := newCtx(t, http.MethodGet, "/api/topup/admin/tiers", nil)
	AdminListTopupTiers(ctx)

	resp := decodeResp(t, recorder)
	if !resp.Success {
		t.Fatalf("expected success: %s", resp.Message)
	}

	var tiers []model.TopupTier
	if err := common.Unmarshal(resp.Data, &tiers); err != nil {
		t.Fatalf("failed to decode tiers: %v", err)
	}
	if len(tiers) != 3 {
		t.Errorf("expected 3 tiers, got %d", len(tiers))
	}
}

func TestAdminUpdateTopupTier(t *testing.T) {
	db := setupSubscriptionTestDB(t)
	tier := seedTopupTier(t, db, "Old Name", 50, 1.0)

	body := map[string]interface{}{
		"title":    "New Name",
		"discount": 0.85,
	}
	ctx, recorder := newCtx(t, http.MethodPut, "/api/topup/admin/tiers/"+strconv.Itoa(tier.Id), body)
	ctx.Params = gin.Params{{Key: "id", Value: strconv.Itoa(tier.Id)}}
	AdminUpdateTopupTier(ctx)

	resp := decodeResp(t, recorder)
	if !resp.Success {
		t.Fatalf("expected success: %s", resp.Message)
	}

	updated, _ := model.GetTopupTierById(tier.Id)
	if updated.Title != "New Name" {
		t.Errorf("expected title 'New Name', got %q", updated.Title)
	}
}

func TestAdminDeleteTopupTier(t *testing.T) {
	db := setupSubscriptionTestDB(t)
	tier := seedTopupTier(t, db, "Delete Me", 10, 1.0)

	ctx, recorder := newCtx(t, http.MethodDelete, "/api/topup/admin/tiers/"+strconv.Itoa(tier.Id), nil)
	ctx.Params = gin.Params{{Key: "id", Value: strconv.Itoa(tier.Id)}}
	AdminDeleteTopupTier(ctx)

	resp := decodeResp(t, recorder)
	if !resp.Success {
		t.Fatalf("expected success: %s", resp.Message)
	}

	_, err := model.GetTopupTierById(tier.Id)
	if err == nil {
		t.Errorf("expected tier to be deleted")
	}
}

func TestAdminTopupTierStatusToggle(t *testing.T) {
	db := setupSubscriptionTestDB(t)
	tier := seedTopupTier(t, db, "Toggle Tier", 20, 1.0)

	body := map[string]interface{}{"enabled": false}
	ctx, recorder := newCtx(t, http.MethodPatch, "/api/topup/admin/tiers/"+strconv.Itoa(tier.Id), body)
	ctx.Params = gin.Params{{Key: "id", Value: strconv.Itoa(tier.Id)}}
	AdminUpdateTopupTierStatus(ctx)

	resp := decodeResp(t, recorder)
	if !resp.Success {
		t.Fatalf("expected success: %s", resp.Message)
	}

	updated, _ := model.GetTopupTierById(tier.Id)
	if updated.Enabled {
		t.Errorf("expected tier to be disabled")
	}
}

func TestGetPublicTopupTiersFiltersDisabled(t *testing.T) {
	db := setupSubscriptionTestDB(t)
	seedTopupTier(t, db, "Enabled", 10, 1.0)
	disabled := seedTopupTier(t, db, "Disabled", 50, 0.9)
	db.Model(&model.TopupTier{}).Where("id = ?", disabled.Id).Update("enabled", false)

	ctx, recorder := newCtx(t, http.MethodGet, "/api/topup/tiers", nil)
	GetPublicTopupTiers(ctx)

	resp := decodeResp(t, recorder)
	if !resp.Success {
		t.Fatalf("expected success: %s", resp.Message)
	}

	var tiers []model.TopupTier
	if err := common.Unmarshal(resp.Data, &tiers); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(tiers) != 1 {
		t.Errorf("expected 1 enabled tier, got %d", len(tiers))
	}
	if tiers[0].Title != "Enabled" {
		t.Errorf("expected 'Enabled' tier, got %q", tiers[0].Title)
	}
}

func TestTopupTierInvalidDiscountNormalized(t *testing.T) {
	setupSubscriptionTestDB(t)

	body := map[string]interface{}{
		"title":    "Bad Discount",
		"amount":   10,
		"discount": 1.5, // > 1, should be normalized to 1.0
	}
	ctx, recorder := newCtx(t, http.MethodPost, "/api/topup/admin/tiers", body)
	AdminCreateTopupTier(ctx)

	resp := decodeResp(t, recorder)
	if !resp.Success {
		t.Fatalf("expected success: %s", resp.Message)
	}

	var tier model.TopupTier
	if err := common.Unmarshal(resp.Data, &tier); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if tier.Discount != 1.0 {
		t.Errorf("expected discount 1.0 (normalized), got %f", tier.Discount)
	}
}
