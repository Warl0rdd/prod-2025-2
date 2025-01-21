package b2c

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v3"
	"solution/cmd/app"
	"solution/internal/adapters/controller/api/validator"
	"solution/internal/adapters/database/postgres"
	"solution/internal/domain/common/errorz"
	"solution/internal/domain/dto"
	"solution/internal/domain/entity"
	"solution/internal/domain/service"
	"time"
)

type UserService interface {
	Create(ctx context.Context, registerReq dto.UserRegister) (*entity.User, error)
	GetByID(ctx context.Context, uuid string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
}

type TokenService interface {
	GenerateAuthTokens(c context.Context, userID string) (*dto.AuthTokens, error)
	GenerateToken(ctx context.Context, userID string, expires time.Time, tokenType string) (*entity.Token, error)
}

type UserHandler struct {
	userService  UserService
	tokenService TokenService
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
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if errValidate := h.validator.ValidateData(userDTO); errValidate != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	user, errCreate := h.userService.Create(c.Context(), userDTO)
	if errors.Is(errCreate, errorz.EmailTaken) {
		return c.Status(fiber.StatusConflict).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Такой email уже зарегистрирован.",
		})
	}

	// Other errors
	if errCreate != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка при создании пользователя.",
		})
	}

	tokens, tokensErr := h.tokenService.GenerateAuthTokens(c.Context(), user.ID)
	if tokensErr != nil || tokens == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка при генерации токенов.",
		})
	}

	response := dto.UserRegisterResponse{
		Token: tokens.Access.Token,
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

func (h UserHandler) login(c fiber.Ctx) error {
	var userDTO dto.UserLogin

	if err := c.Bind().Body(&userDTO); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if errValidate := h.validator.ValidateData(userDTO); errValidate != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	user, errFetch := h.userService.GetByEmail(c.Context(), userDTO.Email)
	if errFetch != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Неверный email или пароль.",
		})
	}

	passErr := user.ComparePassword(userDTO.Password)
	if passErr != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Неверный email или пароль.",
		})
	}

	tokens, tokensErr := h.tokenService.GenerateAuthTokens(c.Context(), user.ID)
	if tokensErr != nil || tokens == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "failed to generate auth tokens",
		})
	}

	response := dto.UserRegisterResponse{
		Token: tokens.Access.Token,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func (h UserHandler) getProfile(c fiber.Ctx) error {
	user := c.Locals("user").(*entity.User)

	profile := dto.UserProfile{
		Email:     user.Email,
		Name:      user.Name,
		Surname:   user.Surname,
		AvatarURL: user.AvatarURL,
		Other: dto.UserOther{
			Age:     user.Age,
			Country: user.Country.String(),
		},
	}

	return c.Status(fiber.StatusOK).JSON(profile)
}

func (h UserHandler) updateProfile(c fiber.Ctx) error {
	user := c.Locals("user").(*entity.User)
	var userDTO dto.UserProfileUpdate

	if err := c.Bind().Body(&userDTO); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	if errValidate := h.validator.ValidateData(userDTO); errValidate != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка в данных запроса.",
		})
	}

	user.Name = userDTO.Name
	user.Surname = userDTO.Surname
	user.AvatarURL = userDTO.AvatarURL
	user.SetPassword(userDTO.Password)

	updatedUser, errUpdate := h.userService.Update(c.Context(), user)
	if errUpdate != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.HTTPResponse{
			Status:  "error",
			Message: "Ошибка при обновлении профиля.",
		})
	}

	profile := dto.UserProfile{
		Email:     updatedUser.Email,
		Name:      updatedUser.Name,
		Surname:   updatedUser.Surname,
		AvatarURL: updatedUser.AvatarURL,
		Other: dto.UserOther{
			Age:     updatedUser.Age,
			Country: updatedUser.Country.String(),
		},
	}

	return c.Status(fiber.StatusOK).JSON(profile)
}

func (h UserHandler) Setup(router fiber.Router, middleware fiber.Handler) {
	userGroup := router.Group("/user")
	userGroup.Post("/auth/sign-up", h.register)
	userGroup.Post("/auth/sign-in", h.login)
	userGroup.Get("/profile", h.getProfile, middleware)
	userGroup.Patch("/profile", h.updateProfile, middleware)
}
