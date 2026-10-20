package types

import "time"

type FlashSaleStatus string

const (
	SCHEDULED FlashSaleStatus = "scheduled"
	ACTIVE    FlashSaleStatus = "active"
	ENDED     FlashSaleStatus = "ended"
	CANCELLED FlashSaleStatus = "cancelled"
)

type FlashSalesType struct {
	Id          int             `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	StartTime   time.Time       `json:"start_time"`
	EndTime     time.Time       `json:"end_time"`
	Status      FlashSaleStatus `json:"status"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type FlashSaleItemsType struct {
	Id           int       `json:"id"`
	FlashSaleId  int       `json:"flash_sale_id"`
	ProductId    int       `json:"product_id"`
	FlashPrice   float32   `json:"flash_price"`
	Stock        int       `json:"stock"`
	SoldQuantity int       `json:"sold_quantity"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type FlashSalesItems struct {
	ProductId  int     `json:"product_id"`
	FlashPrice float32 `json:"flash_price"`
	Stock      int     `json:"stock"`
}

type UserInputFlashSalesType struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	StartTime   string            `json:"start_time"`
	EndTime     string            `json:"end_time"`
	Items       []FlashSalesItems `json:"items"`
}

type FlashSaleWithItems struct {
	FlashSalesType
	Items []FlashSalesItems `json:"items"`
}

type UserInputFlashSaleCheckoutType struct {
	PaymentType PaymentMethodType   `json:"payment_type"`
	FlashSaleId int                 `json:"flash_sale_id"`
	Item        []UserInputItemType `json:"item"`
}
