package controllers

import (
	"net/http"
	"sports-backend-api/database"
	"sports-backend-api/models"
	"sports-backend-api/repositories"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// MatchResultController handles the HTTP requests for Match Results.
type MatchResultController struct {
	resultRepo repositories.MatchResultRepository
	matchRepo  repositories.MatchScheduleRepository
}

// NewMatchResultController creates a new instance of MatchResultController.
func NewMatchResultController() *MatchResultController {
	return &MatchResultController{
		resultRepo: repositories.NewMatchResultRepository(database.DB),
		matchRepo:  repositories.NewMatchScheduleRepository(database.DB),
	}
}

// CreateMatchResult handles the creation of a new match result.
func (c *MatchResultController) CreateMatchResult(ctx *gin.Context) {
	var req models.MatchResultRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validation: Check if the match exists
	match, err := c.matchRepo.GetMatchScheduleByID(req.MatchId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Match schedule not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate match schedule"})
		return
	}

	// Validation: Check if a result for this match already exists.
	exists, err := c.resultRepo.CheckResultExists(req.MatchId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check for existing match result"})
		return
	}
	if exists {
		ctx.JSON(http.StatusConflict, gin.H{"error": "A result for this match already exists"})
		return
	}

	// Determine the winner
	var winnerTeamID int64
	if req.HomeScore > req.AwayScore {
		winnerTeamID = match.HomeTeamId
	} else if req.AwayScore > req.HomeScore {
		winnerTeamID = match.AwayTeamId
	} // If scores are equal, winnerTeamID remains 0, indicating a draw.

	newResult := models.MatchResult{
		MatchId:      req.MatchId,
		HomeScore:    req.HomeScore,
		AwayScore:    req.AwayScore,
		WinnerTeamId: winnerTeamID,
		PlayerScored: req.PlayerScored,
	}

	if err := c.resultRepo.CreateMatchResult(&newResult); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create match result"})
		return
	}

	ctx.JSON(http.StatusCreated, newResult)
}

// GetMatchResultByMatchID retrieves a match result by its match ID.
func (c *MatchResultController) GetMatchResultByMatchID(ctx *gin.Context) {
	matchID, err := strconv.ParseInt(ctx.Param("match_id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid match ID"})
		return
	}
	result, err := c.resultRepo.GetMatchResultByMatchID(matchID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Match result not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve match result"})
		return
	}
	ctx.JSON(http.StatusOK, result)
}
