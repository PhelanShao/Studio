package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultStudioConfig(t *testing.T) {
	cfg := defaultStudioConfig()

	assert.NotNil(t, cfg)
	assert.Equal(t, "sciol", cfg.Server.Platform)
	assert.Equal(t, "api", cfg.Server.Service)
	assert.Equal(t, 48197, cfg.Server.Port)
	assert.Equal(t, "dev", cfg.Server.Env)
}

func TestDefaultStudioConfigFeatures(t *testing.T) {
	cfg := defaultStudioConfig()

	assert.True(t, cfg.Features["workflow_v2"])
	assert.True(t, cfg.Features["realtime_camera"])
	assert.False(t, cfg.Features["new_auth_flow"])
	assert.False(t, cfg.Features["ai_assistant"])
	assert.True(t, cfg.Features["rate_limiting"])
}

func TestDefaultStudioConfigRateLimits(t *testing.T) {
	cfg := defaultStudioConfig()

	assert.True(t, cfg.RateLimits.Enabled)
	assert.Equal(t, 1000, cfg.RateLimits.Global.RequestsPerSecond)
	assert.Equal(t, 100, cfg.RateLimits.Global.Burst)
	assert.Equal(t, 300, cfg.RateLimits.User.RequestsPerMinute)
	assert.Equal(t, 50, cfg.RateLimits.User.Burst)
	assert.Equal(t, 60, cfg.RateLimits.IP.RequestsPerMinute)
	assert.Equal(t, 10, cfg.RateLimits.IP.Burst)
}

func TestDefaultStudioConfigObservability(t *testing.T) {
	cfg := defaultStudioConfig()

	assert.True(t, cfg.Observability.Tracing.Enabled)
	assert.Equal(t, 1.0, cfg.Observability.Tracing.SamplingRate)
	assert.True(t, cfg.Observability.Metrics.Enabled)
	assert.Equal(t, 30, cfg.Observability.Metrics.ExportIntervalSeconds)
	assert.Equal(t, "info", cfg.Observability.Logging.Level)
	assert.Equal(t, "json", cfg.Observability.Logging.Format)
}

func TestGetStudioConfigWithoutLoad(t *testing.T) {
	// Reset for test
	oldConfig := studioConfig
	studioConfig = nil
	defer func() { studioConfig = oldConfig }()

	cfg := GetStudioConfig()
	assert.NotNil(t, cfg)
	assert.Equal(t, "sciol", cfg.Server.Platform)
}

func TestLoadStudioConfigNoFile(t *testing.T) {
	// Try loading from non-existent path
	cfg, err := LoadStudioConfig("/non/existent/path", "dev")

	// Should return defaults without error when file doesn't exist
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "sciol", cfg.Server.Platform)
}

func TestRateLimitTierStruct(t *testing.T) {
	tier := RateLimitTier{
		RequestsPerSecond:  100,
		RequestsPerMinute:  6000,
		Burst:              50,
		ConnectionsPerUser: 10,
	}

	assert.Equal(t, 100, tier.RequestsPerSecond)
	assert.Equal(t, 6000, tier.RequestsPerMinute)
	assert.Equal(t, 50, tier.Burst)
	assert.Equal(t, 10, tier.ConnectionsPerUser)
}

func TestServerConfigStruct(t *testing.T) {
	cfg := ServerConfig{
		Platform:     "test",
		Service:      "api",
		Port:         8080,
		SchedulePort: 8081,
		Env:          "prod",
	}

	assert.Equal(t, "test", cfg.Platform)
	assert.Equal(t, "api", cfg.Service)
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, 8081, cfg.SchedulePort)
	assert.Equal(t, "prod", cfg.Env)
}

