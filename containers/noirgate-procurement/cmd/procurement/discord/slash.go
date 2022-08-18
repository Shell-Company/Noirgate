// handlers for slash commands
package discord

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/Shell-Company/procurement/internal/corporate"
)

var (
	howCommand   = discordgo.ApplicationCommand{Name: "how", Description: "rtfm"}
	shellCommand = discordgo.ApplicationCommand{Name: "shell", Description: "Get a shell"}
	otpCommand   = discordgo.ApplicationCommand{Name: "otp", Description: "Get OTP"}
	lootCommand  = discordgo.ApplicationCommand{Name: "loot", Description: "Get loot bucket"}
	byeCommand   = discordgo.ApplicationCommand{Name: "bye", Description: "Terminate shell"}
	moreCommand  = discordgo.ApplicationCommand{Name: "more", Description: "Get more"}

	warnAfterMinutes = 25
)

func (h *Handler) handleHow(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	ephemeralAckDeferred(s, i)
	content, err := callRpcWithDiscordUserId(h.rpcClient, i, corporate.HowCommand)
	if err != nil {
		ephemeralResponseEdit(s, i, "Error getting help")
		return err
	}
	return ephemeralResponseEdit(s, i, fmt.Sprintf("```\n%s\n```", content))
}

func (h *Handler) handleShell(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	ephemeralAckDeferred(s, i)

	text, err := callRpcWithDiscordUserId(h.rpcClient, i, corporate.ShellCommand)
	if err != nil || text == "" {
		ephemeralResponseEdit(s, i, "Error getting shell")
		return err
	}
	h.scheduleReminder(s, i, time.Minute*time.Duration(warnAfterMinutes))

	_, err = s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
		Content: text,
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Get OTP",
						Style:    discordgo.PrimaryButton,
						CustomID: string(corporate.OtpCommand),
					},
					discordgo.Button{
						Label:    "Get loot bucket",
						Style:    discordgo.PrimaryButton,
						CustomID: string(corporate.LootCommand),
					},
					discordgo.Button{
						Label:    "Terminate shell",
						Style:    discordgo.DangerButton,
						CustomID: string(corporate.ByeCommand),
					},
				},
			},
		},
	})
	return err
}

func (h *Handler) handleOtp(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	f := newRpcCommandHandlerFunc(corporate.OtpCommand, "Error getting OTP")
	return f(h.rpcClient, s, i)
}

func (h *Handler) handleLoot(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	f := newRpcCommandHandlerFunc(corporate.LootCommand, "Error getting loot bucket")
	return f(h.rpcClient, s, i)
}

func (h *Handler) handleBye(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// suppress the alert whether the api call succeeds or not
	defer h.removeReminder(i.Member.User.ID)
	f := newRpcCommandHandlerFunc(corporate.ByeCommand, "Error terminating shell. Is it already terminated?")
	return f(h.rpcClient, s, i)
}

func (h *Handler) handleMore(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	f := newRpcCommandHandlerFunc(corporate.MoreCommand, "Error getting more")
	return f(h.rpcClient, s, i)
}
