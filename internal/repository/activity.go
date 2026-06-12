package repository

import (
	"laci-v3/be/internal/domain"

	"gorm.io/gorm"
)

type ActivityRepository interface {
	FindAll() ([]domain.Activity, error)
	FindByID(id string) (*domain.Activity, error)
	Create(activity *domain.Activity) error
	Update(activity *domain.Activity) error
	Delete(activity *domain.Activity) error
}

type activityRepository struct {
	db *gorm.DB
}

func NewActivityRepository(db *gorm.DB) ActivityRepository {
	return &activityRepository{db: db}
}

func (r *activityRepository) FindAll() ([]domain.Activity, error) {
	var activities []domain.Activity
	err := r.db.Order("start_date asc").Find(&activities).Error
	return activities, err
}

func (r *activityRepository) FindByID(id string) (*domain.Activity, error) {
	var activity domain.Activity
	if err := r.db.First(&activity, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &activity, nil
}

func (r *activityRepository) Create(activity *domain.Activity) error {
	return r.db.Create(activity).Error
}

func (r *activityRepository) Update(activity *domain.Activity) error {
	return r.db.Save(activity).Error
}

func (r *activityRepository) Delete(activity *domain.Activity) error {
	return r.db.Delete(activity).Error
}
