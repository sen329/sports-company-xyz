package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sports-backend-api/models"
	"sports-backend-api/util"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockPlayerRepository is a mock implementation of PlayerRepository
type MockPlayerRepository struct {
	mock.Mock
}

func (m *MockPlayerRepository) CreatePlayer(player *models.Player) error {
	args := m.Called(player)
	return args.Error(0)
}

func (m *MockPlayerRepository) GetPlayerByID(id int64) (*models.PlayerDetail, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PlayerDetail), args.Error(1)
}

func (m *MockPlayerRepository) UpdatePlayer(player *models.Player) error {
	args := m.Called(player)
	return args.Error(0)
}

func (m *MockPlayerRepository) DeletePlayer(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockPlayerRepository) GetPlayerByTeamAndBackNumber(teamID int64, backNumber int) (*models.Player, error) {
	args := m.Called(teamID, backNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Player), args.Error(1)
}

func (m *MockPlayerRepository) GetPlayersByFilter(filter models.PlayerRequest) ([]models.PlayerDetail, int64, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.PlayerDetail), args.Get(1).(int64), args.Error(2)
}

func setupPlayerRouter(repo *MockPlayerRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	// We need a real validator for position testing
	util.NormalizeAndValidatePlayerPosition("Penyerang")

	controller := &PlayerController{
		playerRepo: repo,
	}
	router.POST("/players", controller.CreatePlayer)
	router.GET("/players", controller.GetAllPlayers)
	router.GET("/players/:id", controller.GetPlayerByID)
	router.PUT("/players/:id", controller.UpdatePlayer)
	router.DELETE("/players/:id", controller.DeletePlayer)
	return router
}

