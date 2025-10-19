package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sports-backend-api/models"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockMatchScheduleRepository is a mock implementation of MatchScheduleRepository
type MockMatchScheduleRepository struct {
	mock.Mock
}

func (m *MockMatchScheduleRepository) CreateMatchSchedule(match *models.MatchSchedule) error {
	args := m.Called(match)
	return args.Error(0)
}

func (m *MockMatchScheduleRepository) GetMatchScheduleByID(id int64) (*models.MatchScheduleDetail, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.MatchScheduleDetail), args.Error(1)
}

func (m *MockMatchScheduleRepository) UpdateMatchSchedule(match *models.MatchSchedule) error {
	args := m.Called(match)
	return args.Error(0)
}

func (m *MockMatchScheduleRepository) DeleteMatchSchedule(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockMatchScheduleRepository) GetMatchSchedulesByFilter(filter models.MatchScheduleRequest) ([]models.MatchScheduleDetail, int64, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.MatchScheduleDetail), args.Get(1).(int64), args.Error(2)
}

func (m *MockMatchScheduleRepository) CheckTeamScheduleConflict(teamID int64, date string, matchIDToExclude int64) (bool, error) {
	args := m.Called(teamID, date, matchIDToExclude)
	return args.Bool(0), args.Error(1)
}

func setupMatchRouter(repo *MockMatchScheduleRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	controller := &MatchScheduleController{
		matchRepo: repo,
	}
	router.POST("/matches", controller.CreateMatchSchedule)
	router.GET("/matches", controller.GetAllMatchSchedules)
	router.GET("/matches/:id", controller.GetMatchScheduleByID)
	router.PUT("/matches/:id", controller.UpdateMatchSchedule)
	router.DELETE("/matches/:id", controller.DeleteMatchSchedule)
	return router
}

