package otp

import (
	"log"
	config "noirgate/config"
	"noirgate/sms"
	"strings"
	"time"

	"github.com/kevinburke/twilio-go"
	"github.com/xlzd/gotp"
)

func ValidateGuestTOTP(UserSecret string, PassCode string) bool {
	UserSecret = strings.ToUpper(UserSecret)
	UserMFA := gotp.NewDefaultTOTP(UserSecret)
	validationTime := time.Now()
	validateResult := UserMFA.Verify(PassCode, int(validationTime.Unix())) //.Validate(PassCode, UserSecret)
	return validateResult
}

func SendGuestTOTP(PhoneNumber twilio.PhoneNumber, UserSecret string) {

	UserSecret = strings.ToUpper(UserSecret)
	UserMFA := gotp.NewDefaultTOTP(UserSecret)
	PassCode, expireTime := UserMFA.NowWithExpiration()
	tokenExpiration := time.Unix(expireTime, 0)
	if *config.FlagVerbose {
		log.Println("OTP for ", PhoneNumber, "expires at ", tokenExpiration.String())
		// PassCode := UserMFA.Now()
	}
	sms.SendTwilioMessage(PhoneNumber, "🔑"+PassCode+" use 5 seconds after receiving")
}
