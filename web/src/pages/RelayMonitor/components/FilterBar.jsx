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
import { Input, InputNumber, Select, Button, Tag, Space } from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';

const TIMEOUT_OPTIONS = [
  { value: 60, label: '1 min' },
  { value: 300, label: '5 min' },
  { value: 600, label: '10 min' },
  { value: 1800, label: '30 min' },
];

export default function FilterBar({
  filters,
  setFilters,
  timeout,
  setTimeout,
  isMonitoring,
  isConnected,
  remainingTime,
  traceCount,
  onStart,
  onStop,
  onRenew,
  onClear,
  onConnect,
  onDisconnect,
}) {
  const { t } = useTranslation();

  const formatTime = (seconds) => {
    const m = Math.floor(seconds / 60);
    const s = seconds % 60;
    return `${m}:${s.toString().padStart(2, '0')}`;
  };

  return (
    <div
      style={{
        padding: '16px',
        borderBottom: '1px solid var(--semi-color-border)',
        display: 'flex',
        flexDirection: 'column',
        gap: '12px',
      }}
    >
      {/* Filter row */}
      <div style={{ display: 'flex', gap: '12px', flexWrap: 'wrap', alignItems: 'center' }}>
        <InputNumber
          prefix='Token ID'
          placeholder={t('全部')}
          value={filters.token_id || undefined}
          onChange={(val) => setFilters((f) => ({ ...f, token_id: val || 0 }))}
          disabled={isMonitoring}
          style={{ width: 160 }}
          min={0}
          hideButtons
        />
        <InputNumber
          prefix='User ID'
          placeholder={t('全部')}
          value={filters.user_id || undefined}
          onChange={(val) => setFilters((f) => ({ ...f, user_id: val || 0 }))}
          disabled={isMonitoring}
          style={{ width: 160 }}
          min={0}
          hideButtons
        />
        <Input
          prefix='Model'
          placeholder={t('全部')}
          value={filters.model_name || ''}
          onChange={(val) => setFilters((f) => ({ ...f, model_name: val || '' }))}
          disabled={isMonitoring}
          style={{ width: 220 }}
        />
        <InputNumber
          prefix='Channel'
          placeholder={t('全部')}
          value={filters.channel_id || undefined}
          onChange={(val) => setFilters((f) => ({ ...f, channel_id: val || 0 }))}
          disabled={isMonitoring}
          style={{ width: 160 }}
          min={0}
          hideButtons
        />
        <Select
          value={timeout}
          onChange={setTimeout}
          disabled={isMonitoring}
          style={{ width: 110 }}
          optionList={TIMEOUT_OPTIONS}
        />
      </div>

      {/* Action row */}
      <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
        {!isConnected ? (
          <Button theme='solid' onClick={onConnect}>
            {t('连接')}
          </Button>
        ) : !isMonitoring ? (
          <>
            <Button theme='solid' type='primary' onClick={onStart}>
              {t('开始监控')}
            </Button>
            <Button theme='light' onClick={onDisconnect}>
              {t('断开')}
            </Button>
          </>
        ) : (
          <>
            <Button theme='solid' type='danger' onClick={onStop}>
              {t('停止监控')}
            </Button>
            <Button theme='light' onClick={onRenew}>
              {t('续期')}
            </Button>
          </>
        )}

        {traceCount > 0 && (
          <Button theme='borderless' onClick={onClear}>
            {t('清空')}
          </Button>
        )}

        <Space style={{ marginLeft: 'auto' }}>
          {isConnected && (
            <Tag color='green' size='large'>
              {t('已连接')}
            </Tag>
          )}
          {isMonitoring && remainingTime > 0 && (
            <Tag color='orange' size='large'>
              {t('剩余')} {formatTime(remainingTime)}
            </Tag>
          )}
          {traceCount > 0 && (
            <Tag color='blue' size='large'>
              {traceCount} {t('条')}
            </Tag>
          )}
        </Space>
      </div>
    </div>
  );
}
