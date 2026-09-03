package email

import (
	"fmt"
	"net/smtp"
)

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

type Sender struct {
	cfg Config
}

func NewSender(cfg Config) *Sender {
	return &Sender{cfg: cfg}
}

// SendOTP отправляет красиво оформленный Email с кодом подтверждения
func (s *Sender) SendOTP(toEmail, code string) error {
	auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)

	subject := "Subject: Ваш код входа в Job Prep\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	
	body := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<style>
				body { font-family: Arial, sans-serif; background-color: #f4f4f5; padding: 20px; }
				.card { max-width: 480px; margin: 0 auto; background: #ffffff; padding: 32px; border-radius: 12px; box-shadow: 0 4px 12px rgba(0,0,0,0.05); }
				.code { font-size: 32px; font-weight: bold; letter-spacing: 6px; color: #4F46E5; margin: 24px 0; text-align: center; background: #EEF2FF; padding: 12px; border-radius: 8px; }
				.footer { font-size: 12px; color: #6B7280; text-align: center; margin-top: 24px; }
			</style>
		</head>
		<body>
			<div class="card">
				<h2 style="margin-top:0; color: #111827;">Вход в Job Prep</h2>
				<p style="color: #4B5563;">Используйте следующий код для подтверждения входа в аккаунт. Код действителен в течение 5 минут:</p>
				<div class="code">%s</div>
				<p style="color: #6B7280; font-size: 14px;">Если вы не запрашивали этот код, просто проигнорируйте данное письмо.</p>
				<div class="footer">&copy; Job Prep App. Все права защищены.</div>
			</div>
		</body>
		</html>
	`, code)

	msg := []byte(subject + mime + body)
	addr := fmt.Sprintf("%s:%s", s.cfg.Host, s.cfg.Port)

	return smtp.SendMail(addr, auth, s.cfg.From, []string{toEmail}, msg)
}