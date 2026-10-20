package repository

import (
	"database/sql"
	"flash_drop/types"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type CheckoutRepository struct {
	db *sql.DB
}

func InitCheckoutRepostory(db *sql.DB) *CheckoutRepository {
	return &CheckoutRepository{
		db: db,
	}
}

func (cr *CheckoutRepository) createOrderNumber(userId int) string {
	return fmt.Sprintf("ODR-%d-%d", userId, time.Now().Unix())
}

func (cr *CheckoutRepository) createPaymentReference(userId int) string {
	return fmt.Sprintf("PYMNT-%d-%d", userId, time.Now().Unix())
}

func (cr *CheckoutRepository) CheckoutItem(
	c *gin.Context,
	userId int,
	cartItems []types.UserCartType,
	userInput types.UserInputShippingType,
	paymentMethod types.PaymentMethodType,
) (sql.Result, error) {
	tx, err := cr.db.BeginTx(c, nil)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	var subTotal int
	orderNumber := cr.createOrderNumber(userId)

	for _, item := range cartItems {
		subTotal += item.Price * (item.Quantity)
	}
	totalAmount := subTotal + userInput.ShippingCost

	// Query into 'orders' table
	queryOrder := "INSERT INTO orders (user_id, order_number, subtotal, shipping_cost, total_amount, shipping_name, shipping_phone, shipping_address) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
	res, err := tx.ExecContext(c, queryOrder, userId, orderNumber, subTotal, userInput.ShippingCost, totalAmount, userInput.ShippingName, userInput.Phone, userInput.ShippingAddress)
	if err != nil {
		return nil, err
	}

	orderId, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Query into 'order_items' table
	valString := make([]string, 0, len(cartItems))
	valArgs := make([]any, 0, len(cartItems)*6)
	for _, item := range cartItems {
		valString = append(valString, "(?, ?, ?, ?, ?, ?)")
		valArgs = append(valArgs, orderId, item.ProuctId, item.Name, item.Quantity, item.Price, item.Price*(item.Quantity))
	}

	queryOrderItem := fmt.Sprintf("INSERT INTO order_items (order_id, product_id, product_name, quantity, unit_price, subtotal) VALUES %s", strings.Join(valString, ", "))
	res, err = tx.ExecContext(c, queryOrderItem, valArgs...)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	// Query into 'payments' table
	trReference := cr.createPaymentReference(userId)
	query := `INSERT INTO payments (order_id, payment_method, amount, transaction_reference) VALUES (?, ?, ?, ?)`
	res, err = tx.ExecContext(c, query, orderId, paymentMethod, totalAmount, trReference)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return res, nil
}
