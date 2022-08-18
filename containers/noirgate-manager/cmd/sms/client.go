package sms

import (
	"context"
	"log"
	"net/url"
	"time"

	config "noirgate/config"

	twilio "github.com/kevinburke/twilio-go"
)

var (
	twilioSID   = config.TwilioSID
	twilioToken = config.TwilioToken
	client      = twilio.NewClient(twilioSID, twilioToken, nil)
)

type etcdRequest struct {
	Value string `json:"value"`
}

const (
	HelpMenu = `
🐚NoirGate Shell Company🐚
------------------
HOW         this menu
SHELL       spawn a sandbox
BYE         terminate sandbox
LOOT        create an analysis bucket
OTP         send a new noirgate otp
-------------------
⏳Sandbox as a Service
`
)

func GetTwilioMessage() (Messages []*twilio.Message) {

	// Get all messages between 10:34:00 Oct 26 and 19:25:59 Oct 27, NYC time.
	oak, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		secondsWestOfUTC := int(-1 * (8 * time.Hour).Seconds())
		oak = time.FixedZone("Shell-Company", secondsWestOfUTC)
	}
	Now := time.Now().In(oak).Add(-5 * time.Minute)
	end := time.Now().In(oak).Add(2 * time.Minute)

	Now = time.Now().Add(-5 * time.Minute)
	end = time.Now().Add(2 * time.Minute)

	start := Now
	iter := client.Messages.GetMessagesInRange(start, end, url.Values{})
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	page, err := iter.Next(ctx)
	if err == twilio.NoMoreResults {
		// if err != nil {
		// 	log.Println(err)
		// }
		return Messages
	}
	if err != nil {
		log.Println(err)
	}
	if page != nil {

		Messages = page.Messages
	}
	return Messages

}

func SendTwilioMessage(recipientNumber twilio.PhoneNumber, message string) {

	msg, err := client.Messages.SendMessage(config.TwilioNumber, string(recipientNumber), string(message), nil)
	if err != nil {
		log.Println("Error sending message ", err)
	}
	if *config.FlagVerbose {
		log.Println(msg)
	}

}
