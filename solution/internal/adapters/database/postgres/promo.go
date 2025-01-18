package postgres

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"solution/internal/domain/entity"
)

// promoStorage is a struct that contains a pointer to a gorm.DB instance to interact with promo repository.
type promoStorage struct {
	db *gorm.DB
}

// NewPromoStorage is a function that returns a new instance of promoStorage.
func NewPromoStorage(db *gorm.DB) *promoStorage {
	return &promoStorage{db: db}
}

// Create is a method to create a new Promo in database.
func (s *promoStorage) Create(ctx context.Context, promo entity.Promo) (*entity.Promo, error) {
	err := s.db.Model(&promo).WithContext(ctx).Create(&promo).Error
	return &promo, err
}

// GetByID is a method that returns an error and a pointer to a Promo instance by id.
func (s *promoStorage) GetByID(ctx context.Context, id string) (*entity.Promo, error) {
	var promo *entity.Promo
	err := s.db.WithContext(ctx).Model(&entity.Promo{}).Where("id = ?", id).First(&promo).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return promo, err
}

// GetAll is a method that returns a slice of pointers to all Promo instances.
func (s *promoStorage) GetAll(ctx context.Context, limit, offset int) ([]entity.Promo, error) {
	var promos []entity.Promo
	err := s.db.WithContext(ctx).Model(&entity.Promo{}).Limit(limit).Offset(offset).Find(&promos).Error
	return promos, err
}

// Update is a method to update an existing Promo in database.
func (s *promoStorage) Update(ctx context.Context, promo *entity.Promo) (*entity.Promo, error) {
	err := s.db.WithContext(ctx).Model(&entity.Promo{}).Where("id = ?", promo.PromoID).Updates(&promo).Error
	return promo, err
}
