package monitor

import (
	"sync"
	"sync/atomic"
)

// Hub is the singleton MonitorHub instance.
var Hub = &MonitorHub{
	sessions: make(map[string]*MonitorSession),
}

// MonitorHub manages all active monitoring sessions.
type MonitorHub struct {
	hasActive atomic.Bool
	mu        sync.RWMutex
	sessions  map[string]*MonitorSession
}

// HasActiveSessions returns true if there are any active monitoring sessions.
// This is the fast-path check called on every relay request — it only reads an atomic bool.
func (h *MonitorHub) HasActiveSessions() bool {
	return h.hasActive.Load()
}

// HasMatchingSession checks if any active session matches the given request metadata.
func (h *MonitorHub) HasMatchingSession(tokenID, userID, channelID int, modelName string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, session := range h.sessions {
		if session.MatchesRequest(tokenID, userID, channelID, modelName) {
			return true
		}
	}
	return false
}

// AddSession registers a new monitoring session.
func (h *MonitorHub) AddSession(session *MonitorSession) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.sessions[session.ID] = session
	h.hasActive.Store(true)
}

// RemoveSession removes a monitoring session by ID.
func (h *MonitorHub) RemoveSession(id string) {
	h.mu.Lock()
	session, exists := h.sessions[id]
	if exists {
		delete(h.sessions, id)
	}
	empty := len(h.sessions) == 0
	h.mu.Unlock()

	if empty {
		h.hasActive.Store(false)
	}

	if session != nil {
		session.Close()
	}
}

// GetSession returns a session by ID, or nil if not found.
func (h *MonitorHub) GetSession(id string) *MonitorSession {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.sessions[id]
}

// Broadcast sends a trace to all matching sessions.
func (h *MonitorHub) Broadcast(trace *RelayTrace) {
	h.mu.RLock()
	// Collect matching sessions under read lock
	var matching []*MonitorSession
	for _, session := range h.sessions {
		if session.MatchesRequest(trace.TokenID, trace.UserID, trace.ChannelID, trace.ModelName) {
			matching = append(matching, session)
		}
	}
	h.mu.RUnlock()

	// Send outside the lock to avoid holding it during I/O
	msg := WsDataMessage{
		Type: "trace",
		Data: trace,
	}

	var failedIDs []string
	for _, session := range matching {
		session.IncrementTraceCount()
		if !session.Send(msg) {
			failedIDs = append(failedIDs, session.ID)
		}
	}

	// Clean up failed sessions
	for _, id := range failedIDs {
		h.RemoveSession(id)
	}
}

// SessionCount returns the number of active sessions (for debugging/status).
func (h *MonitorHub) SessionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.sessions)
}
