package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"go.yaml.in/yaml/v3"
)

type Config struct {
	VM   VMConfig    `mapstructure:"vm"   yaml:"vm"`
	Apps []AppConfig `mapstructure:"apps" yaml:"apps"`
}

type VMConfig struct {
	Name               string   `mapstructure:"name"                  yaml:"name"`
	Port               int      `mapstructure:"port"                  yaml:"port,omitempty"`
	Address            string   `mapstructure:"address"               yaml:"address,omitempty"`
	ControlPlaneURL    string   `mapstructure:"control_plane_url"     yaml:"control_plane_url"`
	ControlPlaneAPIKey string   `mapstructure:"control_plane_api_key" yaml:"control_plane_api_key"`
	AuthToken          string   `mapstructure:"auth_token"            yaml:"auth_token"`
	Labels             []string `mapstructure:"labels"                yaml:"labels,omitempty"`
}

type AppConfig struct {
	Name        string      `mapstructure:"name"         yaml:"name"`
	Type        string      `mapstructure:"type"         yaml:"type"`
	Service     string      `mapstructure:"service"      yaml:"service,omitempty"`
	Container   string      `mapstructure:"container"    yaml:"container,omitempty"`
	Environment string      `mapstructure:"environment"  yaml:"environment,omitempty"`
	EnvManaged  bool        `mapstructure:"env_managed"  yaml:"env_managed,omitempty"`
	EnvFile     string      `mapstructure:"env_file"     yaml:"env_file,omitempty"`
	EnvFiles    []string    `mapstructure:"env_files"    yaml:"env_files,omitempty"`
	EnvDir      string      `mapstructure:"env_dir"      yaml:"env_dir,omitempty"`
	AutoRestart bool        `mapstructure:"auto_restart" yaml:"auto_restart,omitempty"`
	DeployDir   string      `mapstructure:"deploy_dir"   yaml:"deploy_dir,omitempty"`
	HealthCheck HealthCheck `mapstructure:"health_check" yaml:"health_check,omitempty"`
}

// AllEnvFiles returns all configured env file paths, with env_file first,
// followed by env_files, then any .env* files found in env_dir.
func (a *AppConfig) AllEnvFiles() []string {
	var files []string
	if a.EnvFile != "" {
		files = append(files, a.EnvFile)
	}
	files = append(files, a.EnvFiles...)
	if a.EnvDir != "" {
		entries, _ := os.ReadDir(a.EnvDir)
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if strings.HasPrefix(name, ".env") {
				files = append(files, filepath.Join(a.EnvDir, name))
			}
		}
	}
	return files
}

type HealthCheck struct {
	Type string `mapstructure:"type" yaml:"type,omitempty"`
	URL  string `mapstructure:"url"  yaml:"url,omitempty"`
	Cmd  string `mapstructure:"cmd"  yaml:"cmd,omitempty"`
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

// Save writes the config back to disk as YAML.
func Save(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	return os.WriteFile(path, data, 0600)
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
