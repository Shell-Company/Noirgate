package corporate

type RpcCommand string

var (
	HowCommand   RpcCommand = "HOW"
	ShellCommand RpcCommand = "SHELL"
	ByeCommand   RpcCommand = "BYE"
	OtpCommand   RpcCommand = "OTP"
	LootCommand  RpcCommand = "LOOT"
	MoreCommand  RpcCommand = "MORE"
)

type PrincipalType string

var (
	SlackPrincipal   PrincipalType = "slack"
	DiscordPrincipal PrincipalType = "discord"
)

type Principal struct {
	Type PrincipalType
	Id   string
}
