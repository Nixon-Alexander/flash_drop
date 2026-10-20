package repository

import (
	"database/sql"
	"flash_drop/types"

	"github.com/gin-gonic/gin"
)

type PaymentRepository struct {
	db *sql.DB
}

func InitPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{
		db: db,
	}
}

func (pr *PaymentRepository) CreatePayment(c *gin.Context, tx *sql.Tx, userId int, totalAmount int64, midtransPayload types.MidtransPaymentRequestType, trReference string) (sql.Result, error) {
	query := `INSERT INTO payments (order_id, payment_method, amount, transaction_reference) VALUES (?, ?, ?, ?)`
	res, err := tx.ExecContext(c, query, midtransPayload.TransactionDetails.OrderId, midtransPayload.PaymentType, totalAmount, trReference)
	if err != nil {
		return nil, err
	}

	return res, nil
}
