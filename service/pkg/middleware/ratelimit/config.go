// Package ratelimit provides distributed rate limiting middleware using Redis.
package ratelimit

import (
	"time"
)

// Config holds the rate limiting configuration.
type Config struct {
	// Enabled determines if rate limiting is active
	Enabled bool

	// Global rate limits (applies to all requests)
	Global TierConfig

	// User rate limits (authenticated requests)
	User TierConfig

	// IP rate limits (unauthenticated requests)
	IP TierConfig

	// API-specific rate limits (path patterns)
	API map[string]TierConfig

	// FallbackToLocal enables local rate limiting when Redis is unavailable
	FallbackToLocal bool
}

// TierConfig defines rate limit settings for a tier.
type TierConfig struct {
	// RequestsPerSecond is the maximum requests per second
	RequestsPerSecond int

	// RequestsPerMinute is the maximum requests per minute
	RequestsPerMinute int

	// Burst is the maximum burst size above the rate limit
	Burst int

	// Window is the time window for the rate limit
	Window time.Duration

	// ConnectionsPerUser limits concurrent connections (for WebSocket)
	ConnectionsPerUser int
}

// DefaultConfig returns the default rate limiting configuration.
func DefaultConfig() *Config {
	return &Config{
		Enabled: true,
		Global: TierConfig{
			RequestsPerSecond: 1000,
			Burst:             100,
			Window:            time.Second,
		},
		User: TierConfig{
			RequestsPerMinute: 300,
			Burst:             50,
			Window:            time.Minute,
		},
		IP: TierConfig{
			RequestsPerMinute: 60,
			Burst:             10,
			Window:            time.Minute,
		},
		API:             make(map[string]TierConfig),
		FallbackToLocal: true,
	}
}

// GetEffectiveLimit returns the effective limit (requests per window).
func (t *TierConfig) GetEffectiveLimit() int {
	if t.RequestsPerSecond > 0 {
		return t.RequestsPerSecond
	}
	return t.RequestsPerMinute
}

// GetEffectiveWindow returns the effective time window.
func (t *TierConfig) GetEffectiveWindow() time.Duration {
	if t.Window > 0 {
		return t.Window
	}
	if t.RequestsPerSecond > 0 {
		return time.Second
	}
	return time.Minute
}

// RateLimitInfo contains rate limit status information for response headers.
type RateLimitInfo struct {
	// Limit is the maximum number of requests allowed in the window
	Limit int

	// Remaining is the number of requests remaining in the current window
	Remaining int

	// Reset is the Unix timestamp when the rate limit resets
	Reset int64

	// RetryAfter is the number of seconds to wait before retrying (when limited)
	RetryAfter int
}

// KeyType represents the type of rate limit key.
type KeyType string

const (
	KeyTypeGlobal KeyType = "global"
	KeyTypeUser   KeyType = "user"
	KeyTypeIP     KeyType = "ip"
	KeyTypeAPI    KeyType = "api"
)

// BuildKey builds a Redis key for rate limiting.
func BuildKey(keyType KeyType, identifier string, path string) string {
	prefix := "ratelimit:"

	switch keyType {
	case KeyTypeGlobal:
		return prefix + "global"
	case KeyTypeUser:
		if path != "" {
			return prefix + "user:" + identifier + ":api:" + path
		}
		return prefix + "user:" + identifier
	case KeyTypeIP:
		return prefix + "ip:" + identifier
	case KeyTypeAPI:
		return prefix + "api:" + path
	default:
		return prefix + "unknown:" + identifier
	}
}

