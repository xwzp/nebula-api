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

import { useState, useRef, useCallback, useEffect } from 'react';

const WS_RECONNECT_DELAY = 3000;

export function useRelayMonitor() {
  const [isConnected, setIsConnected] = useState(false);
  const [isMonitoring, setIsMonitoring] = useState(false);
  const [traces, setTraces] = useState([]);
  const [remainingTime, setRemainingTime] = useState(0);
  const [error, setError] = useState(null);

  const wsRef = useRef(null);
  const reconnectTimerRef = useRef(null);
  const countdownRef = useRef(null);
  const shouldReconnectRef = useRef(false);

  // Countdown timer
  useEffect(() => {
    if (isMonitoring && remainingTime > 0) {
      countdownRef.current = setInterval(() => {
        setRemainingTime((prev) => {
          if (prev <= 1) {
            setIsMonitoring(false);
            return 0;
          }
          return prev - 1;
        });
      }, 1000);
    }
    return () => {
      if (countdownRef.current) {
        clearInterval(countdownRef.current);
      }
    };
  }, [isMonitoring, remainingTime > 0]);

  const buildWsUrl = useCallback(() => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    return `${protocol}//${host}/api/relay-monitor/ws`;
  }, []);

  const handleMessage = useCallback((event) => {
    try {
      const msg = JSON.parse(event.data);
      switch (msg.type) {
        case 'trace':
          setTraces((prev) => {
            const next = [msg.data, ...prev];
            // Keep at most 200 traces in frontend
            if (next.length > 200) next.length = 200;
            return next;
          });
          break;
        case 'status':
          if (msg.data) {
            setIsMonitoring(msg.data.monitoring);
            if (msg.data.remaining) {
              setRemainingTime(msg.data.remaining);
            }
          }
          break;
        case 'timeout':
          setIsMonitoring(false);
          setRemainingTime(0);
          break;
        case 'error':
          setError(msg.data?.message || 'Unknown error');
          break;
        default:
          break;
      }
    } catch {
      // ignore parse errors
    }
  }, []);

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) return;

    shouldReconnectRef.current = true;
    setError(null);

    // Session cookie is sent automatically by the browser for same-origin WebSocket
    const ws = new WebSocket(buildWsUrl());

    ws.onopen = () => {
      setIsConnected(true);
      setError(null);
    };

    ws.onmessage = handleMessage;

    ws.onclose = () => {
      setIsConnected(false);
      setIsMonitoring(false);
      wsRef.current = null;

      if (shouldReconnectRef.current) {
        reconnectTimerRef.current = setTimeout(() => {
          connect();
        }, WS_RECONNECT_DELAY);
      }
    };

    ws.onerror = () => {
      setError('WebSocket connection failed');
    };

    wsRef.current = ws;
  }, [buildWsUrl, handleMessage]);

  const disconnect = useCallback(() => {
    shouldReconnectRef.current = false;
    if (reconnectTimerRef.current) {
      clearTimeout(reconnectTimerRef.current);
    }
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    setIsConnected(false);
    setIsMonitoring(false);
    setRemainingTime(0);
  }, []);

  const sendCommand = useCallback(
    (cmd) => {
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        wsRef.current.send(JSON.stringify(cmd));
      }
    },
    []
  );

  const startMonitoring = useCallback(
    (filters = {}, timeout = 300) => {
      sendCommand({ action: 'start', filters, timeout });
      setRemainingTime(timeout);
      setIsMonitoring(true);
    },
    [sendCommand]
  );

  const stopMonitoring = useCallback(() => {
    sendCommand({ action: 'stop' });
    setIsMonitoring(false);
    setRemainingTime(0);
  }, [sendCommand]);

  const renewTimeout = useCallback(
    (timeout) => {
      sendCommand({ action: 'renew', timeout });
      if (timeout) setRemainingTime(timeout);
    },
    [sendCommand]
  );

  const clearTraces = useCallback(() => {
    setTraces([]);
  }, []);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      shouldReconnectRef.current = false;
      if (reconnectTimerRef.current) {
        clearTimeout(reconnectTimerRef.current);
      }
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, []);

  return {
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
  };
}
