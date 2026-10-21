package controller

import (
	"flash_drop/helper"
	"flash_drop/repository"
	"flash_drop/types"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserController struct {
	logger *zap.Logger
	repo   *repository.AuthRepository
}

func InitUserController(logger *zap.Logger, repo *repository.AuthRepository) *UserController {
	return &UserController{
		logger: logger,
		repo:   repo,
	}
}

func (uc *UserController) Register(c *gin.Context) {
	var in types.UserRegisterType
	if err := c.ShouldBindJSON(&in); err != nil {
		helper.ErrorResponse(c, http.StatusBadRequest, nil, "Invalid request body")
		return
	}
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))

	hash, err := helper.HashPassword(in.Password)
	if err != nil {
		uc.logger.Error("hash password", zap.Error(err))
		helper.ErrorResponse(c, http.StatusInternalServerError, nil, "Something went wrong")
		return
	}
	in.Password = hash

	result, repoErr := uc.repo.Create(c, in)
	if repoErr != nil {
		helper.ErrorResponse(c, repoErr.HttpCode, nil, repoErr.ErrorMessage)
		return
	}
	userID, _ := result.LastInsertId()

	// Role SELALU customer untuk pendaftaran publik
	token, err := helper.CreateToken(in.Name, int(userID), types.Customer)
	if err != nil {
		uc.logger.Error("create token", zap.Error(err))
		helper.ErrorResponse(c, http.StatusInternalServerError, nil, "Something went wrong")
		return
	}

	helper.SuccessResponse(c, http.StatusCreated, map[string]any{
		"id": userID, "role": types.Customer, "email": in.Email, "name": in.Name, "token": token,
	}, "User Registered")
}

func (uc *UserController) LogIn(c *gin.Context) {
	var userInput types.UserLoginType
	err := c.ShouldBind(&userInput)
	if err != nil {
		uc.logger.Error(err.Error())
		helper.ErrorResponse(c, 500, nil, err.Error())
		return
	}

	// Verify password
	var user types.UserType
	row := uc.repo.GetUserByEmail(c, userInput)
	err = row.Scan(&user.Id, &user.Name, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)

	matchedPassword := helper.VerifyPassword(userInput.Password, user.Password)
	if matchedPassword != true {
		helper.ErrorResponse(c, 500, nil, "Email or password is wrong")
		return
	}

	// Create JWT Token
	jwtToken, err := helper.CreateToken(user.Name, user.Id, user.Role)
	if err != nil {
		uc.logger.Error(err.Error())
		helper.ErrorResponse(c, 500, nil, err.Error())
		return
	}

	data := map[string]any{
		"role":  user.Role,
		"email": user.Email,
		"name":  user.Name,
		"token": jwtToken,
	}
	helper.SuccessResponse(c, 200, data, "Logged In !!")
	return
}

func (uc *UserController) VerifyToken(c *gin.Context) {
	var userInput struct {
		Token string `json:"token" binding:"required"`
	}

	err := c.ShouldBind(&userInput)
	if err != nil {
		uc.logger.Error(err.Error())
		helper.ErrorResponse(c, 500, nil, err.Error())
		return
	}

	// Verify JWT Token
	claims, err := helper.VerifyToken(userInput.Token)
	if err != nil {
		uc.logger.Error(err.Error())
		helper.ErrorResponse(c, 401, nil, err.Error())
		return
	}

	helper.SuccessResponse(c, 200, claims, "Verified")
	return
}
