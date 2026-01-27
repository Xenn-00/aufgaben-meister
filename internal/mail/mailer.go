package mail

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Xenn-00/aufgaben-meister/internal/config"
	"github.com/Xenn-00/aufgaben-meister/internal/entity"
	worker_task "github.com/Xenn-00/aufgaben-meister/internal/worker/tasks"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog/log"
)

type Mailer interface {
	SendInvitationEmail(to string, projectName string, inviteLink string) error
	SendReminderAufgabenProgress(aufgabe *entity.ReminderAufgaben) error
	SendReminderAufgabenOverdue(aufgabe *entity.ReminderAufgaben) error
	SendHandoverRequest(aufgabe *worker_task.HandoverRequestNotifyMeister, emailMeister, usernameAssignee string) error
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
			"name":  "Aufgaben Meister - Projekt Einladung",
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

func (m *MailService) SendReminderAufgabenProgress(aufgabe *entity.ReminderAufgaben) error {
	url := m.MailtrapUrl
	payload := map[string]any{
		"from": map[string]string{
			"email": m.DomainSender,
			"name":  "Aufgaben Meister - Erinnerung zum Projektfortschritt",
		},
		"to": []map[string]string{
			{
				"email": aufgabe.EmailAssignee,
			},
		},
		"subject": fmt.Sprintf("%s progress reminder for %s", aufgabe.Title, aufgabe.ProjectName),
		"text": fmt.Sprintf(`
		Hi,

		Just a quick reminder about the task you're currently responsible for.
		Project	: %s
		Task   	: %s
		Status 	: %s
		Priority: %s
		Due at	: %s
		
		This task is approaching it's due time. Please make sure to:
		- update the task progress if you're already working on it, or
		- communicate early if you are blocked or need assistance.

		Keeping task progress up to date helps the whole team stay aligned and avoids last-minute surprises.

		Good luck, and thanks for the effort you've put into this project!

		— Aufgaben Meister
		`, aufgabe.ProjectName, aufgabe.Title, aufgabe.Status, aufgabe.Priority, aufgabe.DueDate.Format("02 Jan 2006 15:04 MST")),
		"category": "Project Progress",
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

func (m *MailService) SendReminderAufgabenOverdue(aufgabe *entity.ReminderAufgaben) error {
	url := m.MailtrapUrl
	payload := map[string]any{
		"from": map[string]string{
			"email": m.DomainSender,
			"name":  "Aufgaben Meister - Überfälligkeitsbenachrichtigung",
		},
		"to": []map[string]string{
			{
				"email": aufgabe.EmailAssignee,
			},
		},
		"subject": fmt.Sprintf("⚠️ Task overdue: %s (%s)", aufgabe.Title, aufgabe.ProjectName),
		"text": fmt.Sprintf(`
		Hi,

		This is a notice regarding a task that has passed its due date and still requires attention.

		Project	: %s
		Task   	: %s
		Status 	: %s
		Priority: %s
		Due at	: %s
		
		This task is now overdue.

		Please take one of the following actions as soon as possible:
		- update the task progress if work is still ongoing,
		- mark the task as completed if it has already been finished, or
		- communicate any blockers or request a handover if you are unable to continue.

		Keeping overdue tasks unattended can impact project timelines and team coordination.
		If you need assistance, it's better to communicate early than let the task stall.

		Thank you for your cooperation.

		— Aufgaben Meister
		`, aufgabe.ProjectName, aufgabe.Title, aufgabe.Status, aufgabe.Priority, aufgabe.DueDate.Format("02 Jan 2006 15:04 MST")),
		"category": "Project Progress",
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

func (m *MailService) SendHandoverRequest(aufgabe *worker_task.HandoverRequestNotifyMeister, emailMeister, usernameAssignee string) error {
	url := m.MailtrapUrl
	payload := map[string]any{
		"from": map[string]string{
			"email": m.DomainSender,
			"name":  "Aufgaben Meister - Überfälligkeitsbenachrichtigung",
		},
		"to": []map[string]string{
			{
				"email": emailMeister,
			},
		},
		"subject": fmt.Sprintf("⚠️ Task overdue: %s (%s)", aufgabe.AufgabeTitle, aufgabe.ProjectName),
		"text": fmt.Sprintf(`
		Hi Meister,

		The assignee has requested a handover for the following task.

		Project	: %s
		Task   	: %s
		Status 	: %s
		Due at	: %s

		Requested by	: %s
		Requested at	: %s

		Please review this request and decide whether to:
		- approve the handover,
		- reassign the task, or
		- discuss further with the assignee.

		Keeping task ownership clear helps avoid stalled or overdue work.

		Thank you for your cooperation.

		— Aufgaben Meister
		`, aufgabe.ProjectName, aufgabe.AufgabeTitle, aufgabe.AufgabeStatus, aufgabe.DueDate.Format("02 Jan 2006 15:04 MST"), usernameAssignee, aufgabe.RequestedAt.Format("02 Jan 2006 15:04 MST")),
		"category": "Project Progress",
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
