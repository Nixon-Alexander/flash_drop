package service

import (
	"database/sql"
	"errors"
	"flash_drop/repository"
	"flash_drop/types"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type CheckoutService struct {
	db             *sql.DB
	logger         *zap.Logger
	orderRepositor *repository.OrderRepository
	paymentService *PaymentService
}

func InitCheckoutService(
	db *sql.DB,
	logger *zap.Logger,
	orderRepositor *repository.OrderRepository,
	paymentService *PaymentService,
) *CheckoutService {
	return &CheckoutService{
		db:             db,
		logger:         logger,
		orderRepositor: orderRepositor,
		paymentService: paymentService,
	}
}

func (cs *CheckoutService) createOrderNumber(userId int) string {
	return fmt.Sprintf("ODR-%d-%d", userId, time.Now().Unix())
}

func (cs *CheckoutService) PurchaseFromOriginalSale(
	c *gin.Context,
	userID int,
	userInput types.UserOrderType,
	shippingDetail types.UserInputShippingType,
) (*types.MidtransPaymentResponse, error) {
	ctx := c.Request.Context()
	tx, err := cs.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var productDetail []types.ProductDetailForCalculationType

	// Query 'products' to take id, name, price, category for all item
	valString := make([]string, 0, len(userInput.Item))
	valArg := make([]any, 0, len(userInput.Item))
	for _, i := range userInput.Item {
		valString = append(valString, "?")
		valArg = append(valArg, i.ProductId)
	}
	queryItem := fmt.Sprintf(`SELECT p.id, p.name, p.price, c.name FROM products p JOIN categories c ON c.id = p.category_id WHERE p.id IN (%s) FOR UPDATE`, strings.Join(valString, ", "))
	productRows, err := tx.QueryContext(ctx, queryItem, valArg...)
	if err != nil {
		return nil, err
	}

	for productRows.Next() {
		pd := types.ProductDetailForCalculationType{}
		err := productRows.Scan(&pd.ProductId, &pd.Name, &pd.Price, &pd.Category)
		if err != nil {
			return nil, err
		}
		productDetail = append(productDetail, pd)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("flash sale is unavailable")
		}
		return nil, err
	}

	// Update all column per item
	var midtransItem []types.UserInputMidtransItemDetails
	querySale := `UPDATE products SET stock = stock - ? WHERE id = ? AND ? <= stock AND is_active = 1`

	for i, item := range userInput.Item {
		result, err := tx.ExecContext(ctx, querySale, item.Quantity, item.ProductId, item.Quantity)
		if err != nil {
			return nil, err
		}

		affected, err := result.RowsAffected()
		if err != nil {
			return nil, err
		}

		if affected != 1 {
			return nil, errors.New("product stock is insufficient or unavailable")
		}

		midtransItem = append(midtransItem, types.UserInputMidtransItemDetails{
			Id:       *item.ProductId,
			Price:    int(productDetail[i].Price * 1000),
			Quantity: *item.Quantity,
			Name:     productDetail[i].Name,
			Category: productDetail[i].Category,
		})
	}

	midtransItem = append(midtransItem, types.UserInputMidtransItemDetails{
		Id:       00000,
		Price:    10000,
		Quantity: 1,
		Name:     "Shipping Cost",
		Category: "Shipping",
	})

	// Create Order
	orderNumber, totalAmount, err := cs.orderRepositor.CreateOrder(c, tx, userID, userInput.Item, productDetail, shippingDetail)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Create Payment
	fmt.Println(totalAmount)
	midtransPayload := types.MidtransPaymentRequestType{
		PaymentType: userInput.PaymentType,
		TransactionDetails: types.UserInputTransactionDetailsType{
			OrderId:     string(orderNumber),
			GrossAmount: int64(totalAmount),
		},
		ItemDetails:     midtransItem,
		ShippingAddress: shippingDetail,
	}

	fmt.Println(midtransPayload)
	tx2, err := cs.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx2.Rollback()
	res, err := cs.paymentService.CreatePayment(c, tx2, int(totalAmount)*10000, userID, midtransPayload)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (cs *CheckoutService) PurchaseFromFlashSale(
	c *gin.Context,
	userID int,
	userInput types.UserInputFlashSaleCheckoutType,
	shippingDetail types.UserInputShippingType,
) (*types.MidtransPaymentResponse, error) {
	ctx := c.Request.Context()
	tx, err := cs.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var (
		flashPrice      float64
		dbProductID     int
		productName     string
		productCategory string
	)

	queryItem := `SELECT fsi.product_id, fsi.flash_price, p.name, p.category
        FROM flash_sale_items AS fsi
        JOIN flash_sales AS fs
            ON fs.id = fsi.flash_sale_id
        JOIN products AS p
            ON p.id = fsi.product_id
        WHERE fsi.flas_sale_id = ?
          AND fsi.product_id = ?
          AND fs.status = 'active'
          AND NOW() >= fs.start_time
          AND NOW() < fs.end_time
          AND p.is_active = 1
        FOR UPDATE
    `

	productRow := tx.QueryRowContext(ctx, queryItem, userInput.FlashSaleId, userInput.Item[0].ProductId)
	err = productRow.Scan(&dbProductID, &flashPrice, &productName)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("flash sale is unavailable")
		}
		return nil, err
	}

	// 2. Reservasi kuota flash sale secara atomik.
	queryFlashSale := `
        UPDATE flash_sale_items AS fsi
        JOIN flash_sales AS fs
            ON fs.id = fsi.flash_sale_id
        SET fsi.sold_quantity = fsi.sold_quantity + ?
        WHERE fsi.flash_sale_id = ?
          AND fsi.product_id = ?
          AND fsi.stock - fsi.sold_quantity >= ?
          AND fs.status = 'active'
          AND NOW() >= fs.start_time
          AND NOW() < fs.end_time
    `

	result, err := tx.ExecContext(ctx, queryFlashSale, userInput.Item[0].Quantity, userInput.FlashSaleId, dbProductID, userInput.Item[0].Quantity)
	if err != nil {
		return nil, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected != 1 {
		return nil, errors.New("flash sale quota is insufficient or unavailable")
	}

	// 3. Kurangi stok produk global secara atomik.
	queryProductStock := `UPDATE products SET stock = stock - ? WHERE id = ? AND is_active = 1 AND stock >= ?`
	result, err = tx.ExecContext(ctx, queryProductStock, userInput.Item[0].Quantity, dbProductID, userInput.Item[0].Quantity)
	if err != nil {
		return nil, err
	}

	affected, err = result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected != 1 {
		return nil, errors.New("product stock is insufficient")
	}

	var productDetail []types.ProductDetailForCalculationType = []types.ProductDetailForCalculationType{
		{
			Price:     float32(flashPrice),
			ProductId: dbProductID,
			Name:      productName,
		},
	}

	// Create Order
	orderID, totalAmount, err := cs.orderRepositor.CreateOrder(c, tx, userID, userInput.Item, productDetail, shippingDetail)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Create Payment
	midtransPayload := types.MidtransPaymentRequestType{
		PaymentType: userInput.PaymentType,
		TransactionDetails: types.UserInputTransactionDetailsType{
			OrderId:     string(orderID),
			GrossAmount: int64(totalAmount) * 10000,
		},
		ItemDetails: []types.UserInputMidtransItemDetails{
			{
				Id:       dbProductID,
				Price:    int(flashPrice) * 1000,
				Quantity: *userInput.Item[0].Quantity,
				Name:     productName,
				Category: productCategory,
			},
		},
		ShippingAddress: shippingDetail,
	}

	tx2, err := cs.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx2.Rollback()
	res, err := cs.paymentService.CreatePayment(c, tx2, int(totalAmount)*10000, userID, midtransPayload)
	if err != nil {
		return nil, err
	}

	if err := tx2.Commit(); err != nil {
		return nil, err
	}

	return res, nil
}
