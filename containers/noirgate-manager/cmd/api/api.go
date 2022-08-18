package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"noirgate/config"
	"noirgate/container"
	"noirgate/core"
	"noirgate/dns"
	"noirgate/loot"
	"noirgate/sms"
	"strings"
	"time"

	"github.com/Jeffail/gabs/v2"
	"github.com/gorilla/mux"
	"github.com/xlzd/gotp"
)

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// A very simple health check.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// In the future we could report back on the status of our DB, or our cache
	// (e.g. Redis) by performing a simple PING, and include them in the response.
	io.WriteString(w, `{"alive": true}\n`)
}

func RouteMessageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	discordUserId := r.Header.Get(*config.FlagAuthHeaderName)
	if discordUserId == "" {
		w.WriteHeader(http.StatusForbidden)
		io.WriteString(w, "Unauthorized\n")
		return
	}

	log.Println(r.RemoteAddr)
	apiRequest := r.Body
	requestBytes, _ := ioutil.ReadAll(apiRequest)

	parsedJSON, _ := gabs.ParseJSON(requestBytes)
	command := parsedJSON.Path("command")
	var message string
	if command.Data() != nil {
		message = strings.ToUpper(command.Data().(string))

		message = strings.Replace(message, " ", "", -1)
	}

	userData, IsExistingUser := core.ActiveUserGates[discordUserId]
	IsNoirgateActive := false
	UserHasLoot := false
	if IsExistingUser {
		IsNoirgateActive = container.IsNoirgateActive(userData.ContainerID)
		UserHasLoot = userData.HasLoot
	}
	w.WriteHeader(http.StatusOK)

	switch message {
	case "HELP", "HOW":
		//
		// Send help menu
		io.WriteString(w, string(sms.HelpMenu))
	// if SHELL spawn a new noirgate container, generate a uuid hostname, generate otp, send noirgate link.
	case "SHELL", "🐚":
		if len(core.ActiveUserGates) >= config.NoirgateMaxUsers {
			io.WriteString(w, "🚫 Sorry, the server is at its max capacity\n")

			break
		}
		if IsNoirgateActive {
			containerHost := fmt.Sprintf("%s.%s.%s", userData.GateID, config.SandboxSubDomain, config.SandboxDomain)
			ShellMessage := fmt.Sprintf("⛔ An active shell has already been provisioned for this user: https://%s \n", containerHost)
			io.WriteString(w, string(ShellMessage))
			break
		}

		// register user
		noirgateUser := core.GuestInfo{
			PhoneNumber: "API-USER",
			GuestType:   "API",
			IPAddress:   r.RemoteAddr,
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
		ShellMessage := fmt.Sprintf("💻 Shell Provisioned: https://%s \n", containerHost)
		// Notify guest the gate is open
		io.WriteString(w, string(ShellMessage))
		core.ActiveUserGates[discordUserId] = noirgateUser

	case "BYE":
		if IsExistingUser && IsNoirgateActive {
			userData.Expired = true
			dns.DeleteNoirgateRecord(userData.GateID)
			// notify user that shell has been terminated
			container.TerminateNoirgate(userData.ContainerID, userData.PhoneNumber)
			io.WriteString(w, "🔥 Burndown Acknowledged: Shell has been terminated\n")

			if UserHasLoot {
				loot.DeleteTemporaryBucket(userData.LootBucketName)
			}
			delete(core.ActiveUserGates, string(discordUserId))

		}
	case "OTP", "🔑":
		if IsExistingUser && IsNoirgateActive {
			APITOTP := SendGuestTOTPAPI(userData.OTPSecret)
			io.WriteString(w, fmt.Sprintf("🔑 %s\n", APITOTP))

		}
	case "LOOT":
		// allow user input to define bucket seed name
		if IsExistingUser && IsNoirgateActive && !UserHasLoot {
			if !config.HasAWSCredentials {
				io.WriteString(w, "🛑 Error: AWS credentials not configured\n")
			} else {
				userData.HasLoot = true
				userData.LootBucketName = fmt.Sprintf("noirgate-s3-sandbox-%s", userData.GateID)
				lootBucket := loot.CreateTemporaryBucket(userData.LootBucketName)
				core.ActiveUserGates[discordUserId] = userData
				io.WriteString(w, fmt.Sprintf("📦 File storage provisioned: use s3 cli to read and write %s\n", lootBucket))
			}
		}
	case "VIP", "🤝":
		if IsExistingUser && IsNoirgateActive && !UserHasLoot {
			userData.ExpireTime = time.Now().Add(8760 * time.Hour)
			userData.VIP = true
			core.ActiveUserGates[discordUserId] = userData
			io.WriteString(w, "👽 Liberare Notitia: VIP mode activated, Shell immortality granted\n")
		}
	}

}

func ListGuestsHandler(w http.ResponseWriter, r *http.Request) {
	// list all guests
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	activeGates := core.ActiveUserGates
	activeUsers, _ := json.Marshal(activeGates)
	io.WriteString(w, fmt.Sprintf("active users:\n %s\n", string(activeUsers)))
}

func StartServer() {
	r := mux.NewRouter()
	// health check endpoint
	r.HandleFunc("/health", HealthCheckHandler)
	r.HandleFunc("/api", RouteMessageHandler)
	r.HandleFunc("/users", ListGuestsHandler)
	serverAddress := fmt.Sprintf("0.0.0.0:%d", *config.FlagServerPort)
	srv := &http.Server{
		Addr: serverAddress,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	defer srv.Shutdown(ctx)
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	log.Fatal(srv.ListenAndServe())

}

func SendGuestTOTPAPI(UserSecret string) string {
	UserSecret = strings.ToUpper(UserSecret)
	UserMFA := gotp.NewDefaultTOTP(UserSecret)
	PassCode, _ := UserMFA.NowWithExpiration()
	return (PassCode + " use within 5 seconds after receiving")
}
