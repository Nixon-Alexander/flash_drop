package types

type MidtransPaymentRequestType struct {
	PaymentType        PaymentMethodType               `json:"payment_type"`
	TransactionDetails UserInputTransactionDetailsType `json:"transaction_details"`
	ItemDetails        []UserInputMidtransItemDetails  `json:"item_details"`
	ShippingAddress    UserInputShippingType           `json:"shipping_address"`
	PaymentOptions     map[string]any                  `json:"-"`
}

type UserInputMidtransItemDetails struct {
	Id       int    `json:"id"`
	Price    int    `json:"price"`
	Quantity int    `json:"quantity"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

type UserInputShippingType struct {
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	City            string `json:"city"`
	PostalCode      string `json:"postal_code"`
	CountryCode     string `json:"country_code"`
	ShippingCost    int    `json:"shipping_cost"`
	ShippingName    string `json:"shipping_name"`
	ShippingAddress string `json:"address"`
}

type UserInputCustomerDetailType struct {
	FirstName        string                `json:"first_name"`
	LastName         string                `json:"last_name"`
	Email            string                `json:"email"`
	Phone            string                `json:"phone"`
	ShippingAddresss UserInputShippingType `json:"shipping_address"`
}

type UserInputTransactionDetailsType struct {
	OrderId     string `json:"order_id"`
	GrossAmount int64  `json:"gross_amount"`
}

type MidtransPaymentResponse struct {
	Token       string `json:"token"`
	RedirectUrl string `json:"redirect_url"`
}
