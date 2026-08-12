package mpesa

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// --- REQUEST & RESPONSE STRUCTS ---

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   string `json:"expires_in"`
}

type STKPushRequest struct {
	BusinessShortCode string `json:"BusinessShortCode"`
	Password          string `json:"Password"`
	Timestamp         string `json:"Timestamp"`
	TransactionType   string `json:"TransactionType"`
	Amount            string `json:"Amount"`
	PartyA            string `json:"PartyA"`
	PartyB            string `json:"PartyB"`
	PhoneNumber       string `json:"PhoneNumber"`
	CallBackURL       string `json:"CallBackURL"`
	AccountReference  string `json:"AccountReference"`
	TransactionDesc   string `json:"TransactionDesc"`
}

type STKPushResponse struct {
	MerchantRequestID   string `json:"MerchantRequestID"`
	CheckoutRequestID   string `json:"CheckoutRequestID"`
	ResponseCode        string `json:"ResponseCode"`
	ResponseDescription string `json:"ResponseDescription"`
	CustomerMessage     string `json:"CustomerMessage"`
}

type CallbackMetadataItem struct {
	Name  string      `json:"Name"`
	Value interface{} `json:"Value"`
}

type CallbackMetadata struct {
	Item []CallbackMetadataItem `json:"Item"`
}

type STKCallbackBody struct {
	MerchantRequestID string           `json:"MerchantRequestID"`
	CheckoutRequestID string           `json:"CheckoutRequestID"`
	ResultCode        int              `json:"ResultCode"`
	ResultDesc        string           `json:"ResultDesc"`
	CallbackMetadata  CallbackMetadata `json:"CallbackMetadata"`
}

type STKCallback struct {
	Body struct {
		StkCallback STKCallbackBody `json:"stkCallback"`
	} `json:"Body"`
}

// --- HELPER FUNCTIONS ---

func getBaseURL() string {
	if os.Getenv("MPESA_ENV") == "production" {
		return "https://api.safaricom.co.ke"
	}
	return "https://sandbox.safaricom.co.ke"
}

// GetAccessToken — fetches a fresh OAuth token from Safaricom
func GetAccessToken() (string, error) {
	consumerKey := os.Getenv("MPESA_CONSUMER_KEY")
	consumerSecret := os.Getenv("MPESA_CONSUMER_SECRET")

	if consumerKey == "" || consumerSecret == "" {
		return "", fmt.Errorf("M-Pesa credentials not configured")
	}

	url := getBaseURL() + "/oauth/v1/generate?grant_type=client_credentials"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	// Base64 encode consumerKey:consumerSecret
	credentials := base64.StdEncoding.EncodeToString([]byte(consumerKey + ":" + consumerSecret))
	req.Header.Set("Authorization", "Basic "+credentials)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("empty access token received: %s", string(body))
	}

	return tokenResp.AccessToken, nil
}

// GeneratePassword — creates the STK Push password
// Password = Base64(ShortCode + Passkey + Timestamp)
func GeneratePassword(timestamp string) string {
	shortCode := os.Getenv("MPESA_SHORTCODE")
	passkey := os.Getenv("MPESA_PASSKEY")
	raw := shortCode + passkey + timestamp
	return base64.StdEncoding.EncodeToString([]byte(raw))
}

// GenerateTimestamp — returns current time in YYYYMMDDHHmmss format
func GenerateTimestamp() string {
	return time.Now().Format("20060102150405")
}

// InitiateSTKPush — sends STK Push request to Safaricom
func InitiateSTKPush(phoneNumber, accountNumber string, amount int64) (*STKPushResponse, error) {
	token, err := GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	shortCode := os.Getenv("MPESA_SHORTCODE")
	callbackURL := os.Getenv("MPESA_CALLBACK_URL")
	timestamp := GenerateTimestamp()
	password := GeneratePassword(timestamp)

	payload := STKPushRequest{
		BusinessShortCode: shortCode,
		Password:          password,
		Timestamp:         timestamp,
		TransactionType:   "CustomerPayBillOnline",
		Amount:            fmt.Sprintf("%d", amount),
		PartyA:            phoneNumber,
		PartyB:            shortCode,
		PhoneNumber:       phoneNumber,
		CallBackURL:       callbackURL,
		AccountReference:  accountNumber,
		TransactionDesc:   "African Vault Deposit",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := getBaseURL() + "/mpesa/stkpush/v1/processrequest"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate STK push: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var stkResp STKPushResponse
	if err := json.Unmarshal(body, &stkResp); err != nil {
		return nil, fmt.Errorf("failed to parse STK response: %w", err)
	}

	if stkResp.ResponseCode != "0" {
		return nil, fmt.Errorf("STK push failed: %s", stkResp.ResponseDescription)
	}

	return &stkResp, nil
}

// ParseCallback — extracts payment details from Safaricom callback
func ParseCallback(body []byte) (*STKCallbackBody, error) {
	var callback STKCallback
	if err := json.Unmarshal(body, &callback); err != nil {
		return nil, fmt.Errorf("failed to parse callback: %w", err)
	}
	return &callback.Body.StkCallback, nil
}

// GetCallbackValue — extracts a specific value from callback metadata
func GetCallbackValue(items []CallbackMetadataItem, name string) interface{} {
	for _, item := range items {
		if item.Name == name {
			return item.Value
		}
	}
	return nil
}