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

// MockTeamHQRepository is a mock implementation of TeamHQRepository
type MockTeamHQRepository struct {
	mock.Mock
}

func (m *MockTeamHQRepository) CreateTeamHQ(team *models.TeamHQ) error {
	args := m.Called(team)
	return args.Error(0)
}

func (m *MockTeamHQRepository) GetTeamHQByID(id int64) (*models.TeamHQ, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TeamHQ), args.Error(1)
}

func (m *MockTeamHQRepository) UpdateTeamHQ(team *models.TeamHQ) error {
	args := m.Called(team)
	return args.Error(0)
}

func (m *MockTeamHQRepository) DeleteTeamHQ(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTeamHQRepository) GetTeamHQsByFilter(filter models.TeamHQRequest) ([]models.TeamHQ, int64, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.TeamHQ), args.Get(1).(int64), args.Error(2)
}

func setupTeamHQRouter(repo *MockTeamHQRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	controller := &TeamHQController{
		teamHQRepo: repo,
	}
	router.POST("/teamhqs", controller.CreateTeamHQ)
	router.GET("/teamhqs", controller.GetAllTeamHQs)
	router.GET("/teamhqs/:id", controller.GetTeamHQByID)
	router.PUT("/teamhqs/:id", controller.UpdateTeamHQ)
	router.DELETE("/teamhqs/:id", controller.DeleteTeamHQ)
	return router
}

func TestCreateTeamHQ(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockTeamHQRepository)
		router := setupTeamHQRouter(mockRepo)

		reqBody := models.TeamHQRequest{Name: "Test Team", City: "Test City"}
		jsonBody, _ := json.Marshal(reqBody)

		mockRepo.On("CreateTeamHQ", mock.AnythingOfType("*models.TeamHQ")).Return(nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/teamhqs", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response models.TeamHQ
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Test Team", response.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository Error", func(t *testing.T) {
		mockRepo := new(MockTeamHQRepository)
		router := setupTeamHQRouter(mockRepo)

		reqBody := models.TeamHQRequest{Name: "Test Team"}
		jsonBody, _ := json.Marshal(reqBody)

		mockRepo.On("CreateTeamHQ", mock.AnythingOfType("*models.TeamHQ")).Return(errors.New("db error"))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/teamhqs", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetAllTeamHQs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockTeamHQRepository)
		router := setupTeamHQRouter(mockRepo)

		teams := []models.TeamHQ{{Id: 1, Name: "Team A"}}
		filter := models.TeamHQRequest{Page: 1, Limit: 10}

		mockRepo.On("GetTeamHQsByFilter", filter).Return(teams, int64(1), nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/teamhqs?page=1&limit=10", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response models.PaginatedTeamHQResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, int64(1), response.TotalRecords)
		assert.Len(t, response.Data, 1)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetTeamHQByID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockTeamHQRepository)
		router := setupTeamHQRouter(mockRepo)

		team := models.TeamHQ{Id: 1, Name: "Test Team"}
		mockRepo.On("GetTeamHQByID", int64(1)).Return(&team, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/teamhqs/1", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response models.TeamHQ
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Test Team", response.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockRepo := new(MockTeamHQRepository)
		router := setupTeamHQRouter(mockRepo)

		mockRepo.On("GetTeamHQByID", int64(1)).Return(nil, gorm.ErrRecordNotFound)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/teamhqs/1", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		mockRepo := new(MockTeamHQRepository)
		router := setupTeamHQRouter(mockRepo)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/teamhqs/abc", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUpdateTeamHQ(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockTeamHQRepository)
		router := setupTeamHQRouter(mockRepo)

		existingTeam := models.TeamHQ{Id: 1, Name: "Old Name"}
		updateReq := models.TeamHQRequest{Name: "New Name"}
		jsonBody, _ := json.Marshal(updateReq)

		updatedTeam := models.TeamHQ{Id: 1, Name: "New Name"}

		mockRepo.On("GetTeamHQByID", int64(1)).Return(&existingTeam, nil)
		mockRepo.On("UpdateTeamHQ", &updatedTeam).Return(nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/teamhqs/1", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response models.TeamHQ
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "New Name", response.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockRepo := new(MockTeamHQRepository)
		router := setupTeamHQRouter(mockRepo)

		updateReq := models.TeamHQRequest{Name: "New Name"}
		jsonBody, _ := json.Marshal(updateReq)

		mockRepo.On("GetTeamHQByID", int64(1)).Return(nil, gorm.ErrRecordNotFound)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/teamhqs/1", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Update Fails", func(t *testing.T) {
		mockRepo := new(MockTeamHQRepository)
		router := setupTeamHQRouter(mockRepo)

		existingTeam := models.TeamHQ{Id: 1, Name: "Old Name"}
		updateReq := models.TeamHQRequest{Name: "New Name"}
		jsonBody, _ := json.Marshal(updateReq)

		updatedTeam := models.TeamHQ{Id: 1, Name: "New Name"}

		mockRepo.On("GetTeamHQByID", int64(1)).Return(&existingTeam, nil)
		mockRepo.On("UpdateTeamHQ", &updatedTeam).Return(errors.New("db error"))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/teamhqs/1", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestDeleteTeamHQ(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockTeamHQRepository)
		router := setupTeamHQRouter(mockRepo)

		mockRepo.On("DeleteTeamHQ", int64(1)).Return(nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/teamhqs/1", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Team HQ deleted successfully", response["message"])
		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository Error", func(t *testing.T) {
		mockRepo := new(MockTeamHQRepository)
		router := setupTeamHQRouter(mockRepo)

		mockRepo.On("DeleteTeamHQ", int64(1)).Return(errors.New("db error"))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/teamhqs/1", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockRepo.AssertExpectations(t)
	})
}
