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

import React, { useState, useCallback } from 'react';
import FilterBar from './FilterBar';
import RequestList from './RequestList';
import RequestDetail from './RequestDetail';
import { useRelayMonitor } from './useRelayMonitor';

export default function MonitorPanel() {
  const {
    isConnected,
    isMonitoring,
    traces,
    remainingTime,
    error,
    connect,
    disconnect,
    startMonitoring,
    stopMonitoring,
    renewTimeout,
    clearTraces,
  } = useRelayMonitor();

  const [filters, setFilters] = useState({
    token_id: 0,
    user_id: 0,
    model_name: '',
    channel_id: 0,
  });
  const [timeout, setTimeout] = useState(300);
  const [selectedIndex, setSelectedIndex] = useState(null);

  const handleStart = useCallback(() => {
    startMonitoring(filters, timeout);
  }, [startMonitoring, filters, timeout]);

  const handleRenew = useCallback(() => {
    renewTimeout(timeout);
  }, [renewTimeout, timeout]);

  const handleClear = useCallback(() => {
    clearTraces();
    setSelectedIndex(null);
  }, [clearTraces]);

  const selectedTrace = selectedIndex !== null ? traces[selectedIndex] : null;

  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        height: 'calc(100vh - 80px)',
        border: '1px solid var(--semi-color-border)',
        borderRadius: 8,
        overflow: 'hidden',
        backgroundColor: 'var(--semi-color-bg-0)',
      }}
    >
      <FilterBar
        filters={filters}
        setFilters={setFilters}
        timeout={timeout}
        setTimeout={setTimeout}
        isMonitoring={isMonitoring}
        isConnected={isConnected}
        remainingTime={remainingTime}
        traceCount={traces.length}
        onStart={handleStart}
        onStop={stopMonitoring}
        onRenew={handleRenew}
        onClear={handleClear}
        onConnect={connect}
        onDisconnect={disconnect}
      />

      {error && (
        <div
          style={{
            padding: '8px 16px',
            backgroundColor: 'var(--semi-color-danger-light-default)',
            color: 'var(--semi-color-danger)',
            fontSize: 13,
          }}
        >
          {error}
        </div>
      )}

      <div style={{ display: 'flex', flex: 1, minHeight: 0 }}>
        {/* Left: Request List */}
        <div
          style={{
            width: 320,
            minWidth: 280,
            borderRight: '1px solid var(--semi-color-border)',
            overflow: 'hidden',
          }}
        >
          <RequestList
            traces={traces}
            selectedIndex={selectedIndex}
            onSelect={setSelectedIndex}
          />
        </div>

        {/* Right: Request Detail */}
        <div style={{ flex: 1, overflow: 'hidden' }}>
          <RequestDetail trace={selectedTrace} />
        </div>
      </div>
    </div>
  );
}
