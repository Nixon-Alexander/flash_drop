package controller

import (
	"crypto/sha512"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	service "flash_drop/Service"
	"flash_drop/types"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PaymentController struct {
	log      *zap.Logger
	checkout *service.CheckoutService
}

type midtransNotification struct {
	OrderID           string `json:"order_id"`
	StatusCode        string `json:"status_code"`
	GrossAmount       string `json:"gross_amount"`
	SignatureKey      string `json:"signature_key"`
	TransactionStatus string `json:"transaction_status"`
	FraudStatus       string `json:"fraud_status"`
}

func InitPaymentController(log *zap.Logger, checkout *service.CheckoutService) *PaymentController {
	return &PaymentController{
		log:      log,
		checkout: checkout,
	}
}

func validSignature(n midtransNotification) bool {
	sum := sha512.Sum512([]byte(n.OrderID + n.StatusCode + n.GrossAmount + os.Getenv("MIDTRANS_KEY")))
	expected := hex.EncodeToString(sum[:])
	return subtle.ConstantTimeCompare([]byte(expected), []byte(n.SignatureKey)) == 1
}

func (pc *PaymentController) Notification(c *gin.Context) {
	var n midtransNotification
	if err := c.ShouldBindJSON(&n); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if !validSignature(n) {
		c.Status(http.StatusUnauthorized)
		return
	}

	ctx := c.Request.Context()
	var err error
	switch n.TransactionStatus {
	case "settlement":
		err = pc.checkout.MarkPaid(ctx, n.OrderID, n.GrossAmount)
	case "capture": // kartu kredit
		if n.FraudStatus == "accept" {
			err = pc.checkout.MarkPaid(ctx, n.OrderID, n.GrossAmount)
		}
	case "expire", "cancel", "deny":
		err = pc.checkout.CancelOrder(ctx, n.OrderID)
	}

	switch {
	case err == nil:
	case errors.Is(err, types.ErrOrderNotFound),
		errors.Is(err, types.ErrAmountMismatch),
		errors.Is(err, types.ErrPaidAfterCancel):
		pc.log.Error("payment notification needs manual review", zap.String("order", n.OrderID), zap.Error(err))
	default:
		pc.log.Error("process notification", zap.String("order", n.OrderID), zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}
