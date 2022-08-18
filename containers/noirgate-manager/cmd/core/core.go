package core

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	config "noirgate/config"
	"noirgate/container"
	"noirgate/dns"
	"noirgate/loot"
	"noirgate/otp"
	"noirgate/sms"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/kevinburke/twilio-go"
	"github.com/xlzd/gotp"
)

var (
	ActiveUserGates   = make(map[string]GuestInfo)
	ProcessedMessages = make(map[string]bool)
	sigs              = make(chan os.Signal, 1)
	exitRequested     = make(chan bool, 1)
	triggerExit       = false
)

type GuestInfo struct {
	GuestType      string
	ContainerID    string
	OTPSecret      string
	OTPTime        time.Time
	PhoneNumber    twilio.PhoneNumber
	ExpireTime     time.Time
	IPAddress      string
	GateID         string
	LootBucketName string
	Expired        bool
	HasLoot        bool
	VIP            bool
}

func ManageGuests() {
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGSEGV)
	go func() {
		sig := <-sigs
		log.Println("Received", sig, "shutting down")
		exitRequested <- true
		triggerExit = true

	}()
	for {
		for k, guest := range ActiveUserGates {
			var containerExpired bool
			var IsNoirgateActive bool
			IsNoirgateActive = container.IsNoirgateActive(guest.ContainerID)

			// Has the users access expired?
			// log.Println(time.Now(), guest.ExpireTime, "status", IsNoirgateActive, containerExpired)
			if time.Now().After(guest.ExpireTime) {
				containerExpired = true
			}
			// Is the container active but past its expiration date?
			if IsNoirgateActive && containerExpired || triggerExit {
				// terminate container
				container.TerminateNoirgate(guest.ContainerID, guest.PhoneNumber)
				// clean up dns
				dns.DeleteNoirgateRecord(guest.GateID)

				// if the user has loot clean up the bucket
				if guest.HasLoot {
					loot.DeleteTemporaryBucket(guest.LootBucketName)
				}
				guest.Expired = true
				if guest.GuestType != "API" {
					sms.SendTwilioMessage(twilio.PhoneNumber(k), "⏰ Session Expired: Shell has been terminated")
				}

				delete(ActiveUserGates, k)

			}

		}

		if triggerExit && (len(ActiveUserGates) == 0) {
			log.Println("Server empty: Shutting down")
			os.Exit(0)
		}
	}
	log.Fatal("Manager exited")
}
func RouteMessage() {
	log.Println("Starting Shell-Company")
	for {
		Messages := sms.GetTwilioMessage()
		// if *config.FlagVerbose {
		// 	log.Println(len(Messages), "in queue")
		// }
		for _, m := range Messages {
			if m.Direction == "Outgoing" {
				break
			}
			// retrieve message ID
			messageID := m.Sid
			messageBody := m.Body
			messageBody = strings.ToUpper(messageBody)
			senderNumber := m.From
			IsNoirgateActive := false
			// check if message has been processed
			var preProcessedMessage bool
			AlreadyProcessedMessage, _ := ProcessedMessages[messageID]
			if !AlreadyProcessedMessage {
				preProcessedMessage = isProcessedMessage(messageID)
			}

			if preProcessedMessage {
				break
			}
			// if so discard and update rate limit cache for senderNumber
			if AlreadyProcessedMessage {
				break
				// if not begin message processing
			} else {
				addProcessedMessage(messageID)
				ProcessedMessages[messageID] = true
			}
			// is this a previously seen user?
			userData, IsExistingUser := ActiveUserGates[string(senderNumber)]
			UserHasLoot := false

			if IsExistingUser {
				IsNoirgateActive = container.IsNoirgateActive(userData.ContainerID)
				UserHasLoot = userData.HasLoot

			}
			// is input one of our menu commands?
			switch messageBody {
			case "HELP", "HOW":
				sms.SendTwilioMessage(senderNumber, sms.HelpMenu)
			// if SHELL spawn a new noirgate container, generate a uuid hostname, generate otp, send noirgate link.
			case "SHELL", "🐚":
				if len(ActiveUserGates) >= config.NoirgateMaxUsers {
					sms.SendTwilioMessage(senderNumber, "🚫 Sorry, the server is at its max capacity")
					break
				}

				log.Printf("%s: from:%s \n (%s)", messageID, senderNumber, messageBody)
				if IsNoirgateActive && !AlreadyProcessedMessage {
					containerHost := fmt.Sprintf("%s.%s.%s", userData.GateID, config.SandboxSubDomain, config.SandboxDomain)
					ShellMessage := fmt.Sprintf("⛔ An active shell has already been provisioned for this number: https://%s", containerHost)
					sms.SendTwilioMessage(senderNumber, ShellMessage)
					break
				}
				// register user
				noirgateUser := GuestInfo{
					PhoneNumber: senderNumber,
					GuestType:   "SMS",
				}
				// Register TOTP
				noirgateUser.OTPSecret = gotp.RandomSecret(32)
				// Attempt to provision new gate
				containerID, noirgateID, containerIP, err := container.ProvisionNoirgate(noirgateUser.OTPSecret)
				if err != nil {
					log.Println("Error provisioning container ", err)
					break
				}
				noirgateUser.ContainerID = containerID
				noirgateUser.GateID = noirgateID
				noirgateUser.ExpireTime = time.Now().Add(30 * time.Minute)
				noirgateUser.Expired = false
				noirgateUser.VIP = false
				noirgateUser.HasLoot = false
				// Add container record to DNS
				dns.AddNoirgateRecord(noirgateID, containerIP)
				containerHost := fmt.Sprintf("%s.%s.%s", noirgateID, config.SandboxSubDomain, config.SandboxDomain)
				ShellMessage := fmt.Sprintf("💻 Shell Provisioned: https://%s", containerHost)
				// Notify guest the gate is open
				sms.SendTwilioMessage(senderNumber, ShellMessage)
				ActiveUserGates[string(noirgateUser.PhoneNumber)] = noirgateUser
			case "GET", "LOOT":
				if !config.HasAWSCredentials {
					sms.SendTwilioMessage(senderNumber, "🛑 Error: AWS credentials not configured\n")
				} else {
					if IsExistingUser && IsNoirgateActive && !UserHasLoot {
						userData.HasLoot = true
						userData.LootBucketName = fmt.Sprintf("noirgate-s3-sandbox-%s", userData.GateID)
						lootBucket := loot.CreateTemporaryBucket(userData.LootBucketName)
						ActiveUserGates[string(userData.PhoneNumber)] = userData
						sms.SendTwilioMessage(senderNumber, fmt.Sprintf("📦 File storage provisioned: use s3 cli to read and write %s", lootBucket))

					}
				}
			case "BYE":
				if IsExistingUser && IsNoirgateActive {
					userData.Expired = true
					userData.ExpireTime = time.Now().Add(-18760 * time.Hour)
					dns.DeleteNoirgateRecord(userData.GateID)
					// notify user that shell has been terminated
					container.TerminateNoirgate(userData.ContainerID, userData.PhoneNumber)
					ActiveUserGates[string(userData.PhoneNumber)] = userData
					if userData.HasLoot {
						loot.DeleteTemporaryBucket(userData.LootBucketName)
					}
					sms.SendTwilioMessage(senderNumber, "🔥 Burndown Acknowledged: Shell has been terminated")
					delete(ActiveUserGates, string(userData.PhoneNumber))

				}
			case "OTP", "🔑":
				if IsExistingUser && IsNoirgateActive {
					otp.SendGuestTOTP(senderNumber, userData.OTPSecret)
				}

			case "MORE":
				if IsExistingUser && IsNoirgateActive {
					userData.ExpireTime = time.Now().Add(30 * time.Minute)
					sms.SendTwilioMessage(senderNumber, "➕ Extension Granted: Shell expiration extended to "+userData.ExpireTime.String())
				}

			case "VIP", "🤝":
				if IsExistingUser && IsNoirgateActive {
					// add allow list of VIP SMS numbers

					userData.ExpireTime = time.Now().Add(8760 * time.Hour)
					userData.VIP = true
					ActiveUserGates[string(userData.PhoneNumber)] = userData
					sms.SendTwilioMessage(senderNumber, "👽 Liberare Notitia: VIP mode activated, Shell immortality granted")
				}

			}

		}
	}
}

