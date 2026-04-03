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

import React from 'react';
import {
  Button,
  Modal,
  Space,
  Tag,
  Typography,
  Badge,
  Tooltip,
} from '@douyinfe/semi-ui';
import { renderQuota } from '../../../helpers';
import { convertUSDToCurrency } from '../../../helpers/render';

const { Text } = Typography;

// ---- Plan variant helpers ----

function formatDuration(plan, t) {
  if (!plan) return '';
  const u = plan.duration_unit || 'month';
  if (u === 'custom') {
    return `${t('自定义')} ${plan.custom_seconds || 0}s`;
  }
  const unitMap = {
    year: t('年'),
    month: t('月'),
    day: t('日'),
    hour: t('小时'),
  };
  return `${plan.duration_value || 0}${unitMap[u] || u}`;
}

function formatResetPeriod(plan, t) {
  const period = plan?.quota_reset_period || 'never';
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

const renderEnabled = (enabled, t) => {
  return enabled ? (
    <Tag
      color='white'
      shape='circle'
      type='light'
      prefixIcon={<Badge dot type='success' />}
    >
      {t('启用')}
    </Tag>
  ) : (
    <Tag
      color='white'
      shape='circle'
      type='light'
      prefixIcon={<Badge dot type='danger' />}
    >
      {t('禁用')}
    </Tag>
  );
};

// ---- Group columns ----

export const getGroupColumns = ({
  t,
  openEditGroup,
  setGroupEnabled,
  deleteGroup,
}) => {
  return [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 60,
      render: (text) => <Text type='tertiary'>#{text}</Text>,
    },
    {
      title: t('套餐组'),
      dataIndex: 'title',
      width: 200,
      render: (text, record) => (
        <div style={{ maxWidth: 180 }}>
          <Text strong ellipsis={{ showTooltip: true }}>
            {text}
          </Text>
          {record.subtitle && (
            <Text
              type='tertiary'
              ellipsis={{ showTooltip: true }}
              style={{ display: 'block', fontSize: 12 }}
            >
              {record.subtitle}
            </Text>
          )}
        </div>
      ),
    },
    {
      title: t('标签'),
      dataIndex: 'tag',
      width: 100,
      render: (text) =>
        text ? (
          <Tag color='blue' shape='circle'>
            {text}
          </Tag>
        ) : (
          <Text type='tertiary'>-</Text>
        ),
    },
    {
      title: t('计费周期'),
      dataIndex: 'plans',
      width: 100,
      render: (plans) => (
        <Text type='secondary'>{(plans || []).length} {t('个')}</Text>
      ),
    },
    {
      title: t('排序'),
      dataIndex: 'sort_order',
      width: 80,
      render: (text) => <Text type='tertiary'>{Number(text || 0)}</Text>,
    },
    {
      title: t('状态'),
      dataIndex: 'enabled',
      width: 80,
      render: (text) => renderEnabled(text, t),
    },
    {
      title: t('操作'),
      dataIndex: 'operate',
      fixed: 'right',
      width: 200,
      render: (_, record) => {
        const isEnabled = record.enabled;
        const handleToggle = () => {
          if (isEnabled) {
            Modal.confirm({
              title: t('确认禁用'),
              content: t('禁用后该套餐组下所有计费周期均不会在用户端展示。是否继续？'),
              centered: true,
              onOk: () => setGroupEnabled(record.id, false),
            });
          } else {
            Modal.confirm({
              title: t('确认启用'),
              content: t('启用后该套餐组将在用户端展示。是否继续？'),
              centered: true,
              onOk: () => setGroupEnabled(record.id, true),
            });
          }
        };

        const handleDelete = () => {
          Modal.confirm({
            title: t('确认删除'),
            content: t('删除套餐组将同时删除其下所有计费周期。此操作不可恢复，是否继续？'),
            centered: true,
            type: 'danger',
            onOk: () => deleteGroup(record.id),
          });
        };

        return (
          <Space spacing={8}>
            <Button
              theme='light'
              type='tertiary'
              size='small'
              onClick={() => openEditGroup(record)}
            >
              {t('编辑')}
            </Button>
            {isEnabled ? (
              <Button
                theme='light'
                type='danger'
                size='small'
                onClick={handleToggle}
              >
                {t('禁用')}
              </Button>
            ) : (
              <Button
                theme='light'
                type='primary'
                size='small'
                onClick={handleToggle}
              >
                {t('启用')}
              </Button>
            )}
            <Button
              theme='light'
              type='danger'
              size='small'
              onClick={handleDelete}
            >
              {t('删除')}
            </Button>
          </Space>
        );
      },
    },
  ];
};

