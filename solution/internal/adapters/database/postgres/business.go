package postgres

import (
	"context"
	"gorm.io/gorm"
	"solution/internal/domain/entity"
)

// businessStorage is a struct that contains a pointer to a gorm.DB instance to interact with business repository.
type businessStorage struct {
	db *gorm.DB
}

// NewBusinessStorage is a function that returns a new instance of businessStorage.
func NewBusinessStorage(db *gorm.DB) *businessStorage {
	return &businessStorage{db: db}
}

// Create is a method to create a new Business in database.
func (s *businessStorage) Create(ctx context.Context, business entity.Business) (*entity.Business, error) {
	err := s.db.WithContext(ctx).Create(&business).Error
	return &business, err
}

// GetByID is a method that returns an error and a pointer to a Business instance by id.
func (s *businessStorage) GetByID(ctx context.Context, id string) (*entity.Business, error) {
	var business *entity.Business
	err := s.db.WithContext(ctx).Model(&entity.Business{}).Where("id = ?", id).First(&business).Error
	return business, err
}

// GetAll is a method that returns a slice of pointers to all Business instances.
func (s *businessStorage) GetAll(ctx context.Context, limit, offset int) ([]entity.Business, error) {
	var businesss []entity.Business
	err := s.db.WithContext(ctx).Model(&entity.Business{}).Limit(limit).Offset(offset).Find(&businesss).Error
	return businesss, err
}

// Update is a method to update an existing Business in database.
func (s *businessStorage) Update(ctx context.Context, business *entity.Business) (*entity.Business, error) {
	err := s.db.WithContext(ctx).Model(&entity.Business{}).Where("id = ?", business.ID).Updates(&business).Error
	return business, err
}

// Delete is a method to delete an existing Business in database.
func (s *businessStorage) Delete(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Unscoped().Delete(&entity.Business{}, "id = ?", id).Error
}

// GetByEmail is a method that returns a pointer to a Business instance and error by email.
func (s *businessStorage) GetByEmail(ctx context.Context, email string) (*entity.Business, error) {
	var business *entity.Business
	err := s.db.WithContext(ctx).Where("email = ?", email).First(&business).Error
	return business, err
}
