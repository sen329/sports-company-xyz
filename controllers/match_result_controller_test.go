package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sports-backend-api/models"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockMatchResultRepository is a mock for MatchResultRepository
type MockMatchResultRepository struct {
	mock.Mock
}

func (m *MockMatchResultRepository) CreateMatchResult(result *models.MatchResult) error {
	args := m.Called(result)
	return args.Error(0)
}

func (m *MockMatchResultRepository) GetMatchResultByMatchID(matchID int64) (*models.MatchResult, error) {
	args := m.Called(matchID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.MatchResult), args.Error(1)
}

func (m *MockMatchResultRepository) CheckResultExists(matchID int64) (bool, error) {
	args := m.Called(matchID)
	return args.Bool(0), args.Error(1)
}

func setupMatchResultRouter(resultRepo *MockMatchResultRepository, matchRepo *MockMatchScheduleRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	controller := &MatchResultController{
		resultRepo: resultRepo,
		matchRepo:  matchRepo,
	}
	router.POST("/match-results", controller.CreateMatchResult)
	router.GET("/match-results/:match_id", controller.GetMatchResultByMatchID)
	return router
}

func TestCreateMatchResult(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		resultRepo := new(MockMatchResultRepository)
		matchRepo := new(MockMatchScheduleRepository)
		router := setupMatchResultRouter(resultRepo, matchRepo)

		reqBody := models.MatchResultRequest{MatchId: 1, HomeScore: 2, AwayScore: 1}
		jsonBody, _ := json.Marshal(reqBody)

		matchRepo.On("GetMatchScheduleByID", int64(1)).Return(&models.MatchScheduleDetail{}, nil)
		resultRepo.On("CheckResultExists", int64(1)).Return(false, nil)
		resultRepo.On("CreateMatchResult", mock.AnythingOfType("*models.MatchResult")).Return(nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/match-results", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response models.MatchResult
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, int64(1), response.MatchId)
		assert.Equal(t, 2, response.HomeScore)
		matchRepo.AssertExpectations(t)
		resultRepo.AssertExpectations(t)
	})

	t.Run("Match Not Found", func(t *testing.T) {
		resultRepo := new(MockMatchResultRepository)
		matchRepo := new(MockMatchScheduleRepository)
		router := setupMatchResultRouter(resultRepo, matchRepo)

		reqBody := models.MatchResultRequest{MatchId: 99}
		jsonBody, _ := json.Marshal(reqBody)

		matchRepo.On("GetMatchScheduleByID", int64(99)).Return(nil, gorm.ErrRecordNotFound)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/match-results", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		matchRepo.AssertExpectations(t)
	})

	t.Run("Result Already Exists", func(t *testing.T) {
		resultRepo := new(MockMatchResultRepository)
		matchRepo := new(MockMatchScheduleRepository)
		router := setupMatchResultRouter(resultRepo, matchRepo)

		reqBody := models.MatchResultRequest{MatchId: 1}
		jsonBody, _ := json.Marshal(reqBody)

		matchRepo.On("GetMatchScheduleByID", int64(1)).Return(&models.MatchScheduleDetail{}, nil)
		resultRepo.On("CheckResultExists", int64(1)).Return(true, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/match-results", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		matchRepo.AssertExpectations(t)
		resultRepo.AssertExpectations(t)
	})
}

func TestGetMatchResultByMatchID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		resultRepo := new(MockMatchResultRepository)
		matchRepo := new(MockMatchScheduleRepository)
		router := setupMatchResultRouter(resultRepo, matchRepo)

		result := models.MatchResult{Id: 1, MatchId: 10}
		resultRepo.On("GetMatchResultByMatchID", int64(10)).Return(&result, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/match-results/10", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response models.MatchResult
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, int64(10), response.MatchId)
		resultRepo.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		resultRepo := new(MockMatchResultRepository)
		matchRepo := new(MockMatchScheduleRepository)
		router := setupMatchResultRouter(resultRepo, matchRepo)

		resultRepo.On("GetMatchResultByMatchID", int64(99)).Return(nil, gorm.ErrRecordNotFound)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/match-results/99", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		resultRepo.AssertExpectations(t)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		resultRepo := new(MockMatchResultRepository)
		matchRepo := new(MockMatchScheduleRepository)
		router := setupMatchResultRouter(resultRepo, matchRepo)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/match-results/abc", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
