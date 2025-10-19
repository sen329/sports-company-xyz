package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sports-backend-api/models"
	"sports-backend-api/repositories"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockMatchResultDetailRepository is a mock for MatchResultDetailRepository
type MockMatchResultDetailRepository struct {
	mock.Mock
}

func (m *MockMatchResultDetailRepository) GetMatchResultDetailByMatchID(matchID int64) (*models.MatchResultDetail, error) {
	args := m.Called(matchID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.MatchResultDetail), args.Error(1)
}

func setupMatchResultDetailRouter(repo repositories.MatchResultDetailRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	controller := &MatchResultDetailController{
		repo: repo,
	}
	router.GET("/match-results-detail/:match_id", controller.GetMatchResultDetailByMatchID)
	return router
}

func TestGetMatchResultDetailByMatchID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockMatchResultDetailRepository)
		router := setupMatchResultDetailRouter(mockRepo)

		detail := models.MatchResultDetail{
			MatchResult:  models.MatchResult{MatchId: 1},
			HomeTeamName: "Team A",
			AwayTeamName: "Team B",
		}
		mockRepo.On("GetMatchResultDetailByMatchID", int64(1)).Return(&detail, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/match-results-detail/1", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response models.MatchResultDetail
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, int64(1), response.MatchId)
		assert.Equal(t, "Team A", response.HomeTeamName)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockRepo := new(MockMatchResultDetailRepository)
		router := setupMatchResultDetailRouter(mockRepo)

		mockRepo.On("GetMatchResultDetailByMatchID", int64(99)).Return(nil, gorm.ErrRecordNotFound)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/match-results-detail/99", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository Error", func(t *testing.T) {
		mockRepo := new(MockMatchResultDetailRepository)
		router := setupMatchResultDetailRouter(mockRepo)

		mockRepo.On("GetMatchResultDetailByMatchID", int64(1)).Return(nil, errors.New("db error"))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/match-results-detail/1", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockRepo.AssertExpectations(t)
	})
}
