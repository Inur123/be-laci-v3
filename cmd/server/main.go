package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"

	"laci-v3/be/internal/config"
	"laci-v3/be/internal/database"
	"laci-v3/be/internal/domain"
	"laci-v3/be/internal/handler"
	"laci-v3/be/internal/middleware"
	"laci-v3/be/internal/repository"
	"laci-v3/be/internal/service"
)

func main() {
	// 1. Load config
	cfg := config.Load()

	// 2. Connect database (PostgreSQL)
	db := database.ConnectPostgres()

	// 3. Auto-migrate schema
	if err := db.AutoMigrate(&domain.Activity{}); err != nil {
		log.Fatalf("Failed to auto-migrate database: %v", err)
	}

	// 4. Dependency Injection
	activityRepo := repository.NewActivityRepository(db)
	activitySvc := service.NewActivityService(activityRepo)
	activityHandler := handler.NewActivityHandler(activitySvc)

	// 5. Echo Setup
	e := echo.New()

	e.Use(echomiddleware.LoggerWithConfig(echomiddleware.LoggerConfig{
		Format: "⇨ [LACI-V3] ${time_custom} | ${status} | ${latency_human} | ${remote_ip} | ${method} ${uri}\n",
		CustomTimeFormat: "2006-01-02 15:04:05",
	}))
	e.Use(echomiddleware.Recover())

	clientURL := cfg.ClientURL
	if clientURL == "" {
		clientURL = "*"
	}
	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins: []string{clientURL},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	// Health Check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "ok",
			"app":    "laci-v3-api",
		})
	})

	// Protected route
	e.GET("/api/data", func(c echo.Context) error {
		user := c.Get("user")
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Halo! Data ini dilindungi oleh SSO-v2",
			"user":    user,
		})
	}, middleware.SSOMiddleware)

	// Activities endpoints
	e.GET("/api/activities", activityHandler.GetActivities)

	apiGroup := e.Group("/api")
	apiGroup.POST("/activities", activityHandler.CreateActivity, middleware.SSOMiddleware)
	apiGroup.PUT("/activities/:id", activityHandler.UpdateActivity, middleware.SSOMiddleware)
	apiGroup.DELETE("/activities/:id", activityHandler.DeleteActivity, middleware.SSOMiddleware)

	port := cfg.AppPort
	if port == "" {
		port = "8081"
	}
	fmt.Println("\n==========================================")
	fmt.Println("         RUNNING: LACI-V3 BACKEND         ")
	fmt.Println("==========================================")
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", port)))
}
