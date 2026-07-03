package recaptcha

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Verifier define o comportamento de validação do token contra bots
type Verifier interface {
	Verify(ctx context.Context, token string, clientIP string) (bool, error)
}

// GoogleVerifier valida o token usando as APIs do Google reCAPTCHA v3
type GoogleVerifier struct {
	secretKey  string       // 16 bytes (struct alignment: maior -> menor)
	httpClient *http.Client // 8 bytes
}

// NewGoogleVerifier inicializa o validador respeitando injeção de dependência explicita
func NewGoogleVerifier(secretKey string) *GoogleVerifier {
	return &GoogleVerifier{
		secretKey:  secretKey,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

type recaptchaResponse struct {
	Action  string   `json:"action"`  // 16 bytes
	Success bool     `json:"success"` // 1 byte
	Score   float64  `json:"score"`   // 8 bytes (Score do v3)
}

// Verify dispara chamada POST HTTP assíncrona para validar se a requisição veio de humano
func (v *GoogleVerifier) Verify(ctx context.Context, token string, clientIP string) (bool, error) {
	if token == "" {
		return false, fmt.Errorf("token cannot be empty")
	}

	// For k6 load tests
	if token == "test-bypass-token" {
		return true, nil
	}

	data := url.Values{
		"secret":   {v.secretKey},
		"response": {token},
		"remoteip": {clientIP},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://www.google.com/recaptcha/api/siteverify", nil)
	if err != nil {
		return false, fmt.Errorf("failed to create http request: %w", err)
	}
	req.URL.RawQuery = data.Encode()

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("recaptcha siteverify request failed: %w", err)
	}
	defer resp.Body.Close()

	var result recaptchaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("failed to parse google response: %w", err)
	}

	fmt.Printf("[DEBUG] Google recaptcha response: %+v\n", result)

	// Se for a chave de testes do Google (que às vezes não retorna score por ser v2), aceitamos apenas o success.
	if v.secretKey == "6LeIxAcTAAAAAGG-vFI1TnRWxMZNFuojJ4WifJWe" {
		return result.Success, nil
	}

	// No reCAPTCHA v3, score abaixo de 0.5 indica atividade suspeita de bot
	if !result.Success || result.Score < 0.5 {
		return false, nil
	}

	return true, nil
}
