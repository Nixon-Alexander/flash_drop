package types

import "errors"

var (
	ErrInvalidItems         = errors.New("invalid items")
	ErrInvalidPaymentType   = errors.New("unsupported payment type")
	ErrProductUnavailable   = errors.New("product is unavailable")
	ErrInsufficientStock    = errors.New("product stock is insufficient")
	ErrFlashSaleUnavailable = errors.New("flash sale is unavailable or quota is insufficient")
	ErrPaymentGateway       = errors.New("payment provider error")
	ErrOrderNotFound        = errors.New("order not found")
	ErrAmountMismatch       = errors.New("payment amount mismatch")
	ErrPaidAfterCancel      = errors.New("payment received for cancelled order")
)
