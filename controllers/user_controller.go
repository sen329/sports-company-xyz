package controllers

import (
	"net/http"
	"os"
	"sports-backend-api/database"
	"sports-backend-api/models"
	"sports-backend-api/repositories"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type UserController struct {
	userRepo repositories.UserRepository
}

func NewUserController() *UserController {
	return &UserController{
		userRepo: repositories.NewUserRepository(database.DB),
	}
}

func (c *UserController) Login(ctx *gin.Context) {
	var req models.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
		return
	}

	user, err := c.userRepo.GetUserByEmail(req.Email)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		ctx.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Invalid credentials"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.UserId,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Failed to generate token"})
		return
	}

	ctx.JSON(http.StatusOK, models.LoginResponse{Token: tokenString})
}

func (c *UserController) Register(ctx *gin.Context) {
	var req models.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
		return
	}

	// Check if user already exists
	_, err := c.userRepo.GetUserByEmail(req.Email)
	if err == nil {
		ctx.JSON(http.StatusConflict, models.ErrorResponse{Message: "User with this email already exists"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Failed to hash password"})
		return
	}

	newUser := models.User{
		UserId:   "user_" + time.Now().Format("20060102150405"), // Simple unique ID generation
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     "user", // Default role
		Status:   "active",
	}

	if err := c.userRepo.CreateUser(&newUser); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Failed to register user"})
		return
	}

	ctx.JSON(http.StatusCreated, models.RegisterResponse{Message: "User registered successfully"})
}
