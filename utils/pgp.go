package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"os"
)

// EncryptPGP шифрует строку с помощью AES-256 GCM и возвращает base64
func EncryptPGP(plainText string) (string, error) {
	key := []byte(os.Getenv("HMAC_SECRET"))
	if len(key) < 32 {
		return "", errors.New("HMAC_SECRET must be at least 32 bytes")
	}
	key = key[:32] // AES-256

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	cipherText := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// DecryptPGP расшифровывает строку из base64 (AES-256 GCM)
func DecryptPGP(cipherBase64 string) (string, error) {
	key := []byte(os.Getenv("HMAC_SECRET"))
	if len(key) < 32 {
		return "", errors.New("HMAC_SECRET must be at least 32 bytes")
	}
	key = key[:32]

	cipherText, err := base64.StdEncoding.DecodeString(cipherBase64)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	if len(cipherText) < gcm.NonceSize() {
		return "", errors.New("ciphertext too short")
	}

	nonce := cipherText[:gcm.NonceSize()]
	data := cipherText[gcm.NonceSize():]

	plainText, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}

func MaskCardNumber(enc string) string {
	plain, err := DecryptPGP(enc)
	if err != nil || len(plain) < 4 {
		return "****"
	}
	last4 := plain[len(plain)-4:]
	return "**** **** **** " + last4
}
