package service

import (
	"cmp"
	"context"
	"database/sql"
	"errors"
	"flash_drop/repository"
	"flash_drop/types"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	maxLines              = 20
	maxQtyPerItem         = 10
	midtransExpiryMinutes = 15
	expiryGrace           = 5 * time.Minute
)

type PaymentGateway interface {
	CreateTransaction(ctx context.Context, o *types.PendingOrder, method types.PaymentMethodType, ship types.UserInputShippingType) (*types.MidtransPaymentResponse, error)
}

type CheckoutInput struct {
	UserID      int
	FlashSaleID *int
	PaymentType types.PaymentMethodType
	Items       []types.UserInputItemType
	Shipping    types.UserInputShippingType
}

type CheckoutService struct {
	db        *sql.DB
	logger    *zap.Logger
	orderRepo *repository.OrderRepository
	payments  PaymentGateway
}

func InitCheckoutService(db *sql.DB, logger *zap.Logger, orderRepo *repository.OrderRepository, payments PaymentGateway) *CheckoutService {
	return &CheckoutService{db: db, logger: logger, orderRepo: orderRepo, payments: payments}
}

func newOrderNumber() string { return "ODR-" + uuid.NewString() }

func calculateShipping(_ types.UserInputShippingType) int64 { return 10000 }

func normalizeItems(items []types.UserInputItemType) ([]types.OrderLine, error) {
	if len(items) == 0 || len(items) > maxLines {
		return nil, types.ErrInvalidItems
	}
	merged := make(map[int]int, len(items))
	for _, it := range items {
		if it.ProductId == nil || it.Quantity == nil || *it.ProductId <= 0 || *it.Quantity <= 0 {
			return nil, types.ErrInvalidItems
		}
		merged[*it.ProductId] += *it.Quantity
		if merged[*it.ProductId] > maxQtyPerItem {
			return nil, types.ErrInvalidItems
		}
	}
	lines := make([]types.OrderLine, 0, len(merged))
	for id, qty := range merged {
		lines = append(lines, types.OrderLine{ProductID: id, Quantity: qty})
	}
	slices.SortFunc(lines, func(a, b types.OrderLine) int { return cmp.Compare(a.ProductID, b.ProductID) })
	return lines, nil
}

func (cs *CheckoutService) createPendingOrder(ctx context.Context, in CheckoutInput, lines []types.OrderLine) (*types.PendingOrder, error) {
	tx, err := cs.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	products, err := cs.orderRepo.LockProducts(ctx, tx, lines)
	if err != nil {
		return nil, err
	}

	var subtotal int64
	for i := range lines {
		l := &lines[i]
		p, ok := products[l.ProductID]
		if !ok {
			return nil, types.ErrProductUnavailable
		}
		l.Name, l.Category, l.UnitPrice = p.Name, p.Category, p.Price

		if in.FlashSaleID != nil {
			price, err := cs.orderRepo.ReserveFlashItem(ctx, tx, *in.FlashSaleID, l.ProductID, l.Quantity)
			if err != nil {
				return nil, err
			}
			l.UnitPrice = price
		}
		if err := cs.orderRepo.DecreaseStock(ctx, tx, l.ProductID, l.Quantity); err != nil {
			return nil, err
		}
		subtotal += l.UnitPrice * int64(l.Quantity)
	}

	shipping := calculateShipping(in.Shipping)
	order := &types.PendingOrder{
		Number:       newOrderNumber(),
		FlashSaleID:  in.FlashSaleID,
		Lines:        lines,
		Subtotal:     subtotal,
		ShippingCost: shipping,
		Total:        subtotal + shipping,
		ExpiresAt:    time.Now().UTC().Add(midtransExpiryMinutes*time.Minute + expiryGrace),
	}
	if err := cs.orderRepo.CreateOrder(ctx, tx, in.UserID, order, in.Shipping); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return order, nil
}

