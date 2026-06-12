package handler

import (
	"errors"
	"net/http"

	"laci-v3/be/internal/domain"
	"laci-v3/be/internal/service"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type ActivityHandler struct {
	svc service.ActivityService
}

func NewActivityHandler(svc service.ActivityService) *ActivityHandler {
	return &ActivityHandler{svc: svc}
}

func (h *ActivityHandler) GetActivities(c echo.Context) error {
	activities, err := h.svc.GetMergedActivities()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch activities"})
	}
	return c.JSON(http.StatusOK, activities)
}

func (h *ActivityHandler) CreateActivity(c echo.Context) error {
	var req domain.CreateActivityRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	activity, err := h.svc.CreateActivity(req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, activity)
}

func (h *ActivityHandler) UpdateActivity(c echo.Context) error {
	id := c.Param("id")
	var req domain.CreateActivityRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	activity, err := h.svc.UpdateActivity(id, req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Activity not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, activity)
}

func (h *ActivityHandler) DeleteActivity(c echo.Context) error {
	id := c.Param("id")
	err := h.svc.DeleteActivity(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Activity not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete activity"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Activity deleted successfully"})
}
