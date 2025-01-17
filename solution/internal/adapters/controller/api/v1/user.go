package v1

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v3"
	"github.com/spf13/viper"
	"solution/cmd/app"
	"solution/internal/adapters/controller/api/validator"
	"solution/internal/adapters/database/postgres"
	"solution/internal/adapters/logger"
	"solution/internal/domain/dto"
	"solution/internal/domain/entity"
	"solution/internal/domain/service"
	"solution/internal/domain/utils/auth"
	"time"
)

type UserService interface {
	Create(ctx context.Context, registerReq dto.UserRegister, code string) (*entity.User, error)
	GetByID(ctx context.Context, uuid string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
}

type TokenService interface {
	GenerateAuthTokens(c context.Context, userID string) (*dto.AuthTokens, error)
	GenerateToken(ctx context.Context, userID string, expires time.Time, tokenType string) (*entity.Token, error)
}

type EmailService interface {
	Send(ctx context.Context, email string, text string, subject string) error
	Check(ctx context.Context, email string) (bool, error)
}

type UserHandler struct {
	userService  UserService
	tokenService TokenService
	emailService EmailService
	validator    *validator.Validator
}

func NewUserHandler(app *app.App) *UserHandler {
	userStorage := postgres.NewUserStorage(app.DB)
	tokenStorage := postgres.NewTokenStorage(app.DB)

	return &UserHandler{
		userService:  service.NewUserService(userStorage),
		tokenService: service.NewTokenService(tokenStorage),
		validator:    app.Validator,
	}
}

func (h UserHandler) register(c fiber.Ctx) error {
	var userDTO dto.UserRegister

	if err := c.Bind().Body(&userDTO); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPError{
			Code:    fiber.StatusBadRequest,
			Message: err.Error(),
		})
	}

	if errValidate := h.validator.ValidateData(userDTO); errValidate != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPError{
			Code:    fiber.StatusBadRequest,
			Message: errValidate.Error(),
		})
	}

	mailValid, mvErr := h.emailService.Check(c.Context(), userDTO.Email)
	if mvErr != nil || !mailValid {
		logger.Log.Errorf("invalid email: %s", userDTO.Email)
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPError{
			Code:    fiber.StatusBadRequest,
			Message: "invalid email",
		})
	}

	code := auth.GenerateCode()
	msErr := h.emailService.Send(c.Context(), userDTO.Email, fmt.Sprintf("Your code is: <b>%s</b>", code), "Verification Code")
	if msErr != nil {
		logger.Log.Errorf("email sending error: %s", msErr.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(dto.HTTPError{
			Code:    fiber.StatusInternalServerError,
			Message: msErr.Error(),
		})
	}

	user, errCreate := h.userService.Create(c.Context(), userDTO, code)
	if errCreate != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.HTTPError{
			Code:    fiber.StatusInternalServerError,
			Message: errCreate.Error(),
		})
	}

	tokens, tokensErr := h.tokenService.GenerateAuthTokens(c.Context(), user.ID)
	if tokensErr != nil || tokens == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.HTTPError{
			Code:    fiber.StatusInternalServerError,
			Message: "failed to generate auth tokens",
		})
	}

	response := dto.UserRegisterResponse{
		User: dto.UserReturn{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
			Role:     user.Role,
		},
		Tokens: *tokens,
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

func (h UserHandler) login(c fiber.Ctx) error {
	var userDTO dto.UserLogin

	if err := c.Bind().Body(&userDTO); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPError{
			Code:    fiber.StatusBadRequest,
			Message: err.Error(),
		})
	}

	if errValidate := h.validator.ValidateData(userDTO); errValidate != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPError{
			Code:    fiber.StatusBadRequest,
			Message: errValidate.Error(),
		})
	}

	user, errFetch := h.userService.GetByEmail(c.Context(), userDTO.Email)
	if errFetch != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.HTTPError{
			Code:    fiber.StatusNotFound,
			Message: "not found",
		})
	}

	passErr := user.ComparePassword(userDTO.Password)
	if passErr != nil {
		return c.Status(fiber.StatusForbidden).JSON(dto.HTTPError{
			Code:    fiber.StatusForbidden,
			Message: "invalid password",
		})
	}

	tokens, tokensErr := h.tokenService.GenerateAuthTokens(c.Context(), user.ID)
	if tokensErr != nil || tokens == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.HTTPError{
			Code:    fiber.StatusInternalServerError,
			Message: "failed to generate auth tokens",
		})
	}

	response := dto.UserRegisterResponse{
		User: dto.UserReturn{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
			Role:     user.Role,
		},
		Tokens: *tokens,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func (h UserHandler) refreshToken(c fiber.Ctx) error {
	var accessTokenDTO dto.Token

	if err := c.Bind().Body(&accessTokenDTO); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPError{
			Code:    fiber.StatusBadRequest,
			Message: err.Error(),
		})
	}

	if errValidate := h.validator.ValidateData(accessTokenDTO); errValidate != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPError{
			Code:    fiber.StatusBadRequest,
			Message: errValidate.Error(),
		})
	}

	userID, errToken := auth.VerifyToken(accessTokenDTO.Token, viper.GetString("service.backend.jwt.secret"), auth.TokenTypeAccess)

	if errToken != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.HTTPError{
			Code:    fiber.StatusUnauthorized,
			Message: errToken.Error(),
		})
	}

	expTime := time.Now().UTC().Add(time.Minute * time.Duration(viper.GetInt("service.backend.jwt.access-token-expiration")))

	newAccess, errNewAccess := h.tokenService.GenerateToken(c.Context(),
		userID,
		expTime,
		auth.TokenTypeAccess)

	if errNewAccess != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.HTTPError{
			Code:    fiber.StatusInternalServerError,
			Message: errNewAccess.Error(),
		})
	}

	response := dto.Token{
		Token:   newAccess.Token,
		Expires: expTime,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func (h UserHandler) Setup(router fiber.Router) {
	userGroup := router.Group("/user")
	userGroup.Post("/register", h.register)
	userGroup.Post("/login", h.login)
	userGroup.Post("/refresh", h.refreshToken)
}
