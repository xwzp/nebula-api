package monitor

import (
	"time"
)

const (
	// DefaultMaxBodyBytes is the default maximum body capture size (256KB).
	// After SanitizeLongStrings processing, most bodies are well under this limit.
	DefaultMaxBodyBytes = 262144
	// DefaultMaxStringLen is the max rune count for individual JSON string values
	// before they are truncated by SanitizeLongStrings.
	DefaultMaxStringLen = 4096
	// DefaultTimeout is the default monitoring session timeout
	DefaultTimeout = 5 * time.Minute
	// DefaultMaxTraces is the maximum number of traces kept per session
	DefaultMaxTraces = 100

	// Context key for storing the relay trace on gin.Context
	ContextKeyRelayTrace = "monitor_relay_trace"
)

// MonitorFilters defines the filter criteria for monitoring.
// Zero values mean "match any".
type MonitorFilters struct {
	TokenID   int    `json:"token_id"`
	UserID    int    `json:"user_id"`
	ModelName string `json:"model_name"`
	ChannelID int    `json:"channel_id"`
}

// CapturedHTTP represents one side of an HTTP exchange.
type CapturedHTTP struct {
	Method     string            `json:"method,omitempty"`
	URL        string            `json:"url,omitempty"`
	StatusCode int               `json:"status_code,omitempty"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	BodyLen    int               `json:"body_len"`
	Truncated  bool              `json:"truncated"`
}

// RelayTrace represents the complete 4-stage capture of a relay request.
type RelayTrace struct {
	TraceID          string        `json:"trace_id"`
	Timestamp        time.Time     `json:"timestamp"`
	ModelName        string        `json:"model_name"`
	ChannelID        int           `json:"channel_id"`
	TokenID          int           `json:"token_id"`
	UserID           int           `json:"user_id"`
	IsStream         bool          `json:"is_stream"`
	StreamEventCount int           `json:"stream_event_count,omitempty"`
	Duration         time.Duration `json:"duration"`
	ClientRequest    *CapturedHTTP `json:"client_request"`
	UpstreamRequest  *CapturedHTTP `json:"upstream_request"`
	UpstreamResponse *CapturedHTTP `json:"upstream_response"`
	ClientResponse   *CapturedHTTP `json:"client_response"`
}

// WsControlMessage is the control message sent by the frontend via WebSocket.
type WsControlMessage struct {
	Action  string          `json:"action"` // start / stop / renew
	Filters *MonitorFilters `json:"filters,omitempty"`
	Timeout int             `json:"timeout,omitempty"` // seconds, 0 means default
}

// WsDataMessage is the data message pushed to the frontend via WebSocket.
type WsDataMessage struct {
	Type string `json:"type"` // trace / status / timeout / error
	Data any    `json:"data,omitempty"`
}

// StatusData is sent as WsDataMessage.Data for status updates.
type StatusData struct {
	Monitoring bool   `json:"monitoring"`
	Remaining  int    `json:"remaining"` // seconds remaining
	TraceCount int    `json:"trace_count"`
	Filters    any    `json:"filters,omitempty"`
	Message    string `json:"message,omitempty"`
}
