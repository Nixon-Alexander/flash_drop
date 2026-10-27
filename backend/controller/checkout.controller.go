package controller

import (
	"errors"
	service "flash_drop/Service"
	"flash_drop/helper"
	"flash_drop/types"
	"fmt"
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

func mapCheckoutError(err error) (int, string) {
	switch {
	case errors.Is(err, types.ErrInvalidItems), errors.Is(err, types.ErrInvalidPaymentType):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, types.ErrProductUnavailable),
		errors.Is(err, types.ErrInsufficientStock),
		errors.Is(err, types.ErrFlashSaleUnavailable):
		return http.StatusConflict, err.Error()
	case errors.Is(err, types.ErrPaymentGateway):
		return http.StatusBadGateway, "Payment provider is unavailable, please try again"
	default:
		return http.StatusInternalServerError, "Something went wrong"
	}
}

func (cc *CheckoutController) buy(c *gin.Context, requireFlashSale bool) {
	userID := c.GetInt("user_id")
	if userID == 0 {
		helper.ErrorResponse(c, http.StatusUnauthorized, nil, "Unauthorized")
		return
	}

	var req types.UserInputBuyItemsType
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorResponse(c, http.StatusBadRequest, nil, "Invalid request body")
		return
	}
	if requireFlashSale && (req.FlashSaleId == nil || *req.FlashSaleId <= 0) {
		helper.ErrorResponse(c, http.StatusBadRequest, nil, "flash_sale_id is required")
		return
	}
	fmt.Println(req)

	res, err := cc.checkoutService.Checkout(c.Request.Context(), service.CheckoutInput{
		UserID:      userID,
		FlashSaleID: req.FlashSaleId,
		PaymentType: req.PaymentType,
		Items:       req.Item,
		Shipping:    req.UserDetail,
	})
	if err != nil {
		status, msg := mapCheckoutError(err)
		fmt.Println(msg)
		if status == http.StatusInternalServerError {
			cc.log.Error("checkout failed", zap.Error(err))
		}
		helper.ErrorResponse(c, status, nil, msg)
		return
	}
	helper.SuccessResponse(c, http.StatusCreated, res, "Order created successfully")
}

func (cc *CheckoutController) BuyItems(c *gin.Context)     { cc.buy(c, false) }
func (cc *CheckoutController) BuyFlashSale(c *gin.Context) { cc.buy(c, true) }
