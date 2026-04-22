package main

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type configFile struct {
	FilePaths struct {
		DataFiles string `yaml:"DataFiles"`
	} `yaml:"FilePaths"`
	Network struct {
		HttpPort  int `yaml:"HttpPort"`
		HttpsPort int `yaml:"HttpsPort"`
		SSHPort   int `yaml:"SSHPort"`
	} `yaml:"Network"`
}

type portSetting struct {
	name  string
	value int
}

func main() {
	if len(os.Args) != 2 {
		os.Stderr.WriteString("usage: dev-preflight <run-needs-sudo>\n")
		os.Exit(2)
	}

	switch os.Args[1] {
	case "run-needs-sudo":
		os.Exit(checkRunNeedsSudo())
	default:
		os.Stderr.WriteString("unknown mode: " + os.Args[1] + "\n")
		os.Exit(2)
	}
}

func checkRunNeedsSudo() int {
	cfg, _, err := loadEffectiveConfig()
	if err != nil {
		return 1
	}

	if os.Geteuid() == 0 {
		return 0
	}

	ports := []portSetting{
		{name: "Network.HttpPort", value: cfg.Network.HttpPort},
		{name: "Network.HttpsPort", value: cfg.Network.HttpsPort},
		{name: "Network.SSHPort", value: cfg.Network.SSHPort},
	}

	privileged := make([]portSetting, 0, len(ports))
	for _, port := range ports {
		if port.value > 0 && port.value < 1024 {
			privileged = append(privileged, port)
		}
	}

	if len(privileged) == 0 {
		return 0
	}
	return 10
}

func loadEffectiveConfig() (configFile, string, error) {
	cfg := configFile{}

	baseBytes, err := os.ReadFile("_datafiles/config.yaml")
	if err != nil {
		return cfg, "", err
	}
	if err := yaml.Unmarshal(baseBytes, &cfg); err != nil {
		return cfg, "", err
	}

	if cfg.FilePaths.DataFiles == "" {
		cfg.FilePaths.DataFiles = "_datafiles/world/default"
	}

	overridePath := os.Getenv("CONFIG_PATH")
	if overridePath == "" {
		overridePath = filepath.Join(cfg.FilePaths.DataFiles, "config-overrides.yaml")
	}

	overrideBytes, err := os.ReadFile(overridePath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, overridePath, nil
		}
		return cfg, overridePath, err
	}

	var override map[string]any
	if err := yaml.Unmarshal(overrideBytes, &override); err != nil {
		return cfg, overridePath, err
	}

	if value, ok := lookupString(override, "FilePaths", "DataFiles"); ok && value != "" {
		cfg.FilePaths.DataFiles = value
	}
	if value, ok := lookupInt(override, "Network", "HttpPort"); ok {
		cfg.Network.HttpPort = value
	}
	if value, ok := lookupInt(override, "Network", "HttpsPort"); ok {
		cfg.Network.HttpsPort = value
	}
	if value, ok := lookupInt(override, "Network", "SSHPort"); ok {
		cfg.Network.SSHPort = value
	}

	return cfg, overridePath, nil
}

func lookupString(m map[string]any, section string, key string) (string, bool) {
	if value, ok := m[section+"."+key]; ok {
		switch typed := value.(type) {
		case string:
			return typed, true
		}
	}

	sectionValue, ok := m[section]
	if !ok {
		return "", false
	}
	sectionMap, ok := sectionValue.(map[string]any)
	if !ok {
		return "", false
	}
	value, ok := sectionMap[key]
	if !ok {
		return "", false
	}
	typed, ok := value.(string)
	return typed, ok
}

func lookupInt(m map[string]any, section string, key string) (int, bool) {
	if value, ok := m[section+"."+key]; ok {
		return asInt(value)
	}

	sectionValue, ok := m[section]
	if !ok {
		return 0, false
	}
	sectionMap, ok := sectionValue.(map[string]any)
	if !ok {
		return 0, false
	}
	value, ok := sectionMap[key]
	if !ok {
		return 0, false
	}
	return asInt(value)
}

func asInt(value any) (int, bool) {
	switch typed := value.(type) {
	case int:
		return typed, true
	case int64:
		return int(typed), true
	case float64:
		return int(typed), true
	default:
		return 0, false
	}
}
