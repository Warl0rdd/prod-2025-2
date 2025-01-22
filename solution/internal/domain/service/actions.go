package service

import "context"

type actionsStorage interface {
	AddLike(ctx context.Context, userID, promoID string) error
	DeleteLike(ctx context.Context, userID, promoID string) error
	AddComment(ctx context.Context, userID, promoID, text string) error
}

type actionsService struct {
	storage actionsStorage
}

func NewActionsService(storage actionsStorage) *actionsService {
	return &actionsService{storage: storage}
}

func (s *actionsService) AddLike(ctx context.Context, userID, promoID string) error {
	return s.storage.AddLike(ctx, userID, promoID)
}

func (s *actionsService) DeleteLike(ctx context.Context, userID, promoID string) error {
	return s.storage.DeleteLike(ctx, userID, promoID)
}

func (s *actionsService) AddComment(ctx context.Context, userID, promoID, text string) error {
	return s.storage.AddComment(ctx, userID, promoID, text)
}
