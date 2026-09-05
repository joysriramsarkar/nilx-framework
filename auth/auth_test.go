package auth

import (
	"testing"
)

func TestAuth(t *testing.T) {
	secret := "secret-jwt-key"
	token, err := SignJWT(Claims{Subject: "user_123", Roles: []string{"admin"}}, secret)
	if err != nil {
		t.Fatalf("failed to sign JWT: %v", err)
	}

	claims, err := VerifyJWT(token, secret)
	if err != nil || claims.Subject != "user_123" {
		t.Errorf("verify JWT failed: claims=%v, err=%v", claims, err)
	}

	// Password
	pwd := "MyPass123"
	h := HashPassword(pwd)
	if !CheckPassword(pwd, h) {
		t.Errorf("password verification failed")
	}
	if CheckPassword("Wrong", h) {
		t.Errorf("wrong password passed verification")
	}
}
