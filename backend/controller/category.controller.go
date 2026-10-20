package controller

import (
	"flash_drop/helper"
	"flash_drop/repository"
	"flash_drop/types"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type CategoryController struct {
	logger *zap.Logger
	repo   *repository.CategoryRepository
}

func InitCategoryController(logger *zap.Logger, repo *repository.CategoryRepository) *CategoryController {
	return &CategoryController{
		logger: logger,
		repo:   repo,
	}
}

func (cc *CategoryController) CreateCategory(c *gin.Context) {
	var userInput types.InputCategoryType
	err := c.ShouldBindJSON(&userInput)
	if err != nil {
		cc.logger.Error(err.Error())
		helper.ErrorResponse(c, 500, err, err.Error())
		return
	}

	// Make slug
	userInput.Slug = helper.MakeSlug(userInput.Name)

	// Insert Data
	result, errRepo := cc.repo.CreateCategory(c, userInput)
	if errRepo != nil {
		cc.logger.Error(errRepo.ErrorMessage)
		helper.ErrorResponse(c, errRepo.HttpCode, errRepo, errRepo.ErrorMessage)
		return
	}
	lastInsertId, err := result.LastInsertId()

	data := map[string]any{
		"id": lastInsertId,
	}

	helper.SuccessResponse(c, 201, data, "Product created")
}

func (cc *CategoryController) GetAllCategory(c *gin.Context) {
	rows, errRepo := cc.repo.GetAllCategory(c)
	if errRepo != nil {
		helper.ErrorResponse(c, 500, errRepo, "Something error please try again later")
		return
	}

	var listCategory []types.CategoryType
	for rows.Next() {
		cat := new(types.CategoryType)
		err := rows.Scan(&cat.Id, &cat.Name, &cat.Slug, &cat.CreatedAt, &cat.UpdatedAt)
		if err != nil {
			cc.logger.Error(err.Error())
			helper.ErrorResponse(c, 500, err, "Something error please try again later")
			return
		}
		listCategory = append(listCategory, *cat)
	}

	helper.SuccessResponse(c, 200, listCategory, "")
}

func (cc *CategoryController) EditCategory(c *gin.Context) {
	categoryId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		cc.logger.Error(err.Error())
		helper.ErrorResponse(c, 500, err, "Error when convert param 'id' to int")
		return
	}

	userInput := types.InputCategoryType{}
	c.ShouldBindJSON(&userInput)

	// Make slug
	userInput.Slug = helper.MakeSlug(userInput.Name)

	result, errRepo := cc.repo.EditCategory(c, userInput, categoryId)
	if errRepo != nil {
		cc.logger.Error(errRepo.ErrorMessage)
		helper.ErrorResponse(c, errRepo.HttpCode, errRepo, "Error when edit category, please try again later")
		return
	}

	helper.SuccessResponse(c, 200, result, "Category edited")
}

func (cc *CategoryController) DeleteCategory(c *gin.Context) {
	categoryId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		cc.logger.Error(err.Error())
		helper.ErrorResponse(c, 500, err, "Error when convert param 'id' to int")
		return
	}

	result, errRepo := cc.repo.DeleteCategory(c, categoryId)
	if errRepo != nil {
		cc.logger.Error(errRepo.ErrorMessage)
		helper.ErrorResponse(c, errRepo.HttpCode, errRepo, "Error when delete category, please try again later")
		return
	}

	helper.SuccessResponse(c, 200, result, "Category Deleted")
}
