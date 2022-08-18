package slack

import (
	"context"
	"fmt"
	"strings"

	"github.com/Shell-Company/procurement/internal/corporate"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

func (h *Handler) handleSlashCommand(ctx context.Context, evt *socketmode.Event, cmd *slack.SlashCommand) error {
	s := strings.ToUpper(strings.TrimSpace(cmd.Text))
	handler, ok := h.slashCommandHandlerFuncs[corporate.RpcCommand(s)]
	if !ok {
		return fmt.Errorf("unknown command, %s", s)
	}
	h.socket.Ack(*evt.Request)
	return handler(ctx, cmd)
}

// newSlashCommandHandlerFunc is a factory for slash command handlers.  it's
// almost identical to newInteractionHandlerFunc, but the payloads from slack
// are just different enough that they need distinct implementations
func newSlashCommandHandlerFunc(h *Handler, c corporate.RpcCommand, errorMessage string) SlashCommandHandlerFunc {
	return func(ctx context.Context, cmd *slack.SlashCommand) error {
		text, err := h.callRpcWithDiscordUserId(ctx, c, cmd.UserID)
		if err != nil {
			return fmt.Errorf("%s, %s", errorMessage, err)
		}
		// seriously, these two function differ by like two "."s, an "_", and a ","
		_, err = h.api.PostEphemeral(cmd.ChannelID, cmd.UserID, slack.MsgOptionText(text, false))
		return err
	}
}

func (h *Handler) handleSlashShell(ctx context.Context, cmd *slack.SlashCommand) error {
	text, err := h.callRpcWithDiscordUserId(ctx, corporate.ShellCommand, cmd.UserID)
	if err != nil {
		return fmt.Errorf("error getting a shell, %s", err)
	}

	msg := slack.MsgOptionBlocks(
		slack.NewSectionBlock(
			&slack.TextBlockObject{
				Type: slack.MarkdownType,
				Text: text,
			},
			nil,
			nil,
		),
		slack.NewActionBlock("actions",
			slack.NewButtonBlockElement(
				string(corporate.OtpCommand),
				"",
				&slack.TextBlockObject{
					Type: slack.PlainTextType,
					Text: "Get OTP",
				},
			),
			slack.NewButtonBlockElement(
				string(corporate.LootCommand),
				"",
				&slack.TextBlockObject{
					Type: slack.PlainTextType,
					Text: "Get loot bucket",
				},
			),
			slack.NewButtonBlockElement(
				string(corporate.ByeCommand),
				"",
				&slack.TextBlockObject{
					Type: slack.PlainTextType,
					Text: "Terminate shell",
				},
			),
		),
	)
	_, _, err = h.api.PostMessage(cmd.ChannelID, slack.MsgOptionReplaceOriginal(cmd.ResponseURL), msg)
	return err

	// TODO(kantatenbot): remind when shell is about to expire
	// reminder := time.AfterFunc(time.Second*10, func() {
	// 	h.api.PostMessage(cmd.ChannelID, slack.MsgOptionReplaceOriginal(cmd.ResponseURL), msg)
	// })
}
