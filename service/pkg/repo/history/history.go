// Package history provides repository operations for execution history records.
package history

import (
	"context"
	"time"

	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/model"
	"github.com/scienceol/studio/service/pkg/repo"
	"gorm.io/gorm"
)

// HistoryRepo defines the interface for history repository operations
type HistoryRepo interface {
	// Workflow Execution History
	CreateWorkflowExecution(ctx context.Context, exec *model.WorkflowExecutionHistory) error
	UpdateWorkflowExecution(ctx context.Context, id int64, updates map[string]interface{}) error
	GetWorkflowExecution(ctx context.Context, id int64) (*model.WorkflowExecutionHistory, error)
	GetWorkflowExecutionByUUID(ctx context.Context, uuid uuid.UUID) (*model.WorkflowExecutionHistory, error)
	ListWorkflowExecutions(ctx context.Context, params *model.HistoryQueryParams) ([]*model.WorkflowExecutionHistory, int64, error)

	// Action Execution History
	CreateActionExecution(ctx context.Context, exec *model.ActionExecutionHistory) error
	CreateActionExecutionBatch(ctx context.Context, execs []*model.ActionExecutionHistory) error
	ListActionExecutions(ctx context.Context, params *model.HistoryQueryParams) ([]*model.ActionExecutionHistory, int64, error)
	ListActionsByWorkflowExecution(ctx context.Context, workflowExecID int64) ([]*model.ActionExecutionHistory, error)

	// Device Event History
	CreateDeviceEvent(ctx context.Context, event *model.DeviceEventHistory) error
	CreateDeviceEventBatch(ctx context.Context, events []*model.DeviceEventHistory) error
	ListDeviceEvents(ctx context.Context, params *model.HistoryQueryParams) ([]*model.DeviceEventHistory, int64, error)

	// Statistics
	GetLabStats(ctx context.Context, labID int64, startTime, endTime *time.Time) (*model.HistoryStats, error)

	// Cleanup
	CleanupOldRecords(ctx context.Context, before time.Time) (int64, error)
}

type historyImpl struct {
	repo.IDOrUUIDTranslate
}

// New creates a new history repository instance
func New() HistoryRepo {
	return &historyImpl{
		IDOrUUIDTranslate: repo.NewBaseDB(),
	}
}

// CreateWorkflowExecution creates a new workflow execution history record
func (h *historyImpl) CreateWorkflowExecution(ctx context.Context, exec *model.WorkflowExecutionHistory) error {
	if err := h.DBWithContext(ctx).Create(exec).Error; err != nil {
		logger.Errorf(ctx, "CreateWorkflowExecution fail: %+v", err)
		return code.CreateDataErr.WithErr(err)
	}
	return nil
}

// UpdateWorkflowExecution updates a workflow execution history record
func (h *historyImpl) UpdateWorkflowExecution(ctx context.Context, id int64, updates map[string]interface{}) error {
	if err := h.DBWithContext(ctx).Model(&model.WorkflowExecutionHistory{}).
		Where("id = ?", id).Updates(updates).Error; err != nil {
		logger.Errorf(ctx, "UpdateWorkflowExecution fail id=%d: %+v", id, err)
		return code.UpdateDataErr.WithErr(err)
	}
	return nil
}

// GetWorkflowExecution retrieves a workflow execution by ID
func (h *historyImpl) GetWorkflowExecution(ctx context.Context, id int64) (*model.WorkflowExecutionHistory, error) {
	var exec model.WorkflowExecutionHistory
	if err := h.DBWithContext(ctx).Where("id = ?", id).First(&exec).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, code.RecordNotFound
		}
		logger.Errorf(ctx, "GetWorkflowExecution fail id=%d: %+v", id, err)
		return nil, code.QueryRecordErr.WithErr(err)
	}
	return &exec, nil
}

