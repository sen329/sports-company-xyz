package controllers

import (
	"fmt"
	"net/http"
	"sports-backend-api/database"
	"sports-backend-api/models"
	"sports-backend-api/repositories"
	"sports-backend-api/util"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// MatchScheduleController handles the HTTP requests for Match Schedules.
type MatchScheduleController struct {
	matchRepo repositories.MatchScheduleRepository
}

// NewMatchScheduleController creates a new instance of MatchScheduleController.
func NewMatchScheduleController() *MatchScheduleController {
	return &MatchScheduleController{
		matchRepo: repositories.NewMatchScheduleRepository(database.DB),
	}
}

// CreateMatchSchedule handles the creation of a new match schedule.
func (c *MatchScheduleController) CreateMatchSchedule(ctx *gin.Context) {
	var req models.MatchScheduleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validation: A team cannot play against itself.
	if req.HomeTeamId == req.AwayTeamId {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Home team and away team cannot be the same"})
		return
	}

	// Validation: Check for schedule conflicts for both teams.
	for _, teamID := range []int64{req.HomeTeamId, req.AwayTeamId} {
		conflict, err := c.matchRepo.CheckTeamScheduleConflict(teamID, req.Date, 0)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate team schedule"})
			return
		}
		if conflict {
			ctx.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("Team with ID %d already has a match scheduled on %s", teamID, req.Date)})
			return
		}
	}

	newMatch := models.MatchSchedule{
		Date:       req.Date,
		Time:       req.Time,
		HomeTeamId: req.HomeTeamId,
		AwayTeamId: req.AwayTeamId,
	}

	if err := c.matchRepo.CreateMatchSchedule(&newMatch); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create match schedule"})
		return
	}

	// Re-fetch the created match to get the team names
	createdMatch, err := c.matchRepo.GetMatchScheduleByID(newMatch.Id)
	if err != nil {
		// Log the error but return the original object as a fallback
		ctx.JSON(http.StatusCreated, newMatch)
		return
	}
	ctx.JSON(http.StatusCreated, createdMatch)
}

// GetAllMatchSchedules retrieves all match schedules, with optional filtering and pagination.
func (c *MatchScheduleController) GetAllMatchSchedules(ctx *gin.Context) {
	var req models.MatchScheduleRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters: " + err.Error()})
		return
	}

	req.Page, req.Limit = util.SetPaginationDefaults(req.Page, req.Limit)

	matches, total, err := c.matchRepo.GetMatchSchedulesByFilter(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve match schedules"})
		return
	}

	totalPages := util.CalculateTotalPages(total, req.Limit)

	// Map MatchScheduleDetail to MatchScheduleResponse
	matchResponses := make([]models.MatchScheduleResponse, len(matches))
	for i, m := range matches {
		matchResponses[i] = models.MatchScheduleResponse{
			Id:           m.Id,
			Date:         m.Date,
			Time:         m.Time,
			HomeTeamName: m.HomeTeamName,
			AwayTeamName: m.AwayTeamName,
		}
	}

	ctx.JSON(http.StatusOK, models.PaginatedMatchScheduleResponse{
		Data:         matchResponses,
		TotalRecords: total,
		CurrentPage:  req.Page,
		PageSize:     req.Limit,
		TotalPages:   totalPages,
	})
}

// GetMatchScheduleByID retrieves a single match schedule by its ID.
func (c *MatchScheduleController) GetMatchScheduleByID(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid match ID"})
		return
	}
	match, err := c.matchRepo.GetMatchScheduleByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Match schedule not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve match schedule"})
		return
	}
	response := models.MatchScheduleResponse{
		Id:           match.Id,
		Date:         match.Date,
		Time:         match.Time,
		HomeTeamName: match.HomeTeamName,
		AwayTeamName: match.AwayTeamName,
	}
	ctx.JSON(http.StatusOK, response)
}

// UpdateMatchSchedule handles updating an existing match schedule.
func (c *MatchScheduleController) UpdateMatchSchedule(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid match ID"})
		return
	}
	var req models.MatchScheduleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Fetch existing match
	match, err := c.matchRepo.GetMatchScheduleByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Match schedule not found"})
		return
	}

	matchToUpdate := &match.MatchSchedule // Get the underlying MatchSchedule model

	// Apply updates from request if fields are provided
	if req.Date != "" {
		matchToUpdate.Date = req.Date
	}
	if req.Time != "" {
		matchToUpdate.Time = req.Time
	}
	if req.HomeTeamId != 0 {
		matchToUpdate.HomeTeamId = req.HomeTeamId
	}
	if req.AwayTeamId != 0 {
		matchToUpdate.AwayTeamId = req.AwayTeamId
	}

	// Validation: A team cannot play against itself.
	if matchToUpdate.HomeTeamId == matchToUpdate.AwayTeamId {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Home team and away team cannot be the same"})
		return
	}

	// Validation: Check for schedule conflicts only if date or teams have changed.
	dateChanged := req.Date != "" && req.Date != match.Date
	teamsChanged := (req.HomeTeamId != 0 && req.HomeTeamId != match.HomeTeamId) || (req.AwayTeamId != 0 && req.AwayTeamId != match.AwayTeamId)

	if dateChanged || teamsChanged {
		for _, teamID := range []int64{matchToUpdate.HomeTeamId, matchToUpdate.AwayTeamId} {
			conflict, err := c.matchRepo.CheckTeamScheduleConflict(teamID, matchToUpdate.Date, id)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate team schedule"})
				return
			}
			if conflict {
				ctx.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("Team with ID %d already has a match scheduled on %s", teamID, matchToUpdate.Date)})
				return
			}
		}
	}

	if err := c.matchRepo.UpdateMatchSchedule(matchToUpdate); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update match schedule"})
		return
	}

	updatedMatch, err := c.matchRepo.GetMatchScheduleByID(match.Id)
	if err != nil {
		// Log the error but return the original object as a fallback
		ctx.JSON(http.StatusOK, match)
		return
	}
	ctx.JSON(http.StatusOK, updatedMatch)
}

// DeleteMatchSchedule handles deleting a match schedule.
func (c *MatchScheduleController) DeleteMatchSchedule(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid match ID"})
		return
	}
	if err := c.matchRepo.DeleteMatchSchedule(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete match schedule"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Match schedule deleted successfully"})
}
