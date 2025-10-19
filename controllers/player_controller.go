package controllers

import (
	"net/http"
	"sports-backend-api/database"
	"sports-backend-api/models"
	"sports-backend-api/repositories"
	"sports-backend-api/util"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PlayerController handles the HTTP requests for Players.
type PlayerController struct {
	playerRepo repositories.PlayerRepository
}

// NewPlayerController creates a new instance of PlayerController.
func NewPlayerController() *PlayerController {
	return &PlayerController{
		playerRepo: repositories.NewPlayerRepository(database.DB),
	}
}

// CreatePlayer handles the creation of a new player.
func (c *PlayerController) CreatePlayer(ctx *gin.Context) {
	var req models.PlayerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validation: Check for valid player position and normalize it.
	canonicalPosition, ok := util.NormalizeAndValidatePlayerPosition(req.Position)
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid player position. Must be one of: Penyerang, Gelandang, Bertahan, Penjaga Gawang"})
		return
	}

	// Validation: Check if a player with the same back number exists on the team.
	_, err := c.playerRepo.GetPlayerByTeamAndBackNumber(req.TeamId, req.BackNumber)
	if err == nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": "A player with this back number already exists on this team"})
		return
	}
	if err != gorm.ErrRecordNotFound {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate player back number"})
		return
	}

	newPlayer := models.Player{
		Name:       req.Name,
		Weight:     req.Weight,
		Height:     req.Height,
		Position:   canonicalPosition, // Use the canonical version
		BackNumber: req.BackNumber,
		TeamId:     req.TeamId,
	}

	if err := c.playerRepo.CreatePlayer(&newPlayer); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create player"})
		return
	}

	ctx.JSON(http.StatusCreated, newPlayer)
}

// GetAllPlayers retrieves all players.
func (c *PlayerController) GetAllPlayers(ctx *gin.Context) {
	var req models.PlayerRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters: " + err.Error()})
		return
	}

	req.Page, req.Limit = util.SetPaginationDefaults(req.Page, req.Limit)

	players, total, err := c.playerRepo.GetPlayersByFilter(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve players"})
		return
	}

	totalPages := util.CalculateTotalPages(total, req.Limit)

	// Map PlayerDetail to PlayerResponse
	playerResponses := make([]models.PlayerResponse, 0, len(players))
	for _, p := range players {
		playerRes := models.PlayerResponse{
			Id:         p.Id,
			Name:       p.Name,
			Weight:     p.Weight,
			Height:     p.Height,
			Position:   p.Position,
			BackNumber: p.BackNumber,
			TeamName:   p.TeamName,
		}

		playerResponses = append(playerResponses, playerRes)
	}

	ctx.JSON(http.StatusOK, models.PaginatedPlayerResponse{
		Data:         playerResponses,
		TotalRecords: total,
		CurrentPage:  req.Page,
		PageSize:     req.Limit,
		TotalPages:   totalPages,
	})
}

// GetPlayerByID retrieves a single player by its ID.
func (c *PlayerController) GetPlayerByID(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid player ID"})
		return
	}

	player, err := c.playerRepo.GetPlayerByID(id)

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Player not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve player"})
		}
		return
	}

	response := models.PlayerResponse{
		Id:         player.Id,
		Name:       player.Name,
		Weight:     player.Weight,
		Height:     player.Height,
		Position:   player.Position,
		BackNumber: player.BackNumber,
		TeamName:   player.TeamName,
	}
	ctx.JSON(http.StatusOK, response)
}

// UpdatePlayer handles updating an existing player's details.
func (c *PlayerController) UpdatePlayer(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid player ID"})
		return
	}
	var req models.PlayerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	player, err := c.playerRepo.GetPlayerByID(id) // This now returns PlayerDetail
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Player not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve player"})
		}
		return
	}

	playerToUpdate := player.Player // Get the underlying Player model

	// Determine the team ID to use for validation. Use the new one if provided, otherwise the existing one.
	teamIDForValidation := playerToUpdate.TeamId
	if req.TeamId != 0 {
		teamIDForValidation = req.TeamId
	}

	// Validation: Check if another player on the same team already has the requested back number.
	if req.BackNumber != 0 && req.BackNumber != playerToUpdate.BackNumber {
		existingPlayer, err := c.playerRepo.GetPlayerByTeamAndBackNumber(teamIDForValidation, req.BackNumber)
		if err != nil && err != gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate player back number"})
			return
		}
		// If a player is found and it's not the same player we are trying to update, then it's a conflict
		if existingPlayer.Id != 0 && existingPlayer.Id != id {
			ctx.JSON(http.StatusConflict, gin.H{"error": "Another player with this back number already exists on this team"})
			return
		}
	}

	// Validation: Check for valid player position if it's being updated.
	if req.Position != "" {
		if canonicalPosition, ok := util.NormalizeAndValidatePlayerPosition(req.Position); ok {
			playerToUpdate.Position = canonicalPosition // Use the canonical version
		} else {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid player position. Must be one of: Penyerang, Gelandang, Bertahan, Penjaga Gawang"})
			return
		}
	}

	// Update fields only if they are provided in the request.
	if req.Name != "" {
		playerToUpdate.Name = req.Name
	}
	if req.Weight != 0 {
		playerToUpdate.Weight = req.Weight
	}
	if req.Height != 0 {
		playerToUpdate.Height = req.Height
	}
	if req.BackNumber != 0 {
		playerToUpdate.BackNumber = req.BackNumber
	}
	if req.TeamId != 0 {
		playerToUpdate.TeamId = req.TeamId
	}

	if err := c.playerRepo.UpdatePlayer(&playerToUpdate); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update player"})
		return
	}

	ctx.JSON(http.StatusOK, playerToUpdate)
}

// DeletePlayer handles the deletion of a player by its ID.
func (c *PlayerController) DeletePlayer(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid player ID"})
		return
	}
	if err := c.playerRepo.DeletePlayer(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete player"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Player deleted successfully"})
}
