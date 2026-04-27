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

import React, { useState, useMemo, useCallback } from 'react';
import {
  Collapse,
  Tag,
  Typography,
  Empty,
  Toast,
  Button,
  Tooltip,
} from '@douyinfe/semi-ui';
import {
  IconCopy,
  IconDownload,
  IconChevronDown,
  IconChevronRight,
} from '@douyinfe/semi-icons';
import { useTranslation } from 'react-i18next';

const { Text } = Typography;

function copyToClipboard(text, t) {
  navigator.clipboard
    .writeText(text)
    .then(() => {
      Toast.success(t('已复制'));
    })
    .catch(() => {
      Toast.error(t('复制失败'));
    });
}

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

// Preview a JSON value: short string for primitives, summary for objects/arrays
function valuePreview(value) {
  if (value === null) return 'null';
  if (value === undefined) return 'undefined';
  if (typeof value === 'string') {
    if (value.length <= 60) return JSON.stringify(value);
    return JSON.stringify(value.substring(0, 57) + '...');
  }
  if (typeof value === 'number' || typeof value === 'boolean') {
    return String(value);
  }
  if (Array.isArray(value)) {
    return `Array(${value.length})`;
  }
  if (typeof value === 'object') {
    const keys = Object.keys(value);
    return `Object{${keys.length} keys}`;
  }
  return String(value);
}

// Tag color based on value type
function typeTagColor(value) {
  if (value === null || value === undefined) return 'grey';
  if (typeof value === 'string') return 'green';
  if (typeof value === 'number') return 'blue';
  if (typeof value === 'boolean') return 'amber';
  if (Array.isArray(value)) return 'cyan';
  if (typeof value === 'object') return 'violet';
  return 'grey';
}

function typeLabel(value) {
  if (value === null) return 'null';
  if (Array.isArray(value)) return 'array';
  return typeof value;
}

// Collapsible JSON key section
function JsonKeySection({ keyName, value, t }) {
  const [expanded, setExpanded] = useState(false);
  const isExpandable = typeof value === 'object' && value !== null;
  const formatted = useMemo(() => JSON.stringify(value, null, 2), [value]);

  const copyText = useMemo(() => {
    return JSON.stringify({ [keyName]: value }, null, 2)
      .slice(1, -2)
      .trim();
  }, [keyName, value]);

  const handleCopy = useCallback(
    (e) => {
      e.stopPropagation();
      copyToClipboard(copyText, t);
    },
    [copyText, t],
  );

  return (
    <div
      style={{
        border: '1px solid var(--semi-color-border)',
        borderRadius: 6,
        marginBottom: 6,
        overflow: 'hidden',
      }}
    >
      {/* Key header row */}
      <div
        onClick={() => isExpandable && setExpanded(!expanded)}
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 8,
          padding: '6px 10px',
          backgroundColor: 'var(--semi-color-fill-0)',
          cursor: isExpandable ? 'pointer' : 'default',
          userSelect: 'none',
        }}
      >
        {isExpandable ? (
          expanded ? (
            <IconChevronDown
              size='small'
              style={{ flexShrink: 0, color: 'var(--semi-color-text-2)' }}
            />
          ) : (
            <IconChevronRight
              size='small'
              style={{ flexShrink: 0, color: 'var(--semi-color-text-2)' }}
            />
          )
        ) : (
          <span style={{ width: 16, flexShrink: 0 }} />
        )}

        <Text
          strong
          style={{
            fontSize: 13,
            fontFamily: 'monospace',
            color: 'var(--semi-color-primary)',
          }}
        >
          {keyName}
        </Text>

        <Tag size='small' color={typeTagColor(value)} style={{ flexShrink: 0 }}>
          {typeLabel(value)}
        </Tag>

        {!expanded && (
          <Text
            type='tertiary'
            style={{
              fontSize: 12,
              fontFamily: 'monospace',
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              whiteSpace: 'nowrap',
              flex: 1,
              minWidth: 0,
            }}
          >
            {valuePreview(value)}
          </Text>
        )}

        <Tooltip content={t('复制此字段')} position='top'>
          <Button
            icon={<IconCopy />}
            size='small'
            theme='borderless'
            type='tertiary'
            onClick={handleCopy}
            style={{ flexShrink: 0, marginLeft: 'auto' }}
          />
        </Tooltip>
      </div>

      {/* Expanded content */}
      {expanded && (
        <pre
          style={{
            margin: 0,
            padding: '8px 10px',
            fontSize: 12,
            fontFamily: 'monospace',
            backgroundColor: 'var(--semi-color-bg-1)',
            overflow: 'auto',
            maxHeight: 500,
            whiteSpace: 'pre-wrap',
            wordBreak: 'break-all',
            lineHeight: 1.5,
            borderTop: '1px solid var(--semi-color-border)',
          }}
        >
          {formatted}
        </pre>
      )}

      {/* Inline display for primitives */}
      {!isExpandable && !expanded && null}
    </div>
  );
}