// GetWorkflowExecutionByUUID retrieves a workflow execution by UUID
func (h *historyImpl) GetWorkflowExecutionByUUID(ctx context.Context, uuid uuid.UUID) (*model.WorkflowExecutionHistory, error) {
	var exec model.WorkflowExecutionHistory
	if err := h.DBWithContext(ctx).Where("uuid = ?", uuid).First(&exec).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, code.RecordNotFound
		}
		logger.Errorf(ctx, "GetWorkflowExecutionByUUID fail uuid=%s: %+v", uuid, err)
		return nil, code.QueryRecordErr.WithErr(err)
	}
	return &exec, nil
}

// ListWorkflowExecutions lists workflow executions with pagination
func (h *historyImpl) ListWorkflowExecutions(ctx context.Context, params *model.HistoryQueryParams) ([]*model.WorkflowExecutionHistory, int64, error) {
	var executions []*model.WorkflowExecutionHistory
	var total int64

	query := h.DBWithContext(ctx).Model(&model.WorkflowExecutionHistory{})
	query = h.applyWorkflowFilters(query, params)

	// Count total
	if err := query.Count(&total).Error; err != nil {
		logger.Errorf(ctx, "ListWorkflowExecutions count fail: %+v", err)
		return nil, 0, code.QueryRecordErr.WithErr(err)
	}

	// Get paginated results
	offset := (params.Page - 1) * params.PageSize
	if err := query.Order("started_at DESC").Offset(offset).Limit(params.PageSize).Find(&executions).Error; err != nil {
		logger.Errorf(ctx, "ListWorkflowExecutions find fail: %+v", err)
		return nil, 0, code.QueryRecordErr.WithErr(err)
	}

	return executions, total, nil
}

func (h *historyImpl) applyWorkflowFilters(query *gorm.DB, params *model.HistoryQueryParams) *gorm.DB {
	if params.LabID > 0 {
		query = query.Where("lab_id = ?", params.LabID)
	}
	if params.UserID != "" {
		query = query.Where("user_id = ?", params.UserID)
	}
	if params.WorkflowID != nil {
		query = query.Where("workflow_id = ?", *params.WorkflowID)
	}
	if params.Status != nil {
		query = query.Where("status = ?", *params.Status)
	}
	if params.StartTime != nil {
		query = query.Where("started_at >= ?", *params.StartTime)
	}
	if params.EndTime != nil {
		query = query.Where("started_at <= ?", *params.EndTime)
	}
	return query
}

// CreateActionExecution creates a new action execution history record
func (h *historyImpl) CreateActionExecution(ctx context.Context, exec *model.ActionExecutionHistory) error {
	if err := h.DBWithContext(ctx).Create(exec).Error; err != nil {
		logger.Errorf(ctx, "CreateActionExecution fail: %+v", err)
		return code.CreateDataErr.WithErr(err)
	}
	return nil
}

// CreateActionExecutionBatch creates multiple action execution records in batch
func (h *historyImpl) CreateActionExecutionBatch(ctx context.Context, execs []*model.ActionExecutionHistory) error {
	if len(execs) == 0 {
		return nil
	}
	if err := h.DBWithContext(ctx).CreateInBatches(execs, 100).Error; err != nil {
		logger.Errorf(ctx, "CreateActionExecutionBatch fail: %+v", err)
		return code.CreateDataErr.WithErr(err)
	}
	return nil
}

// ListActionExecutions lists action executions with pagination
func (h *historyImpl) ListActionExecutions(ctx context.Context, params *model.HistoryQueryParams) ([]*model.ActionExecutionHistory, int64, error) {
	var executions []*model.ActionExecutionHistory
	var total int64

	query := h.DBWithContext(ctx).Model(&model.ActionExecutionHistory{})
	query = h.applyActionFilters(query, params)

	if err := query.Count(&total).Error; err != nil {
		logger.Errorf(ctx, "ListActionExecutions count fail: %+v", err)
		return nil, 0, code.QueryRecordErr.WithErr(err)
	}

	offset := (params.Page - 1) * params.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(params.PageSize).Find(&executions).Error; err != nil {
		logger.Errorf(ctx, "ListActionExecutions find fail: %+v", err)
		return nil, 0, code.QueryRecordErr.WithErr(err)
	}

	return executions, total, nil
}

