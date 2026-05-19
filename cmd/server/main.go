package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type UserResponse struct {
	Success bool `json:"success"`
	Message string `json:"message"`
	Data    struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		Phone     string `json:"phone"`
		Gender    string `json:"gender"`
		Image     string `json:"image"`
		Role      string `json:"role"`
		Verified  bool   `json:"is_verified"`
	} `json:"data"`
}

func main() {
	// Load .env if present
	_ = godotenv.Load()

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:3001"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	// Health Check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "ok",
			"app":    "laci-v3-api",
		})
	})

	// Middleware Oauth2 SSO Verification
	SSOMiddleware := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Authorization header is required"})
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid Authorization header format"})
			}

			token := parts[1]

			// Verify token with SSO Backend
			ssoUserUrl := os.Getenv("SSO_USER_INFO_URL")
			if ssoUserUrl == "" {
				ssoUserUrl = "http://localhost:8080/v1/user/me"
			}

			req, err := http.NewRequest("GET", ssoUserUrl, nil)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create verification request"})
			}
			req.Header.Set("Authorization", "Bearer "+token)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to reach SSO server"})
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid session or token expired"})
			}

			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to read verification response"})
			}

			var userResp UserResponse
			if err := json.Unmarshal(bodyBytes, &userResp); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to parse user info"})
			}

			// Store user in context
			c.Set("user", &userResp.Data)
			return next(c)
		}
	}

	// Protected route
	e.GET("/api/data", func(c echo.Context) error {
		user := c.Get("user")
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Halo! Data ini dilindungi oleh SSO-v2",
			"user":    user,
		})
	}, SSOMiddleware)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8081"
	}
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", port)))
}
