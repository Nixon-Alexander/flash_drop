package controller

import (
	"flash_drop/helper"
	"flash_drop/repository"
	"flash_drop/types"

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
	var userInput types.UserRegisterType
	err := c.ShouldBind(&userInput)
	if err != nil {
		uc.logger.Error(err.Error())
		helper.ErrorResponse(c, 500, nil, err.Error())
		return
	}

	// Hash password
	hashedPassword, err := helper.HashPassword(userInput.Password)
	if err != nil {
		uc.logger.Error(err.Error())
		helper.ErrorResponse(c, 500, nil, err.Error())
		return
	}
	userInput.Password = hashedPassword

	// Insert user
	result, errRepo := uc.repo.Create(c, userInput)
	if errRepo != nil {
		uc.logger.Error(errRepo.ErrorMessage)
		helper.ErrorResponse(c, errRepo.HttpCode, errRepo, errRepo.ErrorMessage)
		return
	}
	lastInsertId, _ := result.LastInsertId()

	// Generate JWT
	jwtToken, err := helper.CreateToken(userInput.Name, int(lastInsertId), userInput.Role)
	if err != nil {
		uc.logger.Error(err.Error())
		helper.ErrorResponse(c, 500, nil, err.Error())
		return
	}

	data := map[string]any{
		"id":    lastInsertId,
		"role":  userInput.Role,
		"email": userInput.Email,
		"name":  userInput.Name,
		"token": jwtToken,
	}

	helper.SuccessResponse(c, 201, data, "User Registered")
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
	err, claims := helper.VerifyToken(userInput.Token)
	if err != nil {
		uc.logger.Error(err.Error())
		helper.ErrorResponse(c, 401, nil, err.Error())
		return
	}

	helper.SuccessResponse(c, 200, claims, "Verified")
	return
}
