package service

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"flash_drop/helper"
	"flash_drop/repository"
	"flash_drop/types"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
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

func (ps *PaymentService) createVANumber() string {
	min := int64(1000000000)
	max := int64(10000000000)

	randomNum := rand.Int64N(max-min) + min
	return string(randomNum)
}

func (ps *PaymentService) bankBCA_VA(vaNumber string) map[string]any {
	return map[string]any{
		"bca_va": map[string]any{
			"va_number": vaNumber,
		},
	}
}

func (ps *PaymentService) bankPermata_VA(vaNumber string, recipientName string) map[string]any {
	return map[string]any{
		"permata_va": map[string]any{
			"va_number":      vaNumber,
			"recipient_name": recipientName,
		},
	}
}

func (ps *PaymentService) bankBNI_VA(vaNumber string) map[string]any {
	return map[string]any{
		"bni_va": map[string]any{
			"va_number": vaNumber,
		},
	}
}

func (ps *PaymentService) bankBRI_VA(vaNumber string) map[string]any {
	return map[string]any{
		"bri_va": map[string]any{
			"va_number": vaNumber,
		},
	}
}

func (ps *PaymentService) bankCIMB_VA(vaNumber string) map[string]any {
	return map[string]any{
		"cimb_va": map[string]any{
			"va_number": vaNumber,
		},
	}
}

func (ps *PaymentService) bankDanamon_VA(vaNumber string) map[string]any {
	return map[string]any{
		"danamon_va": map[string]any{
			"va_number": vaNumber,
		},
	}
}

func (ps *PaymentService) bankBSI_VA(vaNumber string) map[string]any {
	return map[string]any{
		"bsi_va": map[string]any{
			"va_number": vaNumber,
		},
	}
}

func (ps *PaymentService) bankSeabank_VA(vaNumber string) map[string]any {
	return map[string]any{
		"seabank_va": map[string]any{
			"va_number": vaNumber,
		},
	}
}

func (ps *PaymentService) bankSaqu_VA(vaNumber string) map[string]any {
	return map[string]any{
		"saqu_va": map[string]any{
			"va_number": vaNumber,
		},
	}
}

func (ps *PaymentService) gopay(vaNumber string, callbackUrl string, phoneNumber string, countryCode string) map[string]any {
	return map[string]any{
		"gopay": map[string]any{
			"enable_callback":      true,
			"callback_url":         callbackUrl,
			"tokenization":         true,
			"enforce_tokenization": true,
			"phone_number":         phoneNumber,
			"country_code":         countryCode,
		},
	}
}

func (ps *PaymentService) shopeePay(callbackUrl string) map[string]any {
	return map[string]any{
		"shopeepay": map[string]any{
			"callback_url": "http://midtrans.com/finish-url?order_id=order-123",
		},
	}
}

func (ps *PaymentService) ovo(phoneNumber string) map[string]any {
	return map[string]any{
		"ovo": map[string]any{
			"payer_phone_number": phoneNumber,
		},
	}
}

func (ps *PaymentService) CreatePayment(ctx *gin.Context, tx *sql.Tx, totalAmount int, userID int, midtransPayload types.MidtransPaymentRequestType) (*types.MidtransPaymentResponse, error) {
	midtransApi := os.Getenv("MIDTRANS_SNAP_ENDPOINT")
	midtransServerKey := os.Getenv("MIDTRANS_KEY")
	base64AuthString := helper.EncodeBase64(midtransServerKey + ":")

	// Inject payment options
	vaNumber := ps.createVANumber()
	var paymentOptions map[string]any
	switch midtransPayload.PaymentType {
	case types.OVO_EWALLET:
		fmt.Println("ovo")
		paymentOptions = ps.ovo(midtransPayload.ShippingAddress.Phone)

	case types.GOPAY_EWALLET:
		fmt.Println("gopay")
		paymentOptions = ps.gopay(vaNumber, "https://google.com", midtransPayload.ShippingAddress.Phone, midtransPayload.ShippingAddress.PostalCode)

	case types.SHOPEEPAY_EWALLET:
		fmt.Println("shopee")
		paymentOptions = ps.shopeePay("https://google.com")

	case types.PERMATA_VA:
		fmt.Println("permata")
		paymentOptions = ps.bankPermata_VA(vaNumber, midtransPayload.ShippingAddress.FirstName)

	case types.BCA_VA:
		fmt.Println("bca")
		paymentOptions = ps.bankBCA_VA(vaNumber)

	case types.BNI_VA:
		fmt.Println("bni")
		paymentOptions = ps.bankBNI_VA(vaNumber)

	case types.CIMB_VA:
		fmt.Println("cimb")
		paymentOptions = ps.bankCIMB_VA(vaNumber)

	case types.DANAMON_VA:
		fmt.Println("danamon")
		paymentOptions = ps.bankDanamon_VA(vaNumber)

	case types.BSI_VA:
		fmt.Println("bsi")
		paymentOptions = ps.bankBSI_VA(vaNumber)

	case types.SEABANK_VA:
		fmt.Println("seabank")
		paymentOptions = ps.bankSeabank_VA(vaNumber)

	case types.SAQU_VA:
		fmt.Println("saqu")
		paymentOptions = ps.bankSaqu_VA(vaNumber)
	}

	midtransPayload.PaymentOptions = paymentOptions
	// Change Struct into buffer
	var payload bytes.Buffer
	err := json.NewEncoder(&payload).Encode(midtransPayload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", midtransApi, &payload)
	if err != nil {
		return nil, err
	}

	// Add Header
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", base64AuthString))

	// Do request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	fmt.Println("Midtrans status:", resp.StatusCode)
	fmt.Println("Midtrans response:", string(body))

	// Read Response
	var paymentResponse types.MidtransPaymentResponse
	err = json.Unmarshal([]byte(body), &paymentResponse)
	if err != nil {
		return nil, err
	}

	fmt.Println(paymentResponse)
	// Create Payment
	_, err = ps.repo.CreatePayment(ctx, tx, userID, int64(totalAmount), midtransPayload, paymentResponse.Token)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &paymentResponse, nil
}
