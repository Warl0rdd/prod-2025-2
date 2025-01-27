package service

import (
	"context"
	"encoding/json"
	"github.com/biter777/countries"
	"github.com/gofiber/fiber/v3/client"
	"os"
	"solution/internal/adapters/logger"
	"solution/internal/domain/common/errorz"
	"solution/internal/domain/dto"
	"solution/internal/domain/entity"
	"time"
)

type actionsStorage interface {
	AddLike(ctx context.Context, userID, promoID string) error
	DeleteLike(ctx context.Context, userID, promoID string) error
	AddComment(ctx context.Context, userID, promoID, text string) (string, error)
	GetComments(ctx context.Context, promoID string, limit, offset int) ([]dto.Comment, int64, error)
	GetCommentById(ctx context.Context, promoID, commentID string) (dto.Comment, error)
	UpdateComment(ctx context.Context, promoID, commentID, userID, text string) (dto.Comment, error)
	DeleteComment(ctx context.Context, promoID, commentID, userID string) error
}

type activationStorage interface {
	ActivatePromo(ctx context.Context, age int, country countries.CountryCode, promoID, userID string) (string, error)
}

type activationRedisStorage interface {
	Cache(ctx context.Context, email string, until time.Time) error
	CheckCache(ctx context.Context, email string) (bool, error)
}

type actionsService struct {
	actionStorage          actionsStorage
	activationStorage      activationStorage
	activationRedisStorage activationRedisStorage
}

func NewActionsService(actionStorage actionsStorage, activationStorage activationStorage, activationRedisStorage activationRedisStorage) *actionsService {
	return &actionsService{
		actionStorage:          actionStorage,
		activationStorage:      activationStorage,
		activationRedisStorage: activationRedisStorage,
	}
}

func (s *actionsService) AddLike(ctx context.Context, userID, promoID string) error {
	return s.actionStorage.AddLike(ctx, userID, promoID)
}

func (s *actionsService) DeleteLike(ctx context.Context, userID, promoID string) error {
	return s.actionStorage.DeleteLike(ctx, userID, promoID)
}

func (s *actionsService) AddComment(ctx context.Context, userID, promoID, text string) (string, error) {
	return s.actionStorage.AddComment(ctx, userID, promoID, text)
}

func (s *actionsService) GetComments(ctx context.Context, promoID string, limit, offset int) ([]dto.Comment, int64, error) {
	return s.actionStorage.GetComments(ctx, promoID, limit, offset)
}

func (s *actionsService) GetCommentById(ctx context.Context, commentID, promoID string) (dto.Comment, error) {
	return s.actionStorage.GetCommentById(ctx, promoID, commentID)
}

func (s *actionsService) UpdateComment(ctx context.Context, promoID, commentID, userID, text string) (dto.Comment, error) {
	return s.actionStorage.UpdateComment(ctx, promoID, commentID, userID, text)
}

func (s *actionsService) DeleteComment(ctx context.Context, promoID, commentID, userID string) error {
	return s.actionStorage.DeleteComment(ctx, promoID, commentID, userID)
}

func (s *actionsService) Activate(ctx context.Context, user *entity.User, promoID string) (string, error) {
	antiFraudAddress := os.Getenv("ANTIFRAUD_ADDRESS")
	checkCache, cacheErr := s.activationRedisStorage.CheckCache(ctx, user.Email)
	if cacheErr != nil {
		logger.Log.Error(cacheErr)
	}

	if !checkCache {
		cc := client.New()
		cc.SetTimeout(10 * time.Second)
		headers := make(map[string]string)
		headers["Content-Type"] = "application/json"

		resp, err := cc.Post("http://"+antiFraudAddress+"/api/validate", client.Config{
			Ctx:    ctx,
			Header: headers,
			Body: dto.AntiFraudRequest{
				UserEmail: user.Email,
				PromoID:   promoID,
			},
		})

		if err != nil {
			return "", err
		}

		if resp.StatusCode() != 200 {
			newResp, newErr := cc.Post(antiFraudAddress, client.Config{
				Ctx:    ctx,
				Header: headers,
				Body: dto.AntiFraudRequest{
					UserEmail: user.Email,
					PromoID:   promoID,
				},
			})

			if newErr != nil {
				return "", err
			}

			if newResp.StatusCode() != 200 {
				return "", errorz.Forbidden
			}

			var respBody dto.AntiFraudResponse
			if jsonErr := json.Unmarshal(resp.Body(), &respBody); jsonErr != nil {
				return "", jsonErr
			}

			if respBody.CacheUntil != "" {
				until, _ := time.Parse("2006-01-02T15:04:05.000", respBody.CacheUntil)
				cacheCreateErr := s.activationRedisStorage.Cache(ctx, user.Email, until.Add(time.Hour*3)) // UTC+0 to +3
				if cacheCreateErr != nil {
					return "", cacheErr
				}
			}

			if respBody.Ok == false {
				return "", errorz.Forbidden
			}

			return s.activationStorage.ActivatePromo(ctx, user.Age, user.Country, promoID, user.ID)
		}

		var respBody dto.AntiFraudResponse
		if jsonErr := json.Unmarshal(resp.Body(), &respBody); jsonErr != nil {
			return "", jsonErr
		}

		if respBody.CacheUntil != "" {
			until, _ := time.Parse("2006-01-02T15:04:05.000", respBody.CacheUntil)
			cacheCreateErr := s.activationRedisStorage.Cache(ctx, user.Email, until.Add(time.Hour*3)) // UTC+0 to +3
			if cacheCreateErr != nil {
				return "", cacheErr
			}
		}

		if respBody.Ok == false {
			return "", errorz.Forbidden
		}

		return s.activationStorage.ActivatePromo(ctx, user.Age, user.Country, promoID, user.ID)
	}

	return s.activationStorage.ActivatePromo(ctx, user.Age, user.Country, promoID, user.ID)
}
