package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExecutionStatus(t *testing.T) {
	tests := []struct {
		status ExecutionStatus
		want   string
	}{
		{ExecutionStatusPending, "pending"},
		{ExecutionStatusRunning, "running"},
		{ExecutionStatusSuccess, "success"},
		{ExecutionStatusFailed, "failed"},
		{ExecutionStatusCancelled, "cancelled"},
		{ExecutionStatusTimeout, "timeout"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			assert.Equal(t, tt.want, string(tt.status))
		})
	}
}

func TestDeviceEventType(t *testing.T) {
	tests := []struct {
		eventType DeviceEventType
		want      string
	}{
		{DeviceEventStatusChange, "status_change"},
		{DeviceEventDataReceived, "data_received"},
		{DeviceEventError, "error"},
		{DeviceEventConnected, "connected"},
		{DeviceEventDisconnected, "disconnected"},
		{DeviceEventCommandSent, "command_sent"},
		{DeviceEventCommandResult, "command_result"},
	}

	for _, tt := range tests {
		t.Run(string(tt.eventType), func(t *testing.T) {
			assert.Equal(t, tt.want, string(tt.eventType))
		})
	}
}

func TestWorkflowExecutionHistoryTableName(t *testing.T) {
	weh := &WorkflowExecutionHistory{}
	assert.Equal(t, "workflow_execution_history", weh.TableName())
}

func TestActionExecutionHistoryTableName(t *testing.T) {
	aeh := &ActionExecutionHistory{}
	assert.Equal(t, "action_execution_history", aeh.TableName())
}

func TestDeviceEventHistoryTableName(t *testing.T) {
	deh := &DeviceEventHistory{}
	assert.Equal(t, "device_event_history", deh.TableName())
}

func TestNewHistoryQueryParams(t *testing.T) {
	params := NewHistoryQueryParams()

	assert.Equal(t, 1, params.Page)
	assert.Equal(t, 20, params.PageSize)
	assert.Equal(t, int64(0), params.LabID)
	assert.Nil(t, params.WorkflowID)
	assert.Nil(t, params.DeviceID)
	assert.Nil(t, params.Status)
	assert.Nil(t, params.EventType)
	assert.Nil(t, params.StartTime)
	assert.Nil(t, params.EndTime)
}

func TestHistoryQueryParamsWithFilters(t *testing.T) {
	params := NewHistoryQueryParams()
	params.LabID = 123
	params.Page = 2
	params.PageSize = 50

	workflowID := int64(456)
	params.WorkflowID = &workflowID

	status := ExecutionStatusSuccess
	params.Status = &status

	now := time.Now()
	params.StartTime = &now

	assert.Equal(t, int64(123), params.LabID)
	assert.Equal(t, 2, params.Page)
	assert.Equal(t, 50, params.PageSize)
	assert.Equal(t, int64(456), *params.WorkflowID)
	assert.Equal(t, ExecutionStatusSuccess, *params.Status)
	assert.NotNil(t, params.StartTime)
}

func TestHistoryStats(t *testing.T) {
	stats := &HistoryStats{
		TotalExecutions:   100,
		SuccessfulCount:   80,
		FailedCount:       20,
		SuccessRate:       80.0,
		AverageDurationMs: 1500.5,
		TotalActionsCount: 500,
		TotalDeviceEvents: 1000,
	}

	assert.Equal(t, int64(100), stats.TotalExecutions)
	assert.Equal(t, int64(80), stats.SuccessfulCount)
	assert.Equal(t, int64(20), stats.FailedCount)
	assert.Equal(t, 80.0, stats.SuccessRate)
	assert.Equal(t, 1500.5, stats.AverageDurationMs)
	assert.Equal(t, int64(500), stats.TotalActionsCount)
	assert.Equal(t, int64(1000), stats.TotalDeviceEvents)
}

