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

// TeamHQController handles the HTTP requests for Team HQs.
type TeamHQController struct {
	teamHQRepo repositories.TeamHQRepository
}

// NewTeamHQController creates a new instance of TeamHQController.
func NewTeamHQController() *TeamHQController {
	return &TeamHQController{
		teamHQRepo: repositories.NewTeamHQRepository(database.DB),
	}
}

// CreateTeamHQ handles the creation of a new team HQ.
func (c *TeamHQController) CreateTeamHQ(ctx *gin.Context) {
	var req models.TeamHQRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newTeamHQ := models.TeamHQ{
		Name:     req.Name,
		Logo:     req.Logo,
		Location: req.Location,
		City:     req.City,
	}

	if err := c.teamHQRepo.CreateTeamHQ(&newTeamHQ); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create team HQ"})
		return
	}

	ctx.JSON(http.StatusCreated, newTeamHQ)
}

// GetAllTeamHQs retrieves all team HQs, with optional filtering via query parameters.
func (c *TeamHQController) GetAllTeamHQs(ctx *gin.Context) {
	var req models.TeamHQRequest
	// Bind query parameters (e.g., /teamhqs?city=NewYork) to the request struct.
	// Using ShouldBindQuery is safe as it doesn't abort with an error.
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters: " + err.Error()})
		return
	}

	req.Page, req.Limit = util.SetPaginationDefaults(req.Page, req.Limit)

	teams, total, err := c.teamHQRepo.GetTeamHQsByFilter(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve team HQs"})
		return
	}

	totalPages := util.CalculateTotalPages(total, req.Limit)

	ctx.JSON(http.StatusOK, models.PaginatedTeamHQResponse{
		Data:         teams,
		TotalRecords: total,
		CurrentPage:  req.Page,
		PageSize:     req.Limit,
		TotalPages:   totalPages,
	})
}

// GetTeamHQByID retrieves a single team HQ by its ID.
func (c *TeamHQController) GetTeamHQByID(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team HQ ID"})
		return
	}

	team, err := c.teamHQRepo.GetTeamHQByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Team HQ not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve team HQ"})
		return
	}

	ctx.JSON(http.StatusOK, team)
}

// UpdateTeamHQ handles updating an existing team HQ.
func (c *TeamHQController) UpdateTeamHQ(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team HQ ID"})
		return
	}

	var req models.TeamHQRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	team, err := c.teamHQRepo.GetTeamHQByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Team HQ not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve team HQ for update"})
		return
	}

	// Update fields only if they are provided in the request.
	if req.Name != "" {
		team.Name = req.Name
	}
	if req.Logo != "" {
		team.Logo = req.Logo
	}
	if req.Location != "" {
		team.Location = req.Location
	}
	if req.City != "" {
		team.City = req.City
	}

	if err := c.teamHQRepo.UpdateTeamHQ(team); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update team HQ"})
		return
	}

	ctx.JSON(http.StatusOK, team)
}

// DeleteTeamHQ handles deleting a team HQ.
func (c *TeamHQController) DeleteTeamHQ(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team HQ ID"})
		return
	}

	if err := c.teamHQRepo.DeleteTeamHQ(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete team HQ"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Team HQ deleted successfully"})
}