// ListActionsByWorkflowExecution retrieves all actions for a workflow execution
func (h *historyImpl) ListActionsByWorkflowExecution(ctx context.Context, workflowExecID int64) ([]*model.ActionExecutionHistory, error) {
	var executions []*model.ActionExecutionHistory
	if err := h.DBWithContext(ctx).Where("workflow_execution_id = ?", workflowExecID).
		Order("created_at ASC").Find(&executions).Error; err != nil {
		logger.Errorf(ctx, "ListActionsByWorkflowExecution fail: %+v", err)
		return nil, code.QueryRecordErr.WithErr(err)
	}
	return executions, nil
}

func (h *historyImpl) applyActionFilters(query *gorm.DB, params *model.HistoryQueryParams) *gorm.DB {
	if params.LabID > 0 {
		query = query.Where("lab_id = ?", params.LabID)
	}
	if params.DeviceID != nil {
		query = query.Where("device_id = ?", *params.DeviceID)
	}
	if params.Status != nil {
		query = query.Where("status = ?", *params.Status)
	}
	if params.StartTime != nil {
		query = query.Where("created_at >= ?", *params.StartTime)
	}
	if params.EndTime != nil {
		query = query.Where("created_at <= ?", *params.EndTime)
	}
	return query
}

// CreateDeviceEvent creates a new device event history record
func (h *historyImpl) CreateDeviceEvent(ctx context.Context, event *model.DeviceEventHistory) error {
	if err := h.DBWithContext(ctx).Create(event).Error; err != nil {
		logger.Errorf(ctx, "CreateDeviceEvent fail: %+v", err)
		return code.CreateDataErr.WithErr(err)
	}
	return nil
}

// CreateDeviceEventBatch creates multiple device events in batch
func (h *historyImpl) CreateDeviceEventBatch(ctx context.Context, events []*model.DeviceEventHistory) error {
	if len(events) == 0 {
		return nil
	}
	if err := h.DBWithContext(ctx).CreateInBatches(events, 100).Error; err != nil {
		logger.Errorf(ctx, "CreateDeviceEventBatch fail: %+v", err)
		return code.CreateDataErr.WithErr(err)
	}
	return nil
}

// ListDeviceEvents lists device events with pagination
func (h *historyImpl) ListDeviceEvents(ctx context.Context, params *model.HistoryQueryParams) ([]*model.DeviceEventHistory, int64, error) {
	var events []*model.DeviceEventHistory
	var total int64

	query := h.DBWithContext(ctx).Model(&model.DeviceEventHistory{})
	query = h.applyDeviceEventFilters(query, params)

	if err := query.Count(&total).Error; err != nil {
		logger.Errorf(ctx, "ListDeviceEvents count fail: %+v", err)
		return nil, 0, code.QueryRecordErr.WithErr(err)
	}

	offset := (params.Page - 1) * params.PageSize
	if err := query.Order("timestamp DESC").Offset(offset).Limit(params.PageSize).Find(&events).Error; err != nil {
		logger.Errorf(ctx, "ListDeviceEvents find fail: %+v", err)
		return nil, 0, code.QueryRecordErr.WithErr(err)
	}

	return events, total, nil
}

func (h *historyImpl) applyDeviceEventFilters(query *gorm.DB, params *model.HistoryQueryParams) *gorm.DB {
	if params.LabID > 0 {
		query = query.Where("lab_id = ?", params.LabID)
	}
	if params.DeviceID != nil {
		query = query.Where("device_id = ?", *params.DeviceID)
	}
	if params.EventType != nil {
		query = query.Where("event_type = ?", *params.EventType)
	}
	if params.StartTime != nil {
		query = query.Where("timestamp >= ?", *params.StartTime)
	}
	if params.EndTime != nil {
		query = query.Where("timestamp <= ?", *params.EndTime)
	}
	return query
}

