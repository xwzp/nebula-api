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
		&model.SubscriptionPlan{},
		&model.SubscriptionOrder{},
		&model.UserSubscription{},
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

func seedPlan(t *testing.T, db *gorm.DB, title string, priceMonthly float64) *model.SubscriptionPlan {
	t.Helper()
	plan := &model.SubscriptionPlan{
		Title:            title,
		Subtitle:         title + " subtitle",
		Tag:              "popular",
		Features:         `[{"text":"Feature 1","icon":"check","style":"default"}]`,
		PriceMonthly:     priceMonthly,
		Currency:         "USD",
		MonthlyEnabled:   true,
		QuarterlyEnabled: true,
		QuarterlyDiscount: 10,
		YearlyEnabled:    true,
		YearlyDiscount:   20,
		TotalAmount:      500000,
		Enabled:          true,
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

// ---- Subscription Plan Tests ----

func TestAdminCreateSubscriptionPlan(t *testing.T) {
	setupSubscriptionTestDB(t)

	body := map[string]interface{}{
		"title":            "Pro Plan",
		"subtitle":         "Best value",
		"tag":              "推荐",
		"price_monthly":    9.99,
		"monthly_enabled":  true,
		"quarterly_enabled": true,
		"quarterly_discount": 15,
		"yearly_enabled":   true,
		"yearly_discount":  25,
		"total_amount":     500000,
		"features":         `[{"text":"Unlimited access","icon":"check","style":"highlight"}]`,
	}
	ctx, recorder := newCtx(t, http.MethodPost, "/api/subscription/admin/plans", body)
	AdminCreateSubscriptionPlan(ctx)

	resp := decodeResp(t, recorder)
	if !resp.Success {
		t.Fatalf("expected success, got: %s", resp.Message)
	}

	var created model.SubscriptionPlan
	if err := common.Unmarshal(resp.Data, &created); err != nil {
		t.Fatalf("failed to decode created plan: %v", err)
	}
	if created.Title != "Pro Plan" {
		t.Errorf("expected title 'Pro Plan', got %q", created.Title)
	}
	if created.Id <= 0 {
		t.Errorf("expected positive ID, got %d", created.Id)
	}
}

func TestAdminCreatePlanEmptyTitleFails(t *testing.T) {
	setupSubscriptionTestDB(t)

	body := map[string]interface{}{
		"title":         "  ",
		"price_monthly": 9.99,
	}
	ctx, recorder := newCtx(t, http.MethodPost, "/api/subscription/admin/plans", body)
	AdminCreateSubscriptionPlan(ctx)

	resp := decodeResp(t, recorder)
	if resp.Success {
		t.Fatalf("expected failure for empty title")
	}
}

func TestAdminListPlans(t *testing.T) {
	db := setupSubscriptionTestDB(t)
	seedPlan(t, db, "Basic", 9.99)
	seedPlan(t, db, "Pro", 19.99)

	ctx, recorder := newCtx(t, http.MethodGet, "/api/subscription/admin/plans", nil)
	AdminListSubscriptionPlans(ctx)

	resp := decodeResp(t, recorder)
	if !resp.Success {
		t.Fatalf("expected success: %s", resp.Message)
	}

	var plans []model.SubscriptionPlan
	if err := common.Unmarshal(resp.Data, &plans); err != nil {
		t.Fatalf("failed to decode plans: %v", err)
	}
	if len(plans) != 2 {
		t.Fatalf("expected 2 plans, got %d", len(plans))
	}
}

func TestAdminDeletePlanSucceeds(t *testing.T) {
	db := setupSubscriptionTestDB(t)
	plan := seedPlan(t, db, "Delete Me", 9.99)

	ctx, recorder := newCtx(t, http.MethodDelete, "/api/subscription/admin/plans/"+strconv.Itoa(plan.Id), nil)
	ctx.Params = gin.Params{{Key: "id", Value: strconv.Itoa(plan.Id)}}
	AdminDeleteSubscriptionPlan(ctx)

	resp := decodeResp(t, recorder)
	if !resp.Success {
		t.Fatalf("expected success: %s", resp.Message)
	}
}

func TestAdminPlanStatusToggle(t *testing.T) {
	db := setupSubscriptionTestDB(t)
	plan := seedPlan(t, db, "Toggle Me", 9.99)

	enabled := false
	body := map[string]interface{}{"enabled": enabled}
	ctx, recorder := newCtx(t, http.MethodPatch, "/api/subscription/admin/plans/"+strconv.Itoa(plan.Id), body)
	ctx.Params = gin.Params{{Key: "id", Value: strconv.Itoa(plan.Id)}}
	AdminUpdateSubscriptionPlanStatus(ctx)

	resp := decodeResp(t, recorder)
	if !resp.Success {
		t.Fatalf("expected success: %s", resp.Message)
	}
}

// ---- Public API Tests ----

func TestGetPublicSubscriptionPlans(t *testing.T) {
	db := setupSubscriptionTestDB(t)
	seedPlan(t, db, "Starter", 5.99)

	// Also create a disabled plan - should not appear
	disabled := seedPlan(t, db, "Disabled Plan", 19.99)
	db.Model(&model.SubscriptionPlan{}).Where("id = ?", disabled.Id).Update("enabled", false)

	ctx, recorder := newCtx(t, http.MethodGet, "/api/subscription/public-plans", nil)
	GetPublicSubscriptionPlans(ctx)

	resp := decodeResp(t, recorder)
	if !resp.Success {
		t.Fatalf("expected success: %s", resp.Message)
	}

	var plans []PublicSubscriptionPlanDTO
	if err := common.Unmarshal(resp.Data, &plans); err != nil {
		t.Fatalf("failed to decode public plans: %v", err)
	}
	if len(plans) != 1 {
		t.Fatalf("expected 1 enabled plan, got %d", len(plans))
	}
	if plans[0].Title != "Starter" {
		t.Errorf("expected title 'Starter', got %q", plans[0].Title)
	}
	if plans[0].Periods["monthly"].Price != 5.99 {
		t.Errorf("expected monthly price 5.99, got %f", plans[0].Periods["monthly"].Price)
	}
}

func TestGetSubscriptionPlansReturnsEnabled(t *testing.T) {
	db := setupSubscriptionTestDB(t)
	seedPlan(t, db, "Premium", 29.99)

	ctx, recorder := newCtx(t, http.MethodGet, "/api/subscription/plans", nil)
	GetSubscriptionPlans(ctx)

	resp := decodeResp(t, recorder)
	if !resp.Success {
		t.Fatalf("expected success: %s", resp.Message)
	}

	var plans []model.SubscriptionPlan
	if err := common.Unmarshal(resp.Data, &plans); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(plans) != 1 {
		t.Fatalf("expected 1 plan, got %d", len(plans))
	}
	if plans[0].Title != "Premium" {
		t.Errorf("expected title 'Premium', got %q", plans[0].Title)
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
		"discount": 1.5,
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
