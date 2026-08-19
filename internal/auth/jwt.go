package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"
)

type Claims struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
	Issuer    string `json:"iss"`
}

func secret() ([]byte, error) {
	value := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if len(value) < 32 {
		return nil, errors.New("JWT_SECRET must be at least 32 characters")
	}
	return []byte(value), nil
}

func IssueToken(userID uint, username, role string, lifetime time.Duration) (string, error) {
	key, err := secret()
	if err != nil {
		return "", err
	}
	now := time.Now()
	claims := Claims{UserID: userID, Username: username, Role: role, IssuedAt: now.Unix(), ExpiresAt: now.Add(lifetime).Unix(), Issuer: "fintech-labs"}
	enc := base64.RawURLEncoding
	header, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	payload, _ := json.Marshal(claims)
	unsigned := enc.EncodeToString(header) + "." + enc.EncodeToString(payload)
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(unsigned))
	return unsigned + "." + enc.EncodeToString(mac.Sum(nil)), nil
}

func ParseToken(value string) (*Claims, error) {
	key, err := secret()
	if err != nil {
		return nil, err
	}
	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid JWT")
	}
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(parts[0] + "." + parts[1]))
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !hmac.Equal(sig, mac.Sum(nil)) {
		return nil, errors.New("invalid JWT signature")
	}
	data, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errors.New("invalid JWT payload")
	}
	var claims Claims
	if json.Unmarshal(data, &claims) != nil || claims.Issuer != "fintech-labs" || time.Now().Unix() >= claims.ExpiresAt {
		return nil, errors.New("invalid or expired JWT")
	}
	return &claims, nil
}
