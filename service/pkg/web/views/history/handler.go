// Package history provides HTTP handlers for execution history APIs.
package history

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/scienceol/studio/service/pkg/common"
	"github.com/scienceol/studio/service/pkg/common/code"
	"github.com/scienceol/studio/service/pkg/common/uuid"
	"github.com/scienceol/studio/service/pkg/model"
	"github.com/scienceol/studio/service/pkg/repo/history"
)

// Handler handles history-related HTTP requests
type Handler struct {
	repo history.HistoryRepo
}

// NewHandler creates a new history handler
func NewHandler() *Handler {
	return &Handler{
		repo: history.New(),
	}
}

// ListWorkflowExecutionsRequest represents the request for listing workflow executions
type ListWorkflowExecutionsRequest struct {
	LabID      int64  `form:"lab_id" binding:"required"`
	WorkflowID *int64 `form:"workflow_id"`
	Status     string `form:"status"`
	StartTime  string `form:"start_time"`
	EndTime    string `form:"end_time"`
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"page_size,default=20"`
}

// WorkflowExecutionResponse represents a workflow execution in response
type WorkflowExecutionResponse struct {
	UUID           uuid.UUID              `json:"uuid"`
	WorkflowUUID   uuid.UUID              `json:"workflow_uuid"`
	WorkflowName   string                 `json:"workflow_name"`
	Status         model.ExecutionStatus  `json:"status"`
	StepsTotal     int                    `json:"steps_total"`
	StepsCompleted int                    `json:"steps_completed"`
	StepsFailed    int                    `json:"steps_failed"`
	DurationMs     int64                  `json:"duration_ms"`
	ErrorMessage   *string                `json:"error_message,omitempty"`
	StartedAt      time.Time              `json:"started_at"`
	CompletedAt    *time.Time             `json:"completed_at,omitempty"`
}

// ListResponse represents a paginated list response
type ListResponse struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// @Summary 获取工作流执行历史列表
// @Description 获取实验室的工作流执行历史记录
// @Tags History
// @Accept json
// @Produce json
// @Param lab_id query int true "实验室ID"
// @Param workflow_id query int false "工作流ID (可选)"
// @Param status query string false "状态过滤 (pending, running, success, failed, cancelled)"
// @Param start_time query string false "开始时间 (RFC3339格式)"
// @Param end_time query string false "结束时间 (RFC3339格式)"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} common.Resp{data=ListResponse}
// @Router /v1/lab/history/workflow [get]
func (h *Handler) ListWorkflowExecutions(ctx *gin.Context) {
	var req ListWorkflowExecutionsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}

	params := model.NewHistoryQueryParams()
	params.LabID = req.LabID
	params.WorkflowID = req.WorkflowID
	params.Page = req.Page
	params.PageSize = req.PageSize

	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > 100 {
		params.PageSize = 20
	}

	if req.Status != "" {
		status := model.ExecutionStatus(req.Status)
		params.Status = &status
	}

	if req.StartTime != "" {
		if t, err := time.Parse(time.RFC3339, req.StartTime); err == nil {
			params.StartTime = &t
		}
	}
	if req.EndTime != "" {
		if t, err := time.Parse(time.RFC3339, req.EndTime); err == nil {
			params.EndTime = &t
		}
	}

	executions, total, err := h.repo.ListWorkflowExecutions(ctx, params)
	if err != nil {
		common.ReplyErr(ctx, err)
		return
	}

	// Convert to response format
	items := make([]WorkflowExecutionResponse, 0, len(executions))
	for _, e := range executions {
		items = append(items, WorkflowExecutionResponse{
			UUID:           e.UUID,
			WorkflowUUID:   e.WorkflowUUID,
			WorkflowName:   e.WorkflowName,
			Status:         e.Status,
			StepsTotal:     e.StepsTotal,
			StepsCompleted: e.StepsCompleted,
			StepsFailed:    e.StepsFailed,
			DurationMs:     e.DurationMs,
			ErrorMessage:   e.ErrorMessage,
			StartedAt:      e.StartedAt,
			CompletedAt:    e.CompletedAt,
		})
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		totalPages++
	}

	common.ReplyOk(ctx, ListResponse{
		Items:      items,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	})
}

// GetWorkflowExecutionRequest represents the request for getting a workflow execution
type GetWorkflowExecutionRequest struct {
	ExecutionUUID string `uri:"execution_uuid" binding:"required"`
}

// WorkflowExecutionDetailResponse represents detailed workflow execution response
type WorkflowExecutionDetailResponse struct {
	WorkflowExecutionResponse
	Actions []ActionExecutionResponse `json:"actions"`
}

// ActionExecutionResponse represents an action execution in response
type ActionExecutionResponse struct {
	UUID         uuid.UUID              `json:"uuid"`
	DeviceUUID   uuid.UUID              `json:"device_uuid"`
	DeviceName   string                 `json:"device_name"`
	ActionType   string                 `json:"action_type"`
	ActionName   string                 `json:"action_name"`
	Status       model.ExecutionStatus  `json:"status"`
	DurationMs   int64                  `json:"duration_ms"`
	ErrorMessage *string                `json:"error_message,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

// @Summary 获取工作流执行详情
// @Description 获取单次工作流执行的详细信息，包含所有动作
// @Tags History
// @Accept json
// @Produce json
// @Param execution_uuid path string true "执行UUID"
// @Success 200 {object} common.Resp{data=WorkflowExecutionDetailResponse}
// @Router /v1/lab/history/workflow/execution/{execution_uuid} [get]
func (h *Handler) GetWorkflowExecution(ctx *gin.Context) {
	var req GetWorkflowExecutionRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}

	execUUID, err := uuid.FromString(req.ExecutionUUID)
	if err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg("invalid execution UUID"))
		return
	}

	exec, err := h.repo.GetWorkflowExecutionByUUID(ctx, execUUID)
	if err != nil {
		common.ReplyErr(ctx, err)
		return
	}

	// Get associated actions
	actions, err := h.repo.ListActionsByWorkflowExecution(ctx, exec.ID)
	if err != nil {
		common.ReplyErr(ctx, err)
		return
	}

	actionResponses := make([]ActionExecutionResponse, 0, len(actions))
	for _, a := range actions {
		actionResponses = append(actionResponses, ActionExecutionResponse{
			UUID:         a.UUID,
			DeviceUUID:   a.DeviceUUID,
			DeviceName:   a.DeviceName,
			ActionType:   a.ActionType,
			ActionName:   a.ActionName,
			Status:       a.Status,
			DurationMs:   a.DurationMs,
			ErrorMessage: a.ErrorMessage,
			CreatedAt:    a.CreatedAt,
		})
	}

	common.ReplyOk(ctx, WorkflowExecutionDetailResponse{
		WorkflowExecutionResponse: WorkflowExecutionResponse{
			UUID:           exec.UUID,
			WorkflowUUID:   exec.WorkflowUUID,
			WorkflowName:   exec.WorkflowName,
			Status:         exec.Status,
			StepsTotal:     exec.StepsTotal,
			StepsCompleted: exec.StepsCompleted,
			StepsFailed:    exec.StepsFailed,
			DurationMs:     exec.DurationMs,
			ErrorMessage:   exec.ErrorMessage,
			StartedAt:      exec.StartedAt,
			CompletedAt:    exec.CompletedAt,
		},
		Actions: actionResponses,
	})
}