func TestCreatePlayer(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockPlayerRepository)
		router := setupPlayerRouter(mockRepo)

		reqBody := models.PlayerRequest{
			Name:       "Test Player",
			Position:   "Penyerang",
			BackNumber: 10,
			TeamId:     1,
		}
		jsonBody, _ := json.Marshal(reqBody)

		mockRepo.On("GetPlayerByTeamAndBackNumber", reqBody.TeamId, reqBody.BackNumber).Return(nil, gorm.ErrRecordNotFound)
		mockRepo.On("CreatePlayer", mock.AnythingOfType("*models.Player")).Return(nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/players", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response models.Player
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Test Player", response.Name)
		assert.Equal(t, "Penyerang", response.Position)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Invalid Position", func(t *testing.T) {
		mockRepo := new(MockPlayerRepository)
		router := setupPlayerRouter(mockRepo)

		reqBody := models.PlayerRequest{Position: "InvalidPos"}
		jsonBody, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/players", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Back Number Conflict", func(t *testing.T) {
		mockRepo := new(MockPlayerRepository)
		router := setupPlayerRouter(mockRepo)

		reqBody := models.PlayerRequest{Position: "Penyerang", BackNumber: 10, TeamId: 1}
		jsonBody, _ := json.Marshal(reqBody)

		// Simulate finding an existing player
		mockRepo.On("GetPlayerByTeamAndBackNumber", reqBody.TeamId, reqBody.BackNumber).Return(&models.Player{Id: 99}, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/players", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Create Fails", func(t *testing.T) {
		mockRepo := new(MockPlayerRepository)
		router := setupPlayerRouter(mockRepo)

		reqBody := models.PlayerRequest{Position: "Penyerang", BackNumber: 10, TeamId: 1}
		jsonBody, _ := json.Marshal(reqBody)

		mockRepo.On("GetPlayerByTeamAndBackNumber", reqBody.TeamId, reqBody.BackNumber).Return(nil, gorm.ErrRecordNotFound)
		mockRepo.On("CreatePlayer", mock.AnythingOfType("*models.Player")).Return(errors.New("db error"))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/players", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetAllPlayers(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockPlayerRepository)
		router := setupPlayerRouter(mockRepo)

		players := []models.PlayerDetail{
			{Player: models.Player{Id: 1, Name: "Player 1"}, TeamName: "Team A"},
		}
		filter := models.PlayerRequest{Page: 1, Limit: 10}

		mockRepo.On("GetPlayersByFilter", filter).Return(players, int64(1), nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/players?page=1&limit=10", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response models.PaginatedPlayerResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, int64(1), response.TotalRecords)
		assert.Len(t, response.Data, 1)
		assert.Equal(t, "Player 1", response.Data[0].Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository Error", func(t *testing.T) {
		mockRepo := new(MockPlayerRepository)
		router := setupPlayerRouter(mockRepo)

		filter := models.PlayerRequest{Page: 1, Limit: 10}
		mockRepo.On("GetPlayersByFilter", filter).Return(nil, int64(0), errors.New("db error"))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/players?page=1&limit=10", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetPlayerByID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockPlayerRepository)
		router := setupPlayerRouter(mockRepo)

		playerDetail := models.PlayerDetail{
			Player:   models.Player{Id: 1, Name: "Test Player"},
			TeamName: "Team A",
		}
		mockRepo.On("GetPlayerByID", int64(1)).Return(&playerDetail, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/players/1", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response models.PlayerResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, int64(1), response.Id)
		assert.Equal(t, "Test Player", response.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockRepo := new(MockPlayerRepository)
		router := setupPlayerRouter(mockRepo)

		mockRepo.On("GetPlayerByID", int64(1)).Return(nil, gorm.ErrRecordNotFound)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/players/1", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		mockRepo := new(MockPlayerRepository)
		router := setupPlayerRouter(mockRepo)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/players/abc", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUpdatePlayer(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockPlayerRepository)
		router := setupPlayerRouter(mockRepo)

		existingPlayer := models.PlayerDetail{
			Player: models.Player{Id: 1, Name: "Old Name", Position: "Gelandang"},
		}
		updateReq := models.PlayerRequest{Name: "New Name"}
		jsonBody, _ := json.Marshal(updateReq)

		updatedPlayerModel := models.Player{Id: 1, Name: "New Name", Position: "Gelandang"}

		mockRepo.On("GetPlayerByID", int64(1)).Return(&existingPlayer, nil)
		mockRepo.On("UpdatePlayer", &updatedPlayerModel).Return(nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/players/1", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response models.Player
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "New Name", response.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockRepo := new(MockPlayerRepository)
		router := setupPlayerRouter(mockRepo)

		updateReq := models.PlayerRequest{Name: "New Name"}
		jsonBody, _ := json.Marshal(updateReq)

		mockRepo.On("GetPlayerByID", int64(1)).Return(nil, gorm.ErrRecordNotFound)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/players/1", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Update with Back Number Conflict", func(t *testing.T) {
		mockRepo := new(MockPlayerRepository)
		router := setupPlayerRouter(mockRepo)

		existingPlayer := models.PlayerDetail{
			Player: models.Player{Id: 1, BackNumber: 10, TeamId: 1},
		}
		updateReq := models.PlayerRequest{BackNumber: 7} // Changing the back number
		jsonBody, _ := json.Marshal(updateReq)

		conflictingPlayer := models.Player{Id: 2, BackNumber: 7, TeamId: 1}

		mockRepo.On("GetPlayerByID", int64(1)).Return(&existingPlayer, nil)
		mockRepo.On("GetPlayerByTeamAndBackNumber", int64(1), 7).Return(&conflictingPlayer, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/players/1", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestDeletePlayer(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockPlayerRepository)
		router := setupPlayerRouter(mockRepo)

		mockRepo.On("DeletePlayer", int64(1)).Return(nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/players/1", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Player deleted successfully", response["message"])
		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository Error", func(t *testing.T) {
		mockRepo := new(MockPlayerRepository)
		router := setupPlayerRouter(mockRepo)

		mockRepo.On("DeletePlayer", int64(1)).Return(errors.New("db error"))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/players/1", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockRepo.AssertExpectations(t)
	})
}
