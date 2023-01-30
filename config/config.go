package config

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/jbsmith7741/toml"
)

// Config represents a configuration parse
type Config struct {
	BlenderPath    string `toml:"blender_path" desc:"Blender Path to start Blender from"`
	EQPath         string `toml:"eq_path" desc:"EverQuest Path to copy converted zones to"`
	IsEQCopy       bool   `toml:"eq_copy" desc:"copy eqgzi output to eq path"`
	LastZone       string `toml:"last_zone" desc:"Last zone selected"`
	ServerPath     string `toml:"server_path" desc:"EQEmu Server Path, if any"`
	IsServerCopy   bool   `toml:"server_copy" desc:"copy eqgzi output to server path"`
	EQGZIVersion   string `toml:"eqgzi_version" desc:"Last downloaded EQGZI version"`
	LanternVersion string `toml:"lantern_version" desc:"Last downloaded LanternExtractor version"`
}

// NewConfig creates a new configuration
func New(ctx context.Context) (*Config, error) {
	var f *os.File
	cfg := Config{}
	path := "eqgzi-manager.conf"

	isNewConfig := false
	fi, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("config info: %w", err)
		}
		f, err = os.Create(path)
		if err != nil {
			return nil, fmt.Errorf("create eqgzi-manager.conf: %w", err)
		}
		fi, err = os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("new config info: %w", err)
		}
		isNewConfig = true
	}
	if !isNewConfig {
		f, err = os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("open config: %w", err)
		}
	}

	defer f.Close()
	if fi.IsDir() {
		return nil, fmt.Errorf("eqgzi-manager.conf is a directory, should be a file")
	}

	if isNewConfig {
		enc := toml.NewEncoder(f)
		cfg = getDefaultConfig()
		err = enc.Encode(cfg)
		if err != nil {
			return nil, fmt.Errorf("encode default: %w", err)
		}
		return &cfg, nil
	}

	_, err = toml.DecodeReader(f, &cfg)
	if err != nil {
		return nil, fmt.Errorf("decode eqgzi-manager.conf: %w", err)
	}

	return &cfg, nil
}

// Verify returns an error if configuration appears off
func (c *Config) Verify() error {

	return nil
}

func getDefaultConfig() Config {
	cfg := Config{}
	if runtime.GOOS == "darwin" {
		_, err := os.Stat("/Applications/Blender.app")
		if err == nil {
			cfg.BlenderPath = "/Applications/Blender.app/Contents/MacOS/Blender"
		}
	}

	return cfg
}

func (c *Config) Save() error {
	w, err := os.Create("eqgzi-manager.conf")
	if err != nil {
		return fmt.Errorf("create eqgzi-manager.conf: %w", err)
	}
	defer w.Close()

	enc := toml.NewEncoder(w)
	err = enc.Encode(c)
	if err != nil {
		return fmt.Errorf("encode default: %w", err)
	}
	return nil
}