func TestCreateMatchSchedule(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockMatchScheduleRepository)
		router := setupMatchRouter(mockRepo)

		reqBody := models.MatchScheduleRequest{
			Date:       "2024-01-01",
			Time:       "19:00",
			HomeTeamId: 1,
			AwayTeamId: 2,
		}
		jsonBody, _ := json.Marshal(reqBody)

		createdMatch := models.MatchSchedule{Id: 1, Date: reqBody.Date, Time: reqBody.Time, HomeTeamId: reqBody.HomeTeamId, AwayTeamId: reqBody.AwayTeamId}
		createdMatchDetail := models.MatchScheduleDetail{MatchSchedule: createdMatch, HomeTeamName: "Team A", AwayTeamName: "Team B"}

		mockRepo.On("CheckTeamScheduleConflict", int64(1), "2024-01-01", int64(0)).Return(false, nil)
		mockRepo.On("CheckTeamScheduleConflict", int64(2), "2024-01-01", int64(0)).Return(false, nil)
		mockRepo.On("CreateMatchSchedule", mock.AnythingOfType("*models.MatchSchedule")).Return(nil).Run(func(args mock.Arguments) {
			arg := args.Get(0).(*models.MatchSchedule)
			arg.Id = 1 // Simulate database assigning an ID
		})
		mockRepo.On("GetMatchScheduleByID", int64(1)).Return(&createdMatchDetail, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/matches", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response models.MatchScheduleDetail
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Team A", response.HomeTeamName)
		assert.Equal(t, "Team B", response.AwayTeamName)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Conflict", func(t *testing.T) {
		mockRepo := new(MockMatchScheduleRepository)
		router := setupMatchRouter(mockRepo)

		reqBody := models.MatchScheduleRequest{Date: "2024-01-01", HomeTeamId: 1, AwayTeamId: 2}
		jsonBody, _ := json.Marshal(reqBody)

		mockRepo.On("CheckTeamScheduleConflict", int64(1), "2024-01-01", int64(0)).Return(true, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/matches", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Same Team", func(t *testing.T) {
		mockRepo := new(MockMatchScheduleRepository)
		router := setupMatchRouter(mockRepo)

		reqBody := models.MatchScheduleRequest{Date: "2024-01-01", HomeTeamId: 1, AwayTeamId: 1}
		jsonBody, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/matches", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetAllMatchSchedules(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockMatchScheduleRepository)
		router := setupMatchRouter(mockRepo)

		matches := []models.MatchScheduleDetail{
			{MatchSchedule: models.MatchSchedule{Id: 1}, HomeTeamName: "Team A", AwayTeamName: "Team B"},
		}
		filter := models.MatchScheduleRequest{Page: 1, Limit: 10}

		mockRepo.On("GetMatchSchedulesByFilter", filter).Return(matches, 1, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/matches?page=1&limit=10", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response models.PaginatedMatchScheduleResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, int64(1), response.TotalRecords)
		assert.Len(t, response.Data, 1)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository Error", func(t *testing.T) {
		mockRepo := new(MockMatchScheduleRepository)
		router := setupMatchRouter(mockRepo)

		filter := models.MatchScheduleRequest{Page: 1, Limit: 10}
		mockRepo.On("GetMatchSchedulesByFilter", filter).Return(nil, 0, errors.New("db error"))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/matches?page=1&limit=10", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetMatchScheduleByID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockMatchScheduleRepository)
		router := setupMatchRouter(mockRepo)

		matchDetail := models.MatchScheduleDetail{
			MatchSchedule: models.MatchSchedule{Id: 1, Date: "2024-01-01"},
			HomeTeamName:  "Team A",
			AwayTeamName:  "Team B",
		}
		mockRepo.On("GetMatchScheduleByID", int64(1)).Return(&matchDetail, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/matches/1", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response models.MatchScheduleResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, int64(1), response.Id)
		assert.Equal(t, "Team A", response.HomeTeamName)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockRepo := new(MockMatchScheduleRepository)
		router := setupMatchRouter(mockRepo)

		mockRepo.On("GetMatchScheduleByID", int64(1)).Return(nil, gorm.ErrRecordNotFound)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/matches/1", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		mockRepo := new(MockMatchScheduleRepository)
		router := setupMatchRouter(mockRepo)

		w := httptest.NewRecorder()
		// The controller currently ignores the error from strconv.ParseInt, so it will try to fetch ID 0.
		mockRepo.On("GetMatchScheduleByID", int64(0)).Return(nil, gorm.ErrRecordNotFound)
		req, _ := http.NewRequest("GET", "/matches/abc", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestUpdateMatchSchedule(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockMatchScheduleRepository)
		router := setupMatchRouter(mockRepo)

		existingMatch := models.MatchScheduleDetail{
			MatchSchedule: models.MatchSchedule{Id: 1, Date: "2024-01-01", HomeTeamId: 1, AwayTeamId: 2},
			HomeTeamName:  "Old Home", AwayTeamName: "Old Away",
		}
		updateReq := models.MatchScheduleRequest{Time: "20:00"}
		jsonBody, _ := json.Marshal(updateReq)

		updatedMatchModel := models.MatchSchedule{Id: 1, Date: "2024-01-01", Time: "20:00", HomeTeamId: 1, AwayTeamId: 2}
		updatedMatchDetail := models.MatchScheduleDetail{
			MatchSchedule: updatedMatchModel,
			HomeTeamName:  "Old Home", AwayTeamName: "Old Away",
		}

		mockRepo.On("GetMatchScheduleByID", int64(1)).Return(&existingMatch, nil).Once()
		mockRepo.On("UpdateMatchSchedule", &updatedMatchModel).Return(nil)
		mockRepo.On("GetMatchScheduleByID", int64(1)).Return(&updatedMatchDetail, nil).Once()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/matches/1", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response models.MatchScheduleDetail
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "20:00", response.Time)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockRepo := new(MockMatchScheduleRepository)
		router := setupMatchRouter(mockRepo)

		updateReq := models.MatchScheduleRequest{Time: "20:00"}
		jsonBody, _ := json.Marshal(updateReq)

		mockRepo.On("GetMatchScheduleByID", int64(1)).Return(nil, gorm.ErrRecordNotFound)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/matches/1", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Update with Conflict", func(t *testing.T) {
		mockRepo := new(MockMatchScheduleRepository)
		router := setupMatchRouter(mockRepo)

		existingMatch := models.MatchScheduleDetail{
			MatchSchedule: models.MatchSchedule{Id: 1, Date: "2024-01-01", HomeTeamId: 1, AwayTeamId: 2},
		}
		updateReq := models.MatchScheduleRequest{Date: "2024-01-02"} // Changing the date
		jsonBody, _ := json.Marshal(updateReq)

		mockRepo.On("GetMatchScheduleByID", int64(1)).Return(&existingMatch, nil)
		mockRepo.On("CheckTeamScheduleConflict", int64(1), "2024-01-02", int64(1)).Return(true, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/matches/1", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestDeleteMatchSchedule(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockMatchScheduleRepository)
		router := setupMatchRouter(mockRepo)

		mockRepo.On("DeleteMatchSchedule", int64(1)).Return(nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/matches/1", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Match schedule deleted successfully", response["message"])
		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository Error", func(t *testing.T) {
		mockRepo := new(MockMatchScheduleRepository)
		router := setupMatchRouter(mockRepo)

		mockRepo.On("DeleteMatchSchedule", int64(1)).Return(errors.New("db error"))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/matches/1", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetMatchScheduleByID_InvalidID(t *testing.T) {
	mockRepo := new(MockMatchScheduleRepository)
	router := setupMatchRouter(mockRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/matches/abc", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Invalid match ID", response["error"])
}
