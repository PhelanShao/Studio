package model

import (
	"time"

	"github.com/scienceol/studio/service/pkg/common/uuid"
	"gorm.io/datatypes"
)

// ExecutionStatus represents the status of an execution
type ExecutionStatus string

const (
	ExecutionStatusPending   ExecutionStatus = "pending"
	ExecutionStatusRunning   ExecutionStatus = "running"
	ExecutionStatusSuccess   ExecutionStatus = "success"
	ExecutionStatusFailed    ExecutionStatus = "failed"
	ExecutionStatusCancelled ExecutionStatus = "cancelled"
	ExecutionStatusTimeout   ExecutionStatus = "timeout"
)

// WorkflowExecutionHistory records the history of workflow executions
type WorkflowExecutionHistory struct {
	BaseModel
	LabID          int64           `gorm:"type:bigint;not null;index:idx_weh_lab" json:"lab_id"`
	UserID         string          `gorm:"type:varchar(120);not null;index:idx_weh_user" json:"user_id"`
	WorkflowID     int64           `gorm:"type:bigint;not null;index:idx_weh_workflow" json:"workflow_id"`
	WorkflowUUID   uuid.UUID       `gorm:"type:uuid;not null" json:"workflow_uuid"`
	WorkflowName   string          `gorm:"type:varchar(255);not null" json:"workflow_name"`
	Status         ExecutionStatus `gorm:"type:varchar(50);not null;default:'pending';index:idx_weh_status" json:"status"`
	StepsTotal     int             `gorm:"type:int;not null;default:0" json:"steps_total"`
	StepsCompleted int             `gorm:"type:int;not null;default:0" json:"steps_completed"`
	StepsFailed    int             `gorm:"type:int;not null;default:0" json:"steps_failed"`
	DurationMs     int64           `gorm:"type:bigint;default:0" json:"duration_ms"`
	ErrorMessage   *string         `gorm:"type:text" json:"error_message"`
	Result         datatypes.JSON  `gorm:"type:jsonb" json:"result"`
	StartedAt      time.Time       `gorm:"not null;index:idx_weh_started" json:"started_at"`
	CompletedAt    *time.Time      `json:"completed_at"`
	Metadata       datatypes.JSON  `gorm:"type:jsonb" json:"metadata"`
}

func (*WorkflowExecutionHistory) TableName() string {
	return "workflow_execution_history"
}

// ActionExecutionHistory records the history of device action executions
type ActionExecutionHistory struct {
	BaseModel
	WorkflowExecutionID *int64          `gorm:"type:bigint;index:idx_aeh_wf_exec" json:"workflow_execution_id"`
	LabID               int64           `gorm:"type:bigint;not null;index:idx_aeh_lab" json:"lab_id"`
	DeviceID            int64           `gorm:"type:bigint;not null;index:idx_aeh_device" json:"device_id"`
	DeviceUUID          uuid.UUID       `gorm:"type:uuid;not null" json:"device_uuid"`
	DeviceName          string          `gorm:"type:varchar(255);not null" json:"device_name"`
	ActionType          string          `gorm:"type:varchar(100);not null;index:idx_aeh_action" json:"action_type"`
	ActionName          string          `gorm:"type:varchar(255);not null" json:"action_name"`
	Input               datatypes.JSON  `gorm:"type:jsonb" json:"input"`
	Output              datatypes.JSON  `gorm:"type:jsonb" json:"output"`
	Status              ExecutionStatus `gorm:"type:varchar(50);not null;default:'pending';index:idx_aeh_status" json:"status"`
	DurationMs          int64           `gorm:"type:bigint;default:0" json:"duration_ms"`
	ErrorMessage        *string         `gorm:"type:text" json:"error_message"`
	Metadata            datatypes.JSON  `gorm:"type:jsonb" json:"metadata"`
}

func (*ActionExecutionHistory) TableName() string {
	return "action_execution_history"
}

// DeviceEventType represents the type of device event
type DeviceEventType string

const (
	DeviceEventStatusChange  DeviceEventType = "status_change"
	DeviceEventDataReceived  DeviceEventType = "data_received"
	DeviceEventError         DeviceEventType = "error"
	DeviceEventConnected     DeviceEventType = "connected"
	DeviceEventDisconnected  DeviceEventType = "disconnected"
	DeviceEventCommandSent   DeviceEventType = "command_sent"
	DeviceEventCommandResult DeviceEventType = "command_result"
)

// DeviceEventHistory records device events
type DeviceEventHistory struct {
	BaseModel
	LabID     int64           `gorm:"type:bigint;not null;index:idx_deh_lab" json:"lab_id"`
	DeviceID  int64           `gorm:"type:bigint;not null;index:idx_deh_device" json:"device_id"`
	DeviceUUID uuid.UUID      `gorm:"type:uuid;not null" json:"device_uuid"`
	EventType DeviceEventType `gorm:"type:varchar(50);not null;index:idx_deh_type" json:"event_type"`
	EventData datatypes.JSON  `gorm:"type:jsonb" json:"event_data"`
	Timestamp time.Time       `gorm:"not null;index:idx_deh_time" json:"timestamp"`
}

func (*DeviceEventHistory) TableName() string {
	return "device_event_history"
}

// HistoryQueryParams represents query parameters for history queries
type HistoryQueryParams struct {
	LabID      int64
	UserID     string
	WorkflowID *int64
	DeviceID   *int64
	Status     *ExecutionStatus
	EventType  *DeviceEventType
	StartTime  *time.Time
	EndTime    *time.Time
	Page       int
	PageSize   int
}

// NewHistoryQueryParams creates a new HistoryQueryParams with defaults
func NewHistoryQueryParams() *HistoryQueryParams {
	return &HistoryQueryParams{
		Page:     1,
		PageSize: 20,
	}
}

// HistoryStats represents aggregated statistics
type HistoryStats struct {
	TotalExecutions    int64   `json:"total_executions"`
	SuccessfulCount    int64   `json:"successful_count"`
	FailedCount        int64   `json:"failed_count"`
	SuccessRate        float64 `json:"success_rate"`
	AverageDurationMs  float64 `json:"average_duration_ms"`
	TotalActionsCount  int64   `json:"total_actions_count"`
	TotalDeviceEvents  int64   `json:"total_device_events"`
}

