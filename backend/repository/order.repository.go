package repository

import (
	"database/sql"
	"flash_drop/types"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type OrderRepository struct {
	db *sql.DB
}

func InitOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{
		db: db,
	}
}

func (or *OrderRepository) createOrderNumber(userId int) string {
	return fmt.Sprintf("ODR-%d-%d", userId, time.Now().Unix())
}

func (or *OrderRepository) CreateOrder(
	ctx *gin.Context,
	tx *sql.Tx,
	userID int,
	items []types.UserInputItemType,
	product []types.ProductDetailForCalculationType,
	shippingDetail types.UserInputShippingType,
) (string, int, error) {
	var subTotal int
	for _, item := range items {
		index := slices.IndexFunc(product, func(p types.ProductDetailForCalculationType) bool {
			return p.ProductId == *item.ProductId
		})
		subTotal = subTotal + int(product[index].Price*1000)
	}

	totalAmount := subTotal + shippingDetail.ShippingCost

	orderNumber := or.createOrderNumber(userID)

	queryOrder := `
        INSERT INTO orders (
            user_id,
            order_number,
            subtotal,
            shipping_cost,
            total_amount,
            shipping_name,
            shipping_phone,
            shipping_address
        )
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `

	result, err := tx.ExecContext(
		ctx,
		queryOrder,
		userID,
		orderNumber,
		subTotal,
		shippingDetail.ShippingCost,
		totalAmount,
		shippingDetail.ShippingName,
		shippingDetail.Phone,
		shippingDetail.ShippingAddress,
	)
	if err != nil {
		return "", 0, err
	}

	orderID, err := result.LastInsertId()
	orderIDint := int(orderID)
	if err != nil {
		return "", 0, err
	}

	// Insert into 'order_items'
	valString := make([]string, 0, len(items))
	valArgs := make([]any, 0, len(items)*6)
	for _, item := range items {
		index := slices.IndexFunc(product, func(p types.ProductDetailForCalculationType) bool {
			return p.ProductId == *item.ProductId
		})
		p := product[index]

		valString = append(valString, "(?, ?, ?, ?, ?, ?)")
		valArgs = append(valArgs, orderIDint, item.ProductId, p.Name, item.Quantity, p.Price, p.Price*float32(*item.Quantity))
	}

	queryOrderItem := fmt.Sprintf("INSERT INTO order_items (order_id, product_id, product_name, quantity, unit_price, subtotal) VALUES %s", strings.Join(valString, ", "))
	_, err = tx.ExecContext(ctx, queryOrderItem, valArgs...)
	if err != nil {
		return "", 0, err
	}
	if err != nil {
		return "", 0, err
	}

	return orderNumber, totalAmount, nil
}

func (or *OrderRepository) UpdateStatus(c *gin.Context, tx *sql.Tx, status types.OrderStatus, orderNumber string) (sql.Result, error) {
	query := `UPDATE orders SET status = ? WHERE order_number = ?`
	res, err := tx.ExecContext(c, query, status, orderNumber)
	if err != nil {
		return nil, err
	}

	return res, nil
}
