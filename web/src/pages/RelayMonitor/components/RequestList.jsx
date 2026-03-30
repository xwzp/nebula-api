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
import { Tag } from '@douyinfe/semi-ui';

function formatTimestamp(ts) {
  const d = new Date(ts);
  return d.toLocaleTimeString('zh-CN', { hour12: false });
}

function formatDuration(ns) {
  // duration comes as nanoseconds
  const ms = ns / 1e6;
  if (ms < 1000) return `${ms.toFixed(0)}ms`;
  return `${(ms / 1000).toFixed(1)}s`;
}

function getStatusColor(trace) {
  if (!trace.upstream_response) return 'grey';
  const code = trace.upstream_response.status_code;
  if (code >= 200 && code < 300) return 'green';
  if (code >= 400 && code < 500) return 'orange';
  if (code >= 500) return 'red';
  return 'grey';
}

export default function RequestList({ traces, selectedIndex, onSelect }) {
  if (traces.length === 0) {
    return (
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          height: '100%',
          color: 'var(--semi-color-text-2)',
          fontSize: 14,
        }}
      >
        No traces captured yet
      </div>
    );
  }

  return (
    <div style={{ overflow: 'auto', height: '100%' }}>
      {traces.map((trace, index) => (
        <div
          key={trace.trace_id || index}
          onClick={() => onSelect(index)}
          style={{
            padding: '10px 14px',
            cursor: 'pointer',
            borderBottom: '1px solid var(--semi-color-border)',
            backgroundColor:
              selectedIndex === index
                ? 'var(--semi-color-primary-light-default)'
                : 'transparent',
            transition: 'background-color 0.15s',
          }}
          onMouseEnter={(e) => {
            if (selectedIndex !== index) {
              e.currentTarget.style.backgroundColor = 'var(--semi-color-fill-0)';
            }
          }}
          onMouseLeave={(e) => {
            if (selectedIndex !== index) {
              e.currentTarget.style.backgroundColor = 'transparent';
            }
          }}
        >
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '8px',
              marginBottom: '4px',
            }}
          >
            <span style={{ fontFamily: 'monospace', fontSize: 12, color: 'var(--semi-color-text-2)' }}>
              {formatTimestamp(trace.timestamp)}
            </span>
            <Tag size='small' color={getStatusColor(trace)}>
              {trace.upstream_response?.status_code || '...'}
            </Tag>
            {trace.is_stream && (
              <Tag size='small' color='cyan'>
                SSE
              </Tag>
            )}
          </div>
          <div
            style={{
              fontSize: 13,
              fontWeight: 500,
              whiteSpace: 'nowrap',
              overflow: 'hidden',
              textOverflow: 'ellipsis',
            }}
          >
            {trace.model_name}
          </div>
          <div
            style={{
              fontSize: 12,
              color: 'var(--semi-color-text-2)',
              display: 'flex',
              gap: '8px',
            }}
          >
            <span>ch:{trace.channel_id}</span>
            {trace.duration > 0 && <span>{formatDuration(trace.duration)}</span>}
            <span style={{ marginLeft: 'auto', fontFamily: 'monospace' }}>
              {trace.trace_id?.substring(0, 8)}
            </span>
          </div>
        </div>
      ))}
    </div>
  );
}
