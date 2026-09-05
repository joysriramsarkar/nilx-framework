package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type Claims struct {
	Subject   string   `json:"sub"`
	Roles     []string `json:"roles"`
	ExpiresAt int64    `json:"exp"`
}

func SignJWT(claims Claims, secret string) (string, error) {
	if claims.ExpiresAt == 0 {
		claims.ExpiresAt = time.Now().Add(24 * time.Hour).Unix()
	}
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	hJSON, _ := json.Marshal(header)
	cJSON, _ := json.Marshal(claims)

	payload := base64.RawURLEncoding.EncodeToString(hJSON) + "." + base64.RawURLEncoding.EncodeToString(cJSON)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return payload + "." + sig, nil
}

func VerifyJWT(token, secret string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token")
	}

	payload := parts[0] + "." + parts[1]
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, err
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return nil, errors.New("invalid signature")
	}

	cJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	var claims Claims
	if err := json.Unmarshal(cJSON, &claims); err != nil {
		return nil, err
	}

	if time.Now().Unix() > claims.ExpiresAt {
		return nil, errors.New("expired token")
	}

	return &claims, nil
}

func HashPassword(pwd string) string {
	salt := make([]byte, 16)
	_, _ = rand.Read(salt)
	mac := hmac.New(sha256.New, salt)
	mac.Write([]byte(pwd))
	return hex.EncodeToString(salt) + ":" + hex.EncodeToString(mac.Sum(nil))
}

func CheckPassword(pwd, stored string) bool {
	parts := strings.Split(stored, ":")
	if len(parts) != 2 {
		return false
	}
	salt, _ := hex.DecodeString(parts[0])
	expected, _ := hex.DecodeString(parts[1])

	mac := hmac.New(sha256.New, salt)
	mac.Write([]byte(pwd))
	return hmac.Equal(mac.Sum(nil), expected)
}
