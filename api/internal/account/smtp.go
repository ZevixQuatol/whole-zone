package account

import (
	"context"
	"fmt"
	"net/smtp"
	"net/url"
)

type SMTPMailer struct {
	addr     string
	host     string
	user     string
	password string
	from     string
	resetURL string
}

func NewSMTPMailer(host, port, user, password, from, appURL string) *SMTPMailer {
	return &SMTPMailer{
		addr: host + ":" + port, host: host, user: user, password: password,
		from: from, resetURL: appURL + "/reset-password",
	}
}

func (m *SMTPMailer) SendPasswordReset(ctx context.Context, email, token string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	link := m.resetURL + "?token=" + url.QueryEscape(token)
	message := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: HuanYu password reset\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n请使用以下链接重置密码，链接将在短时间后失效：\r\n%s\r\n",
		m.from, email, link,
	))
	var auth smtp.Auth
	if m.user != "" {
		auth = smtp.PlainAuth("", m.user, m.password, m.host)
	}
	return smtp.SendMail(m.addr, auth, m.from, []string{email}, message)
}
