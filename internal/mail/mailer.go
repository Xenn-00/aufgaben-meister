package mail

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Xenn-00/aufgaben-meister/internal/config"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog/log"
)

type Mailer interface {
	SendInvitationEmail(to string, projectName string, inviteLink string) error
}

type MailService struct {
	DomainSender string
	MailtrapUrl  string
	MailAPI      string
}

func NewMailer(cfg *config.AppConfig) Mailer {
	if cfg.APP.State == "prod" {
		return &MailService{
			DomainSender: cfg.MAILTRAP.API.MailtrapDomain,
			MailtrapUrl:  cfg.MAILTRAP.API.MailtrapURL,
			MailAPI:      cfg.MAILTRAP.API.MailtrapTokenAPI,
		}
	}
	return &MailService{
		DomainSender: cfg.MAILTRAP.Sandbox.SandboxDomain,
		MailtrapUrl:  cfg.MAILTRAP.Sandbox.SandboxURL,
		MailAPI:      cfg.MAILTRAP.Sandbox.SandboxAPI,
	}
}

func (m *MailService) SendInvitationEmail(to string, projectName string, inviteLink string) error {
	log.Info().Msg("Mailer: Send Invitation email hit.")
	url := m.MailtrapUrl
	log.Info().Str("url", url).Msg("Mailer: target URL")

	payload := map[string]any{
		"from": map[string]string{
			"email": m.DomainSender,
			"name":  "Aufgaben Meister - Project Einladung",
		},
		"to": []map[string]string{
			{
				"email": to,
			},
		},
		"subject": fmt.Sprintf("Invitation to join project %s", projectName),
		"text": fmt.Sprintf(
			"You are invited to join project \"%s\".\n\nAccept invitation:\n%s\n\nThis link expires in 7 days.",
			projectName,
			inviteLink,
		),
		"category": "Project Invitation",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Error().Err(err).Msg("Error when marshalling payload body.")
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		log.Error().Err(err).Msg("Error when send the request.")
		return err
	}

	req.Header.Set("Authorization", "Bearer "+m.MailAPI)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("Error when get response from server.")
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("mailtrap send failed: status=%d body=%s",
			resp.StatusCode,
			string(respBody))
	}

	return nil
}
