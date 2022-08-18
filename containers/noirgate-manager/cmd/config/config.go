package config

import (
	"flag"
	"os"
	"strconv"
)

var (
	FlagVerbose             = flag.Bool("v", false, "Enable verbose messages")
	FlagAPI                 = flag.Bool("w", false, "Enable API Server")
	FlagTXT                 = flag.Bool("txt", false, "Enable SMS Server")
	FlagServerPort          = flag.Int("p", 31337, "API endpoint port")
	FlagAuthHeaderName      = flag.String("header", "X-Discord-UserId", "Header to authenticate against for API requests")
	FlagClientDockerNetwork = flag.String("network", "noirgate-compose_clients", "Docker network to use for the client")

	TwilioSID        = os.Getenv("TWILIO_SID")
	TwilioToken      = os.Getenv("TWILIO_TOKEN")
	TwilioNumber     = os.Getenv("TWILIO_NUMBER")
	ImageName        = os.Getenv("NOIRGATE_IMAGE")
	SandboxSubDomain = os.Getenv("NOIRGATE_SUB")
	SandboxDomain    = os.Getenv("NOIRGATE_TLD")
	NoirgateETCDHost = os.Getenv("NOIRGATE_ETCD_HOST")
	// convert env var string to int
	NoirgateMaxUsers, _ = strconv.Atoi(os.Getenv("NOIRGATE_MAX_USERS"))
	// NoirgateMaxUsers     = os.Getenv("NOIRGATE_MAX_USERS")
	NoirgateServerSecret = os.Getenv("NOIRGATE_SERVER_SECRET")
	// checks for AWS variables or config file
	HasAWSCredentialsVar   = os.Getenv("AWS_ACCESS_KEY_ID") != "" && os.Getenv("AWS_SECRET_ACCESS_KEY") != ""
	_, configFileReadError = os.ReadFile("~/.aws/config")
	// error to bool
	HasAWSCredentialsConfig = configFileReadError == nil

	HasAWSCredentials = HasAWSCredentialsVar || HasAWSCredentialsConfig
)
