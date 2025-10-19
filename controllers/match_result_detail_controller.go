package controllers

import (
	"net/http"
	"sports-backend-api/database"
	"sports-backend-api/repositories"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// MatchResultDetailController handles the HTTP requests for detailed Match Results.
type MatchResultDetailController struct {
	repo repositories.MatchResultDetailRepository
}

// NewMatchResultDetailController creates a new instance of MatchResultDetailController.
func NewMatchResultDetailController() *MatchResultDetailController {
	return &MatchResultDetailController{
		repo: repositories.NewMatchResultDetailRepository(database.DB),
	}
}

// GetMatchResultDetailByMatchID retrieves a detailed match result by its match ID.
func (c *MatchResultDetailController) GetMatchResultDetailByMatchID(ctx *gin.Context) {
	matchID, err := strconv.ParseInt(ctx.Param("match_id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid match ID"})
		return
	}

	result, err := c.repo.GetMatchResultDetailByMatchID(matchID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Match result not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve detailed match result"})
		return
	}
	ctx.JSON(http.StatusOK, result)
}
