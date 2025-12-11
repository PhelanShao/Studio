package ratelimit

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// SlidingWindowLimiter implements a sliding window rate limiter using Redis.
type SlidingWindowLimiter struct {
	client *redis.Client
}

// NewSlidingWindowLimiter creates a new sliding window rate limiter.
func NewSlidingWindowLimiter(client *redis.Client) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		client: client,
	}
}

// Allow checks if a request is allowed under the rate limit.
// Returns (allowed, remaining, resetTime, error).
func (l *SlidingWindowLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, int, int64, error) {
	now := time.Now()
	windowStart := now.Add(-window)
	windowEnd := now.Add(window)

	// Use a Lua script for atomic operations
	script := redis.NewScript(`
		local key = KEYS[1]
		local now = tonumber(ARGV[1])
		local window_start = tonumber(ARGV[2])
		local limit = tonumber(ARGV[3])
		local window_ms = tonumber(ARGV[4])

		-- Remove old entries outside the window
		redis.call('ZREMRANGEBYSCORE', key, '-inf', window_start)

		-- Count current requests in window
		local count = redis.call('ZCARD', key)

		if count < limit then
			-- Add the current request
			redis.call('ZADD', key, now, now .. ':' .. math.random())
			-- Set expiration
			redis.call('PEXPIRE', key, window_ms)
			return {1, limit - count - 1, 0}
		else
			-- Get the oldest entry to calculate retry time
			local oldest = redis.call('ZRANGE', key, 0, 0, 'WITHSCORES')
			local retry_after = 0
			if oldest[2] then
				retry_after = math.ceil((tonumber(oldest[2]) + window_ms - now) / 1000)
				if retry_after < 0 then retry_after = 0 end
			end
			return {0, 0, retry_after}
		end
	`)

	nowMs := now.UnixMilli()
	windowStartMs := windowStart.UnixMilli()
	windowMs := window.Milliseconds()

	result, err := script.Run(ctx, l.client, []string{key},
		nowMs, windowStartMs, limit, windowMs).Slice()
	if err != nil {
		return false, 0, 0, err
	}

	allowed := result[0].(int64) == 1
	remaining := int(result[1].(int64))
	// retryAfter from Lua script is available in result[2] if needed
	_ = result[2].(int64) // retryAfter - currently using windowEnd for reset time

	resetTime := windowEnd.Unix()

	return allowed, remaining, resetTime, nil
}

// TokenBucketLimiter implements a token bucket rate limiter using Redis.
// This is an alternative to sliding window for different use cases.
type TokenBucketLimiter struct {
	client *redis.Client
}

// NewTokenBucketLimiter creates a new token bucket rate limiter.
func NewTokenBucketLimiter(client *redis.Client) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		client: client,
	}
}

// Allow checks if a request is allowed under the token bucket rate limit.
func (l *TokenBucketLimiter) Allow(ctx context.Context, key string, rate float64, burst int) (bool, int, int64, error) {
	now := time.Now()

	// Token bucket Lua script
	script := redis.NewScript(`
		local key = KEYS[1]
		local rate = tonumber(ARGV[1])
		local burst = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])

		local bucket = redis.call('HMGET', key, 'tokens', 'last_update')
		local tokens = tonumber(bucket[1]) or burst
		local last_update = tonumber(bucket[2]) or now

		-- Calculate tokens to add based on time elapsed
		local elapsed = now - last_update
		tokens = math.min(burst, tokens + (elapsed * rate))

		local allowed = 0
		local remaining = 0

		if tokens >= 1 then
			tokens = tokens - 1
			allowed = 1
			remaining = math.floor(tokens)
		end

		-- Update bucket
		redis.call('HMSET', key, 'tokens', tokens, 'last_update', now)
		redis.call('EXPIRE', key, math.ceil(burst / rate) + 1)

		return {allowed, remaining}
	`)

	result, err := script.Run(ctx, l.client, []string{key},
		rate, burst, float64(now.UnixMilli())/1000).Slice()
	if err != nil {
		return false, 0, 0, err
	}

	allowed := result[0].(int64) == 1
	remaining := int(result[1].(int64))
	resetTime := now.Add(time.Duration(float64(burst)/rate) * time.Second).Unix()

	return allowed, remaining, resetTime, nil
}

// GetCurrentCount returns the current request count for a key.
func (l *SlidingWindowLimiter) GetCurrentCount(ctx context.Context, key string, window time.Duration) (int64, error) {
	now := time.Now()
	windowStart := now.Add(-window)

	count, err := l.client.ZCount(ctx, key,
		strconv.FormatInt(windowStart.UnixMilli(), 10),
		strconv.FormatInt(now.UnixMilli(), 10)).Result()
	if err != nil {
		return 0, err
	}

	return count, nil
}

