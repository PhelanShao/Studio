package ratelimit

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/scienceol/studio/service/pkg/features"
	"github.com/scienceol/studio/service/pkg/middleware/otel"
)

const (
	// HeaderRateLimitLimit is the header for rate limit maximum
	HeaderRateLimitLimit = "X-RateLimit-Limit"
	// HeaderRateLimitRemaining is the header for remaining requests
	HeaderRateLimitRemaining = "X-RateLimit-Remaining"
	// HeaderRateLimitReset is the header for reset timestamp
	HeaderRateLimitReset = "X-RateLimit-Reset"
	// HeaderRetryAfter is the header for retry delay
	HeaderRetryAfter = "Retry-After"

	// UserIDContextKey is the context key for user ID (set by auth middleware)
	UserIDContextKey = "user_id"
)

// Limiter is the rate limiter interface.
type Limiter interface {
	Allow(key string, limit int, window time.Duration) (bool, int, int64, error)
}

// RateLimitMiddleware holds the rate limiting middleware state.
type RateLimitMiddleware struct {
	config        *Config
	redisClient   *redis.Client
	slidingWindow *SlidingWindowLimiter
	localLimiter  *LocalLimiter
	useLocal      bool
	mu            sync.RWMutex
}

// LocalLimiter provides in-memory rate limiting as fallback.
type LocalLimiter struct {
	counters map[string]*localCounter
	mu       sync.RWMutex
}

type localCounter struct {
	count     int
	resetTime time.Time
}

// NewLocalLimiter creates a new local rate limiter.
func NewLocalLimiter() *LocalLimiter {
	return &LocalLimiter{
		counters: make(map[string]*localCounter),
	}
}

// Allow checks if a request is allowed (local implementation).
func (l *LocalLimiter) Allow(key string, limit int, window time.Duration) (bool, int, int64) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	counter, exists := l.counters[key]

	if !exists || now.After(counter.resetTime) {
		// Create new counter or reset expired one
		l.counters[key] = &localCounter{
			count:     1,
			resetTime: now.Add(window),
		}
		return true, limit - 1, now.Add(window).Unix()
	}

	if counter.count >= limit {
		return false, 0, counter.resetTime.Unix()
	}

	counter.count++
	return true, limit - counter.count, counter.resetTime.Unix()
}

// New creates a new rate limiting middleware.
func New(redisClient *redis.Client, config *Config) *RateLimitMiddleware {
	if config == nil {
		config = DefaultConfig()
	}

	m := &RateLimitMiddleware{
		config:       config,
		redisClient:  redisClient,
		localLimiter: NewLocalLimiter(),
		useLocal:     redisClient == nil,
	}

	if redisClient != nil {
		m.slidingWindow = NewSlidingWindowLimiter(redisClient)
	}

	return m
}

// Middleware returns the Gin middleware handler.
func (m *RateLimitMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if rate limiting feature is enabled
		if !features.IsEnabled(features.FeatureRateLimiting) {
			c.Next()
			return
		}

		if !m.config.Enabled {
			c.Next()
			return
		}

		// Determine rate limit tier and key
		tierConfig, key := m.determineTierAndKey(c)
		if tierConfig == nil {
			c.Next()
			return
		}

		// Check rate limit
		allowed, remaining, resetTime := m.checkLimit(c, key, tierConfig)

		// Set rate limit headers
		c.Header(HeaderRateLimitLimit, strconv.Itoa(tierConfig.GetEffectiveLimit()))
		c.Header(HeaderRateLimitRemaining, strconv.Itoa(remaining))
		c.Header(HeaderRateLimitReset, strconv.FormatInt(resetTime, 10))

		if !allowed {
			// Calculate retry after
			retryAfter := resetTime - time.Now().Unix()
			if retryAfter < 1 {
				retryAfter = 1
			}
			c.Header(HeaderRetryAfter, strconv.FormatInt(retryAfter, 10))

			// Record metric
			metrics := otel.GetMetrics()
			metrics.RecordHTTPRequest(c.Request.Context(), c.Request.Method, c.FullPath(), http.StatusTooManyRequests, "")

			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"retry_after": retryAfter,
			})
			return
		}

		c.Next()
	}
}

func (m *RateLimitMiddleware) determineTierAndKey(c *gin.Context) (*TierConfig, string) {
	path := c.FullPath()
	if path == "" {
		path = c.Request.URL.Path
	}

	// Check for API-specific rate limits first
	for pattern, tier := range m.config.API {
		if matchPath(pattern, path) {
			tierCopy := tier
			return &tierCopy, BuildKey(KeyTypeAPI, "", path)
		}
	}

	// Check if user is authenticated
	if userID, exists := c.Get(UserIDContextKey); exists {
		if uid, ok := userID.(string); ok && uid != "" {
			return &m.config.User, BuildKey(KeyTypeUser, uid, "")
		}
	}

	// Fall back to IP-based rate limiting
	clientIP := c.ClientIP()
	return &m.config.IP, BuildKey(KeyTypeIP, clientIP, "")
}

func (m *RateLimitMiddleware) checkLimit(c *gin.Context, key string, tier *TierConfig) (bool, int, int64) {
	limit := tier.GetEffectiveLimit()
	window := tier.GetEffectiveWindow()

	m.mu.RLock()
	useLocal := m.useLocal
	m.mu.RUnlock()

	if useLocal || m.slidingWindow == nil {
		return m.localLimiter.Allow(key, limit, window)
	}

	// Try Redis-based rate limiting
	allowed, remaining, resetTime, err := m.slidingWindow.Allow(c.Request.Context(), key, limit, window)
	if err != nil {
		// Log error and fall back to local limiter if configured
		if m.config.FallbackToLocal {
			m.mu.Lock()
			m.useLocal = true
			m.mu.Unlock()

			// Schedule recovery check
			go m.scheduleRedisRecoveryCheck()

			return m.localLimiter.Allow(key, limit, window)
		}
		// If no fallback, allow the request (fail open)
		return true, limit, time.Now().Add(window).Unix()
	}

	return allowed, remaining, resetTime
}

func (m *RateLimitMiddleware) scheduleRedisRecoveryCheck() {
	time.Sleep(30 * time.Second)

	if m.redisClient == nil {
		return
	}

	// Try to ping Redis
	ctx := context.Background()
	if err := m.redisClient.Ping(ctx).Err(); err == nil {
		m.mu.Lock()
		m.useLocal = false
		m.mu.Unlock()
	} else {
		// Schedule another check
		go m.scheduleRedisRecoveryCheck()
	}
}

// matchPath checks if a path matches a pattern.
// Supports simple wildcard patterns like "/api/v1/edge/*"
func matchPath(pattern, path string) bool {
	if pattern == path {
		return true
	}

	// Handle wildcard patterns
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		return strings.HasPrefix(path, prefix)
	}

	// Handle exact patterns with parameters
	if strings.Contains(pattern, ":") {
		patternParts := strings.Split(pattern, "/")
		pathParts := strings.Split(path, "/")

		if len(patternParts) != len(pathParts) {
			return false
		}

		for i, part := range patternParts {
			if strings.HasPrefix(part, ":") {
				continue // Parameter placeholder, matches anything
			}
			if part != pathParts[i] {
				return false
			}
		}
		return true
	}

	return false
}

// SetConfig updates the rate limiter configuration.
func (m *RateLimitMiddleware) SetConfig(config *Config) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = config
}

// IsUsingLocalFallback returns true if currently using local fallback.
func (m *RateLimitMiddleware) IsUsingLocalFallback() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.useLocal
}
