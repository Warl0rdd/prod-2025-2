package b2c

//
//import (
//	"context"
//	"github.com/gofiber/fiber/v3"
//	"solution/cmd/app"
//	"solution/internal/adapters/controller/api/validator"
//	"solution/internal/adapters/database/postgres"
//	"solution/internal/domain/dto"
//	"solution/internal/domain/entity"
//	"solution/internal/domain/service"
//	"time"
//)
//
//type UserService interface {
//	Create(ctx context.Context, registerReq dto.UserRegister) (*entity.User, error)
//	GetByID(ctx context.Context, uuid string) (*entity.User, error)
//	Update(ctx context.Context, user *entity.User) (*entity.User, error)
//	GetByEmail(ctx context.Context, email string) (*entity.User, error)
//}
//
//type TokenService interface {
//	GenerateAuthTokens(c context.Context, userID string) (*dto.AuthTokens, error)
//	GenerateToken(ctx context.Context, userID string, expires time.Time, tokenType string) (*entity.Token, error)
//}
//
//type UserHandler struct {
//	userService  UserService
//	tokenService TokenService
//	validator    *validator.Validator
//}
//
//func NewUserHandler(app *app.App) *UserHandler {
//	userStorage := postgres.NewUserStorage(app.DB)
//	tokenStorage := postgres.NewTokenStorage(app.DB)
//
//	return &UserHandler{
//		userService:  service.NewUserService(userStorage),
//		tokenService: service.NewTokenService(tokenStorage),
//		validator:    app.Validator,
//	}
//}
//
//func (h UserHandler) Setup(router fiber.Router, middleware fiber.Handler) {
//	userGroup := router.Group("/user")
//}
