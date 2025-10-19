package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"sports-backend-api/models"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// MockUserRepository is a mock for UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserByID(id uint) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) UpdateUser(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteUser(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func setupUserRouter(repo *MockUserRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	controller := &UserController{
		userRepo: repo,
	}
	router.POST("/login", controller.Login)
	router.POST("/register", controller.Register)
	return router
}

func TestLogin(t *testing.T) {
	// Set a dummy JWT secret for testing
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := &models.User{
		UserId:   "user_123",
		Email:    "test@example.com",
		Password: string(hashedPassword),
		Role:     "user",
	}

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		router := setupUserRouter(mockRepo)

		mockRepo.On("GetUserByEmail", "test@example.com").Return(user, nil)

		reqBody := models.LoginRequest{Email: "test@example.com", Password: "password123"}
		jsonBody, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response models.LoginResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotEmpty(t, response.Token)
		mockRepo.AssertExpectations(t)
	})

	t.Run("User Not Found", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		router := setupUserRouter(mockRepo)

		mockRepo.On("GetUserByEmail", "notfound@example.com").Return(nil, gorm.ErrRecordNotFound)

		reqBody := models.LoginRequest{Email: "notfound@example.com", Password: "password123"}
		jsonBody, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Incorrect Password", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		router := setupUserRouter(mockRepo)

		mockRepo.On("GetUserByEmail", "test@example.com").Return(user, nil)

		reqBody := models.LoginRequest{Email: "test@example.com", Password: "wrongpassword"}
		jsonBody, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestRegister(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		router := setupUserRouter(mockRepo)

		mockRepo.On("GetUserByEmail", "newuser@example.com").Return(nil, gorm.ErrRecordNotFound)
		mockRepo.On("CreateUser", mock.AnythingOfType("*models.User")).Return(nil)

		reqBody := models.RegisterRequest{Name: "New User", Email: "newuser@example.com", Password: "password123"}
		jsonBody, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response models.RegisterResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "User registered successfully", response.Message)
		mockRepo.AssertExpectations(t)
	})

	t.Run("User Already Exists", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		router := setupUserRouter(mockRepo)

		existingUser := &models.User{Email: "existing@example.com"}
		mockRepo.On("GetUserByEmail", "existing@example.com").Return(existingUser, nil)

		reqBody := models.RegisterRequest{Name: "Existing User", Email: "existing@example.com", Password: "password123"}
		jsonBody, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Create User Fails", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		router := setupUserRouter(mockRepo)

		mockRepo.On("GetUserByEmail", "newuser@example.com").Return(nil, gorm.ErrRecordNotFound)
		mockRepo.On("CreateUser", mock.AnythingOfType("*models.User")).Return(errors.New("db error"))

		reqBody := models.RegisterRequest{Name: "New User", Email: "newuser@example.com", Password: "password123"}
		jsonBody, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockRepo.AssertExpectations(t)
	})
}
