package email

import (
	"crypto/tls"
	"fmt"

	gomail "gopkg.in/gomail.v2"
)

// Mailer defines the contract for sending emails.
type Mailer interface {
	SendOTP(to, name, code string) error
}

// GmailMailer uses Gmail SMTP with an App Password.
type GmailMailer struct {
	fromName string
	fromAddr string
	password string // Gmail App Password (16 chars)
	host     string
	port     int
}

// NewGmailMailer creates a GmailMailer.
// host is smtp.gmail.com, port is 587.
func NewGmailMailer(fromName, fromAddr, appPassword string) *GmailMailer {
	return &GmailMailer{
		fromName: fromName,
		fromAddr: fromAddr,
		password: appPassword,
		host:     "smtp.gmail.com",
		port:     587,
	}
}

// SendOTP sends a clean, branded OTP email for password reset.
func (m *GmailMailer) SendOTP(to, name, code string) error {
	msg := gomail.NewMessage()
	msg.SetAddressHeader("From", m.fromAddr, m.fromName)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", "Your Altar Password Reset Code")
	msg.SetBody("text/html", buildOTPEmailHTML(name, code))

	dialer := gomail.NewDialer(m.host, m.port, m.fromAddr, m.password)
	// Port 587 uses STARTTLS (explicit TLS upgrade), NOT implicit SSL.
	// gomail defaults SSL=true which would attempt direct TLS on port 587 and fail.
	dialer.SSL = false
	dialer.TLSConfig = &tls.Config{ServerName: m.host}

	if err := dialer.DialAndSend(msg); err != nil {
		return fmt.Errorf("email send failed: %w", err)
	}
	return nil
}

func buildOTPEmailHTML(name, code string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
  <title>Password Reset – Altar</title>
</head>
<body style="margin:0;padding:0;background:#0A0A1A;font-family:'Segoe UI',Helvetica,Arial,sans-serif;">
  <table width="100%%" cellpadding="0" cellspacing="0" style="background:#0A0A1A;padding:40px 0;">
    <tr>
      <td align="center">
        <table width="560" cellpadding="0" cellspacing="0"
          style="background:#111126;border-radius:16px;overflow:hidden;box-shadow:0 8px 40px rgba(0,0,0,0.6);">

          <!-- Header -->
          <tr>
            <td style="background:linear-gradient(135deg,#6C63FF,#3ECFCF);padding:36px 40px;text-align:center;">
              <h1 style="margin:0;color:#fff;font-size:28px;letter-spacing:2px;font-weight:700;">Altar</h1>
              <p style="margin:6px 0 0;color:rgba(255,255,255,0.8);font-size:13px;letter-spacing:1px;">Faith · Growth · Community</p>
            </td>
          </tr>

          <!-- Body -->
          <tr>
            <td style="padding:40px;">
              <p style="margin:0 0 8px;color:#A0A0C0;font-size:14px;">Hello, <strong style="color:#E0E0FF;">%s</strong></p>
              <p style="margin:0 0 28px;color:#7070A0;font-size:14px;line-height:1.6;">
                We received a request to reset your Altar password. Use the code below — it expires in <strong style="color:#A0A0C0;">10 minutes</strong>.
              </p>

              <!-- OTP Box -->
              <div style="background:#1A1A3E;border:1px solid #6C63FF40;border-radius:12px;padding:28px;text-align:center;margin-bottom:28px;">
                <p style="margin:0 0 8px;color:#7070A0;font-size:11px;letter-spacing:3px;text-transform:uppercase;">Your Reset Code</p>
                <p style="margin:0;font-size:48px;font-weight:700;letter-spacing:10px;color:#6C63FF;font-family:monospace;">%s</p>
              </div>

              <p style="margin:0 0 6px;color:#5A5A80;font-size:13px;line-height:1.6;">
                If you didn't request this, you can safely ignore this email. Your password will not change.
              </p>
            </td>
          </tr>

          <!-- Footer -->
          <tr>
            <td style="padding:20px 40px 32px;border-top:1px solid #1E1E40;text-align:center;">
              <p style="margin:0;color:#3A3A60;font-size:12px;">
                &copy; 2025 Altar. All rights reserved.
              </p>
            </td>
          </tr>

        </table>
      </td>
    </tr>
  </table>
</body>
</html>`, name, code)
}
