package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"laci-v3/be/internal/config"
	"laci-v3/be/internal/domain"
	"laci-v3/be/internal/repository"

	"github.com/google/uuid"
)

type ActivityService interface {
	GetMergedActivities() ([]domain.Activity, error)
	CreateActivity(req domain.CreateActivityRequest) (*domain.Activity, error)
	UpdateActivity(id string, req domain.CreateActivityRequest) (*domain.Activity, error)
	DeleteActivity(id string) error
}

type activityService struct {
	repo repository.ActivityRepository
}

func NewActivityService(repo repository.ActivityRepository) ActivityService {
	return &activityService{repo: repo}
}

// GoogleCalendarResponse maps the Google Calendar API events list response
type GoogleCalendarResponse struct {
	Items []GoogleCalendarEvent `json:"items"`
}

type GoogleCalendarEvent struct {
	ID          string `json:"id"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
	Location    string `json:"location"`
	Start       struct {
		Date     string `json:"date"`
		DateTime string `json:"dateTime"`
	} `json:"start"`
	End struct {
		Date     string `json:"date"`
		DateTime string `json:"dateTime"`
	} `json:"end"`
}

func (s *activityService) FetchGoogleCalendarEvents() ([]domain.Activity, error) {
	cfg := config.Get()
	apiKey := cfg.GoogleAPIKey
	calendarID := cfg.GoogleCalendarID

	if apiKey == "" || calendarID == "" {
		log.Println("Google Calendar API Key or Calendar ID is not configured. Skipping Google Calendar sync.")
		return []domain.Activity{}, nil
	}

	timeMin := time.Now().AddDate(-1, 0, 0).Format(time.RFC3339)
	apiURL := fmt.Sprintf(
		"https://www.googleapis.com/calendar/v3/calendars/%s/events?key=%s&singleEvents=true&orderBy=startTime&timeMin=%s",
		url.PathEscape(calendarID), url.QueryEscape(apiKey), url.QueryEscape(timeMin),
	)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Google API returned status %d: %s", resp.StatusCode, string(body))
	}

	var googleResp GoogleCalendarResponse
	if err := json.NewDecoder(resp.Body).Decode(&googleResp); err != nil {
		return nil, err
	}

	activities := make([]domain.Activity, 0, len(googleResp.Items))
	for _, item := range googleResp.Items {
		var startDate, endDate, startTime, endTime string
		var description *string

		if item.Description != "" {
			desc := item.Description
			description = &desc
		}

		if item.Start.DateTime != "" {
			parsedTime, err := time.Parse(time.RFC3339, item.Start.DateTime)
			if err == nil {
				startDate = parsedTime.Format("2006-01-02")
				startTime = parsedTime.Format("15:04")
			}
		} else if item.Start.Date != "" {
			startDate = item.Start.Date
		}

		if item.End.DateTime != "" {
			parsedTime, err := time.Parse(time.RFC3339, item.End.DateTime)
			if err == nil {
				endDate = parsedTime.Format("2006-01-02")
				endTime = parsedTime.Format("15:04")
			}
		} else if item.End.Date != "" {
			t, err := time.Parse("2006-01-02", item.End.Date)
			if err == nil {
				endDate = t.AddDate(0, 0, -1).Format("2006-01-02")
			} else {
				endDate = item.End.Date
			}
		}

		act := domain.Activity{
			ID:          "google_" + item.ID,
			Title:       item.Summary,
			StartDate:   startDate,
			Location:    item.Location,
			ColorLabel:  "#3B82F6", // Default blue for Google Calendar events (PHBI)
			Description: description,
			Source:      "google_calendar",
		}

		if endDate != "" && endDate != startDate {
			act.EndDate = &endDate
		}
		if startTime != "" {
			act.StartTime = &startTime
		}
		if endTime != "" && endTime != startTime {
			act.EndTime = &endTime
		}

		activities = append(activities, act)
	}

	return activities, nil
}

func (s *activityService) GetMergedActivities() ([]domain.Activity, error) {
	localActivities, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}

	googleActivities, err := s.FetchGoogleCalendarEvents()
	if err != nil {
		log.Printf("Error fetching Google Calendar events: %v", err)
	}

	allActivities := append(localActivities, googleActivities...)
	return allActivities, nil
}

func (s *activityService) CreateActivity(req domain.CreateActivityRequest) (*domain.Activity, error) {
	if req.Title == "" || req.StartDate == "" {
		return nil, fmt.Errorf("title and start date are required")
	}

	// Clean fields
	if req.EndDate != nil && *req.EndDate == "" {
		req.EndDate = nil
	}
	if req.StartTime != nil && *req.StartTime == "" {
		req.StartTime = nil
	}
	if req.EndTime != nil && *req.EndTime == "" {
		req.EndTime = nil
	}
	if req.Description != nil && *req.Description == "" {
		req.Description = nil
	}

	activity := &domain.Activity{
		ID:          uuid.New().String(),
		Title:       req.Title,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Location:    req.Location,
		ColorLabel:  req.ColorLabel,
		Description: req.Description,
		Source:      "local",
	}

	if err := s.repo.Create(activity); err != nil {
		return nil, err
	}

	return activity, nil
}

func (s *activityService) UpdateActivity(id string, req domain.CreateActivityRequest) (*domain.Activity, error) {
	activity, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if req.Title == "" || req.StartDate == "" {
		return nil, fmt.Errorf("title and start date are required")
	}

	// Clean fields
	if req.EndDate != nil && *req.EndDate == "" {
		req.EndDate = nil
	}
	if req.StartTime != nil && *req.StartTime == "" {
		req.StartTime = nil
	}
	if req.EndTime != nil && *req.EndTime == "" {
		req.EndTime = nil
	}
	if req.Description != nil && *req.Description == "" {
		req.Description = nil
	}

	activity.Title = req.Title
	activity.StartDate = req.StartDate
	activity.EndDate = req.EndDate
	activity.StartTime = req.StartTime
	activity.EndTime = req.EndTime
	activity.Location = req.Location
	activity.ColorLabel = req.ColorLabel
	activity.Description = req.Description

	if err := s.repo.Update(activity); err != nil {
		return nil, err
	}

	return activity, nil
}

func (s *activityService) DeleteActivity(id string) error {
	activity, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	return s.repo.Delete(activity)
}
