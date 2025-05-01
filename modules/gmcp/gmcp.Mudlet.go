package gmcp

import (
	"embed"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/plugins"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/users"
	"gopkg.in/yaml.v3"
)

var (
	//go:embed files/*
	files embed.FS
)

// MudletConfig holds the configuration for Mudlet clients
type MudletConfig struct {
	Version string `json:"version" yaml:"version"`
	URL     string `json:"url" yaml:"url"`
	MapURL  string `json:"map_url" yaml:"map_url"`
}

// GMCPMudletModule handles Mudlet-specific GMCP functionality
type GMCPMudletModule struct {
	plug   *plugins.Plugin
	config MudletConfig
}

// GMCPMudletDetected is an event fired when a Mudlet client is detected
type GMCPMudletDetected struct {
	ConnectionId uint64
	UserId       int
}

func (g GMCPMudletDetected) Type() string { return `GMCPMudletDetected` }

func init() {
	// Set up a default configuration first
	g := GMCPMudletModule{
		plug: plugins.New(`gmcp.Mudlet`, `1.0`),
		config: MudletConfig{
			Version: "1",                                                                     // Default value
			URL:     "URL for the auto-install package goes in the mudlet-config.yaml file.", // Default value
		},
	}

	// Attach embedded filesystem without logging errors
	_ = g.plug.AttachFileSystem(files)

	// Try to load config silently
	configData, err := files.ReadFile("files/datafiles/mudlet-config.yaml")
	if err == nil {
		// Only try to unmarshal if we successfully read the file
		_ = yaml.Unmarshal(configData, &g.config)
	}

	// Register event listeners
	events.RegisterListener(events.PlayerSpawn{}, g.playerSpawnHandler)
	events.RegisterListener(GMCPMudletDetected{}, g.mudletDetectedHandler)

	// Register the Mudlet-specific user command - set as hidden (true for first bool)
	g.plug.AddUserCommand("mudletmap", g.sendMapCommand, true, false)
}

// sendMapCommand is a user command that sends the map URL to Mudlet clients
func (g *GMCPMudletModule) sendMapCommand(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {
	// Only send if the client is Mudlet
	connId := user.ConnectionId()
	if gmcpData, ok := gmcpModule.cache.Get(connId); ok && gmcpData.Client.IsMudlet {
		// Send the map URL
		g.sendMudletMapConfig(user.UserId)

		// Return true to indicate the command was handled (but don't show any output to the user)
		return true, nil
	}

	// Return false to indicate the command wasn't handled (if not a Mudlet client)
	// This allows other handlers to potentially process it
	return false, nil
}

// sendMudletMapConfig sends the Mudlet map configuration via GMCP
func (g *GMCPMudletModule) sendMudletMapConfig(userId int) {
	if userId < 1 {
		return
	}

	// Create a payload for the Client.Map message
	mapConfig := map[string]string{
		"url": g.config.MapURL,
	}

	// Send the Client.Map message
	events.AddToQueue(GMCPOut{
		UserId:  userId,
		Module:  "Client.Map",
		Payload: mapConfig,
	})

	mudlog.Debug("GMCP", "type", "Mudlet", "action", "Sent Mudlet map config", "userId", userId)
}

// playerSpawnHandler sends Mudlet-specific GMCP when a player connects
func (g *GMCPMudletModule) playerSpawnHandler(e events.Event) events.ListenerReturn {
	evt, typeOk := e.(events.PlayerSpawn)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "PlayerSpawn", "Actual Type", e.Type())
		return events.Cancel
	}

	// Check if the client is Mudlet
	if gmcpData, ok := gmcpModule.cache.Get(evt.ConnectionId); ok {
		if gmcpData.Client.IsMudlet {
			// Send Mudlet-specific GMCP
			g.sendMudletConfig(evt.UserId)
		}
	}

	return events.Continue
}

// mudletDetectedHandler handles the event when a Mudlet client is detected
func (g *GMCPMudletModule) mudletDetectedHandler(e events.Event) events.ListenerReturn {
	evt, typeOk := e.(GMCPMudletDetected)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "GMCPMudletDetected", "Actual Type", e.Type())
		return events.Cancel
	}

	if evt.UserId > 0 {
		g.sendMudletConfig(evt.UserId)
	}

	return events.Continue
}

// sendMudletConfig sends the Mudlet configuration via GMCP
func (g *GMCPMudletModule) sendMudletConfig(userId int) {
	if userId < 1 {
		return
	}

	// Create a GUI payload with only version and url (not map_url)
	guiPayload := struct {
		Version string `json:"version"`
		URL     string `json:"url"`
	}{
		Version: g.config.Version,
		URL:     g.config.URL,
	}

	// Send the Client.GUI message with only version and URL
	events.AddToQueue(GMCPOut{
		UserId:  userId,
		Module:  "Client.GUI",
		Payload: guiPayload,
	})

	mudlog.Debug("GMCP", "type", "Mudlet", "action", "Sent Mudlet package config", "userId", userId)
}
