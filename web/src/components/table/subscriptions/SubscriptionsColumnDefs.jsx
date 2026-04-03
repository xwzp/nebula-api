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

export const getColumns = ({
  t,
  openEditPlan,
  setPlanEnabled,
  deletePlan,
}) => {
  return [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 60,
      render: (text) => <Text type='tertiary'>#{text}</Text>,
    },
    {
      title: t('标题'),
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
      title: t('月基准价'),
      dataIndex: 'price_monthly',
      width: 120,
      render: (text, record) => (
        <Text strong style={{ color: 'var(--semi-color-success)' }}>
          {convertUSDToCurrency(Number(text || 0), 2)}
          {record.currency && record.currency !== 'USD' && (
            <Text type='tertiary' style={{ fontSize: 11, marginLeft: 4 }}>
              {record.currency}
            </Text>
          )}
        </Text>
      ),
    },
    {
      title: t('付款周期'),
      width: 220,
      render: (_, record) => {
        const periods = [];
        if (record.monthly_enabled) {
          periods.push({ label: t('月付'), color: 'blue' });
        }
        if (record.quarterly_enabled) {
          const discount = Number(record.quarterly_discount || 0);
          periods.push({
            label: discount > 0 ? `${t('季付')} -${discount}%` : t('季付'),
            color: 'green',
          });
        }
        if (record.yearly_enabled) {
          const discount = Number(record.yearly_discount || 0);
          periods.push({
            label: discount > 0 ? `${t('年付')} -${discount}%` : t('年付'),
            color: 'orange',
          });
        }
        if (periods.length === 0) {
          return <Text type='tertiary'>-</Text>;
        }
        return (
          <Space spacing={4}>
            {periods.map((p, i) => (
              <Tag key={i} color={p.color} shape='circle' size='small'>
                {p.label}
              </Tag>
            ))}
          </Space>
        );
      },
    },
    {
      title: t('额度'),
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
              content: t('禁用后该订阅套餐不再展示给用户。是否继续？'),
              centered: true,
              onOk: () => setPlanEnabled(record.id, false),
            });
          } else {
            Modal.confirm({
              title: t('确认启用'),
              content: t('启用后该订阅套餐将展示给用户。是否继续？'),
              centered: true,
              onOk: () => setPlanEnabled(record.id, true),
            });
          }
        };

        const handleDelete = () => {
          Modal.confirm({
            title: t('确认删除'),
            content: t('删除后该订阅套餐将不可恢复。如有用户正在使用此套餐的活跃订阅，将无法删除。是否继续？'),
            centered: true,
            type: 'danger',
            onOk: () => deletePlan(record.id),
          });
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
