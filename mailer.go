package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
)

type EmailRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /email", sendEmail)

	log.Println("email server running on port 8081")
	http.ListenAndServe(":8081", mux)
}

type Config struct {
	SMTPHost  string
	SMTPPort  int
	FromEmail string
	Password  string
}

// export MAILER_SMTP_HOST=smtp.gmail.com
// export MAILER_FROM_EMAIL=spaciery@gmail.com
// export MAILER_EMAIL_PASSWORD="dhaskdjfhwerj"

func loadConfig() Config {

	return Config{
		SMTPHost:  os.Getenv("MAILER_SMTP_HOST"),
		SMTPPort:  587,
		FromEmail: os.Getenv("MAILER_FROM_EMAIL"),
		Password:  os.Getenv("MAILER_EMAIL_PASSWORD"),
	}

}

func sendEmail(w http.ResponseWriter, r *http.Request) {

	var req EmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	config := loadConfig()

	log.Printf("from %s", config.FromEmail)
	log.Printf("host %s", config.SMTPHost)

	auth := smtp.PlainAuth("", config.FromEmail, config.Password, config.SMTPHost)

	to := []string{req.To}

	htmlBody := fmt.Sprintf(`
		<html>
		<head>
			<style>
				body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
				.container {color: black}
				.footer { margin-top: 20px; border-top: 1px solid #ddd; padding-top: 10px; font-size: 12px; color: #777; }
				.footer img { width: 50px; height: auto; vertical-align: middle; margin-left: 10px; }
			</style>
		</head>
		<body>
			<div class="container">
				<p>%s</p>
				<div class="footer">
					<b>Spaciery - Sustainable innovations</b>
					<img src="https://avatars.githubusercontent.com/u/179962257?v=4" alt="Spaciery logo">
				</div>
			</div>
		</body>
		</html>
	`, req.Body)

	boundary := "nhb9mhbvuygvygvtfvtiugiugiug"
	mimeHeaders := "MIME-version: 1.0;\nContent-Type: multipart/alternative; boundary=" + boundary + "\n\n"
	body := strings.Join([]string{
		"--" + boundary,
		"Content-Type: text/plain; charset=\"UTF-8\"",
		"",
		req.Body,
		"--" + boundary,
		"Content-Type: text/html; charset=\"UTF-8\"",
		"",
		htmlBody,
		"--" + boundary + "--",
	}, "\r\n")

	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"From: %s\r\n"+
		"Subject: %s\r\n"+
		"%s\r\n"+
		"%s\r\n", req.To, config.FromEmail, req.Subject, mimeHeaders, body))

	err := smtp.SendMail(fmt.Sprintf("%s:%d", config.SMTPHost, config.SMTPPort), auth, config.FromEmail, to, msg)
	if err != nil {
		log.Printf("Error sending email: %v", err)
		http.Error(w, "Failed to send email", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Email sent successfully"))
}
