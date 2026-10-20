package controller

import (
	"errors"
	service "flash_drop/Service"
	"flash_drop/helper"
	"flash_drop/types"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type CheckoutController struct {
	log             *zap.Logger
	checkoutService *service.CheckoutService
}

func InitCheckoutController(
	log *zap.Logger,
	checkoutService *service.CheckoutService,
) *CheckoutController {
	return &CheckoutController{
		log:             log,
		checkoutService: checkoutService,
	}
}

func (cc *CheckoutController) BuyItems(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		helper.ErrorResponse(c, http.StatusUnauthorized, errors.New("user is not authorized"), "Unauthorized")
		return
	}

	userID, ok := userIDVal.(float64)
	userIDint := int(userID)
	if !ok {
		helper.ErrorResponse(c, http.StatusUnauthorized, errors.New("invalid user ID"), "Unauthorized")
		return
	}

	var userInput types.UserInputBuyItemsType

	if err := c.ShouldBindJSON(&userInput); err != nil {
		cc.log.Error(err.Error())
		helper.ErrorResponse(c, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	if len(userInput.Item) == 0 {
		cc.log.Error("Item cant be empty")
		helper.ErrorResponse(c, http.StatusBadRequest, errors.New("items cannot be empty"), "Invalid items")
		return
	}

	// var orderID int
	var res *types.MidtransPaymentResponse
	var err error

	if userInput.FlashSaleId != nil {
		if *userInput.FlashSaleId <= 0 {
			helper.ErrorResponse(c, http.StatusBadRequest, errors.New("invalid flash sale ID"), "Invalid flash sale")
			return
		}

		flashSaleInput := types.UserInputFlashSaleCheckoutType{
			FlashSaleId: *userInput.FlashSaleId,
			Item:        userInput.Item,
		}

		res, err = cc.checkoutService.PurchaseFromFlashSale(c, userIDint, flashSaleInput, userInput.UserDetail)
	} else {
		originalInput := types.UserOrderType{
			PaymentType: userInput.PaymentType,
			Item:        userInput.Item,
		}

		res, err = cc.checkoutService.PurchaseFromOriginalSale(c, userIDint, originalInput, userInput.UserDetail)
	}

	if err != nil {
		cc.log.Error(err.Error())
		helper.ErrorResponse(c, http.StatusBadRequest, err, "Failed to create order")
		return
	}

	helper.SuccessResponse(c, http.StatusCreated, *res, "Order created successfully")
}
