package middleware

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"laci-v3/be/internal/config"

	"github.com/labstack/echo/v4"
)

type UserResponse struct {
	Success bool `json:"success"`
	Message string `json:"message"`
	Data    UserData `json:"data"`
}

type UserData struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Gender   string `json:"gender"`
	Image    string `json:"image"`
	Role     string `json:"role"`
	Verified bool   `json:"is_verified"`
}

func SSOMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
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

		cfg := config.Get()
		ssoApiUrl := cfg.SSOAPIURL
		if ssoApiUrl == "" {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "SSO_API_URL env is not configured"})
		}
		ssoUserUrl := fmt.Sprintf("%s/v1/user/me", strings.TrimSuffix(ssoApiUrl, "/"))

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

		c.Set("user", &userResp.Data)
		return next(c)
	}
}
