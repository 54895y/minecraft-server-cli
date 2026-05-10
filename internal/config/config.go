package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	DefaultUserAgent = "54895y/minecraft-server-cli/dev (https://github.com/54895y/minecraft-server-cli)"
)

type Manager struct {
	v *viper.Viper
}

func NewManager(appName string) *Manager {
	v := viper.New()
	v.SetEnvPrefix("MCSERVER")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	v.SetDefault("download.threads", 8)
	v.SetDefault("download.timeout", "30s")
	v.SetDefault("download.retries", 3)
	v.SetDefault("download.user_agent", DefaultUserAgent)
	v.SetDefault("core.source", "official")
	v.SetDefault("github.proxy", "none")
	v.SetDefault("paths.output_dir", ".")
	v.SetDefault("mirrors.fastmirror.paper", "https://www.fastmirror.net")
	v.SetDefault("mirrors.msl.base", "https://dl.mslmc.cn")

	configDir, err := os.UserConfigDir()
	if err != nil || configDir == "" {
		configDir = "."
	}
	v.SetDefault("config.file", filepath.Join(configDir, appName, "config.yaml"))

	return &Manager{v: v}
}

func (m *Manager) BindPFlag(key string, flag *pflag.Flag) error {
	return m.v.BindPFlag(key, flag)
}

func (m *Manager) Load() error {
	configFile := m.v.GetString("config.file")
	if configFile == "" {
		return nil
	}
	m.v.SetConfigFile(configFile)
	err := m.v.ReadInConfig()
	if err == nil {
		return nil
	}
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		return nil
	}
	if os.IsNotExist(err) {
		return nil
	}
	return fmt.Errorf("read config: %w", err)
}

func (m *Manager) Path() string {
	return m.v.GetString("config.file")
}

func (m *Manager) Set(key string, value any) {
	m.v.Set(key, value)
}

func (m *Manager) GetString(key string) string {
	return m.v.GetString(key)
}

func (m *Manager) GetInt(key string) int {
	return m.v.GetInt(key)
}

func (m *Manager) GetStringMapString(key string) map[string]string {
	return m.v.GetStringMapString(key)
}

func (m *Manager) AllSettings() map[string]any {
	return m.v.AllSettings()
}

func (m *Manager) Write() error {
	configFile := m.Path()
	if configFile == "" {
		return fmt.Errorf("config file path is empty")
	}
	if err := os.MkdirAll(filepath.Dir(configFile), 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	if err := m.v.WriteConfigAs(configFile); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

func (m *Manager) Reset() error {
	configFile := m.Path()
	if configFile == "" {
		return nil
	}
	if err := os.Remove(configFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove config: %w", err)
	}
	return nil
}