// ListDeviceEventsRequest represents the request for listing device events
type ListDeviceEventsRequest struct {
	LabID     int64  `form:"lab_id" binding:"required"`
	DeviceID  *int64 `form:"device_id"`
	EventType string `form:"event_type"`
	StartTime string `form:"start_time"`
	EndTime   string `form:"end_time"`
	Page      int    `form:"page,default=1"`
	PageSize  int    `form:"page_size,default=20"`
}

// DeviceEventResponse represents a device event in response
type DeviceEventResponse struct {
	UUID       uuid.UUID           `json:"uuid"`
	DeviceUUID uuid.UUID           `json:"device_uuid"`
	EventType  model.DeviceEventType `json:"event_type"`
	EventData  interface{}         `json:"event_data"`
	Timestamp  time.Time           `json:"timestamp"`
}

// @Summary 获取设备事件历史
// @Description 获取实验室的设备事件历史记录
// @Tags History
// @Accept json
// @Produce json
// @Param lab_id query int true "实验室ID"
// @Param device_id query int false "设备ID (可选)"
// @Param event_type query string false "事件类型过滤"
// @Param start_time query string false "开始时间 (RFC3339格式)"
// @Param end_time query string false "结束时间 (RFC3339格式)"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} common.Resp{data=ListResponse}
// @Router /v1/lab/history/device [get]
func (h *Handler) ListDeviceEvents(ctx *gin.Context) {
	var req ListDeviceEventsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg(err.Error()))
		return
	}

	params := model.NewHistoryQueryParams()
	params.LabID = req.LabID
	params.DeviceID = req.DeviceID
	params.Page = req.Page
	params.PageSize = req.PageSize

	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > 100 {
		params.PageSize = 20
	}

	if req.EventType != "" {
		eventType := model.DeviceEventType(req.EventType)
		params.EventType = &eventType
	}

	if req.StartTime != "" {
		if t, err := time.Parse(time.RFC3339, req.StartTime); err == nil {
			params.StartTime = &t
		}
	}
	if req.EndTime != "" {
		if t, err := time.Parse(time.RFC3339, req.EndTime); err == nil {
			params.EndTime = &t
		}
	}

	events, total, err := h.repo.ListDeviceEvents(ctx, params)
	if err != nil {
		common.ReplyErr(ctx, err)
		return
	}

	items := make([]DeviceEventResponse, 0, len(events))
	for _, e := range events {
		items = append(items, DeviceEventResponse{
			UUID:       e.UUID,
			DeviceUUID: e.DeviceUUID,
			EventType:  e.EventType,
			EventData:  e.EventData,
			Timestamp:  e.Timestamp,
		})
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		totalPages++
	}

	common.ReplyOk(ctx, ListResponse{
		Items:      items,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	})
}

