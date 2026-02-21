package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	VM   VMConfig    `mapstructure:"vm"`
	Apps []AppConfig `mapstructure:"apps"`
}

type VMConfig struct {
	Name                string   `mapstructure:"name"`
	Port                int      `mapstructure:"port"`
	Address             string   `mapstructure:"address"`               // how control plane reaches this agent
	ControlPlaneURL     string   `mapstructure:"control_plane_url"`
	ControlPlaneAPIKey  string   `mapstructure:"control_plane_api_key"` // API key for authenticating to control plane
	AuthToken           string   `mapstructure:"auth_token"`
	Labels              []string `mapstructure:"labels"`
}

type AppConfig struct {
	Name        string      `mapstructure:"name"`
	Type        string      `mapstructure:"type"`        // "systemd", "docker"
	Service     string      `mapstructure:"service"`     // systemd only
	Container   string      `mapstructure:"container"`   // docker only
	Environment string      `mapstructure:"environment"` // optional label, e.g. "production"
	EnvFile     string      `mapstructure:"env_file"`
	HealthCheck HealthCheck `mapstructure:"health_check"`
}

type HealthCheck struct {
	Type string `mapstructure:"type"` // "http", "command"
	URL  string `mapstructure:"url"`
	Cmd  string `mapstructure:"cmd"`
}

func Load(path string) (*Config, error) {
	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

func validate(cfg *Config) error {
	if cfg.VM.Name == "" {
		return fmt.Errorf("vm.name is required")
	}
	if cfg.VM.AuthToken == "" {
		return fmt.Errorf("vm.auth_token is required")
	}
	if cfg.VM.ControlPlaneURL == "" {
		return fmt.Errorf("vm.control_plane_url is required")
	}
	if cfg.VM.ControlPlaneAPIKey == "" {
		return fmt.Errorf("vm.control_plane_api_key is required")
	}
	if cfg.VM.Port == 0 {
		cfg.VM.Port = 9000
	}
	if cfg.VM.Address == "" {
		cfg.VM.Address = fmt.Sprintf("http://localhost:%d", cfg.VM.Port)
	}
	for i, app := range cfg.Apps {
		if app.Name == "" {
			return fmt.Errorf("apps[%d].name is required", i)
		}
		switch app.Type {
		case "systemd":
			if app.Service == "" {
				return fmt.Errorf("apps[%d].service is required for type 'systemd'", i)
			}
		case "docker":
			if app.Container == "" {
				return fmt.Errorf("apps[%d].container is required for type 'docker'", i)
			}
		default:
			return fmt.Errorf("apps[%d].type must be 'systemd' or 'docker', got %q", i, app.Type)
		}
	}
	return nil
}
