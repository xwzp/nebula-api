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

import React, { useState } from 'react';
import { Collapse, Tag, Typography, Empty } from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';

const { Text } = Typography;

function HeadersTable({ headers }) {
  if (!headers || Object.keys(headers).length === 0) {
    return <Text type='tertiary'>No headers</Text>;
  }

  return (
    <div style={{ overflowX: 'auto' }}>
      <table
        style={{
          width: '100%',
          borderCollapse: 'collapse',
          fontSize: 12,
          fontFamily: 'monospace',
        }}
      >
        <tbody>
          {Object.entries(headers).map(([key, value]) => (
            <tr
              key={key}
              style={{ borderBottom: '1px solid var(--semi-color-border)' }}
            >
              <td
                style={{
                  padding: '4px 8px',
                  fontWeight: 600,
                  whiteSpace: 'nowrap',
                  verticalAlign: 'top',
                  width: '1%',
                  color: 'var(--semi-color-text-1)',
                }}
              >
                {key}
              </td>
              <td
                style={{
                  padding: '4px 8px',
                  wordBreak: 'break-all',
                  color:
                    value === '[REDACTED]'
                      ? 'var(--semi-color-danger)'
                      : 'var(--semi-color-text-2)',
                }}
              >
                {value}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function BodyDisplay({ body, bodyLen, truncated }) {
  const [expanded, setExpanded] = useState(false);
  const { t } = useTranslation();

  if (!body) {
    return <Text type='tertiary'>No body</Text>;
  }

  // Try to format as JSON
  let formatted = body;
  let isJson = false;
  try {
    const parsed = JSON.parse(body);
    formatted = JSON.stringify(parsed, null, 2);
    isJson = true;
  } catch {
    // not JSON, display as-is
  }

  const displayContent = expanded ? formatted : formatted.substring(0, 2000);
  const showExpandButton = formatted.length > 2000 && !expanded;

  return (
    <div>
      {truncated && (
        <Tag color='amber' size='small' style={{ marginBottom: 8 }}>
          {t('已截断')} ({bodyLen} bytes, {t('显示前')} {body.length} bytes)
        </Tag>
      )}
      <pre
        style={{
          margin: 0,
          padding: '8px',
          fontSize: 12,
          fontFamily: 'monospace',
          backgroundColor: 'var(--semi-color-fill-0)',
          borderRadius: 4,
          overflow: 'auto',
          maxHeight: expanded ? 'none' : 400,
          whiteSpace: 'pre-wrap',
          wordBreak: 'break-all',
          lineHeight: 1.5,
        }}
      >
        {displayContent}
      </pre>
      {showExpandButton && (
        <Text
          link
          onClick={() => setExpanded(true)}
          style={{ fontSize: 12, marginTop: 4, display: 'inline-block' }}
        >
          {t('展开全部')}
        </Text>
      )}
    </div>
  );
}

function StagePanel({ title, tag, captured }) {
  const { t } = useTranslation();

  if (!captured) {
    return (
      <Collapse.Panel header={title} itemKey={title}>
        <Text type='tertiary'>{t('暂无数据')}</Text>
      </Collapse.Panel>
    );
  }

  const extra = (
    <div style={{ display: 'flex', gap: 6, alignItems: 'center' }}>
      {captured.method && (
        <Tag size='small' color='blue'>
          {captured.method}
        </Tag>
      )}
      {captured.status_code > 0 && (
        <Tag
          size='small'
          color={captured.status_code < 300 ? 'green' : captured.status_code < 500 ? 'orange' : 'red'}
        >
          {captured.status_code}
        </Tag>
      )}
      {tag}
    </div>
  );

  return (
    <Collapse.Panel header={title} itemKey={title} extra={extra}>
      {captured.url && (
        <div style={{ marginBottom: 8 }}>
          <Text strong style={{ fontSize: 12 }}>
            URL:{' '}
          </Text>
          <Text
            copyable
            style={{
              fontSize: 12,
              fontFamily: 'monospace',
              wordBreak: 'break-all',
            }}
          >
            {captured.url}
          </Text>
        </div>
      )}

      <div style={{ marginBottom: 12 }}>
        <Text strong style={{ fontSize: 13, marginBottom: 4, display: 'block' }}>
          Headers
        </Text>
        <HeadersTable headers={captured.headers} />
      </div>

      <div>
        <Text strong style={{ fontSize: 13, marginBottom: 4, display: 'block' }}>
          Body
        </Text>
        <BodyDisplay
          body={captured.body}
          bodyLen={captured.body_len}
          truncated={captured.truncated}
        />
      </div>
    </Collapse.Panel>
  );
}

export default function RequestDetail({ trace }) {
  const { t } = useTranslation();

  if (!trace) {
    return (
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          height: '100%',
        }}
      >
        <Empty description={t('选择一个请求查看详情')} />
      </div>
    );
  }

  const stages = [
    {
      title: `1. ${t('客户端')} \u2192 ${t('网关')}`,
      captured: trace.client_request,
    },
    {
      title: `2. ${t('网关')} \u2192 ${t('上游')}`,
      captured: trace.upstream_request,
    },
    {
      title: `3. ${t('上游')} \u2192 ${t('网关')}`,
      captured: trace.upstream_response,
    },
    {
      title: `4. ${t('网关')} \u2192 ${t('客户端')}`,
      captured: trace.client_response,
    },
  ];

  return (
    <div style={{ padding: '12px', overflow: 'auto', height: '100%' }}>
      <div
        style={{
          marginBottom: 12,
          display: 'flex',
          gap: 8,
          alignItems: 'center',
          flexWrap: 'wrap',
        }}
      >
        <Tag color='blue'>{trace.model_name}</Tag>
        <Tag color='grey'>ch:{trace.channel_id}</Tag>
        <Tag color='grey'>token:{trace.token_id}</Tag>
        <Tag color='grey'>user:{trace.user_id}</Tag>
        {trace.is_stream && (
          <Tag color='cyan'>
            SSE {trace.stream_event_count > 0 ? `(${trace.stream_event_count} events)` : ''}
          </Tag>
        )}
        <Text
          copyable
          style={{ fontSize: 12, fontFamily: 'monospace', marginLeft: 'auto' }}
        >
          {trace.trace_id}
        </Text>
      </div>

      <Collapse defaultActiveKey={stages.map((s) => s.title)} keepDOM={false}>
        {stages.map((stage) => (
          <StagePanel
            key={stage.title}
            title={stage.title}
            captured={stage.captured}
          />
        ))}
      </Collapse>
    </div>
  );
}
