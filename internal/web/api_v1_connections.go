package web

import (
	"net/http"
)

type connectionsData struct {
	TelnetPorts          []int `json:"telnet_ports"`
	WebSocketPort        int   `json:"websocket_port"`
	SSHPort              int   `json:"ssh_port"`
	TelnetConnections    int   `json:"telnet_connections"`
	WebSocketConnections int   `json:"websocket_connections"`
	SSHConnections       int   `json:"ssh_connections"`
}

// GET /admin/api/v1/connections
func apiV1GetConnections(w http.ResponseWriter, r *http.Request) {
	s := GetStats()

	writeJSON(w, http.StatusOK, APIResponse[connectionsData]{
		Success: true,
		Data: connectionsData{
			TelnetPorts:          s.TelnetPorts,
			WebSocketPort:        s.WebSocketPort,
			SSHPort:              s.SSHPort,
			TelnetConnections:    s.TelnetConnections,
			WebSocketConnections: s.WebSocketConnections,
			SSHConnections:       s.SSHConnections,
		},
	})
}
