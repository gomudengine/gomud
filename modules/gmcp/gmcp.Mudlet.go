package gmcp

import (
	"embed"
	"strings"

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
	// Mapper configuration
	MapperVersion string `json:"mapper_version" yaml:"mapper_version"`
	MapperURL     string `json:"mapper_url" yaml:"mapper_url"`

	// UI configuration
	UIVersion string `json:"ui_version" yaml:"ui_version"`
	UIURL     string `json:"ui_url" yaml:"ui_url"`

	// Map data configuration
	MapVersion string `json:"map_version" yaml:"map_version"`
	MapURL     string `json:"map_url" yaml:"map_url"`
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
			MapperVersion: "1",                                                                                          // Default value
			MapperURL:     "https://github.com/GoMudEngine/MudletMapper/releases/latest/download/GoMud-Mapper.mpackage", // Default value
			UIVersion:     "1",                                                                                          // Default value
			UIURL:         "https://thiswillbetheuiurl.com",                                                             // Default value
			MapVersion:    "1",                                                                                          // Default value
			MapURL:        "https://github.com/GoMudEngine/MudletMapper/releases/latest/download/gomud.dat",             // Default value
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

	// Register the Mudlet-specific user commands - set as hidden (true for first bool)
	g.plug.AddUserCommand("mudletmap", g.sendMapCommand, true, false)
	g.plug.AddUserCommand("mudletui", g.sendUICommand, true, false)
}

// sendUICommand is a user command that sends UI-related GMCP messages to Mudlet clients
func (g *GMCPMudletModule) sendUICommand(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {
	// Only send if the client is Mudlet
	connId := user.ConnectionId()
	if gmcpData, ok := gmcpModule.cache.Get(connId); ok && gmcpData.Client.IsMudlet {
		// Process command arguments
		args := strings.Fields(rest)
		if len(args) > 0 {
			switch args[0] {
			case "install":
				// Send UI install message
				g.sendMudletUIInstall(user.UserId)
			case "remove":
				// Send UI remove message
				g.sendMudletUIRemove(user.UserId)
			case "update":
				// Send UI update message
				g.sendMudletUIUpdate(user.UserId)
			default:
				// Unknown command
				user.SendText("Usage: mudletui install|remove|update")
			}
		} else {
			// No arguments provided
			user.SendText("Usage: mudletui install|remove|update")
		}

		// Return true to indicate the command was handled
		return true, nil
	}

	// Return false to indicate the command wasn't handled (if not a Mudlet client)
	return false, nil
}

// sendMudletUIInstall sends the UI installation GMCP message
func (g *GMCPMudletModule) sendMudletUIInstall(userId int) {
	if userId < 1 {
		return
	}

	// Create a payload for UI installation
	payload := struct {
		Version string `json:"version"`
		URL     string `json:"url"`
	}{
		Version: g.config.UIVersion,
		URL:     g.config.UIURL,
	}

	// Send the Client.GUI message
	events.AddToQueue(GMCPOut{
		UserId:  userId,
		Module:  "Client.GUI",
		Payload: payload,
	})

	mudlog.Debug("GMCP", "type", "Mudlet", "action", "Sent Mudlet UI install config", "userId", userId)
}

// sendMudletUIRemove sends the UI remove GMCP message
func (g *GMCPMudletModule) sendMudletUIRemove(userId int) {
	if userId < 1 {
		return
	}

	// Create a payload for UI removal
	payload := struct {
		GoMudUI string `json:"gomudui"`
	}{
		GoMudUI: "remove",
	}

	// Send the Client.GUI message
	events.AddToQueue(GMCPOut{
		UserId:  userId,
		Module:  "Client.GUI",
		Payload: payload,
	})

	mudlog.Debug("GMCP", "type", "Mudlet", "action", "Sent Mudlet UI remove command", "userId", userId)
}

// sendMudletUIUpdate sends the UI update GMCP message
func (g *GMCPMudletModule) sendMudletUIUpdate(userId int) {
	if userId < 1 {
		return
	}

	// Create a payload for UI update
	payload := struct {
		GoMudUI string `json:"gomudui"`
	}{
		GoMudUI: "update",
	}

	// Send the Client.GUI message
	events.AddToQueue(GMCPOut{
		UserId:  userId,
		Module:  "Client.GUI",
		Payload: payload,
	})

	mudlog.Debug("GMCP", "type", "Mudlet", "action", "Sent Mudlet UI update command", "userId", userId)
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

	// Create a GUI payload with mapper version and url
	guiPayload := struct {
		Version string `json:"version"`
		URL     string `json:"url"`
	}{
		Version: g.config.MapperVersion,
		URL:     g.config.MapperURL,
	}

	// Send the Client.GUI message with mapper version and URL
	events.AddToQueue(GMCPOut{
		UserId:  userId,
		Module:  "Client.GUI",
		Payload: guiPayload,
	})

	mudlog.Debug("GMCP", "type", "Mudlet", "action", "Sent Mudlet package config", "userId", userId)
}
