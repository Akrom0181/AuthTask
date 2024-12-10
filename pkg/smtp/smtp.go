package smtp

import (
	"task/config"
	"log"
	"net/smtp"
)

func SendMail(toEmail string, msg string) error {

	// Compose the email message
	from := "akromjonotaboyev@gmail.com"
	to := []string{toEmail}
	subject := "Register for Khorezm_Shashlik"
	message := msg

	// Create the email message
	body := "To: " + to[0] + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" + message

	auth := smtp.PlainAuth("", config.SmtpUsername, config.SmtpPassword, config.SmtpServer)

	log.Println("Connecting to SMTP server...")
	err := smtp.SendMail(config.SmtpServer+":"+config.SmtpPort, auth, from, to, []byte(body))
	if err != nil {
		log.Println("Error sending mail:", err)
		return err
	}
	log.Println("Email sent successfully")
	return nil
}
