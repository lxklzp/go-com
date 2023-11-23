package internet

import (
	"crypto/tls"
	"github.com/jordan-wright/email"
	"go-com/global"
	"net/smtp"
)

type emailCfg struct {
	FromEmail string
	ToEmail   []string
	SmtpHost  string
	SmtpPort  string
	SmtpPass  string
}

// Email 发送邮件
func Email() {
	// 发送邮件
	var err error
	cfg := emailCfg{}
	e := email.NewEmail()
	e.From = cfg.FromEmail
	e.To = cfg.ToEmail
	e.Subject = ""
	e.Text = []byte("")
	_, err = e.AttachFile("")
	if err != nil {
		global.Log.Panic(err)
	}
	err = e.SendWithTLS(cfg.SmtpHost+":"+cfg.SmtpPort, smtp.PlainAuth("", cfg.FromEmail, cfg.SmtpPass, cfg.SmtpHost), &tls.Config{ServerName: cfg.SmtpHost})
	if err != nil {
		global.Log.Panic(err)
	}
}