func addProcessedMessage(messageId string) {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.DisableKeepAlives = true
	httpClient := http.Client{Transport: t}
	defer httpClient.CloseIdleConnections()
	httpRequestURL := fmt.Sprintf("http://%s:2379/v2/keys/shellcompany/messages/%s/", config.NoirgateETCDHost, messageId)
	params := url.Values{}
	params.Add("value", messageId)
	body := strings.NewReader(params.Encode())

	req, err := http.NewRequest("PUT", httpRequestURL, body)
	if err != nil {
		// handle err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpResponse, err := httpClient.Do(req)
	responseBytes, _ := ioutil.ReadAll(httpResponse.Body)
	log.Println("Adding incoming message to execution log", string(responseBytes))

}

func isProcessedMessage(messageId string) (exists bool) {
	exists = true
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.DisableKeepAlives = true
	httpClient := http.Client{Transport: t}
	defer httpClient.CloseIdleConnections()
	httpRequestURL := fmt.Sprintf("http://%s:2379/v2/keys/shellcompany/messages/%s/", config.NoirgateETCDHost, messageId)
	httpResponse, err := httpClient.Get(httpRequestURL)
	if err != nil {
		log.Fatal("Error reading from etcd", err)
	}
	if httpResponse.StatusCode == 404 {
		if *config.FlagVerbose {
			log.Println("Message ", messageId, "is new")
		}
		return false
	} else {
		if *config.FlagVerbose {

			log.Println("Message ", messageId, " exists in message history")
		}
		return exists
	}
}
