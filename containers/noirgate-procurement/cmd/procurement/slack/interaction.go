package slack

import (
	"context"
	"fmt"
	"strings"

	"github.com/Shell-Company/procurement/internal/corporate"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

// handleInteractionCommand handles "block interaction" events (eg clicks from
// buttons that we made using message "blocks"). We use the "ActionID" of the
// block interaction to dispatch to individual handlers
func (h *Handler) handleInteractionCommand(ctx context.Context, evt *socketmode.Event, cmd *slack.InteractionCallback) error {
	actionId := cmd.ActionCallback.BlockActions[0].ActionID
	s := strings.ToLower(strings.TrimSpace(actionId))
	handler, ok := h.interactionHandlerFuncs[corporate.RpcCommand(s)]
	if !ok {
		return fmt.Errorf("unknown command, %s", s)
	}
	h.socket.Ack(*evt.Request)
	h.logger.Sugar().Infof("handling interaction %s", actionId)
	return handler(ctx, cmd)
}

func newRpcInteractionHandlerFunc(h *Handler, c corporate.RpcCommand, errorMsg string) InteractionHandlerFunc {
	return func(ctx context.Context, cmd *slack.InteractionCallback) error {
		text, err := h.callRpcWithDiscordUserId(ctx, c, cmd.User.ID)
		if err != nil {
			return fmt.Errorf("%s, %s", errorMsg, err.Error())
		}
		_, err = h.api.PostEphemeral(cmd.Channel.ID, cmd.User.ID, slack.MsgOptionText(text, false))
		return err
	}
}

func (h *Handler) handleInteractionBye(ctx context.Context, cmd *slack.InteractionCallback) error {
	f := newRpcInteractionHandlerFunc(h, corporate.ByeCommand, "error terminating shell")
	h.api.PostEphemeral(
		cmd.Channel.ID,
		cmd.User.ID,
		slack.MsgOptionReplaceOriginal(cmd.ResponseURL),
		slack.MsgOptionText("Shell is scheduled for termination", false),
	)
	return f(ctx, cmd)
}
