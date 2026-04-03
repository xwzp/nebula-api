/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import React, { useEffect, useMemo, useRef, useState } from 'react';
import {
  Badge,
  Button,
  Card,
  Divider,
  RadioGroup,
  Radio,
  Select,
  Skeleton,
  Space,
  Tag,
  Tooltip,
  Typography,
} from '@douyinfe/semi-ui';
import { API, showError, showSuccess, renderQuota } from '../../helpers';
import { getCurrencyConfig } from '../../helpers/render';
import { RefreshCw, Sparkles } from 'lucide-react';
import SubscriptionPurchaseModal from './modals/SubscriptionPurchaseModal';
import ScanPayModal from './modals/WechatPayModal';
import {
  PERIOD_MONTHLY, PERIOD_QUARTERLY, PERIOD_YEARLY,
  formatPeriodLabel, calcPeriodPrice, getStripePriceId,
  getCreemProductId, getPeriodFeatures, parseFeatures,
} from '../../helpers/subscriptionFormat';

const { Text } = Typography;

function getEnabledPeriods(plan) {
  const periods = [];
  if (plan?.monthly_enabled) periods.push(PERIOD_MONTHLY);
  if (plan?.quarterly_enabled) periods.push(PERIOD_QUARTERLY);
  if (plan?.yearly_enabled) periods.push(PERIOD_YEARLY);
  return periods;
}

function getDefaultPeriod(plan) {
  if (plan?.monthly_enabled) return PERIOD_MONTHLY;
  if (plan?.quarterly_enabled) return PERIOD_QUARTERLY;
  if (plan?.yearly_enabled) return PERIOD_YEARLY;
  return PERIOD_MONTHLY;
}

// Filter for epay methods
function getEpayMethods(payMethods = []) {
  return (payMethods || []).filter(
    (m) => m?.type && m.type !== 'stripe' && m.type !== 'creem',
  );
}

// Submit epay form
function submitEpayForm({ url, params }) {
  const form = document.createElement('form');
  form.action = url;
  form.method = 'POST';
  const isSafari =
    navigator.userAgent.indexOf('Safari') > -1 &&
    navigator.userAgent.indexOf('Chrome') < 1;
  if (!isSafari) form.target = '_blank';
  Object.keys(params || {}).forEach((key) => {
    const input = document.createElement('input');
    input.type = 'hidden';
    input.name = key;
    input.value = params[key];
    form.appendChild(input);
  });
  document.body.appendChild(form);
  form.submit();
  document.body.removeChild(form);
}

function getPaymentError(res, t) {
  return typeof res.data?.data === 'string'
    ? res.data.data
    : res.data?.message || t('支付失败');
}

