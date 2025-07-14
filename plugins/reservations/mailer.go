package reservations

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mailgun/mailgun-go/v5"
)

var (
	mgDomain = "mg.hyvinvointikeskusluxus.com"
	mgAPIKey = ""
)
func SetupMailer(){
	mgAPIKey = os.Getenv("MG_API_KEY")
	fmt.Println("apikey ",mgAPIKey)
}
func getMailgunClient() *mailgun.Client {
	mg := mailgun.NewMailgun(mgAPIKey)
	mg.SetAPIBase(mailgun.APIBaseEU) // Optional, for EU users
	return mg
}

func sendMail(subject, to, plainText, htmlContent string) error {
	mg := getMailgunClient()
	sender := fmt.Sprintf("hyvinvointikeskusluxus <noreply@%s>",mgDomain)
	message := mailgun.NewMessage(mgDomain, sender, subject, plainText, to)
	message.SetHTML(htmlContent)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_,  err := mg.Send(ctx, message)
	return err
}
