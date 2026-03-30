package monitor

import (
	"bytes"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	relaycommon "github.com/QuantumNous/new-api/relay/common"

	"github.com/gin-gonic/gin"
)

// base64 pattern: data:image/xxx;base64,<data>
var base64ImageRegex = regexp.MustCompile(`"data:image/[^;]+;base64,[A-Za-z0-9+/=]{100,}"`)

// TruncateBody truncates the data to maxBytes and returns the string + truncated flag.
func TruncateBody(data []byte, maxBytes int) (string, bool) {
	if len(data) <= maxBytes {
		return string(data), false
	}
	return string(data[:maxBytes]), true
}

// SanitizeBase64 replaces base64 image data with a placeholder to reduce body size.
func SanitizeBase64(data []byte) []byte {
	return base64ImageRegex.ReplaceAllFunc(data, func(match []byte) []byte {
		// Calculate original base64 data size (approximate)
		sizeKB := len(match) / 1024
		if sizeKB < 1 {
			sizeKB = 1
		}
		return []byte(`"[base64 image, ` + strings.TrimRight(strings.TrimRight(
			// Simple int to string without importing strconv
			string([]byte{byte('0' + sizeKB/100%10), byte('0' + sizeKB/10%10), byte('0' + sizeKB%10)}),
			"0"), "") + `KB]"`)
	})
}

// CaptureHeaders converts http.Header to a simple map, redacting sensitive values.
func CaptureHeaders(h http.Header) map[string]string {
	result := make(map[string]string, len(h))
	for key, values := range h {
		lowerKey := strings.ToLower(key)
		if lowerKey == "authorization" || lowerKey == "x-api-key" || lowerKey == "api-key" {
			result[key] = "[REDACTED]"
		} else {
			result[key] = strings.Join(values, ", ")
		}
	}
	return result
}

// CaptureReaderBody reads up to maxBytes from the reader, returns the captured bytes
// and a replacement reader that reconstitutes the original stream.
// This solves the io.Reader read-once problem.
func CaptureReaderBody(r io.Reader, maxBytes int) (captured []byte, replacement io.Reader, err error) {
	if r == nil {
		return nil, nil, nil
	}

	buf := make([]byte, maxBytes)
	n, readErr := io.ReadAtLeast(r, buf, 1)
	if n == 0 {
		if readErr == io.EOF || readErr == io.ErrUnexpectedEOF {
			return nil, bytes.NewReader(nil), nil
		}
		return nil, r, readErr
	}

	captured = buf[:n]
	// Reconstitute the reader: captured bytes + remaining unconsumed bytes
	replacement = io.MultiReader(bytes.NewReader(captured), r)
	return captured, replacement, nil
}

// CaptureBodyFromBytes captures body from a byte slice (e.g., BodyStorage.Bytes()).
func CaptureBodyFromBytes(data []byte, maxBytes int) *CapturedHTTP {
	if data == nil {
		return &CapturedHTTP{
			Headers: make(map[string]string),
			Body:    "",
			BodyLen: 0,
		}
	}
	sanitized := SanitizeBase64(data)
	body, truncated := TruncateBody(sanitized, maxBytes)
	return &CapturedHTTP{
		Body:      body,
		BodyLen:   len(data),
		Truncated: truncated,
	}
}

// InitTrace creates a new RelayTrace on the gin context if monitoring is active
// and the request matches any session's filters.
// channelID is passed explicitly because info.ChannelId may not yet reflect the
// current retry's channel (it is re-read from context inside InitChannelMeta).
// Returns true if a trace was created (monitoring is active for this request).
func InitTrace(c *gin.Context, info *relaycommon.RelayInfo, channelID int) bool {
	if !Hub.HasActiveSessions() {
		return false
	}

	tokenID := info.TokenId
	userID := info.UserId
	modelName := info.OriginModelName

	// Check if any session matches this request
	if !Hub.HasMatchingSession(tokenID, userID, channelID, modelName) {
		return false
	}

	trace := &RelayTrace{
		TraceID:   common.GetContextKeyString(c, common.RequestIdKey),
		Timestamp: time.Now(),
		ModelName: modelName,
		ChannelID: channelID,
		TokenID:   tokenID,
		UserID:    userID,
		IsStream:  info.IsStream,
	}

	c.Set(ContextKeyRelayTrace, trace)
	return true
}

