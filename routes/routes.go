package routes

import (
	"os"
	"sports-backend-api/controllers"

	"sports-backend-api/routes/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes() *gin.Engine {
	router := gin.Default()
	router.Use(gin.Logger())
	// Define your routes here
	v1 := router.Group("/api/v1")
	// Create an instance of the user controller
	userController := controllers.NewUserController()

	// User routes
	userRoutes := v1.Group("/users")
	userRoutes.POST("/login", userController.Login)
	userRoutes.POST("/register", userController.Register)
	// Add more routes as needed
	// ...

	teamHQController := controllers.NewTeamHQController()
	teamHQRoutes := v1.Group("/teamhqs")
	teamHQRoutes.Use(middleware.AuthMiddleware())
	{
		teamHQRoutes.GET("/", teamHQController.GetAllTeamHQs)
		teamHQRoutes.GET("/:id", teamHQController.GetTeamHQByID)
	}

	teamHQRoutesAdmin := v1.Group("/teamhqs/admin")
	teamHQRoutesAdmin.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("admin", "superadmin"))
	{
		teamHQRoutesAdmin.POST("/", teamHQController.CreateTeamHQ)
		teamHQRoutesAdmin.PUT("/:id", teamHQController.UpdateTeamHQ)
		teamHQRoutesAdmin.DELETE("/:id", teamHQController.DeleteTeamHQ)
	}

	playerController := controllers.NewPlayerController()
	playerRoutes := v1.Group("/players")
	playerRoutes.Use(middleware.AuthMiddleware())
	{
		playerRoutes.GET("/", playerController.GetAllPlayers)
		playerRoutes.GET("/:id", playerController.GetPlayerByID)
	}
	playerRoutesAdmin := v1.Group("/players/admin")
	playerRoutesAdmin.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("admin", "superadmin"))
	{
		playerRoutesAdmin.POST("/", playerController.CreatePlayer)
		playerRoutesAdmin.PUT("/:id", playerController.UpdatePlayer)
		playerRoutesAdmin.DELETE("/:id", playerController.DeletePlayer)
	}

	matchController := controllers.NewMatchScheduleController()
	matchRoutes := v1.Group("/matches")
	matchRoutes.Use(middleware.AuthMiddleware())
	{
		matchRoutes.GET("/", matchController.GetAllMatchSchedules)
		matchRoutes.GET("/:id", matchController.GetMatchScheduleByID)
	}
	matchRoutesAdmin := v1.Group("/matches/admin")
	matchRoutesAdmin.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("admin", "superadmin"))
	{
		matchRoutesAdmin.POST("/", matchController.CreateMatchSchedule)
		matchRoutesAdmin.PUT("/:id", matchController.UpdateMatchSchedule)
		matchRoutesAdmin.DELETE("/:id", matchController.DeleteMatchSchedule)
	}

	matchResultController := controllers.NewMatchResultController()
	matchResultRoutes := v1.Group("/match-results")
	matchResultRoutes.Use(middleware.AuthMiddleware())
	{
		matchResultRoutes.GET("/:match_id", matchResultController.GetMatchResultByMatchID)
	}
	matchResultRoutesAdmin := v1.Group("/match-results/admin")
	matchResultRoutesAdmin.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("admin", "superadmin"))
	{
		matchResultRoutesAdmin.POST("/", matchResultController.CreateMatchResult)
	}

	matchResultDetailController := controllers.NewMatchResultDetailController()
	matchResultDetailRoutes := v1.Group("/match-results-detail")
	matchResultDetailRoutes.Use(middleware.AuthMiddleware())
	{
		matchResultDetailRoutes.GET("/:match_id", matchResultDetailController.GetMatchResultDetailByMatchID)
	}

	// Run the server
	router.Run(":" + os.Getenv("PORT"))
	return router
}
