package util

import (
	"fmt"

	"github.com/Shell-Company/procurement/internal/corporate/client"
	"github.com/spf13/cobra"
)

type RpcConfig struct {
	RpcClient string `mapstructure:"rpc-client"`
	RpcTarget string `mapstructure:"rpc-target"`
}

func GetRpcClient(cmd *cobra.Command, cfg *RpcConfig) (client.RpcClient, error) {
	t := cfg.RpcClient
	if t != string(client.DefaultClient) && t != string(client.LambdaClient) {
		return nil, fmt.Errorf(`invalid rpc proxy %s`, t)
	}
	return client.NewRpcClient(cmd.Context(), client.ClientType(t), cfg.RpcTarget)
}

func AddRpcFlags(cmd *cobra.Command, cfg *RpcConfig) {
	cmd.Flags().StringVar(&cfg.RpcClient, "rpc-client", string(client.DefaultClient), `"direct" or "lambda"`)
	cmd.Flags().StringVar(&cfg.RpcTarget, "rpc-target", "", `url if direct, function name if lambda`)
}