// GetTrace retrieves the RelayTrace from the gin context, if any.
func GetTrace(c *gin.Context) *RelayTrace {
	val, exists := c.Get(ContextKeyRelayTrace)
	if !exists {
		return nil
	}
	trace, ok := val.(*RelayTrace)
	if !ok {
		return nil
	}
	return trace
}

// CaptureClientRequest captures Stage 1: Client -> Gateway.
func CaptureClientRequest(c *gin.Context, bodyBytes []byte) {
	trace := GetTrace(c)
	if trace == nil {
		return
	}

	captured := CaptureBodyFromBytes(bodyBytes, DefaultMaxBodyBytes)
	captured.Method = c.Request.Method
	captured.URL = c.Request.URL.String()
	captured.Headers = CaptureHeaders(c.Request.Header)
	trace.ClientRequest = captured
}

// CaptureUpstreamRequest captures Stage 2: Gateway -> Upstream.
func CaptureUpstreamRequest(c *gin.Context, req *http.Request, bodyBytes []byte) {
	trace := GetTrace(c)
	if trace == nil {
		return
	}

	captured := CaptureBodyFromBytes(bodyBytes, DefaultMaxBodyBytes)
	captured.Method = req.Method
	captured.URL = req.URL.String()
	captured.Headers = CaptureHeaders(req.Header)
	trace.UpstreamRequest = captured
}

// CaptureUpstreamResponse captures Stage 3: Upstream -> Gateway.
func CaptureUpstreamResponse(c *gin.Context, resp *http.Response, bodyBytes []byte) {
	trace := GetTrace(c)
	if trace == nil {
		return
	}

	captured := CaptureBodyFromBytes(bodyBytes, DefaultMaxBodyBytes)
	captured.StatusCode = resp.StatusCode
	captured.Headers = CaptureHeaders(resp.Header)
	trace.UpstreamResponse = captured
}

// CaptureClientResponse captures Stage 4: Gateway -> Client.
func CaptureClientResponse(c *gin.Context, writer *CapturingResponseWriter) {
	trace := GetTrace(c)
	if trace == nil {
		return
	}

	bodyBytes := writer.CapturedBody()
	captured := CaptureBodyFromBytes(bodyBytes, DefaultMaxBodyBytes)
	captured.StatusCode = writer.Status()
	captured.Headers = CaptureHeaders(writer.Header())
	trace.ClientResponse = captured
}

// FinalizeAndBroadcast finalizes the trace and broadcasts it to matching sessions.
func FinalizeAndBroadcast(c *gin.Context) {
	trace := GetTrace(c)
	if trace == nil {
		return
	}
	trace.Duration = time.Since(trace.Timestamp)
	Hub.Broadcast(trace)
}

// CapturingResponseWriter wraps gin.ResponseWriter to capture the response body.
type CapturingResponseWriter struct {
	gin.ResponseWriter
	body        *bytes.Buffer
	maxBytes    int
	capturedLen int
}

// NewCapturingResponseWriter creates a new capturing writer wrapping the original.
func NewCapturingResponseWriter(w gin.ResponseWriter) *CapturingResponseWriter {
	return &CapturingResponseWriter{
		ResponseWriter: w,
		body:           &bytes.Buffer{},
		maxBytes:       DefaultMaxBodyBytes,
	}
}

// Write captures up to maxBytes of the response body.
func (w *CapturingResponseWriter) Write(data []byte) (int, error) {
	if w.body.Len() < w.maxBytes {
		remaining := w.maxBytes - w.body.Len()
		if len(data) <= remaining {
			w.body.Write(data)
		} else {
			w.body.Write(data[:remaining])
		}
	}
	w.capturedLen += len(data)
	return w.ResponseWriter.Write(data)
}

// CapturedBody returns the captured body bytes.
func (w *CapturingResponseWriter) CapturedBody() []byte {
	return w.body.Bytes()
}

// TotalWritten returns the total number of bytes written to the client.
func (w *CapturingResponseWriter) TotalWritten() int {
	return w.capturedLen
}
