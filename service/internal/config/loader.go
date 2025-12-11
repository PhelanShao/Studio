package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// StudioConfig represents the full YAML configuration structure.
type StudioConfig struct {
	Server        ServerConfig        `mapstructure:"server"`
	Features      map[string]bool     `mapstructure:"features"`
	RateLimits    RateLimitsConfig    `mapstructure:"rate_limits"`
	Observability ObservabilityConfig `mapstructure:"observability"`
	Workflow      WorkflowConfig      `mapstructure:"workflow"`
	Material      MaterialConfig      `mapstructure:"material"`
	Security      SecurityConfig      `mapstructure:"security"`
}

// ServerConfig from YAML
type ServerConfig struct {
	Platform     string `mapstructure:"platform"`
	Service      string `mapstructure:"service"`
	Port         int    `mapstructure:"port"`
	SchedulePort int    `mapstructure:"schedule_port"`
	Env          string `mapstructure:"env"`
}

// RateLimitsConfig from YAML
type RateLimitsConfig struct {
	Enabled bool                       `mapstructure:"enabled"`
	Global  RateLimitTier              `mapstructure:"global"`
	User    RateLimitTier              `mapstructure:"user"`
	IP      RateLimitTier              `mapstructure:"ip"`
	API     map[string]RateLimitTier   `mapstructure:"api"`
}

// RateLimitTier defines rate limit settings for a tier
type RateLimitTier struct {
	RequestsPerSecond   int `mapstructure:"requests_per_second"`
	RequestsPerMinute   int `mapstructure:"requests_per_minute"`
	Burst               int `mapstructure:"burst"`
	ConnectionsPerUser  int `mapstructure:"connections_per_user"`
}

// ObservabilityConfig from YAML
type ObservabilityConfig struct {
	Tracing TracingConfig `mapstructure:"tracing"`
	Metrics MetricsConfig `mapstructure:"metrics"`
	Logging LoggingConfig `mapstructure:"logging"`
}

// TracingConfig from YAML
type TracingConfig struct {
	Enabled      bool    `mapstructure:"enabled"`
	SamplingRate float64 `mapstructure:"sampling_rate"`
}

// MetricsConfig from YAML
type MetricsConfig struct {
	Enabled              bool `mapstructure:"enabled"`
	ExportIntervalSeconds int `mapstructure:"export_interval_seconds"`
}

// LoggingConfig from YAML
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// WorkflowConfig from YAML
type WorkflowConfig struct {
	MaxConcurrentExecutions int           `mapstructure:"max_concurrent_executions"`
	DefaultTimeoutSeconds   int           `mapstructure:"default_timeout_seconds"`
	MaxRetryAttempts        int           `mapstructure:"max_retry_attempts"`
	Queue                   QueueConfig   `mapstructure:"queue"`
}

// QueueConfig from YAML
type QueueConfig struct {
	Name       string `mapstructure:"name"`
	MaxWorkers int    `mapstructure:"max_workers"`
}

// MaterialConfig from YAML
type MaterialConfig struct {
	SyncIntervalSeconds int `mapstructure:"sync_interval_seconds"`
	MaxDevicesPerLab    int `mapstructure:"max_devices_per_lab"`
}

// SecurityConfig from YAML
type SecurityConfig struct {
	Validation ValidationConfig `mapstructure:"validation"`
	CORS       CORSConfig       `mapstructure:"cors"`
}

// ValidationConfig from YAML
type ValidationConfig struct {
	MaxRequestBodySizeMB int  `mapstructure:"max_request_body_size_mb"`
	SanitizeInput        bool `mapstructure:"sanitize_input"`
}

// CORSConfig from YAML
type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAgeHours      int      `mapstructure:"max_age_hours"`
}

var studioConfig *StudioConfig
var configViper *viper.Viper

// LoadStudioConfig loads the YAML configuration files.
// It loads the base config and then overlays environment-specific config.
func LoadStudioConfig(configPath string, env string) (*StudioConfig, error) {
	configViper = viper.New()

	// Set config file locations
	if configPath == "" {
		configPath = "config"
	}

	// Try to find config directory
	possiblePaths := []string{
		configPath,
		filepath.Join(".", "config"),
		filepath.Join("..", "config"),
		filepath.Join("service", "config"),
	}

	var foundPath string
	for _, p := range possiblePaths {
		if _, err := os.Stat(filepath.Join(p, "studio.yaml")); err == nil {
			foundPath = p
			break
		}
	}

	if foundPath == "" {
		// No config file found, use defaults
		studioConfig = defaultStudioConfig()
		return studioConfig, nil
	}

	// Load base configuration
	configViper.SetConfigName("studio")
	configViper.SetConfigType("yaml")
	configViper.AddConfigPath(foundPath)

	if err := configViper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read base config: %w", err)
	}

	// Load environment-specific configuration overlay
	if env != "" {
		envConfigPath := filepath.Join(foundPath, fmt.Sprintf("studio.%s.yaml", strings.ToLower(env)))
		if _, err := os.Stat(envConfigPath); err == nil {
			envViper := viper.New()
			envViper.SetConfigFile(envConfigPath)
			if err := envViper.ReadInConfig(); err == nil {
				// Merge environment config into base config
				if err := configViper.MergeConfigMap(envViper.AllSettings()); err != nil {
					return nil, fmt.Errorf("failed to merge env config: %w", err)
				}
			}
		}
	}

	// Allow environment variables to override config
	configViper.AutomaticEnv()
	configViper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Unmarshal into struct
	studioConfig = &StudioConfig{}
	if err := configViper.Unmarshal(studioConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return studioConfig, nil
}

// GetStudioConfig returns the loaded studio configuration.
func GetStudioConfig() *StudioConfig {
	if studioConfig == nil {
		return defaultStudioConfig()
	}
	return studioConfig
}

// GetConfigViper returns the viper instance used for configuration.
// This is useful for feature flags which need access to the viper instance.
func GetConfigViper() *viper.Viper {
	return configViper
}

func defaultStudioConfig() *StudioConfig {
	return &StudioConfig{
		Server: ServerConfig{
			Platform:     "sciol",
			Service:      "api",
			Port:         48197,
			SchedulePort: 48198,
			Env:          "dev",
		},
		Features: map[string]bool{
			"workflow_v2":          true,
			"realtime_camera":      true,
			"new_auth_flow":        false,
			"ai_assistant":         false,
			"advanced_scheduling":  false,
			"enhanced_tracing":     true,
			"business_metrics":     true,
			"rate_limiting":        true,
			"request_validation":   true,
		},
		RateLimits: RateLimitsConfig{
			Enabled: true,
			Global: RateLimitTier{
				RequestsPerSecond: 1000,
				Burst:             100,
			},
			User: RateLimitTier{
				RequestsPerMinute: 300,
				Burst:             50,
			},
			IP: RateLimitTier{
				RequestsPerMinute: 60,
				Burst:             10,
			},
		},
		Observability: ObservabilityConfig{
			Tracing: TracingConfig{
				Enabled:      true,
				SamplingRate: 1.0,
			},
			Metrics: MetricsConfig{
				Enabled:               true,
				ExportIntervalSeconds: 30,
			},
			Logging: LoggingConfig{
				Level:  "info",
				Format: "json",
			},
		},
	}
}

