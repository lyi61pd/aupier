package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config 应用配置结构
type Config struct {
	OutputDir          string `json:"output_dir"`
	RecordHotkey       string `json:"record_hotkey"`
	PlayLastClipHotkey string `json:"play_last_clip_hotkey"`
	SampleRate         int    `json:"sample_rate"`
	Channels           int    `json:"channels"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		OutputDir:          "./recordings",
		RecordHotkey:       "Ctrl+Shift+R",
		PlayLastClipHotkey: "Ctrl+Shift+P",
		SampleRate:         44100,
		Channels:           2,
	}
}

// Load 从文件加载配置
func Load(filename string) (*Config, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		cfg := DefaultConfig()
		if err := cfg.Save(filename); err != nil {
			return nil, fmt.Errorf("failed to save default config: %w", err)
		}
		return cfg, nil
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	return &cfg, nil
}

// Save 保存配置到文件
func (c *Config) Save(filename string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetAbsOutputDir 获取输出目录的绝对路径
func (c *Config) GetAbsOutputDir() (string, error) {
	return filepath.Abs(c.OutputDir)
}