// GetLabStats retrieves aggregated statistics for a lab
func (h *historyImpl) GetLabStats(ctx context.Context, labID int64, startTime, endTime *time.Time) (*model.HistoryStats, error) {
	stats := &model.HistoryStats{}

	// Workflow execution stats
	wfQuery := h.DBWithContext(ctx).Model(&model.WorkflowExecutionHistory{}).Where("lab_id = ?", labID)
	if startTime != nil {
		wfQuery = wfQuery.Where("started_at >= ?", *startTime)
	}
	if endTime != nil {
		wfQuery = wfQuery.Where("started_at <= ?", *endTime)
	}

	wfQuery.Count(&stats.TotalExecutions)
	wfQuery.Where("status = ?", model.ExecutionStatusSuccess).Count(&stats.SuccessfulCount)
	wfQuery.Where("status = ?", model.ExecutionStatusFailed).Count(&stats.FailedCount)

	if stats.TotalExecutions > 0 {
		stats.SuccessRate = float64(stats.SuccessfulCount) / float64(stats.TotalExecutions) * 100
	}

	// Average duration
	var avgDuration struct{ Avg float64 }
	h.DBWithContext(ctx).Model(&model.WorkflowExecutionHistory{}).
		Where("lab_id = ? AND duration_ms > 0", labID).
		Select("AVG(duration_ms) as avg").Scan(&avgDuration)
	stats.AverageDurationMs = avgDuration.Avg

	// Action count
	actionQuery := h.DBWithContext(ctx).Model(&model.ActionExecutionHistory{}).Where("lab_id = ?", labID)
	if startTime != nil {
		actionQuery = actionQuery.Where("created_at >= ?", *startTime)
	}
	if endTime != nil {
		actionQuery = actionQuery.Where("created_at <= ?", *endTime)
	}
	actionQuery.Count(&stats.TotalActionsCount)

	// Device event count
	eventQuery := h.DBWithContext(ctx).Model(&model.DeviceEventHistory{}).Where("lab_id = ?", labID)
	if startTime != nil {
		eventQuery = eventQuery.Where("timestamp >= ?", *startTime)
	}
	if endTime != nil {
		eventQuery = eventQuery.Where("timestamp <= ?", *endTime)
	}
	eventQuery.Count(&stats.TotalDeviceEvents)

	return stats, nil
}

// CleanupOldRecords removes records older than the specified time
func (h *historyImpl) CleanupOldRecords(ctx context.Context, before time.Time) (int64, error) {
	var totalDeleted int64

	// Cleanup workflow executions
	result := h.DBWithContext(ctx).Where("started_at < ?", before).Delete(&model.WorkflowExecutionHistory{})
	if result.Error != nil {
		logger.Errorf(ctx, "CleanupOldRecords workflow fail: %+v", result.Error)
		return 0, code.DeleteDataErr.WithErr(result.Error)
	}
	totalDeleted += result.RowsAffected

	// Cleanup action executions
	result = h.DBWithContext(ctx).Where("created_at < ?", before).Delete(&model.ActionExecutionHistory{})
	if result.Error != nil {
		logger.Errorf(ctx, "CleanupOldRecords action fail: %+v", result.Error)
		return totalDeleted, code.DeleteDataErr.WithErr(result.Error)
	}
	totalDeleted += result.RowsAffected

	// Cleanup device events
	result = h.DBWithContext(ctx).Where("timestamp < ?", before).Delete(&model.DeviceEventHistory{})
	if result.Error != nil {
		logger.Errorf(ctx, "CleanupOldRecords device fail: %+v", result.Error)
		return totalDeleted, code.DeleteDataErr.WithErr(result.Error)
	}
	totalDeleted += result.RowsAffected

	return totalDeleted, nil
}