// GetLabStatsRequest represents the request for getting lab stats
type GetLabStatsRequest struct {
	LabID     int64  `uri:"lab_id" binding:"required"`
	StartTime string `form:"start_time"`
	EndTime   string `form:"end_time"`
}

// @Summary 获取实验室使用统计
// @Description 获取实验室的工作流执行统计数据
// @Tags History
// @Accept json
// @Produce json
// @Param lab_id path int true "实验室ID"
// @Param start_time query string false "开始时间 (RFC3339格式)"
// @Param end_time query string false "结束时间 (RFC3339格式)"
// @Success 200 {object} common.Resp{data=model.HistoryStats}
// @Router /v1/lab/{lab_id}/stats [get]
func (h *Handler) GetLabStats(ctx *gin.Context) {
	labIDStr := ctx.Param("lab_id")
	labID, err := strconv.ParseInt(labIDStr, 10, 64)
	if err != nil {
		common.ReplyErr(ctx, code.ParamErr.WithMsg("invalid lab_id"))
		return
	}

	var startTime, endTime *time.Time
	if st := ctx.Query("start_time"); st != "" {
		if t, err := time.Parse(time.RFC3339, st); err == nil {
			startTime = &t
		}
	}
	if et := ctx.Query("end_time"); et != "" {
		if t, err := time.Parse(time.RFC3339, et); err == nil {
			endTime = &t
		}
	}

	stats, err := h.repo.GetLabStats(ctx, labID, startTime, endTime)
	if err != nil {
		common.ReplyErr(ctx, err)
		return
	}

	common.ReplyOk(ctx, stats)
}

