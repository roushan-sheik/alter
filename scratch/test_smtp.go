//go:build ignore

package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load("../.env")

	fromAddr := os.Getenv("SMTP_FROM_ADDRESS")
	appPass := os.Getenv("SMTP_APP_PASSWORD")
	fromName := os.Getenv("SMTP_FROM_NAME")

	fmt.Printf("=== SMTP DIAGNOSTIC ===\n")
	fmt.Printf("From Address : %q\n", fromAddr)
	fmt.Printf("From Name    : %q\n", fromName)
	fmt.Printf("App Password : %q  (len=%d)\n", appPass, len(appPass))
	fmt.Println()

	if fromAddr == "" || appPass == "" {
		log.Fatal("SMTP_FROM_ADDRESS or SMTP_APP_PASSWORD is empty — check .env loading")
	}
	if len(appPass) != 16 {
		log.Fatalf("App Password length is %d — expected 16 (no spaces). Got: %q", len(appPass), appPass)
	}

	host := "smtp.gmail.com"
	port := "587"
	addr := host + ":" + port

	fmt.Printf("Dialing %s ...\n", addr)
	conn, err := smtp.Dial(addr)
	if err != nil {
		log.Fatalf("DIAL FAILED: %v", err)
	}
	defer conn.Close()
	fmt.Println("✅ TCP connection OK")

	tlsCfg := &tls.Config{ServerName: host}
	if err := conn.StartTLS(tlsCfg); err != nil {
		log.Fatalf("STARTTLS FAILED: %v\n\n👉 This means port 587 STARTTLS is blocked by your network/firewall.", err)
	}
	fmt.Println("✅ STARTTLS handshake OK")

	auth := smtp.PlainAuth("", fromAddr, appPass, host)
	if err := conn.Auth(auth); err != nil {
		log.Fatalf(`AUTH FAILED: %v

👉 Possible causes:
   1. App Password is wrong/revoked — generate a new one at:
      https://myaccount.google.com/apppasswords
   2. 2-Step Verification is NOT enabled on %s
      (App Passwords require 2FA — go to myaccount.google.com/security)
   3. You generated the App Password for a DIFFERENT Google account
      than the one in SMTP_FROM_ADDRESS`, err, fromAddr)
	}
	fmt.Println("✅ SMTP AUTH OK — credentials are valid!")

	// Send a test email to yourself
	to := fromAddr
	msg := []byte("To: " + to + "\r\nSubject: ZICK SMTP Test\r\n\r\nIf you see this, SMTP is working correctly.\r\n")
	if err := conn.Mail(fromAddr); err != nil {
		log.Fatalf("MAIL FROM failed: %v", err)
	}
	if err := conn.Rcpt(to); err != nil {
		log.Fatalf("RCPT TO failed: %v", err)
	}
	w, err := conn.Data()
	if err != nil {
		log.Fatalf("DATA failed: %v", err)
	}
	if _, err := w.Write(msg); err != nil {
		log.Fatalf("WRITE failed: %v", err)
	}
	if err := w.Close(); err != nil {
		log.Fatalf("CLOSE failed: %v", err)
	}
	fmt.Printf("\n✅ Test email sent to %s — check your inbox!\n", to)
}
