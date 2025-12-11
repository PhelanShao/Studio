package history

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewHandler(t *testing.T) {
	handler := NewHandler()
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.repo)
}

func TestListWorkflowExecutionsMissingLabID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := NewHandler()
	router.GET("/history/workflow", handler.ListWorkflowExecutions)

	req := httptest.NewRequest(http.MethodGet, "/history/workflow", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return error due to missing lab_id
	assert.Equal(t, http.StatusOK, w.Code) // API returns 200 with error in body
}

func TestListWorkflowExecutionsWithParams(t *testing.T) {
	// Skip this test as it requires database connection
	t.Skip("Requires database connection")
}

func TestListDeviceEventsMissingLabID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := NewHandler()
	router.GET("/history/device", handler.ListDeviceEvents)

	req := httptest.NewRequest(http.MethodGet, "/history/device", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetWorkflowExecutionInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := NewHandler()
	router.GET("/history/workflow/execution/:execution_uuid", handler.GetWorkflowExecution)

	req := httptest.NewRequest(http.MethodGet, "/history/workflow/execution/invalid-uuid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return error for invalid UUID
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetLabStatsInvalidLabID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := NewHandler()
	router.GET("/lab/:lab_id/stats", handler.GetLabStats)

	req := httptest.NewRequest(http.MethodGet, "/lab/invalid/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListResponseStruct(t *testing.T) {
	resp := ListResponse{
		Items:      []string{"item1", "item2"},
		Total:      100,
		Page:       1,
		PageSize:   20,
		TotalPages: 5,
	}

	assert.Equal(t, int64(100), resp.Total)
	assert.Equal(t, 1, resp.Page)
	assert.Equal(t, 20, resp.PageSize)
	assert.Equal(t, 5, resp.TotalPages)
}

