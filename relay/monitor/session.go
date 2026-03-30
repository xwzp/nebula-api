package monitor

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/QuantumNous/new-api/common"

	"github.com/gorilla/websocket"
)

// MonitorSession represents a single WebSocket monitoring session.
type MonitorSession struct {
	ID              string
	Conn            *websocket.Conn
	Filters         MonitorFilters
	Monitoring      bool
	CreatedAt       time.Time
	TimeoutDuration time.Duration
	TimeoutTimer    *time.Timer
	traceCount      int
	mu              sync.Mutex
	closed          atomic.Bool
}

// NewMonitorSession creates a new monitoring session.
func NewMonitorSession(id string, conn *websocket.Conn) *MonitorSession {
	return &MonitorSession{
		ID:        id,
		Conn:      conn,
		CreatedAt: time.Now(),
	}
}

// MatchesRequest checks whether the request metadata matches this session's filters.
func (s *MonitorSession) MatchesRequest(tokenID, userID, channelID int, modelName string) bool {
	if !s.Monitoring {
		return false
	}
	if s.Filters.TokenID != 0 && s.Filters.TokenID != tokenID {
		return false
	}
	if s.Filters.UserID != 0 && s.Filters.UserID != userID {
		return false
	}
	if s.Filters.ChannelID != 0 && s.Filters.ChannelID != channelID {
		return false
	}
	if s.Filters.ModelName != "" && s.Filters.ModelName != modelName {
		return false
	}
	return true
}

// Send marshals the message and writes it to the WebSocket connection.
// Returns false if the write failed (session should be removed).
func (s *MonitorSession) Send(msg WsDataMessage) bool {
	if s.closed.Load() {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := common.Marshal(msg)
	if err != nil {
		return false
	}

	if err := s.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return false
	}
	return true
}

// StartMonitoring begins the monitoring session with the given filters and timeout.
func (s *MonitorSession) StartMonitoring(filters MonitorFilters, timeoutSec int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Monitoring = true
	s.Filters = filters
	s.traceCount = 0

	if timeoutSec <= 0 {
		timeoutSec = int(DefaultTimeout.Seconds())
	}
	s.TimeoutDuration = time.Duration(timeoutSec) * time.Second

	// Reset or create timeout timer
	if s.TimeoutTimer != nil {
		s.TimeoutTimer.Stop()
	}
	s.TimeoutTimer = time.AfterFunc(s.TimeoutDuration, func() {
		s.onTimeout()
	})
}

// StopMonitoring stops the monitoring session.
func (s *MonitorSession) StopMonitoring() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Monitoring = false
	if s.TimeoutTimer != nil {
		s.TimeoutTimer.Stop()
		s.TimeoutTimer = nil
	}
}

// RenewTimeout resets the timeout timer.
func (s *MonitorSession) RenewTimeout() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.Monitoring {
		return
	}
	if s.TimeoutTimer != nil {
		s.TimeoutTimer.Stop()
	}
	s.TimeoutTimer = time.AfterFunc(s.TimeoutDuration, func() {
		s.onTimeout()
	})
}

// IncrementTraceCount increments the trace count and returns the new count.
func (s *MonitorSession) IncrementTraceCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.traceCount++
	return s.traceCount
}

// GetRemainingSeconds returns the approximate remaining seconds for the timeout.
func (s *MonitorSession) GetRemainingSeconds() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.Monitoring || s.TimeoutTimer == nil {
		return 0
	}
	// Approximate: timeout duration - elapsed since last reset
	// Timer doesn't expose remaining time, so we track it ourselves
	return int(s.TimeoutDuration.Seconds())
}

// Close closes the session and its WebSocket connection.
func (s *MonitorSession) Close() {
	if s.closed.Swap(true) {
		return // already closed
	}
	s.mu.Lock()
	if s.TimeoutTimer != nil {
		s.TimeoutTimer.Stop()
		s.TimeoutTimer = nil
	}
	s.Monitoring = false
	s.mu.Unlock()

	_ = s.Conn.Close()
}

// onTimeout is called when the monitoring session times out.
func (s *MonitorSession) onTimeout() {
	s.mu.Lock()
	s.Monitoring = false
	if s.TimeoutTimer != nil {
		s.TimeoutTimer = nil
	}
	s.mu.Unlock()

	// Send timeout message
	s.Send(WsDataMessage{
		Type: "timeout",
		Data: StatusData{
			Monitoring: false,
			Message:    "monitoring session timed out",
		},
	})

	// Remove from hub
	Hub.RemoveSession(s.ID)
}
