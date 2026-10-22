package repository

import (
	"context"
	"database/sql"
	"flash_drop/types"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type ProductSnapshot struct {
	Name     string
	Category string
	Price    int64
}

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

func (r *OrderRepository) LockProducts(ctx context.Context, tx *sql.Tx, lines []types.OrderLine) (map[int]ProductSnapshot, error) {
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(lines)), ",")
	args := make([]any, len(lines))
	for i, l := range lines {
		args[i] = l.ProductID
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT p.id, p.name, c.name, CAST(p.price AS SIGNED)
		FROM products p
		JOIN categories c ON c.id = p.category_id
		WHERE p.id IN (`+placeholders+`) AND p.is_active = 1
		ORDER BY p.id
		FOR UPDATE OF p`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[int]ProductSnapshot, len(lines))
	for rows.Next() {
		var id int
		var s ProductSnapshot
		if err := rows.Scan(&id, &s.Name, &s.Category, &s.Price); err != nil {
			return nil, err
		}
		out[id] = s
	}
	return out, rows.Err()
}

func (r *OrderRepository) DecreaseStock(ctx context.Context, tx *sql.Tx, productID, qty int) error {
	res, err := tx.ExecContext(ctx,
		`UPDATE products SET stock = stock - ? WHERE id = ? AND is_active = 1 AND stock >= ?`,
		qty, productID, qty)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n != 1 {
		return types.ErrInsufficientStock
	}
	return nil
}

func (r *OrderRepository) ReserveFlashItem(ctx context.Context, tx *sql.Tx, flashSaleID, productID, qty int) (int64, error) {
	res, err := tx.ExecContext(ctx, `
		UPDATE flash_sale_items fsi
		JOIN flash_sales fs ON fs.id = fsi.flash_sale_id
		SET fsi.sold_quantity = fsi.sold_quantity + ?
		WHERE fsi.flash_sale_id = ? AND fsi.product_id = ?
		  AND fsi.stock - fsi.sold_quantity >= ?
		  AND fs.status = 'active'
		  AND NOW() >= fs.start_time AND NOW() < fs.end_time`,
		qty, flashSaleID, productID, qty)
	if err != nil {
		return 0, err
	}
	if n, _ := res.RowsAffected(); n != 1 {
		return 0, types.ErrFlashSaleUnavailable
	}

	var price int64
	err = tx.QueryRowContext(ctx,
		`SELECT CAST(flash_price AS SIGNED) FROM flash_sale_items WHERE flash_sale_id = ? AND product_id = ?`,
		flashSaleID, productID).Scan(&price)
	return price, err
}

func (r *OrderRepository) CreateOrder(ctx context.Context, tx *sql.Tx, userID int, o *types.PendingOrder, ship types.UserInputShippingType) error {
	res, err := tx.ExecContext(ctx, `
		INSERT INTO orders (user_id, flash_sale_id, order_number, subtotal, shipping_cost, total_amount,
			status, expires_at, shipping_name, shipping_phone, shipping_address)
		VALUES (?, ?, ?, ?, ?, ?, 'pending', ?, ?, ?, ?)`,
		userID, o.FlashSaleID, o.Number, o.Subtotal, o.ShippingCost, o.Total,
		o.ExpiresAt, ship.ShippingName, ship.Phone, ship.ShippingAddress)
	if err != nil {
		return err
	}
	orderID, err := res.LastInsertId()
	if err != nil {
		return err
	}

	placeholders := make([]string, 0, len(o.Lines))
	args := make([]any, 0, len(o.Lines)*6)
	for _, l := range o.Lines {
		placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?)")
		args = append(args, orderID, l.ProductID, l.Name, l.Quantity, l.UnitPrice, l.UnitPrice*int64(l.Quantity))
	}
	_, err = tx.ExecContext(ctx,
		`INSERT INTO order_items (order_id, product_id, product_name, quantity, unit_price, subtotal) VALUES `+
			strings.Join(placeholders, ", "), args...)
	return err
}

func (or *OrderRepository) UpdateStatus(c *gin.Context, tx *sql.Tx, status types.OrderStatus, orderNumber string) (sql.Result, error) {
	query := `UPDATE orders SET status = ? WHERE order_number = ?`
	res, err := tx.ExecContext(c, query, status, orderNumber)
	if err != nil {
		return nil, err
	}

	return res, nil
}
