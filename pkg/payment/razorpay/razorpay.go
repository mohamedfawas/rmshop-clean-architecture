package razorpay

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"github.com/razorpay/razorpay-go"
)

// Service struct encapsulates the Razorpay client and secret key for payment operations.
type Service struct {
	client    *razorpay.Client // Razorpay client instance for making API calls
	secretKey string           // Secret key for verifying payment signatures
	keyID     string
}

// Order struct represents a payment order with relevant details.
type Order struct {
	ID       string `json:"id"` // Unique identifier for the order
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

// creates a Razorpay client that will be used for subsequent API calls.
func NewService(keyID, keySecret string) *Service {
	client := razorpay.NewClient(keyID, keySecret)
	return &Service{
		client:    client,
		secretKey: keySecret,
		keyID:     keyID,
	}
}

func (s *Service) CreateOrder(amount int64, currency string) (*Order, error) {
	data := map[string]interface{}{
		"amount":   amount,
		"currency": currency,
	}
	body, err := s.client.Order.Create(data, nil)
	if err != nil {
		return nil, err
	}

	// Since body is already a map[string]interface{}, we can directly access its fields
	id, ok := body["id"].(string)
	if !ok {
		return nil, errors.New("invalid order id type")
	}

	// Amount comes as float64 from the API, so we need to convert it to int64
	amountFloat, ok := body["amount"].(float64)
	if !ok {
		return nil, errors.New("invalid amount type")
	}
	amount = int64(amountFloat)

	currency, ok = body["currency"].(string)
	if !ok {
		return nil, errors.New("invalid currency type")
	}

	order := &Order{
		ID:       id,
		Amount:   amount,
		Currency: currency,
	}

	return order, nil
}

func (s *Service) VerifyPaymentSignature(attributes map[string]interface{}) error {
	orderId, ok := attributes["razorpay_order_id"].(string)
	if !ok {
		return errors.New("invalid order id")
	}

	paymentId, ok := attributes["razorpay_payment_id"].(string)
	if !ok {
		return errors.New("invalid payment id")
	}

	signature, ok := attributes["razorpay_signature"].(string)
	if !ok {
		return errors.New("invalid signature")
	}

	// Create the message
	message := orderId + "|" + paymentId

	// Create a new HMAC by defining the hash type and the key (as byte array)
	h := hmac.New(sha256.New, []byte(s.secretKey))

	// Write Data to it
	h.Write([]byte(message))

	// Get result and encode as hexadecimal string
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	if signature != expectedSignature {
		return errors.New("invalid signature")
	}

	return nil
}

func (s *Service) GetKeyID() string {
	return s.keyID
}
