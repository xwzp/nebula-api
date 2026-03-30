package controller

import (
	"net/http"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/relay/monitor"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var monitorUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func RelayMonitorWs(c *gin.Context) {
	conn, err := monitorUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to upgrade to websocket: " + err.Error()})
		return
	}

	sessionID := common.GetUUID()
	session := monitor.NewMonitorSession(sessionID, conn)
	monitor.Hub.AddSession(session)

	// Send initial status
	session.Send(monitor.WsDataMessage{
		Type: "status",
		Data: monitor.StatusData{
			Monitoring: false,
			Message:    "connected, send {\"action\":\"start\"} to begin monitoring",
		},
	})

	// Read loop for control messages
	defer func() {
		monitor.Hub.RemoveSession(sessionID)
	}()

	for {
		_, message, readErr := conn.ReadMessage()
		if readErr != nil {
			// Client disconnected
			return
		}

		var ctrl monitor.WsControlMessage
		if jsonErr := common.Unmarshal(message, &ctrl); jsonErr != nil {
			session.Send(monitor.WsDataMessage{
				Type: "error",
				Data: map[string]string{"message": "invalid control message: " + jsonErr.Error()},
			})
			continue
		}

		switch ctrl.Action {
		case "start":
			filters := monitor.MonitorFilters{}
			if ctrl.Filters != nil {
				filters = *ctrl.Filters
			}
			session.StartMonitoring(filters, ctrl.Timeout)
			remaining := int(monitor.DefaultTimeout.Seconds())
			if ctrl.Timeout > 0 {
				remaining = ctrl.Timeout
			}
			session.Send(monitor.WsDataMessage{
				Type: "status",
				Data: monitor.StatusData{
					Monitoring: true,
					Remaining:  remaining,
					Filters:    filters,
					Message:    "monitoring started",
				},
			})

		case "stop":
			session.StopMonitoring()
			session.Send(monitor.WsDataMessage{
				Type: "status",
				Data: monitor.StatusData{
					Monitoring: false,
					Message:    "monitoring stopped",
				},
			})

		case "renew":
			session.RenewTimeout()
			remaining := int(monitor.DefaultTimeout.Seconds())
			if ctrl.Timeout > 0 {
				remaining = ctrl.Timeout
				// Update timeout duration if provided
				session.StartMonitoring(session.Filters, ctrl.Timeout)
			}
			session.Send(monitor.WsDataMessage{
				Type: "status",
				Data: monitor.StatusData{
					Monitoring: true,
					Remaining:  remaining,
					Message:    "timeout renewed",
				},
			})

		case "ping":
			session.Send(monitor.WsDataMessage{
				Type: "status",
				Data: monitor.StatusData{
					Monitoring: session.Monitoring,
					TraceCount: session.GetRemainingSeconds(),
					Message:    "pong",
				},
			})

		default:
			session.Send(monitor.WsDataMessage{
				Type: "error",
				Data: map[string]string{"message": "unknown action: " + ctrl.Action},
			})
		}
	}
}

// GetRelayMonitorStatus returns the current monitoring status (REST endpoint).
func GetRelayMonitorStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"active_sessions": monitor.Hub.SessionCount(),
		"server_time":     time.Now().Unix(),
	})
}
