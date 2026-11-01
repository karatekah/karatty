package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

// Resend sends verification codes as plain text over the Resend email API.
type Resend struct {
	APIKey string
	From   string
	Client *http.Client
}

func (m Resend) SendCode(ctx context.Context, to, code, purpose string) error {
	if m.APIKey == "" || m.From == "" {
		return errors.New("email provider is not configured")
	}
	subject := "Verifica tu correo"
	intro := "Usa este código para crear tu cuenta."
	if purpose == "recovery" {
		subject = "Recuperación de cuenta"
		intro = "Usa este código junto con tu código de recuperación para registrar una clave SSH nueva."
	}
	data, err := json.Marshal(map[string]any{
		"from": m.From, "to": []string{to}, "subject": subject,
		"text": intro + "\n\n" + code + "\n\nCaduca en 10 minutos. Si no lo solicitaste, ignora este correo.",
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.resend.com/emails", bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+m.APIKey)
	req.Header.Set("Content-Type", "application/json")
	client := m.Client
	if client == nil {
		client = &http.Client{Timeout: 8 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return errors.New("email delivery unavailable")
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New("email delivery unavailable")
	}
	return nil
}
