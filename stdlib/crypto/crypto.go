// Package crypto provides hashing, encoding, and random generator functions for NilLang.
package crypto

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// Sha256 computes the SHA-256 hash of a string, returning hexadecimal representation.
func Sha256(data string) string {
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:])
}

// Md5 computes the MD5 hash of a string.
func Md5(data string) string {
	h := md5.Sum([]byte(data))
	return hex.EncodeToString(h[:])
}

// Base64Encode encodes a string to Base64.
func Base64Encode(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}

// Base64Decode decodes a Base64 string.
func Base64Decode(encoded string) (string, error) {
	bytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("crypto.base64Decode error: %w", err)
	}
	return string(bytes), nil
}

// UUID generates a random RFC 4122 version 4 UUID string.
func UUID() string {
	var u [16]byte
	_, _ = rand.Read(u[:])
	u[6] = (u[6] & 0x0f) | 0x40 // Version 4
	u[8] = (u[8] & 0x3f) | 0x80 // Variant 10
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		u[0:4], u[4:6], u[6:8], u[8:10], u[10:16])
}

// RandomHex generates a random hexadecimal string of the specified byte length.
func RandomHex(length int) string {
	b := make([]byte, length)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
