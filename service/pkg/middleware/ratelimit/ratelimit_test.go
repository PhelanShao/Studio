package ratelimit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBuildKey(t *testing.T) {
	tests := []struct {
		name       string
		keyType    KeyType
		identifier string
		path       string
		expected   string
	}{
		{
			name:       "global key",
			keyType:    KeyTypeGlobal,
			identifier: "",
			path:       "",
			expected:   "ratelimit:global",
		},
		{
			name:       "user key",
			keyType:    KeyTypeUser,
			identifier: "user123",
			path:       "",
			expected:   "ratelimit:user:user123",
		},
		{
			name:       "user with path key",
			keyType:    KeyTypeUser,
			identifier: "user123",
			path:       "/api/v1/test",
			expected:   "ratelimit:user:user123:api:/api/v1/test",
		},
		{
			name:       "IP key",
			keyType:    KeyTypeIP,
			identifier: "192.168.1.1",
			path:       "",
			expected:   "ratelimit:ip:192.168.1.1",
		},
		{
			name:       "API key",
			keyType:    KeyTypeAPI,
			identifier: "",
			path:       "/api/v1/workflow",
			expected:   "ratelimit:api:/api/v1/workflow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildKey(tt.keyType, tt.identifier, tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTierConfigGetEffectiveLimit(t *testing.T) {
	// Test with requests per second
	tier1 := TierConfig{
		RequestsPerSecond: 100,
		RequestsPerMinute: 6000,
	}
	assert.Equal(t, 100, tier1.GetEffectiveLimit())

	// Test with only requests per minute
	tier2 := TierConfig{
		RequestsPerSecond: 0,
		RequestsPerMinute: 300,
	}
	assert.Equal(t, 300, tier2.GetEffectiveLimit())
}

func TestTierConfigGetEffectiveWindow(t *testing.T) {
	// Test with explicit window
	tier1 := TierConfig{
		Window: 5 * time.Second,
	}
	assert.Equal(t, 5*time.Second, tier1.GetEffectiveWindow())

	// Test with requests per second
	tier2 := TierConfig{
		RequestsPerSecond: 100,
	}
	assert.Equal(t, time.Second, tier2.GetEffectiveWindow())

	// Test with requests per minute
	tier3 := TierConfig{
		RequestsPerMinute: 300,
	}
	assert.Equal(t, time.Minute, tier3.GetEffectiveWindow())
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.True(t, cfg.Enabled)
	assert.True(t, cfg.FallbackToLocal)
	assert.Equal(t, 1000, cfg.Global.RequestsPerSecond)
	assert.Equal(t, 100, cfg.Global.Burst)
	assert.Equal(t, 300, cfg.User.RequestsPerMinute)
	assert.Equal(t, 60, cfg.IP.RequestsPerMinute)
	assert.NotNil(t, cfg.API)
}

func TestMatchPath(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		path     string
		expected bool
	}{
		{"exact match", "/api/v1/test", "/api/v1/test", true},
		{"no match", "/api/v1/test", "/api/v1/other", false},
		{"wildcard match", "/api/v1/edge/*", "/api/v1/edge/device", true},
		{"wildcard match nested", "/api/v1/edge/*", "/api/v1/edge/device/action", true},
		{"wildcard no match", "/api/v1/edge/*", "/api/v1/other/device", false},
		{"param match", "/api/v1/lab/:uuid", "/api/v1/lab/123", true},
		{"param no match length", "/api/v1/lab/:uuid", "/api/v1/lab/123/extra", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchPath(tt.pattern, tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLocalLimiter(t *testing.T) {
	limiter := NewLocalLimiter()

	// First request should be allowed
	allowed, remaining, _ := limiter.Allow("test:key", 5, time.Minute)
	assert.True(t, allowed)
	assert.Equal(t, 4, remaining)

	// Use up the remaining quota
	for i := 0; i < 4; i++ {
		allowed, _, _ = limiter.Allow("test:key", 5, time.Minute)
		assert.True(t, allowed)
	}

	// Next request should be denied
	allowed, remaining, _ = limiter.Allow("test:key", 5, time.Minute)
	assert.False(t, allowed)
	assert.Equal(t, 0, remaining)
}

