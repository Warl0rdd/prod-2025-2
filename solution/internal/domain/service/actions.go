package service

import (
	"context"
	"solution/internal/domain/dto"
)

type actionsStorage interface {
	AddLike(ctx context.Context, userID, promoID string) error
	DeleteLike(ctx context.Context, userID, promoID string) error
	AddComment(ctx context.Context, userID, promoID, text string) error
	GetComments(ctx context.Context, promoID string, limit, offset int) ([]dto.Comment, error)
	GetCommentById(ctx context.Context, promoID, commentID string) (dto.Comment, error)
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

func (s *actionsService) GetComments(ctx context.Context, promoID string, limit, offset int) ([]dto.Comment, error) {
	return s.storage.GetComments(ctx, promoID, limit, offset)
}

func (s *actionsService) GetCommentById(ctx context.Context, commentID, promoID string) (dto.Comment, error) {
	return s.storage.GetCommentById(ctx, commentID, promoID)
}