function BodyDisplay({ body, bodyLen, truncated }) {
  const { t } = useTranslation();
  const [expanded, setExpanded] = useState(false);

  if (!body) {
    return <Text type='tertiary'>No body</Text>;
  }

  // Try to parse as JSON
  let parsed = null;
  let isJson = false;
  try {
    parsed = JSON.parse(body);
    isJson = true;
  } catch {
    // not JSON
  }

  // If JSON object, render collapsible key sections
  if (
    isJson &&
    typeof parsed === 'object' &&
    parsed !== null &&
    !Array.isArray(parsed)
  ) {
    const keys = Object.keys(parsed);
    const fullFormatted = JSON.stringify(parsed, null, 2);

    return (
      <div>
        {truncated && (
          <Tag color='amber' size='small' style={{ marginBottom: 8 }}>
            {t('已截断')} ({bodyLen} bytes, {t('显示前')} {body.length} bytes)
          </Tag>
        )}

        {/* Toolbar: copy all */}
        <div style={{ display: 'flex', gap: 6, marginBottom: 8 }}>
          <Button
            icon={<IconCopy />}
            size='small'
            theme='light'
            onClick={() => copyToClipboard(fullFormatted, t)}
          >
            {t('复制全部')}
          </Button>
          <Text type='tertiary' style={{ fontSize: 12, lineHeight: '32px' }}>
            {keys.length} {t('个字段')}
          </Text>
        </div>

        {/* Collapsible key sections */}
        {keys.map((key) => (
          <JsonKeySection key={key} keyName={key} value={parsed[key]} t={t} />
        ))}
      </div>
    );
  }

  // Fallback: non-JSON or JSON array — show raw
  let formatted = body;
  if (isJson) {
    formatted = JSON.stringify(parsed, null, 2);
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
      <div style={{ marginBottom: 8 }}>
        <Button
          icon={<IconCopy />}
          size='small'
          theme='light'
          onClick={() => copyToClipboard(formatted, t)}
        >
          {t('复制全部')}
        </Button>
      </div>
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
          color={
            captured.status_code < 300
              ? 'green'
              : captured.status_code < 500
                ? 'orange'
                : 'red'
          }
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
        <Text
          strong
          style={{ fontSize: 13, marginBottom: 4, display: 'block' }}
        >
          Headers
        </Text>
        <HeadersTable headers={captured.headers} />
      </div>

      <div>
        <Text
          strong
          style={{ fontSize: 13, marginBottom: 4, display: 'block' }}
        >
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

function downloadTraceAsJson(trace, t) {
  const data = {
    trace_id: trace.trace_id,
    timestamp: trace.timestamp,
    model_name: trace.model_name,
    channel_id: trace.channel_id,
    token_id: trace.token_id,
    user_id: trace.user_id,
    is_stream: trace.is_stream,
    stream_event_count: trace.stream_event_count,
    stages: {},
  };

  // Parse body strings back to JSON objects for cleaner output
  const parseBody = (captured) => {
    if (!captured) return null;
    const result = { ...captured };
    if (result.body) {
      try {
        result.body = JSON.parse(result.body);
      } catch {
        // keep as string
      }
    }
    return result;
  };

  data.stages.client_request = parseBody(trace.client_request);
  data.stages.upstream_request = parseBody(trace.upstream_request);
  data.stages.upstream_response = parseBody(trace.upstream_response);
  data.stages.client_response = parseBody(trace.client_response);

  const jsonStr = JSON.stringify(data, null, 2);
  const blob = new Blob([jsonStr], { type: 'application/json' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  const ts = trace.timestamp
    ? new Date(trace.timestamp)
        .toISOString()
        .replace(/[:.]/g, '-')
        .substring(0, 19)
    : 'unknown';
  a.download = `trace_${trace.model_name || 'unknown'}_${ts}.json`;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
  Toast.success(t('下载成功'));
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
            SSE{' '}
            {trace.stream_event_count > 0
              ? `(${trace.stream_event_count} events)`
              : ''}
          </Tag>
        )}
        <Text copyable style={{ fontSize: 12, fontFamily: 'monospace' }}>
          {trace.trace_id}
        </Text>
        <Button
          icon={<IconDownload />}
          size='small'
          theme='light'
          type='primary'
          onClick={() => downloadTraceAsJson(trace, t)}
          style={{ marginLeft: 'auto' }}
        >
          {t('下载')}
        </Button>
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
