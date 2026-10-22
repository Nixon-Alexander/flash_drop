package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flash_drop/repository"
	"flash_drop/types"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
)

type PaymentService struct {
	logger *zap.Logger
	repo   *repository.PaymentRepository
}

func InitPaymentService(logger *zap.Logger, repo *repository.PaymentRepository) *PaymentService {
	return &PaymentService{
		logger: logger,
		repo:   repo,
	}
}

var midtransClient = &http.Client{Timeout: 10 * time.Second}

var enabledChannels = map[types.PaymentMethodType]string{
	types.GOPAY_EWALLET:     "gopay",
	types.SHOPEEPAY_EWALLET: "shopeepay",
	types.PERMATA_VA:        "permata_va",
	types.BCA_VA:            "bca_va",
	types.BNI_VA:            "bni_va",
	types.CIMB_VA:           "cimb_va",
}

type snapItem struct {
	ID       string `json:"id"`
	Price    int64  `json:"price"`
	Quantity int    `json:"quantity"`
	Name     string `json:"name"`
	Category string `json:"category,omitempty"`
}

type snapAddress struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Address     string `json:"address"`
	City        string `json:"city"`
	PostalCode  string `json:"postal_code"`
	CountryCode string `json:"country_code"`
}

type snapCustomer struct {
	FirstName       string      `json:"first_name"`
	LastName        string      `json:"last_name"`
	Email           string      `json:"email"`
	Phone           string      `json:"phone"`
	ShippingAddress snapAddress `json:"shipping_address"`
}

type snapRequest struct {
	TransactionDetails struct {
		OrderID     string `json:"order_id"`
		GrossAmount int64  `json:"gross_amount"`
	} `json:"transaction_details"`
	ItemDetails     []snapItem   `json:"item_details"`
	CustomerDetails snapCustomer `json:"customer_details"`
	EnabledPayments []string     `json:"enabled_payments"`
	Expiry          struct {
		Unit     string `json:"unit"`
		Duration int    `json:"duration"`
	} `json:"expiry"`
}

func isSupportedPayment(m types.PaymentMethodType) bool {
	_, ok := enabledChannels[m]
	return ok
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

func (ps *PaymentService) CreateTransaction(ctx context.Context, o *types.PendingOrder, method types.PaymentMethodType, ship types.UserInputShippingType) (*types.MidtransPaymentResponse, error) {

	channel, ok := enabledChannels[method]
	if !ok {
		return nil, types.ErrInvalidPaymentType
	}

	items := make([]snapItem, 0, len(o.Lines)+1)
	var sum int64
	for _, l := range o.Lines {
		items = append(items, snapItem{
			ID:       strconv.Itoa(l.ProductID),
			Price:    l.UnitPrice,
			Quantity: l.Quantity,
			Name:     truncate(l.Name, 50),
			Category: truncate(l.Category, 50),
		})
		sum += l.UnitPrice * int64(l.Quantity)
	}
	items = append(items, snapItem{
		ID:       "SHIPPING",
		Price:    o.ShippingCost,
		Quantity: 1,
		Name:     "Shipping Cost",
		Category: "Shipping",
	})
	sum += o.ShippingCost
	if sum != o.Total {
		return nil, fmt.Errorf("item total %d does not match order total %d", sum, o.Total)
	}

	var req snapRequest
	req.TransactionDetails.OrderID = o.Number
	req.TransactionDetails.GrossAmount = o.Total
	req.ItemDetails = items
	req.EnabledPayments = []string{channel}
	req.Expiry.Unit = "minutes"
	req.Expiry.Duration = midtransExpiryMinutes
	req.CustomerDetails = snapCustomer{
		FirstName: ship.FirstName, LastName: ship.LastName, Email: ship.Email, Phone: ship.Phone,
		ShippingAddress: snapAddress{
			FirstName: ship.FirstName, LastName: ship.LastName, Email: ship.Email, Phone: ship.Phone,
			Address: ship.ShippingAddress, City: ship.City, PostalCode: ship.PostalCode, CountryCode: "IDN",
		},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, os.Getenv("MIDTRANS_SNAP_ENDPOINT"), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.SetBasicAuth(os.Getenv("MIDTRANS_KEY"), "")

	resp, err := midtransClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("midtrans request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		ps.logger.Error("midtrans non-2xx", zap.Int("status", resp.StatusCode),
			zap.String("order", o.Number), zap.ByteString("body", raw))
		return nil, fmt.Errorf("midtrans returned status %d", resp.StatusCode)
	}

	var out types.MidtransPaymentResponse
	if err := json.Unmarshal(raw, &out); err != nil || out.Token == "" {
		return nil, errors.New("midtrans returned no token")
	}

	if err := ps.repo.Create(ctx, o.Number, method, o.Total, out.Token); err != nil {
		return nil, fmt.Errorf("save payment: %w", err)
	}
	return &out, nil
}
