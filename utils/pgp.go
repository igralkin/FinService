package utils

import "fmt"

// EncryptPGP – временная заглушка (эмуляция шифрования)
func EncryptPGP(plaintext string) (string, error) {
	return fmt.Sprintf("pgp{%s}", plaintext), nil
}

// DecryptPGP – временная заглушка (эмуляция расшифровки)
func DecryptPGP(ciphertext string) (string, error) {
	// Снимаем обёртку вида pgp{...}
	if len(ciphertext) >= 5 && ciphertext[:4] == "pgp{" && ciphertext[len(ciphertext)-1:] == "}" {
		return ciphertext[4 : len(ciphertext)-1], nil
	}
	return ciphertext, nil
}

func MaskCardNumber(cipher string) string {
	decrypted, err := DecryptPGP(cipher)
	if err != nil || len(decrypted) < 4 {
		return "**** **** **** ????"
	}
	return "**** **** **** " + decrypted[len(decrypted)-4:]
}