func (cs *CheckoutService) CancelOrder(ctx context.Context, orderNumber string) error {
	tx, err := cs.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var (
		orderID     int64
		status      string
		flashSaleID sql.NullInt64
	)
	err = tx.QueryRowContext(ctx,
		`SELECT id, status, flash_sale_id FROM orders WHERE order_number = ? FOR UPDATE`,
		orderNumber).Scan(&orderID, &status, &flashSaleID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if types.OrderStatus(status) != types.PENDING_ORDER {
		return nil
	}

	rows, err := tx.QueryContext(ctx, `SELECT product_id, quantity FROM order_items WHERE order_id = ? ORDER BY product_id`, orderID)
	if err != nil {
		return err
	}
	type item struct{ productID, qty int }
	var items []item
	for rows.Next() {
		var it item
		if err := rows.Scan(&it.productID, &it.qty); err != nil {
			rows.Close()
			return err
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	for _, it := range items {
		if _, err := tx.ExecContext(ctx,
			`UPDATE products SET stock = stock + ? WHERE id = ?`, it.qty, it.productID); err != nil {
			return err
		}
		if flashSaleID.Valid {
			if _, err := tx.ExecContext(ctx, `
				UPDATE flash_sale_items SET sold_quantity = sold_quantity - ?
				WHERE flash_sale_id = ? AND product_id = ? AND sold_quantity >= ?`,
				it.qty, flashSaleID.Int64, it.productID, it.qty); err != nil {
				return err
			}
		}
	}

	if _, err := tx.ExecContext(ctx, `UPDATE orders SET status = 'cancelled' WHERE id = ?`, orderID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE payments SET status = 'expired' WHERE order_id = ? AND status = 'pending'`, orderNumber); err != nil {
		return err
	}
	return tx.Commit()
}

func (cs *CheckoutService) ExpireOverdueOrders(ctx context.Context) {
	rows, err := cs.db.QueryContext(ctx, `
		SELECT order_number FROM orders
		WHERE status = 'pending' AND expires_at < NOW()
		ORDER BY expires_at LIMIT 100`)
	if err != nil {
		cs.logger.Error("query overdue orders", zap.Error(err))
		return
	}
	var numbers []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err == nil {
			numbers = append(numbers, n)
		}
	}
	rows.Close()

	for _, n := range numbers {
		fmt.Println("order cancelled")
		if err := cs.CancelOrder(ctx, n); err != nil {
			cs.logger.Error("expire order", zap.String("order", n), zap.Error(err))
		}
	}
}

func (cs *CheckoutService) StartExpiryWorker(ctx context.Context, every time.Duration) {
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			cs.ExpireOverdueOrders(ctx)
		}
	}
}

func (cs *CheckoutService) MarkPaid(ctx context.Context, orderNumber, grossAmount string) error {
	tx, err := cs.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var (
		orderID int64
		status  string
		total   int64
	)
	err = tx.QueryRowContext(ctx, `SELECT id, status, CAST(total_amount AS SIGNED) FROM orders WHERE order_number = ? FOR UPDATE`, orderNumber).Scan(&orderID, &status, &total)
	if errors.Is(err, sql.ErrNoRows) {
		return types.ErrOrderNotFound
	}
	if err != nil {
		return err
	}
	fmt.Println(orderID, status, total)
	switch types.OrderStatus(status) {
	case types.PAID_ORDER:
		return nil
	case types.CANCELLED_ORDER:
		return types.ErrPaidAfterCancel
	}

	if _, err := tx.ExecContext(ctx, `UPDATE orders SET status = 'paid' WHERE id = ?`, orderID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE payments SET status = 'paid', paid_at = NOW() WHERE order_id = ?`, orderNumber); err != nil {
		return err
	}
	return tx.Commit()
}

func (cs *CheckoutService) Checkout(ctx context.Context, in CheckoutInput) (*types.MidtransPaymentResponse, error) {
	if !isSupportedPayment(in.PaymentType) {
		return nil, types.ErrInvalidPaymentType
	}
	lines, err := normalizeItems(in.Items)
	if err != nil {
		return nil, err
	}

	order, err := cs.createPendingOrder(ctx, in, lines)
	if err != nil {
		return nil, err
	}

	res, err := cs.payments.CreateTransaction(ctx, order, in.PaymentType, in.Shipping)
	if err != nil {
		cs.logger.Error("create payment failed", zap.String("order", order.Number), zap.Error(err))
		if cerr := cs.CancelOrder(context.WithoutCancel(ctx), order.Number); cerr != nil {
			cs.logger.Error("compensation failed", zap.String("order", order.Number), zap.Error(cerr))
		}
		return nil, types.ErrPaymentGateway
	}
	return res, nil
}
