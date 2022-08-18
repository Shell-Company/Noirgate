package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/xlzd/gotp"
)

func ValidateGuestTOTP(UserSecret string, PassCode string) bool {
	UserSecret = strings.ToUpper(UserSecret)
	UserMFA := gotp.NewDefaultTOTP(UserSecret)
	validationTime := time.Now()
	validateResult := UserMFA.Verify(PassCode, int(validationTime.Unix()))
	return validateResult
}

func main() {
	OTPSecret, _ := os.LookupEnv("OTP")
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Enter noirgate OTP:")
	fmt.Println("---------------------")

	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)

		validOTP := ValidateGuestTOTP(OTPSecret, text)
		if validOTP {
			log.Println("Access Granted")
			return
		} else {
			log.Println("Access Denied")
		}

	}

}
