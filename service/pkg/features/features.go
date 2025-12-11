// Package features provides feature flag functionality for Studio.
package features

import (
	"sync"

	"github.com/spf13/viper"
)

// Known feature flag names
const (
	// Core features
	FeatureWorkflowV2       = "workflow_v2"
	FeatureRealtimeCamera   = "realtime_camera"
	FeatureNewAuthFlow      = "new_auth_flow"
	
	// Experimental features
	FeatureAIAssistant       = "ai_assistant"
	FeatureAdvancedScheduling = "advanced_scheduling"
	
	// Observability features
	FeatureEnhancedTracing = "enhanced_tracing"
	FeatureBusinessMetrics = "business_metrics"
	
	// Security features
	FeatureRateLimiting     = "rate_limiting"
	FeatureRequestValidation = "request_validation"
)

// Manager manages feature flags.
type Manager struct {
	v      *viper.Viper
	mu     sync.RWMutex
	cache  map[string]bool
	prefix string
}

var (
	globalManager *Manager
	managerOnce   sync.Once
)

// Init initializes the global feature flag manager with a Viper instance.
// The viper instance should already have loaded the configuration.
func Init(v *viper.Viper) {
	managerOnce.Do(func() {
		globalManager = &Manager{
			v:      v,
			cache:  make(map[string]bool),
			prefix: "features.",
		}
		globalManager.refreshCache()
	})
}

// GetManager returns the global feature flag manager.
// Returns nil if not initialized.
func GetManager() *Manager {
	return globalManager
}

// IsEnabled checks if a feature is enabled.
// This is a convenience function that uses the global manager.
func IsEnabled(feature string) bool {
	if globalManager == nil {
		return false
	}
	return globalManager.IsEnabled(feature)
}

// IsEnabled checks if a feature is enabled.
func (m *Manager) IsEnabled(feature string) bool {
	m.mu.RLock()
	if val, ok := m.cache[feature]; ok {
		m.mu.RUnlock()
		return val
	}
	m.mu.RUnlock()

	// If not in cache, check viper directly
	return m.v.GetBool(m.prefix + feature)
}

// GetAll returns all feature flags and their values.
func (m *Manager) GetAll() map[string]bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]bool, len(m.cache))
	for k, v := range m.cache {
		result[k] = v
	}
	return result
}

// Refresh refreshes the feature flag cache from the viper configuration.
func (m *Manager) Refresh() {
	m.refreshCache()
}

func (m *Manager) refreshCache() {
	m.mu.Lock()
	defer m.mu.Unlock()

	featuresMap := m.v.GetStringMap("features")
	for key, val := range featuresMap {
		if boolVal, ok := val.(bool); ok {
			m.cache[key] = boolVal
		}
	}
}

// AllFeatures returns a list of all known feature flag names.
func AllFeatures() []string {
	return []string{
		FeatureWorkflowV2,
		FeatureRealtimeCamera,
		FeatureNewAuthFlow,
		FeatureAIAssistant,
		FeatureAdvancedScheduling,
		FeatureEnhancedTracing,
		FeatureBusinessMetrics,
		FeatureRateLimiting,
		FeatureRequestValidation,
	}
}

// DefaultValues returns the default values for all feature flags.
func DefaultValues() map[string]bool {
	return map[string]bool{
		FeatureWorkflowV2:         true,
		FeatureRealtimeCamera:     true,
		FeatureNewAuthFlow:        false,
		FeatureAIAssistant:        false,
		FeatureAdvancedScheduling: false,
		FeatureEnhancedTracing:    true,
		FeatureBusinessMetrics:    true,
		FeatureRateLimiting:       true,
		FeatureRequestValidation:  true,
	}
}

