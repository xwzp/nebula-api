// Period type constants
export const PERIOD_MONTHLY = 'monthly';
export const PERIOD_QUARTERLY = 'quarterly';
export const PERIOD_YEARLY = 'yearly';

export function formatPeriodLabel(period, t) {
  const labels = {
    [PERIOD_MONTHLY]: t('月付'),
    [PERIOD_QUARTERLY]: t('季付'),
    [PERIOD_YEARLY]: t('年付'),
  };
  return labels[period] || period;
}

export function parseFeatures(featuresStr) {
  if (!featuresStr) return [];
  try {
    return JSON.parse(featuresStr);
  } catch {
    return [];
  }
}

export function calcPeriodPrice(plan, periodType) {
  const base = Number(plan?.price_monthly || 0);
  switch (periodType) {
    case PERIOD_MONTHLY:
      return base;
    case PERIOD_QUARTERLY:
      return Math.round(base * 3 * (1 - (plan?.quarterly_discount || 0) / 100) * 100) / 100;
    case PERIOD_YEARLY:
      return Math.round(base * 12 * (1 - (plan?.yearly_discount || 0) / 100) * 100) / 100;
    default:
      return base;
  }
}

export function getStripePriceId(plan, periodType) {
  switch (periodType) {
    case PERIOD_MONTHLY: return plan?.monthly_stripe_price_id || '';
    case PERIOD_QUARTERLY: return plan?.quarterly_stripe_price_id || '';
    case PERIOD_YEARLY: return plan?.yearly_stripe_price_id || '';
    default: return '';
  }
}

export function getCreemProductId(plan, periodType) {
  switch (periodType) {
    case PERIOD_MONTHLY: return plan?.monthly_creem_product_id || '';
    case PERIOD_QUARTERLY: return plan?.quarterly_creem_product_id || '';
    case PERIOD_YEARLY: return plan?.yearly_creem_product_id || '';
    default: return '';
  }
}

// Returns parsed features array for a period, falling back to shared features.
export function getPeriodFeatures(plan, periodType) {
  let raw = '';
  switch (periodType) {
    case PERIOD_MONTHLY: raw = plan?.monthly_features; break;
    case PERIOD_QUARTERLY: raw = plan?.quarterly_features; break;
    case PERIOD_YEARLY: raw = plan?.yearly_features; break;
  }
  const parsed = parseFeatures(raw);
  return parsed.length > 0 ? parsed : parseFeatures(plan?.features);
}

export function sortBySortOrderThenAmount(items) {
  return [...items].sort((a, b) => {
    if ((b.sort_order || 0) !== (a.sort_order || 0)) {
      return (b.sort_order || 0) - (a.sort_order || 0);
    }
    return (a.amount || 0) - (b.amount || 0);
  });
}

// Kept for backward compatibility with user-facing pages
// (SubscriptionPlansCard, SubscriptionPurchaseModal, Home)
export function formatSubscriptionDuration(plan, t) {
  const unit = plan?.duration_unit || 'month';
  const value = plan?.duration_value || 1;
  const unitLabels = {
    year: t('年'),
    month: t('个月'),
    day: t('天'),
    hour: t('小时'),
    custom: t('自定义'),
  };
  if (unit === 'custom') {
    const seconds = plan?.custom_seconds || 0;
    if (seconds >= 86400) return `${Math.floor(seconds / 86400)} ${t('天')}`;
    if (seconds >= 3600) return `${Math.floor(seconds / 3600)} ${t('小时')}`;
    return `${seconds} ${t('秒')}`;
  }
  return `${value} ${unitLabels[unit] || unit}`;
}

export function formatSubscriptionResetPeriod(plan, t) {
  const period = plan?.quota_reset_period || 'never';
  if (period === 'never') return t('不重置');
  if (period === 'daily') return t('每天');
  if (period === 'weekly') return t('每周');
  if (period === 'monthly') return t('每月');
  if (period === 'custom') {
    const seconds = Number(plan?.quota_reset_custom_seconds || 0);
    if (seconds >= 86400) return `${Math.floor(seconds / 86400)} ${t('天')}`;
    if (seconds >= 3600) return `${Math.floor(seconds / 3600)} ${t('小时')}`;
    if (seconds >= 60) return `${Math.floor(seconds / 60)} ${t('分钟')}`;
    return `${seconds} ${t('秒')}`;
  }
  return t('不重置');
}
