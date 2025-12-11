package features

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestFeatureFlags(t *testing.T) {
	// Create a new viper instance with test configuration
	v := viper.New()
	v.Set("features.workflow_v2", true)
	v.Set("features.realtime_camera", true)
	v.Set("features.new_auth_flow", false)
	v.Set("features.ai_assistant", false)
	v.Set("features.rate_limiting", true)

	// Initialize with our test viper
	manager := &Manager{
		v:      v,
		cache:  make(map[string]bool),
		prefix: "features.",
	}
	manager.refreshCache()

	// Test IsEnabled for enabled features
	assert.True(t, manager.IsEnabled("workflow_v2"), "workflow_v2 should be enabled")
	assert.True(t, manager.IsEnabled("realtime_camera"), "realtime_camera should be enabled")
	assert.True(t, manager.IsEnabled("rate_limiting"), "rate_limiting should be enabled")

	// Test IsEnabled for disabled features
	assert.False(t, manager.IsEnabled("new_auth_flow"), "new_auth_flow should be disabled")
	assert.False(t, manager.IsEnabled("ai_assistant"), "ai_assistant should be disabled")

	// Test GetAll
	all := manager.GetAll()
	assert.Len(t, all, 5, "should have 5 features")
	assert.True(t, all["workflow_v2"])
	assert.False(t, all["new_auth_flow"])
}

func TestDefaultValues(t *testing.T) {
	defaults := DefaultValues()

	assert.True(t, defaults[FeatureWorkflowV2])
	assert.True(t, defaults[FeatureRealtimeCamera])
	assert.False(t, defaults[FeatureNewAuthFlow])
	assert.False(t, defaults[FeatureAIAssistant])
	assert.True(t, defaults[FeatureRateLimiting])
}

func TestAllFeatures(t *testing.T) {
	features := AllFeatures()

	assert.Contains(t, features, FeatureWorkflowV2)
	assert.Contains(t, features, FeatureRealtimeCamera)
	assert.Contains(t, features, FeatureNewAuthFlow)
	assert.Contains(t, features, FeatureAIAssistant)
	assert.Contains(t, features, FeatureRateLimiting)
	assert.Len(t, features, 9, "should have 9 known features")
}

func TestManagerRefresh(t *testing.T) {
	v := viper.New()
	v.Set("features.test_feature", false)

	manager := &Manager{
		v:      v,
		cache:  make(map[string]bool),
		prefix: "features.",
	}
	manager.refreshCache()

	assert.False(t, manager.IsEnabled("test_feature"))

	// Update config
	v.Set("features.test_feature", true)
	manager.Refresh()

	assert.True(t, manager.IsEnabled("test_feature"))
}

func TestIsEnabledWithoutInit(t *testing.T) {
	// Reset global manager for this test
	oldManager := globalManager
	globalManager = nil
	defer func() { globalManager = oldManager }()

	// Should return false when not initialized
	assert.False(t, IsEnabled("any_feature"))
}

