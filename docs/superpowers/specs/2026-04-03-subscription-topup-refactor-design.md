# Subscription & Topup Tier Management Refactor

## Overview

Refactor the subscription management system from flat independent plans to a **plan group model** (group → billing cycle variants), and extract pay-as-you-go topup tier configuration from JSON system settings into a dedicated **topup tier** table with its own admin management page.

## Goals

1. **Subscription plan groups**: Group billing cycle variants (monthly, yearly, quarterly, etc.) under a single plan group with shared display information (title, subtitle, tag, feature list).
2. **Topup tier management**: Replace `AmountOptions` / `AmountDiscount` JSON config with a `topup_tier` database table and dedicated admin CRUD page.
3. **Structured feature lists**: Both plan groups and topup tiers support a structured feature/advantage list for homepage card display (icon type + style per item).
4. **Clean migration**: No data migration needed — no active subscriptions exist on production. Topup tier is a new table; old JSON config is removed.

## Non-Goals

- i18n support for new text fields (not needed now)
- Changing the homepage layout (keep current vertical scroll: subscriptions first, then topup)
- Subscription upgrade/downgrade logic (future work)
- Changes to payment gateway integrations (Stripe, WeChat, Alipay, etc.)

---

## Data Model

### New Table: `subscription_plan_group`

Represents a logical subscription product (e.g., "Pro Plan") with shared display properties.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | int | PK, auto-increment | Primary key |
| title | varchar(128) | NOT NULL | Plan group name, e.g., "Pro" |
| subtitle | varchar(255) | | Subtitle, e.g., "For professional users" |
| tag | varchar(64) | | Card badge label, e.g., "Most Popular" (empty = hidden) |
| features | text (JSON) | | Structured feature list (see below) |
| sort_order | int | default 0 | Display order (higher = first) |
| enabled | bool | default true | Visibility toggle |
| created_at | datetime | | |
| updated_at | datetime | | |

### Modified Table: `subscription_plan`

Each row becomes a billing cycle variant under a group. Changes from current schema:

| Change | Column | Details |
|--------|--------|---------|
| **Add** | group_id | int, NOT NULL, FK → subscription_plan_group.id |
| **Remove** | title | Moved to group |
| **Remove** | subtitle | Moved to group |
| **Keep** | price_amount | Price for this specific variant |
| **Keep** | currency | |
| **Keep** | duration_unit, duration_value, custom_seconds | Billing cycle definition |
| **Keep** | total_amount | Quota allocation (can differ per variant) |
| **Keep** | quota_reset_period, quota_reset_custom_seconds | Quota reset rules (can differ per variant) |
| **Keep** | upgrade_group | User group upgrade (can differ per variant) |
| **Keep** | max_purchase_per_user | |
| **Keep** | stripe_price_id, creem_product_id | Payment integration IDs |
| **Keep** | enabled | Independent enable/disable per variant |
| **Keep** | sort_order | Order within the group |

### New Table: `topup_tier`

Replaces `PaymentSetting.AmountOptions` and `PaymentSetting.AmountDiscount`.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | int | PK, auto-increment | Primary key |
| title | varchar(128) | NOT NULL | Tier name, e.g., "Starter Pack" |
| subtitle | varchar(255) | | Subtitle, e.g., "For light usage" |
| tag | varchar(64) | | Card badge, e.g., "Best Value" (empty = hidden) |
| amount | bigint | NOT NULL | Topup amount in USD |
| discount | decimal(10,4) | default 1.0 | Discount rate: 1.0 = full price, 0.95 = 5% off |
| bonus_quota | bigint | default 0 | Extra quota granted on top of standard conversion |
| features | text (JSON) | | Structured feature list (same format as plan group) |
| sort_order | int | default 0 | Display order (higher = first) |
| enabled | bool | default true | Visibility toggle |
| created_at | datetime | | |
| updated_at | datetime | | |

### Shared: Feature List JSON Structure

Used by both `subscription_plan_group.features` and `topup_tier.features`:

```json
[
  {"text": "Unlimited GPT-4 conversations", "icon": "check", "style": "default"},
  {"text": "Priority support", "icon": "check", "style": "highlight"},
  {"text": "API access", "icon": "x", "style": "disabled"},
  {"text": "Beta features included", "icon": "info", "style": "default"}
]
```

- `icon`: `"check"` | `"x"` | `"info"`
- `style`: `"default"` | `"highlight"` | `"disabled"`

---

## API Design

### Subscription Admin APIs

```
GET    /api/subscription/admin/groups              — List all groups (with nested variants)
POST   /api/subscription/admin/groups              — Create group
PUT    /api/subscription/admin/groups/:id           — Update group
DELETE /api/subscription/admin/groups/:id           — Delete group (must have no variants)
PATCH  /api/subscription/admin/groups/:id           — Toggle group enabled

POST   /api/subscription/admin/groups/:id/plans     — Add variant to group
PUT    /api/subscription/admin/plans/:id            — Update variant
DELETE /api/subscription/admin/plans/:id            — Delete variant
PATCH  /api/subscription/admin/plans/:id            — Toggle variant enabled
```

