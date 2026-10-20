package controller

import (
	"flash_drop/helper"
	"flash_drop/repository"
	"flash_drop/types"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type CartController struct {
	log  *zap.Logger
	repo *repository.CartRepository
}

func InitCartController(log *zap.Logger, repo *repository.CartRepository) *CartController {
	return &CartController{log: log, repo: repo}
}

func (cc *CartController) getUserId(c *gin.Context) (int, bool) {
	v, exist := c.Get("user_id")
	if !exist {
		cc.log.Error("user_id not found")
		helper.ErrorResponse(c, 401, nil, "unauthorized")
		return 0, false
	}

	switch id := v.(type) {
	case float64:
		return int(id), true
	case int:
		return id, true
	default:
		cc.log.Error("invalid user_id type")
		helper.ErrorResponse(c, 500, nil, "invalid user_id type")
		return 0, false
	}
}

func (cc *CartController) GetItems(c *gin.Context) {
	userId, ok := cc.getUserId(c)
	if !ok {
		return
	}

	rows, err := cc.repo.GetItems(c, userId)
	if err != nil {
		cc.log.Error(err.Error())
		helper.ErrorResponse(c, 500, err, "Error when getting data")
		return
	}

	var items []types.UserCartType
	for rows.Next() {
		var item types.UserCartType
		err = rows.Scan(&item.CartItemsId, &item.CartId, &item.ProuctId, &item.Quantity, &item.CategoryName, &item.Name, &item.Slug, &item.Price, &item.ImageUrl)
		if err != nil {
			cc.log.Error(err.Error())
			helper.ErrorResponse(c, 500, err, "Error when getting data")
			return
		}

		items = append(items, item)
	}

	helper.SuccessResponse(c, 200, items, "Success getting data")
}

func (cc *CartController) AddItems(c *gin.Context) {
	userId, ok := cc.getUserId(c)
	if !ok {
		return
	}

	var req types.UserInputUpdateCartItemsType
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorResponse(c, 400, nil, "invalid request body")
		return
	}

	if len(req.Item) == 0 {
		helper.ErrorResponse(c, 400, nil, "item cannot be empty")
		return
	}
	for _, it := range req.Item {
		if it.ProductId == nil || it.Quantity == nil {
			helper.ErrorResponse(c, 400, nil, "product_id and quantity are required")
			return
		}
		if *it.Quantity <= 0 {
			helper.ErrorResponse(c, 400, nil, "quantity must be greater than 0")
			return
		}
	}

	if errRes := cc.repo.AddItems(c, userId, req.Item); errRes != nil {
		cc.log.Error("failed to add items", zap.String("error", errRes.ErrorMessage))
		helper.ErrorResponse(c, errRes.HttpCode, nil, errRes.ErrorMessage)
		return
	}

	helper.SuccessResponse(c, 200, nil, "Success add items into your cart")
}

func (cc *CartController) UpdateItems(c *gin.Context) {
	userId, ok := cc.getUserId(c)
	if !ok {
		return
	}

	var req types.UserInputUpdateCartItemsType
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorResponse(c, 400, nil, "invalid request body")
		return
	}

	if len(req.Item) == 0 {
		helper.ErrorResponse(c, 400, nil, "item cannot be empty")
		return
	}

	if errRes := cc.repo.UpdateCart(c, userId, req.Item); errRes != nil {
		cc.log.Error("failed to update cart", zap.String("error", errRes.ErrorMessage))
		helper.ErrorResponse(c, errRes.HttpCode, nil, errRes.ErrorMessage)
		return
	}

	helper.SuccessResponse(c, 200, nil, "Success update your cart")
}