const SubscriptionPlansCard = ({
  t,
  loading = false,
  plans = [],
  payMethods = [],
  enableOnlineTopUp = false,
  enableStripeTopUp = false,
  enableCreemTopUp = false,
  enableWechatTopUp = false,
  enableAlipayTopUp = false,
  billingPreference,
  onChangeBillingPreference,
  activeSubscriptions = [],
  allSubscriptions = [],
  reloadSubscriptionSelf,
  withCard = true,
  defaultOpenPlanId = null,
}) => {
  const [open, setOpen] = useState(false);
  // selectedPlan stores { plan, periodType }
  const [selectedPlan, setSelectedPlan] = useState(null);
  const [paying, setPaying] = useState(false);
  const [selectedEpayMethod, setSelectedEpayMethod] = useState('');
  const [refreshing, setRefreshing] = useState(false);

  // Track the selected period per plan card (keyed by plan.id)
  const [periodSelections, setPeriodSelections] = useState({});

  const [scanPayData, setScanPayData] = useState(null);

  const epayMethods = useMemo(() => getEpayMethods(payMethods), [payMethods]);

  // Initialize period selections when plans load
  useEffect(() => {
    if (plans.length === 0) return;
    setPeriodSelections((prev) => {
      const next = { ...prev };
      plans.forEach((p) => {
        if (p?.id && !next[p.id]) {
          next[p.id] = getDefaultPeriod(p);
        }
      });
      return next;
    });
  }, [plans]);

  const getSelectedPeriod = (planId) =>
    periodSelections[planId] || PERIOD_MONTHLY;

  const handlePeriodChange = (planId, period) => {
    setPeriodSelections((prev) => ({ ...prev, [planId]: period }));
  };

  const defaultOpenHandledRef = useRef(null);
  useEffect(() => {
    if (!defaultOpenPlanId || loading || plans.length === 0) return;
    if (defaultOpenHandledRef.current === defaultOpenPlanId) return;
    defaultOpenHandledRef.current = defaultOpenPlanId;
    const targetPlan = plans.find((p) => p?.id === defaultOpenPlanId);
    if (targetPlan) {
      openBuy(targetPlan, getSelectedPeriod(targetPlan.id));
    }
  }, [defaultOpenPlanId, loading, plans]);

  const openBuy = (plan, periodType) => {
    setSelectedPlan({ plan, periodType });
    setSelectedEpayMethod(epayMethods?.[0]?.type || '');
    setOpen(true);
  };

  const closeBuy = () => {
    setOpen(false);
    setSelectedPlan(null);
    setPaying(false);
  };

  const handleRefresh = async () => {
    setRefreshing(true);
    try {
      await reloadSubscriptionSelf?.();
    } finally {
      setRefreshing(false);
    }
  };

  const payStripe = async () => {
    const plan = selectedPlan?.plan;
    const periodType = selectedPlan?.periodType;
    const stripePriceId = getStripePriceId(plan, periodType);
    if (!stripePriceId) {
      showError(t('该套餐未配置 Stripe'));
      return;
    }
    setPaying(true);
    try {
      const res = await API.post('/api/subscription/stripe/pay', {
        plan_id: plan.id,
        period_type: periodType,
      });
      if (res.data?.message === 'success') {
        window.open(res.data.data?.pay_link, '_blank');
        showSuccess(t('已打开支付页面'));
        closeBuy();
      } else {
        showError(getPaymentError(res, t));
      }
    } catch (e) {
      showError(t('支付请求失败'));
    } finally {
      setPaying(false);
    }
  };

  const payCreem = async () => {
    const plan = selectedPlan?.plan;
    const periodType = selectedPlan?.periodType;
    const creemProductId = getCreemProductId(plan, periodType);
    if (!creemProductId) {
      showError(t('该套餐未配置 Creem'));
      return;
    }
    setPaying(true);
    try {
      const res = await API.post('/api/subscription/creem/pay', {
        plan_id: plan.id,
        period_type: periodType,
      });
      if (res.data?.message === 'success') {
        window.open(res.data.data?.checkout_url, '_blank');
        showSuccess(t('已打开支付页面'));
        closeBuy();
      } else {
        showError(getPaymentError(res, t));
      }
    } catch (e) {
      showError(t('支付请求失败'));
    } finally {
      setPaying(false);
    }
  };

  const payEpay = async () => {
    if (!selectedEpayMethod) {
      showError(t('请选择支付方式'));
      return;
    }
    setPaying(true);
    try {
      const res = await API.post('/api/subscription/epay/pay', {
        plan_id: selectedPlan.plan.id,
        period_type: selectedPlan.periodType,
        payment_method: selectedEpayMethod,
      });
      if (res.data?.message === 'success') {
        submitEpayForm({ url: res.data.url, params: res.data.data });
        showSuccess(t('已发起支付'));
        closeBuy();
      } else {
        showError(getPaymentError(res, t));
      }
    } catch (e) {
      showError(t('支付请求失败'));
    } finally {
      setPaying(false);
    }
  };

  const payScanMethod = async (method) => {
    setPaying(true);
    try {
      const res = await API.post(`/api/subscription/${method}/pay`, {
        plan_id: selectedPlan.plan.id,
        period_type: selectedPlan.periodType,
      });
      if (res.data?.message === 'success' && res.data.data?.code_url) {
        closeBuy();
        setScanPayData({
          method,
          codeUrl: res.data.data.code_url,
          money: res.data.data.pay_money || 0,
        });
      } else {
        showError(getPaymentError(res, t));
      }
    } catch (e) {
      showError(t('支付请求失败'));
    } finally {
      setPaying(false);
    }
  };

  // Current subscription info - supports multiple subscriptions
  const hasActiveSubscription = activeSubscriptions.length > 0;
  const hasAnySubscription = allSubscriptions.length > 0;
  const disableSubscriptionPreference = !hasActiveSubscription;
  const isSubscriptionPreference =
    billingPreference === 'subscription_first' ||
    billingPreference === 'subscription_only';
  const displayBillingPreference =
    disableSubscriptionPreference && isSubscriptionPreference
      ? 'wallet_first'
      : billingPreference;
  const subscriptionPreferenceLabel =
    billingPreference === 'subscription_only' ? t('仅用订阅') : t('优先订阅');

  const planPurchaseCountMap = useMemo(() => {
    const map = new globalThis.Map();
    (allSubscriptions || []).forEach((sub) => {
      const planId = sub?.subscription?.plan_id;
      if (!planId) return;
      map.set(planId, (map.get(planId) || 0) + 1);
    });
    return map;
  }, [allSubscriptions]);

  const planTitleMap = useMemo(() => {
    const map = new globalThis.Map();
    (plans || []).forEach((p) => {
      if (!p?.id) return;
      map.set(p.id, p.title || '');
    });
    return map;
  }, [plans]);

  const getPlanPurchaseCount = (planId) =>
    planPurchaseCountMap.get(planId) || 0;

  // Calculate remaining days for a single subscription
  const getRemainingDays = (sub) => {
    if (!sub?.subscription?.end_time) return 0;
    const now = Date.now() / 1000;
    const remaining = sub.subscription.end_time - now;
    return Math.max(0, Math.ceil(remaining / 86400));
  };

  // Calculate usage percent for a single subscription
  const getUsagePercent = (sub) => {
    const total = Number(sub?.subscription?.amount_total || 0);
    const used = Number(sub?.subscription?.amount_used || 0);
    if (total <= 0) return 0;
    return Math.round((used / total) * 100);
  };

  const cardContent = (
    <>
      {/* Card header */}
      {loading ? (
        <div className='space-y-4'>
          {/* My subscriptions skeleton */}
          <Card className='!rounded-xl w-full' bodyStyle={{ padding: '12px' }}>
            <div className='flex items-center justify-between mb-3'>
              <Skeleton.Title active style={{ width: 100, height: 20 }} />
              <Skeleton.Button active style={{ width: 24, height: 24 }} />
            </div>
            <div className='space-y-2'>
              <Skeleton.Paragraph active rows={2} />
            </div>
          </Card>
          {/* Plan list skeleton */}
          <div className='grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-2 xl:grid-cols-3 gap-5 w-full px-1'>
            {[1, 2, 3].map((i) => (
              <Card
                key={i}
                className='!rounded-xl w-full h-full'
                bodyStyle={{ padding: 16 }}
              >
                <Skeleton.Title
                  active
                  style={{ width: '60%', height: 24, marginBottom: 8 }}
                />
                <Skeleton.Paragraph
                  active
                  rows={1}
                  style={{ marginBottom: 12 }}
                />
                <div className='text-center py-4'>
                  <Skeleton.Title
                    active
                    style={{ width: '40%', height: 32, margin: '0 auto' }}
                  />
                </div>
                <Skeleton.Paragraph active rows={3} style={{ marginTop: 12 }} />
                <Skeleton.Button
                  active
                  block
                  style={{ marginTop: 16, height: 32 }}
                />
              </Card>
            ))}
          </div>
        </div>
      ) : (
        <Space vertical style={{ width: '100%' }} spacing={8}>
          {/* Current subscription status */}
          <Card className='!rounded-xl w-full' bodyStyle={{ padding: '12px' }}>
            <div className='flex items-center justify-between mb-2 gap-3'>
              <div className='flex items-center gap-2 flex-1 min-w-0'>
                <Text strong>{t('我的订阅')}</Text>
                {hasActiveSubscription ? (
                  <Tag
                    color='white'
                    size='small'
                    shape='circle'
                    prefixIcon={<Badge dot type='success' />}
                  >
                    {activeSubscriptions.length} {t('个生效中')}
                  </Tag>
                ) : (
                  <Tag color='white' size='small' shape='circle'>
                    {t('无生效')}
                  </Tag>
                )}
                {allSubscriptions.length > activeSubscriptions.length && (
                  <Tag color='white' size='small' shape='circle'>
                    {allSubscriptions.length - activeSubscriptions.length}{' '}
                    {t('个已过期')}
                  </Tag>
                )}
              </div>
              <div className='flex items-center gap-2'>
                <Select
                  value={displayBillingPreference}
                  onChange={onChangeBillingPreference}
                  size='small'
                  optionList={[
                    {
                      value: 'subscription_first',
                      label: disableSubscriptionPreference
                        ? `${t('优先订阅')} (${t('无生效')})`
                        : t('优先订阅'),
                      disabled: disableSubscriptionPreference,
                    },
                    { value: 'wallet_first', label: t('优先钱包') },
                    {
                      value: 'subscription_only',
                      label: disableSubscriptionPreference
                        ? `${t('仅用订阅')} (${t('无生效')})`
                        : t('仅用订阅'),
                      disabled: disableSubscriptionPreference,
                    },
                    { value: 'wallet_only', label: t('仅用钱包') },
                  ]}
                />
                <Button
                  size='small'
                  theme='light'
                  type='tertiary'
                  icon={
                    <RefreshCw
                      size={12}
                      className={refreshing ? 'animate-spin' : ''}
                    />
                  }
                  onClick={handleRefresh}
                  loading={refreshing}
                />
              </div>
            </div>
            {disableSubscriptionPreference && isSubscriptionPreference && (
              <Text type='tertiary' size='small'>
                {t('已保存偏好为')}
                {subscriptionPreferenceLabel}
                {t('，当前无生效订阅，将自动使用钱包')}
              </Text>
            )}

            {hasAnySubscription ? (
              <>
                <Divider margin={8} />
                <div className='max-h-64 overflow-y-auto pr-1 semi-table-body'>
                  {allSubscriptions.map((sub, subIndex) => {
                    const isLast = subIndex === allSubscriptions.length - 1;
                    const subscription = sub.subscription;
                    const totalAmount = Number(subscription?.amount_total || 0);
                    const usedAmount = Number(subscription?.amount_used || 0);
                    const remainAmount =
                      totalAmount > 0
                        ? Math.max(0, totalAmount - usedAmount)
                        : 0;
                    const planTitle =
                      planTitleMap.get(subscription?.plan_id) || '';
                    const remainDays = getRemainingDays(sub);
                    const usagePercent = getUsagePercent(sub);
                    const now = Date.now() / 1000;
                    const isExpired = (subscription?.end_time || 0) < now;
                    const isCancelled = subscription?.status === 'cancelled';
                    const isActive =
                      subscription?.status === 'active' && !isExpired;

                    return (
                      <div key={subscription?.id || subIndex}>
                        {/* Subscription summary */}
                        <div className='flex items-center justify-between text-xs mb-2'>
                          <div className='flex items-center gap-2'>
                            <span className='font-medium'>
                              {planTitle
                                ? `${planTitle} · ${t('订阅')} #${subscription?.id}`
                                : `${t('订阅')} #${subscription?.id}`}
                            </span>
                            {isActive ? (
                              <Tag
                                color='white'
                                size='small'
                                shape='circle'
                                prefixIcon={<Badge dot type='success' />}
                              >
                                {t('生效')}
                              </Tag>
                            ) : isCancelled ? (
                              <Tag color='white' size='small' shape='circle'>
                                {t('已作废')}
                              </Tag>
                            ) : (
                              <Tag color='white' size='small' shape='circle'>
                                {t('已过期')}
                              </Tag>
                            )}
                          </div>
                          {isActive && (
                            <span className='text-gray-500'>
                              {t('剩余')} {remainDays} {t('天')}
                            </span>
                          )}
                        </div>
                        <div className='text-xs text-gray-500 mb-2'>
                          {isActive
                            ? t('至')
                            : isCancelled
                              ? t('作废于')
                              : t('过期于')}{' '}
                          {new Date(
                            (subscription?.end_time || 0) * 1000,
                          ).toLocaleString()}
                        </div>
                        <div className='text-xs text-gray-500 mb-2'>
                          {t('总额度')}:{' '}
                          {totalAmount > 0 ? (
                            <Tooltip
                              content={`${t('原生额度')}：${usedAmount}/${totalAmount} · ${t('剩余')} ${remainAmount}`}
                            >
                              <span>
                                {renderQuota(usedAmount)}/
                                {renderQuota(totalAmount)} · {t('剩余')}{' '}
                                {renderQuota(remainAmount)}
                              </span>
                            </Tooltip>
                          ) : (
                            t('不限')
                          )}
                          {totalAmount > 0 && (
                            <span className='ml-2'>
                              {t('已用')} {usagePercent}%
                            </span>
                          )}
                        </div>
                        {!isLast && <Divider margin={12} />}
                      </div>
                    );
                  })}
                </div>
              </>
            ) : (
              <div className='text-xs text-gray-500'>
                {t('购买套餐后即可享受模型权益')}
              </div>
            )}
          </Card>

          {/* Available plans - pricing cards */}
          {plans.length > 0 ? (
            <div className='grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-2 xl:grid-cols-3 gap-5 w-full px-1'>
              {plans.map((p, index) => {
                const totalAmount = Number(p?.total_amount || 0);
                const { symbol, rate } = getCurrencyConfig();
                const enabledPeriods = getEnabledPeriods(p);
                const currentPeriod = getSelectedPeriod(p.id);
                const periodPrice = calcPeriodPrice(p, currentPeriod);
                const convertedPrice = periodPrice * rate;
                const displayPrice = convertedPrice.toFixed(
                  Number.isInteger(convertedPrice) ? 0 : 2,
                );
                const isPopular = !!p.tag;
                const limit = Number(p?.max_purchase_per_user || 0);
                const limitLabel = limit > 0 ? `${t('限购')} ${limit}` : null;
                const totalLabel =
                  totalAmount > 0
                    ? `${t('总额度')}: ${renderQuota(totalAmount)}`
                    : `${t('总额度')}: ${t('不限')}`;
                const upgradeLabel = p?.upgrade_group
                  ? `${t('升级分组')}: ${p.upgrade_group}`
                  : null;

                // Period-specific features
                const features = getPeriodFeatures(p, currentPeriod);

                // Build benefits list
                const planBenefits = [
                  totalAmount > 0
                    ? {
                        label: totalLabel,
                        tooltip: `${t('原生额度')}：${totalAmount}`,
                      }
                    : { label: totalLabel },
                  limitLabel ? { label: limitLabel } : null,
                  upgradeLabel ? { label: upgradeLabel } : null,
                ].filter(Boolean);

                // Show discount tag for quarterly/yearly
                const discountTag = (() => {
                  if (currentPeriod === PERIOD_QUARTERLY && p.quarterly_discount > 0) {
                    return `${t('省')} ${p.quarterly_discount}%`;
                  }
                  if (currentPeriod === PERIOD_YEARLY && p.yearly_discount > 0) {
                    return `${t('省')} ${p.yearly_discount}%`;
                  }
                  return null;
                })();

                return (
                  <Card
                    key={p?.id}
                    className={`!rounded-xl transition-all hover:shadow-lg w-full h-full ${
                      isPopular ? 'ring-2 ring-purple-500' : ''
                    }`}
                    bodyStyle={{ padding: 0 }}
                  >
                    <div className='p-4 h-full flex flex-col'>
                      {/* Tag */}
                      {isPopular && (
                        <div className='mb-2'>
                          <Tag color='purple' shape='circle' size='small'>
                            <Sparkles size={10} className='mr-1' />
                            {p.tag}
                          </Tag>
                        </div>
                      )}
                      {/* Plan title */}
                      <div className='mb-3'>
                        <Typography.Title
                          heading={5}
                          ellipsis={{ rows: 1, showTooltip: true }}
                          style={{ margin: 0 }}
                        >
                          {p.title || t('订阅套餐')}
                        </Typography.Title>
                        {p.subtitle && (
                          <Text
                            type='tertiary'
                            size='small'
                            ellipsis={{ rows: 1, showTooltip: true }}
                            style={{ display: 'block' }}
                          >
                            {p.subtitle}
                          </Text>
                        )}
                      </div>

                      {/* Period selector */}
                      {enabledPeriods.length > 1 && (
                        <div className='mb-3'>
                          <RadioGroup
                            type='button'
                            size='small'
                            value={currentPeriod}
                            onChange={(e) => handlePeriodChange(p.id, e.target.value)}
                            style={{ width: '100%' }}
                          >
                            {enabledPeriods.map((period) => (
                              <Radio key={period} value={period}>
                                {formatPeriodLabel(period, t)}
                              </Radio>
                            ))}
                          </RadioGroup>
                        </div>
                      )}

                      {/* Price area */}
                      <div className='py-2'>
                        <div className='flex items-baseline justify-start gap-1'>
                          <span className='text-xl font-bold text-purple-600'>
                            {symbol}
                          </span>
                          <span className='text-3xl font-bold text-purple-600'>
                            {displayPrice}
                          </span>
                          <span className='text-sm text-gray-400 ml-1'>
                            / {formatPeriodLabel(currentPeriod, t)}
                          </span>
                        </div>
                        {discountTag && (
                          <Tag color='green' size='small' shape='circle' className='mt-1'>
                            {discountTag}
                          </Tag>
                        )}
                      </div>

                      {/* Features from JSON */}
                      {features.length > 0 && (
                        <div className='flex flex-col items-start gap-1 pb-2'>
                          {features.map((item, fi) => (
                            <div
                              key={fi}
                              className='w-full flex justify-start'
                            >
                              <div className='flex items-center gap-2 text-xs text-gray-500'>
                                <Badge dot type='tertiary' />
                                <span>{typeof item === 'string' ? item : item.text || ''}</span>
                              </div>
                            </div>
                          ))}
                        </div>
                      )}

                      {/* Plan benefits */}
                      <div className='flex flex-col items-start gap-1 pb-2'>
                        {planBenefits.map((item) => {
                          const content = (
                            <div className='flex items-center gap-2 text-xs text-gray-500'>
                              <Badge dot type='tertiary' />
                              <span>{item.label}</span>
                            </div>
                          );
                          if (!item.tooltip) {
                            return (
                              <div
                                key={item.label}
                                className='w-full flex justify-start'
                              >
                                {content}
                              </div>
                            );
                          }
                          return (
                            <Tooltip key={item.label} content={item.tooltip}>
                              <div className='w-full flex justify-start'>
                                {content}
                              </div>
                            </Tooltip>
                          );
                        })}
                      </div>

                      <div className='mt-auto'>
                        <Divider margin={12} />

                        {/* Purchase button */}
                        {(() => {
                          const count = getPlanPurchaseCount(p?.id);
                          const reached = limit > 0 && count >= limit;
                          const tip = reached
                            ? t('已达到购买上限') + ` (${count}/${limit})`
                            : '';
                          const buttonEl = (
                            <Button
                              theme='outline'
                              type='primary'
                              block
                              disabled={reached}
                              onClick={() => {
                                if (!reached) openBuy(p, currentPeriod);
                              }}
                            >
                              {reached ? t('已达上限') : t('立即订阅')}
                            </Button>
                          );
                          return reached ? (
                            <Tooltip content={tip} position='top'>
                              {buttonEl}
                            </Tooltip>
                          ) : (
                            buttonEl
                          );
                        })()}
                      </div>
                    </div>
                  </Card>
                );
              })}
            </div>
          ) : (
            <div className='text-center text-gray-400 text-sm py-4'>
              {t('暂无可购买套餐')}
            </div>
          )}
        </Space>
      )}
    </>
  );

  return (
    <>
      {withCard ? (
        <Card className='!rounded-2xl shadow-sm border-0'>{cardContent}</Card>
      ) : (
        <div className='space-y-3'>{cardContent}</div>
      )}

      {/* Purchase confirmation modal */}
      <SubscriptionPurchaseModal
        t={t}
        visible={open}
        onCancel={closeBuy}
        selectedPlan={selectedPlan}
        paying={paying}
        selectedEpayMethod={selectedEpayMethod}
        setSelectedEpayMethod={setSelectedEpayMethod}
        epayMethods={epayMethods}
        enableOnlineTopUp={enableOnlineTopUp}
        enableStripeTopUp={enableStripeTopUp}
        enableCreemTopUp={enableCreemTopUp}
        enableWechatTopUp={enableWechatTopUp}
        enableAlipayTopUp={enableAlipayTopUp}
        purchaseLimitInfo={
          selectedPlan?.plan?.id
            ? {
                limit: Number(selectedPlan?.plan?.max_purchase_per_user || 0),
                count: getPlanPurchaseCount(selectedPlan?.plan?.id),
              }
            : null
        }
        onPayStripe={payStripe}
        onPayCreem={payCreem}
        onPayEpay={payEpay}
        onPayWechat={() => payScanMethod('wechat')}
        onPayAlipay={() => payScanMethod('alipay')}
      />

      {/* Wechat/Alipay scan pay modal */}
      <ScanPayModal
        visible={!!scanPayData}
        onCancel={() => setScanPayData(null)}
        codeUrl={scanPayData?.codeUrl || ''}
        payMoney={scanPayData?.money || 0}
        topUpCount={0}
        paymentMethod={scanPayData?.method || 'wechat'}
      />
    </>
  );
};

export default SubscriptionPlansCard;
