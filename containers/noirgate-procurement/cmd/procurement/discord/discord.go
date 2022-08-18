package discord

import (
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/Shell-Company/procurement/cmd/procurement/util"
)

type discordConfig struct {
	GuildId        string `mapstructure:"guild"`
	Token          string `mapstructure:"token"`
	ShouldRegister bool   `mapstructure:"register"`
	ShouldReset    bool   `mapstructure:"reset"`
	WarnAfter      int    `mapstructure:"warnafter"`
	util.RpcConfig
}

func NewDiscordCommand() *cobra.Command {
	discordCfg := &discordConfig{}

	cmd := &cobra.Command{
		Use: "discord",
		PreRun: func(cmd *cobra.Command, args []string) {
			setupFromEnv(cmd, discordCfg)
		},
		Run: func(cmd *cobra.Command, args []string) {
			logger, _ := zap.NewProduction()

			rpc, err := util.GetRpcClient(cmd, &discordCfg.RpcConfig)
			if err != nil {
				logger.Sugar().Errorf("error getting rpc client, %s", err.Error())
			}
			handler := NewHandler(logger, rpc)
			logger.Sugar().Infof("initializing with options: rpc client: %s; register: %v; reset: %v; warnafter: %d",
				discordCfg.RpcClient,
				discordCfg.ShouldRegister,
				discordCfg.ShouldReset,
				discordCfg.WarnAfter,
			)

			session, err := discordgo.New("Bot " + discordCfg.Token)
			if err != nil {
				logger.Sugar().Fatalf("Invalid bot token %s", err.Error())
			}

			// register event handlers
			session.AddHandler(handler.HandleInteraction)
			session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
				logger.Info("bot is running")
			})

			// connect to discord
			err = session.Open()
			if err != nil {
				logger.Sugar().Fatalf("Error getting session %v", err)
			}
			defer session.Close()

			// register slash commands
			if discordCfg.ShouldRegister {
				for _, h := range handler.Commands {
					_, err := session.ApplicationCommandCreate(session.State.User.ID, discordCfg.GuildId, h)
					if err != nil {
						logger.Sugar().Panicf("Error creating command '%s' command, %s", h.Name, err.Error())
					}
					logger.Sugar().Infof("Registered command '%s'", h.Name)
				}
			}

			sc := make(chan os.Signal, 1)
			signal.Notify(sc, syscall.SIGTERM, syscall.SIGINT, os.Interrupt)
			<-sc
			logger.Info("Shutting down (CTRL-C to force)")

			if discordCfg.ShouldReset {
				// Delete slash commands. This is mostly useful if you create
				// global commands (guildId == ""). Recreating a global command
				// without deleting the existing version first can cause users
				// to hit stale cache on Discord's side, ie, dead slash commands
				existing, err := session.ApplicationCommands(session.State.User.ID, discordCfg.GuildId)
				if err != nil {
					logger.Sugar().Infof("error deregistering commands,", err.Error())
				}
				for _, e := range existing {
					err = session.ApplicationCommandDelete(e.ApplicationID, discordCfg.GuildId, e.ID)
					if err != nil {
						logger.Sugar().Errorf("error deregistering command %s, %s", e.Name, err.Error())
					}
					logger.Sugar().Infof("deregistered command %s", e.Name)
				}
			}
		},
	}

	bindFlags(cmd, discordCfg)

	return cmd
}

func bindFlags(cmd *cobra.Command, config *discordConfig) {
	cmd.Flags().StringVarP(&config.Token, "token", "t", "", "bot token")
	cmd.Flags().StringVarP(&config.GuildId, "guild", "g", "", "guild id. leave empty to register global commands")
	cmd.Flags().BoolVar(&config.ShouldRegister, "register", false, "register commands on startup")
	cmd.Flags().BoolVar(&config.ShouldReset, "reset", false, "delete commands on shutdown")
	cmd.Flags().IntVar(&config.WarnAfter, "shell-warning-timeout", 0, "when to warn about shell deletion (seconds) (0 or negative for never)")
	util.AddRpcFlags(cmd, &config.RpcConfig)
}

func setupFromEnv(cmd *cobra.Command, config *discordConfig) {
	viper.BindPFlags(cmd.Flags())
	viper.SetEnvPrefix("PROCUREMENT_DISCORD")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
	viper.Unmarshal(config)
}
