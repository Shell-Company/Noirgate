package slack

import (
	"context"
	"fmt"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"go.uber.org/zap"

	"github.com/Shell-Company/procurement/internal/corporate"
	"github.com/Shell-Company/procurement/internal/corporate/client"
)

type SlashCommandHandlerFunc func(ctx context.Context, cmd *slack.SlashCommand) error

type InteractionHandlerFunc func(ctx context.Context, cmd *slack.InteractionCallback) error

type Handler struct {
	slashCommandHandlerFuncs map[corporate.RpcCommand]SlashCommandHandlerFunc
	interactionHandlerFuncs  map[corporate.RpcCommand]InteractionHandlerFunc
	api                      *slack.Client
	socket                   *socketmode.Client
	logger                   *zap.Logger
	rpcClient                client.RpcClient
	// reminders                map[string]*time.Timer
}

func NewHandler(logger *zap.Logger, api *slack.Client, socket *socketmode.Client, rpcClient client.RpcClient) *Handler {
	h := &Handler{
		slashCommandHandlerFuncs: make(map[corporate.RpcCommand]SlashCommandHandlerFunc),
		interactionHandlerFuncs:  make(map[corporate.RpcCommand]InteractionHandlerFunc),
		// reminders:                make(map[string]*time.Timer),
		api:       api,
		socket:    socket,
		rpcClient: rpcClient,
		logger:    logger,
	}

	h.slashCommandHandlerFuncs[corporate.ShellCommand] = h.handleSlashShell
	h.slashCommandHandlerFuncs[corporate.ByeCommand] = newSlashCommandHandlerFunc(h, corporate.ByeCommand, "error terminating shell")
	h.slashCommandHandlerFuncs[corporate.HowCommand] = newSlashCommandHandlerFunc(h, corporate.HowCommand, "error getting manual")
	h.slashCommandHandlerFuncs[corporate.LootCommand] = newSlashCommandHandlerFunc(h, corporate.LootCommand, "error getting loot bucket")
	h.slashCommandHandlerFuncs[corporate.OtpCommand] = newSlashCommandHandlerFunc(h, corporate.OtpCommand, "error getting otp")

	h.interactionHandlerFuncs[corporate.ByeCommand] = h.handleInteractionBye
	h.interactionHandlerFuncs[corporate.LootCommand] = newRpcInteractionHandlerFunc(h, corporate.LootCommand, "error getting loot bucket")
	h.interactionHandlerFuncs[corporate.OtpCommand] = newRpcInteractionHandlerFunc(h, corporate.OtpCommand, "error getting otp")

	return h
}

func (h *Handler) Run() {
	ctx := context.Background()
	go func() {
		for evt := range h.socket.Events {
			h.handleEvent(ctx, &evt)
		}
	}()
	h.socket.Run()
}

func (h *Handler) handleEvent(ctx context.Context, evt *socketmode.Event) {
	defer func() {
		if p := recover(); p != nil {
			h.logger.Sugar().Warnf("recovered from panic, %s", p)
		}
	}()
	switch evt.Type {
	case socketmode.EventTypeConnecting:
		h.logger.Info("connecting to slack")
	case socketmode.EventTypeConnectionError:
		h.logger.Sugar().Errorf("connection failed, %+v", evt.Data)
	case socketmode.EventTypeConnected:
		h.logger.Info("connected")
	case socketmode.EventTypeInteractive:
		cmd, ok := evt.Data.(slack.InteractionCallback)
		if !ok {
			h.logger.Sugar().Infof("ignoring malformed interaction %+v", evt)
			return
		}
		if err := h.handleInteractionCommand(ctx, evt, &cmd); err != nil {
			h.logger.Sugar().Errorf("error handling interaction, %s", err.Error())
		}
	case socketmode.EventTypeSlashCommand:
		cmd, ok := evt.Data.(slack.SlashCommand)
		if !ok {
			h.logger.Sugar().Infof("ignoring malformed slash command %+v", evt)
			return
		}
		h.logger.Sugar().Infof("handling command %s, subcommand", cmd.Command, cmd.Text)
		if err := h.handleSlashCommand(ctx, evt, &cmd); err != nil {
			h.logger.Sugar().Errorf("error handling slash command, %s", err.Error())
		}
	}
}

func (h *Handler) callRpcWithDiscordUserId(ctx context.Context, c corporate.RpcCommand, uid string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*15)
	defer cancel()
	text, err := h.rpcClient.Call(ctx, c, corporate.Principal{
		Type: corporate.DiscordPrincipal,
		Id:   uid,
	})
	if text == "" {
		return "", fmt.Errorf("empty response from api")
	}
	return text, err
}