// ---- Plan variant columns (used in expanded row) ----

export const getPlanVariantColumns = ({
  t,
  openEditPlan,
  setPlanEnabled,
  enableEpay,
}) => {
  return [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 60,
      render: (text) => <Text type='tertiary'>#{text}</Text>,
    },
    {
      title: t('周期'),
      width: 100,
      render: (_, record) => (
        <Text type='secondary'>{formatDuration(record, t)}</Text>
      ),
    },
    {
      title: t('价格'),
      dataIndex: 'price_amount',
      width: 100,
      render: (text) => (
        <Text strong style={{ color: 'var(--semi-color-success)' }}>
          {convertUSDToCurrency(Number(text || 0), 2)}
        </Text>
      ),
    },
    {
      title: t('总额度'),
      dataIndex: 'total_amount',
      width: 100,
      render: (text) => {
        const total = Number(text || 0);
        return (
          <Text type={total > 0 ? 'secondary' : 'tertiary'}>
            {total > 0 ? (
              <Tooltip content={`${t('原生额度')}：${total}`}>
                <span>{renderQuota(total)}</span>
              </Tooltip>
            ) : (
              t('不限')
            )}
          </Text>
        );
      },
    },
    {
      title: t('升级分组'),
      dataIndex: 'upgrade_group',
      width: 100,
      render: (text) => (
        <Text type={text ? 'secondary' : 'tertiary'}>
          {text || t('不升级')}
        </Text>
      ),
    },
    {
      title: t('购买上限'),
      dataIndex: 'max_purchase_per_user',
      width: 90,
      render: (text) => {
        const limit = Number(text || 0);
        return (
          <Text type={limit > 0 ? 'secondary' : 'tertiary'}>
            {limit > 0 ? limit : t('不限')}
          </Text>
        );
      },
    },
    {
      title: t('重置'),
      width: 80,
      render: (_, record) => {
        const period = record?.quota_reset_period || 'never';
        const isNever = period === 'never';
        return (
          <Text type={isNever ? 'tertiary' : 'secondary'}>
            {formatResetPeriod(record, t)}
          </Text>
        );
      },
    },
    {
      title: t('支付渠道'),
      width: 180,
      render: (_, record) => {
        const hasStripe = !!record?.stripe_price_id;
        const hasCreem = !!record?.creem_product_id;
        const hasEpay = !!enableEpay;
        return (
          <Space spacing={4}>
            {hasStripe && (
              <Tag color='violet' shape='circle'>
                Stripe
              </Tag>
            )}
            {hasCreem && (
              <Tag color='cyan' shape='circle'>
                Creem
              </Tag>
            )}
            {hasEpay && (
              <Tag color='light-green' shape='circle'>
                {t('易支付')}
              </Tag>
            )}
          </Space>
        );
      },
    },
    {
      title: t('状态'),
      dataIndex: 'enabled',
      width: 80,
      render: (text) => renderEnabled(text, t),
    },
    {
      title: t('操作'),
      dataIndex: 'operate',
      fixed: 'right',
      width: 140,
      render: (_, record) => {
        const isEnabled = record.enabled;
        const handleToggle = () => {
          if (isEnabled) {
            Modal.confirm({
              title: t('确认禁用'),
              content: t('禁用后该计费周期不再展示。是否继续？'),
              centered: true,
              onOk: () => setPlanEnabled(record.id, false),
            });
          } else {
            Modal.confirm({
              title: t('确认启用'),
              content: t('启用后该计费周期将展示给用户。是否继续？'),
              centered: true,
              onOk: () => setPlanEnabled(record.id, true),
            });
          }
        };

        return (
          <Space spacing={8}>
            <Button
              theme='light'
              type='tertiary'
              size='small'
              onClick={() => openEditPlan(record)}
            >
              {t('编辑')}
            </Button>
            {isEnabled ? (
              <Button
                theme='light'
                type='danger'
                size='small'
                onClick={handleToggle}
              >
                {t('禁用')}
              </Button>
            ) : (
              <Button
                theme='light'
                type='primary'
                size='small'
                onClick={handleToggle}
              >
                {t('启用')}
              </Button>
            )}
          </Space>
        );
      },
    },
  ];
};
