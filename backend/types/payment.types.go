package types

import "time"

type PaymentMethodType string

const (
	GOPAY_EWALLET     PaymentMethodType = "gopay"
	OVO_EWALLET       PaymentMethodType = "ovo"
	SHOPEEPAY_EWALLET PaymentMethodType = "shopeepay"
	PERMATA_VA        PaymentMethodType = "permata_va"
	BCA_VA            PaymentMethodType = "bca_va"
	BNI_VA            PaymentMethodType = "bni_va"
	BRI_VA            PaymentMethodType = "bri_va"
	CIMB_VA           PaymentMethodType = "cimb_va"
	DANAMON_VA        PaymentMethodType = "danamon_va"
	BSI_VA            PaymentMethodType = "bsi_va"
	SEABANK_VA        PaymentMethodType = "seabank_va"
	SAQU_VA           PaymentMethodType = "saqu_va"
)

type PaymentStatus string

const (
	PENDING_PAYMENT PaymentStatus = "pending"
	PAID_PAYMENT    PaymentStatus = "paid"
	FAILED_PAYMENT  PaymentStatus = "failed"
	EXPIRED_PAYMENT PaymentStatus = "expired"
)

type PaymentType struct {
	Id                   int               `json:"id"`
	OrderId              int               `json:"order_id"`
	PaymentMethod        PaymentMethodType `json:"payment_method"`
	Amount               float32           `json:"amount"`
	Status               PaymentStatus     `json:"status"`
	TransactionReference string            `json:"transaction_reference"`
	PaidAt               time.Time         `json:"paid_at"`
	CreatedAt            time.Time         `json:"created_at"`
	UpdatedAt            time.Time         `json:"updated_at"`
}
