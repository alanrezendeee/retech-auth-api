package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"time"
)

// ValidateHMAC valida a assinatura HMAC do body
func ValidateHMAC(body []byte, timestamp int64, signature, secret string) error {
	if secret == "" {
		return fmt.Errorf("secret não configurado")
	}

	// Validar timestamp (evitar replay attacks)
	now := time.Now().Unix()
	maxAge := int64(300) // 5 minutos
	if timestamp < now-maxAge || timestamp > now+60 {
		return fmt.Errorf("timestamp inválido ou muito antigo")
	}

	// Calcular HMAC esperado: HMAC-SHA256(body + timestamp)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	mac.Write([]byte(fmt.Sprintf("%d", timestamp)))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	// Comparar assinaturas de forma segura (constant-time)
	if !hmac.Equal([]byte(expectedSignature), []byte(signature)) {
		return fmt.Errorf("assinatura HMAC inválida")
	}

	return nil
}

// CalculateHMAC calcula a assinatura HMAC (usado no CLI)
func CalculateHMAC(body []byte, timestamp int64, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	mac.Write([]byte(fmt.Sprintf("%d", timestamp)))
	return hex.EncodeToString(mac.Sum(nil))
}

// ReadBody lê o body sem consumir (para usar no middleware)
func ReadBody(reader io.ReadCloser) ([]byte, error) {
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return body, nil
}

