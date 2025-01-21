package service

import "context"

type actionsStorage interface {
	AddLike(ctx context.Context, userID, promoID string) error
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