### Subscription Public APIs

```
GET    /api/subscription/public-plans              — Returns groups with enabled variants, for homepage cards
```

Response shape:
```json
[
  {
    "id": 1,
    "title": "Pro",
    "subtitle": "For professional users",
    "tag": "Most Popular",
    "features": [...],
    "sort_order": 10,
    "plans": [
      {
        "id": 101,
        "price_amount": 29.00,
        "duration_unit": "month",
        "duration_value": 1,
        "total_amount": 0,
        ...
      },
      {
        "id": 102,
        "price_amount": 290.00,
        "duration_unit": "year",
        "duration_value": 1,
        "total_amount": 0,
        ...
      }
    ]
  }
]
```

### Subscription User APIs (unchanged)

Existing payment and self-query endpoints remain unchanged. They reference `plan_id` directly, which still exists as a variant ID.

### Topup Tier Admin APIs

```
GET    /api/topup/admin/tiers                      — List all tiers
POST   /api/topup/admin/tiers                      — Create tier
PUT    /api/topup/admin/tiers/:id                   — Update tier
PATCH  /api/topup/admin/tiers/:id                   — Toggle enabled
DELETE /api/topup/admin/tiers/:id                   — Delete tier
```

### Topup Tier Public API

```
GET    /api/topup/tiers                            — List enabled tiers (for homepage and topup page)
```

This replaces the tier-related fields currently returned by `GET /api/user/topup/info`. The `topup/info` endpoint continues to return payment method configuration, min topup, group ratio, etc. — just no longer `amount_options` or `amount_discount`.

---

## Admin Frontend

### Subscription Management Page (refactored)

- **List view**: Shows plan groups as primary rows (title, tag, # of variants, sort order, enabled)
- **Expand/click group**: Shows variant table (billing cycle, price, quota, payment IDs, enabled)
- **Create flow**: "Create Group" button → sidebar/modal form with:
  - Title, subtitle, tag, feature list editor, sort order, enabled
- **Add variant**: Within a group, "Add Plan" button → sidebar/modal form with:
  - Billing cycle (duration unit + value), price, quota settings, payment IDs, upgrade group, enabled
- **Feature list editor**: Dynamic list with per-item fields: text input, icon dropdown (check/x/info), style dropdown (default/highlight/disabled). Add/remove/reorder rows.

### Topup Tier Management Page (new)

- **Menu position**: Parallel to "Subscription Management" and "Channel Management" in admin sidebar
- **List view**: Table of tiers (title, amount, discount, tag, bonus quota, sort order, enabled)
- **Create/Edit**: Sidebar/modal form with:
  - Title, subtitle, tag
  - Amount (USD), discount rate, bonus quota
  - Feature list editor (same component as subscription)
  - Sort order, enabled

### System Settings Cleanup

- Remove `AmountOptions` JSON field from payment settings page
- Remove `AmountDiscount` JSON field from payment settings page
- Keep other payment settings (Price, MinTopUp, TopupGroupRatio, gateway configs, etc.)

---

## Backend Changes

### Model Layer

- New model: `SubscriptionPlanGroup` in `model/subscription.go`
- New model: `TopupTier` in `model/topup_tier.go`
- Modify `SubscriptionPlan`: remove `Title`/`Subtitle`, add `GroupID`
- Auto-migrate both new tables on startup
- Add CRUD methods for both new models
- Add cache layer for `TopupTier` (similar to existing plan cache)

### Controller Layer

- New controller functions for group CRUD in `controller/subscription.go`
- Refactor `GetPublicPlans` to return group-aggregated response
- New controller: `controller/topup_tier.go` for tier CRUD
- Modify `controller/topup.go`:
  - `getPayMoney()` reads discount from `topup_tier` table instead of `PaymentSetting`
  - Topup info endpoint no longer returns `amount_options`/`amount_discount`

### Router Layer

- Register new group admin routes in `router/api-router.go`
- Register new topup tier routes in `router/api-router.go`

### Settings Layer

- Remove `AmountOptions` and `AmountDiscount` from `PaymentSetting` struct
- Remove related option sync logic

---

## Affected Existing Endpoints

| Endpoint | Change |
|----------|--------|
| `GET /api/subscription/public-plans` | Returns group-structured data instead of flat plan list |
| `GET /api/user/topup/info` | No longer returns `amount_options` or `amount_discount` |
| `POST /api/user/pay` (and other pay endpoints) | Reads discount from `topup_tier` table |
| All existing subscription admin plan endpoints | Replaced by group + variant endpoints |
| Existing subscription payment endpoints | Unchanged (still reference plan_id) |

---

## Homepage Frontend Changes

- **Subscription section**: Refactor to consume group-structured API response. Render cards per group with year/month toggle switching between variants within the same group.
- **Topup section**: Refactor to consume `GET /api/topup/tiers` instead of extracting tiers from `topup/info`. Render cards with title, subtitle, tag, features, and pricing (showing discount if applicable).
- **Shared card component**: Extract a reusable card component for the structured feature list rendering (check/x/info icons with default/highlight/disabled styles), used by both subscription and topup cards.
