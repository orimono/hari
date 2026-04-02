package config

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/orimono/shutter/internal/logger"
)

type Duration time.Duration

type Config struct {
	AgentID          string          `json:"agent_id"`
	ServerURL        string          `json:"server_url"`
	ReaderTimeout    Duration        `json:"reader_timeout"`
	WriterTimeout    Duration        `json:"writer_timeout"`
	PingInterval     Duration        `json:"ping_interval"`
	PongTimeout      Duration        `json:"pong_timeout"`
	MaxRetryCount    int             `json:"max_retry_count"`
	MaxTimeoutCount  int             `json:"max_timeout_count"`
	RetryInterval    Duration        `json:"retry_interval"`
	MaxRetryInterval Duration        `json:"max_retry_interval"`
	WorkerCount      int             `json:"worker_count"`
	LogLevel         logger.LogLevel `json:"log_level"`
	Tags             []string        `json:"tags"`
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(dur)
	return nil
}

func parse(raw []byte, cfg *Config) error {
	err := json.Unmarshal(raw, cfg)
	if err != nil {
		slog.Error("Failed to parse JSON config.", "message", err)
	}
	return err
}

func ReadFromFile(path string, cfg *Config) error {
	content, err := os.ReadFile(path)
	if err != nil {
		slog.Error("Failed to read JSON from file", "path", path)
		return err
	}
	return parse(content, cfg)
}

func Load() (*Config, error) {
	cfgPath := os.Getenv("SHUTTER_CONFIG_PATH")
	if cfgPath == "" {
		home := os.Getenv("SHUTTER_HOME")
		if home == "" {
			home = "."
		}
		cfgPath = filepath.Join(home, "config.json")
	}

	absPath, err := filepath.Abs(cfgPath)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := ReadFromFile(absPath, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		slog.Error("Failed to load configuration", "message", err)
		panic("critical configuration error")
	}
	return cfg
}
