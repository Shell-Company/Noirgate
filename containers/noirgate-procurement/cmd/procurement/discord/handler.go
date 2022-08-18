package discord

import (
	"context"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"

	"github.com/Shell-Company/procurement/internal/corporate"
	"github.com/Shell-Company/procurement/internal/corporate/client"
)

type commandHandlerFunc func(s *discordgo.Session, i *discordgo.InteractionCreate) error

type Handler struct {
	Commands []*discordgo.ApplicationCommand

	handlerFuncs map[string]commandHandlerFunc
	reminders    map[string]*time.Timer

	rpcClient client.RpcClient
	logger    *zap.Logger
}

func NewHandler(logger *zap.Logger, rpcClient client.RpcClient) *Handler {
	h := &Handler{
		handlerFuncs: make(map[string]commandHandlerFunc),
		reminders:    make(map[string]*time.Timer),

		rpcClient: rpcClient,
		logger:    logger,
	}
	h.addCommand(&byeCommand, h.handleBye)
	h.addCommand(&howCommand, h.handleHow)
	h.addCommand(&lootCommand, h.handleLoot)
	h.addCommand(&moreCommand, h.handleMore)
	h.addCommand(&otpCommand, h.handleOtp)
	h.addCommand(&shellCommand, h.handleShell)
	return h
}

func (h *Handler) addCommand(c *discordgo.ApplicationCommand, f commandHandlerFunc) {
	h.handlerFuncs[c.Name] = f
	h.Commands = append(h.Commands, c)
}

func (h *Handler) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		cmdName := i.ApplicationCommandData().Name
		h.handleApplicationCommand(s, i, cmdName)
	case discordgo.InteractionMessageComponent:
		h.handleButtonClick(s, i)
	default:
		h.logger.Sugar().Info("Unknown interaction type, %+v", i)
		// if you send weird to me im logging it that's on you
	}
}

func (h *Handler) handleApplicationCommand(s *discordgo.Session, i *discordgo.InteractionCreate, cmdName string) {
	h.logger.Sugar().Infof("Handling %s command", cmdName)
	if handler, ok := h.handlerFuncs[cmdName]; ok {
		if err := handler(s, i); err != nil {
			h.logger.Sugar().Errorf("Error handling command %s, %s", cmdName, err.Error())
		}
	} else {
		h.logger.Sugar().Info("Unknown command ", cmdName)
	}
}

func (h *Handler) handleButtonClick(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// sort of a hack due to how we make the action components
	cmdName := strings.ToLower(i.MessageComponentData().CustomID)
	switch cmdName {
	// TODO(kantatenbot): delete the original message with the buttons when the command is "bye"
	default:
		h.handleApplicationCommand(s, i, cmdName)
	}
}

type unboundCommandHandlerFunc func(rpcClient client.RpcClient, s *discordgo.Session, i *discordgo.InteractionCreate) error

func newRpcCommandHandlerFunc(command corporate.RpcCommand, failureMessage string) unboundCommandHandlerFunc {
	return func(rpcClient client.RpcClient, s *discordgo.Session, i *discordgo.InteractionCreate) error {
		ephemeralAckDeferred(s, i)
		content, err := callRpcWithDiscordUserId(rpcClient, i, command)
		if err != nil || content == "" {
			ephemeralResponseEdit(s, i, failureMessage)
			return err
		}
		return ephemeralResponseEdit(s, i, content)
	}
}

func callRpcWithDiscordUserId(rpcClient client.RpcClient, i *discordgo.InteractionCreate, c corporate.RpcCommand) (string, error) {
	return rpcClient.Call(context.Background(), c, corporate.Principal{
		Type: corporate.DiscordPrincipal,
		Id:   i.Member.User.ID,
	})
}

func ephemeralAckDeferred(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: 1 << 6, // ephemeral
		},
	})
}

func ephemeralResponseEdit(s *discordgo.Session, i *discordgo.InteractionCreate, content string) error {
	_, err := s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
		Content: content,
	})
	return err
}
