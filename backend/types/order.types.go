package types

import "time"

type OrderStatus string

type OrderCategory string

const (
	ORDER_TYPE_ORIGINAL   OrderCategory = "original"
	ORDER_TYPE_FLASH_SALE OrderCategory = "flash_sale"
)

const (
	PENDING_ORDER    OrderStatus = "pending"
	PAID_ORDER       OrderStatus = "paid"
	PROCESSING_ORDER OrderStatus = "processing"
	SHIPPED_ORDER    OrderStatus = "shipped"
	COMPLETED_ORDER  OrderStatus = "completed"
	CANCELLED_ORDER  OrderStatus = "cancelled"
)

type OrderType struct {
	Id              int         `json:"id"`
	UserId          int         `json:"user_id"`
	OrderNumber     string      `json:"order_number"`
	Subtotal        float32     `json:"subtotal"`
	ShippingCost    int         `json:"shipping_cost"`
	TotalAmount     float32     `json:"total_amount"`
	Status          OrderStatus `json:"status"`
	ShippingName    string      `json:"shipping_name"`
	ShippingPhone   string      `json:"shipping_phone"`
	ShippingAddress string      `json:"shipping_address"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

type OrderItemsType struct {
	Id          int       `json:"id"`
	OrderId     int       `json:"order_id"`
	ProductId   int       `json:"product_id"`
	ProductName string    `json:"product_name"`
	Quantity    int       `json:"quantity"`
	UnitPrice   float32   `json:"unit_price"`
	Subtotal    float32   `json:"subtotal"`
	CreatedAt   time.Time `json:"created_at"`
}

type UserOrderType struct {
	PaymentType PaymentMethodType   `json:"payment_type"`
	Item        []UserInputItemType `json:"item"`
}

type OrderLine struct {
	ProductID int
	Quantity  int
	Name      string
	Category  string
	UnitPrice int64
}

type PendingOrder struct {
	Number       string
	FlashSaleID  *int
	Lines        []OrderLine
	Subtotal     int64
	ShippingCost int64
	Total        int64
	ExpiresAt    time.Time
}

type UserInputItemType struct {
	ProductId *int `json:"product_id" binding:"required,gt=0"`
	Quantity  *int `json:"quantity" binding:"required,gt=0,lte=10"`
}

type UserInputBuyItemsType struct {
	UserDetail  UserInputShippingType `json:"shipping_detail"`
	PaymentType PaymentMethodType     `json:"payment_type" binding:"required"`
	FlashSaleId *int                  `json:"flash_sale_id"`
	Item        []UserInputItemType   `json:"items" binding:"required,min=1,max=20,dive"`
}
