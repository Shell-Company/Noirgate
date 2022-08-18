package slack

import (
	"strings"

	"github.com/Shell-Company/procurement/cmd/procurement/util"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type slackConfig struct {
	AppToken     string `mapstructure:"app-token"`
	BotUserToken string `mapstructure:"bot-user-token"`
	WarnAfter    int    `mapstructure:"warnafter"`
	util.RpcConfig
}

func NewSlackCommand() *cobra.Command {
	config := &slackConfig{}

	cmd := &cobra.Command{
		Use: "slack",
		PreRun: func(cmd *cobra.Command, args []string) {
			setupFromEnv(cmd, config)
		},
		Run: func(cmd *cobra.Command, args []string) {
			logger, _ := zap.NewProduction()
			rpcClient, err := util.GetRpcClient(cmd, &config.RpcConfig)
			if err != nil {
				logger.Sugar().Errorf("error getting rpc client, %s", err.Error())
			}
			api := slack.New(config.BotUserToken, slack.OptionAppLevelToken(config.AppToken))
			socket := socketmode.New(api)
			handler := NewHandler(logger, api, socket, rpcClient)
			handler.Run()
		},
	}

	bindFlags(cmd, config)

	return cmd
}

func bindFlags(cmd *cobra.Command, config *slackConfig) {
	cmd.Flags().StringVar(&config.AppToken, "app-token", "", "app token (starts with xapp-)")
	cmd.Flags().StringVar(&config.BotUserToken, "bot-user-token", "", "bot token (starts with xbot-)")
	cmd.Flags().IntVar(&config.WarnAfter, "shell-warning-timeout", 0, "when to warn about shell deletion (seconds) (0 or negative for never)")
	util.AddRpcFlags(cmd, &config.RpcConfig)
}

func setupFromEnv(cmd *cobra.Command, config *slackConfig) {
	viper.BindPFlags(cmd.Flags())
	viper.SetEnvPrefix("PROCUREMENT_SLACK")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
	viper.Unmarshal(config)
}
