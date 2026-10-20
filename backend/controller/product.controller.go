package controller

import (
	"flash_drop/helper"
	"flash_drop/repository"
	"flash_drop/types"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ProductController struct {
	logger *zap.Logger
	repo   *repository.ProductRepository
}

func InitProductController(logger *zap.Logger, repo *repository.ProductRepository) *ProductController {
	return &ProductController{
		logger: logger,
		repo:   repo,
	}
}

func (pc *ProductController) GetAllProduct(c *gin.Context) {
	rows, errRepo := pc.repo.GetAllProduct(c)
	if errRepo != nil {
		helper.ErrorResponse(c, 500, errRepo, "Something error please try again later")
		return
	}

	var listProduct []types.ProductType
	for rows.Next() {
		prd := new(types.ProductType)
		err := rows.Scan(&prd.Id, &prd.CategoryId, &prd.Name, &prd.Slug, &prd.Description, &prd.Price, &prd.Stock, &prd.ImageUrl, &prd.IsActive, &prd.CreatedAt, &prd.UpdatedAt)
		if err != nil {
			pc.logger.Error(err.Error())
			helper.ErrorResponse(c, 500, err, "Something error please try again later")
			return
		}
		listProduct = append(listProduct, *prd)
	}

	helper.SuccessResponse(c, 200, listProduct, "")
}

func (pc *ProductController) CreateProduct(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		pc.logger.Error(err.Error())
		helper.ErrorResponse(c, 400, err, "Failed to parse multipart form")
		return
	}

	var userInput types.InputProductType
	err = c.ShouldBind(&userInput)
	if err != nil {
		pc.logger.Error(err.Error())
		helper.ErrorResponse(c, 500, err, err.Error())
		return
	}

	// Make slug
	userInput.Slug = helper.MakeSlug(userInput.Name)

	// Insert image
	files := form.File["image_url"]
	uploadedFiles, err := helper.UploadFile(c, files)
	if err != nil {
		pc.logger.Error(err.Error())
		helper.ErrorResponse(c, 404, err, err.Error())
		return
	}
	userInput.ImageUrl = uploadedFiles

	// Insert Data
	result, errRepo := pc.repo.CreateProduct(c, userInput)
	if errRepo != nil {
		pc.logger.Error(errRepo.ErrorMessage)
		helper.ErrorResponse(c, errRepo.HttpCode, errRepo, errRepo.ErrorMessage)
		return
	}
	lastInsertId, err := result.LastInsertId()

	data := map[string]any{
		"id": lastInsertId,
	}

	helper.SuccessResponse(c, 201, data, "Product created")
}

func (pc *ProductController) EditProduct(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		pc.logger.Error(err.Error())
		helper.ErrorResponse(c, 400, err, "Failed to parse multipart form")
		return
	}

	var userInput types.InputProductType
	err = c.ShouldBind(&userInput)
	productId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		pc.logger.Error(err.Error())
		helper.ErrorResponse(c, 500, err, "Error when convert param 'id' to int")
		return
	}

	// Make slug
	userInput.Slug = helper.MakeSlug(userInput.Name)

	// Update Image
	// -- Get Old Data
	var product types.ProductType
	row := pc.repo.GetProductById(c, productId)

	err = row.Scan(
		&product.Id,
		&product.CategoryId,
		&product.Name,
		&product.Slug,
		&product.Description,
		&product.Price,
		&product.Stock,
		&product.ImageUrl,
		&product.IsActive,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		pc.logger.Error(err.Error())
		helper.ErrorResponse(c, 500, err, "Error when getting old data")
		return
	}

	// -- Insert New Data
	files := form.File["image_url"]
	uploadedFiles, err := helper.EditFile(c, product.ImageUrl, files)
	userInput.ImageUrl = uploadedFiles

	// Edit Data
	result, errRepo := pc.repo.EditProduct(c, userInput, productId)
	if errRepo != nil {
		pc.logger.Error(errRepo.ErrorMessage)
		helper.ErrorResponse(c, errRepo.HttpCode, errRepo, "Error when edit category, please try again later")
		return
	}

	helper.SuccessResponse(c, 200, result, "Category edited")
}

func (pc *ProductController) DeleteProduct(c *gin.Context) {
	productId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		pc.logger.Error(err.Error())
		helper.ErrorResponse(c, 500, err, "Error when convert param 'id' to int")
		return
	}

	result, errRepo := pc.repo.DeleteProduct(c, productId)
	if errRepo != nil {
		pc.logger.Error(errRepo.ErrorMessage)
		helper.ErrorResponse(c, errRepo.HttpCode, errRepo, "Error when delete product, please try again later")
		return
	}

	helper.SuccessResponse(c, 200, result, "Product Deleted")
}
