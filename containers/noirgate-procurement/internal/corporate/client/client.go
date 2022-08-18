package client

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

type ClientType string

var (
	DefaultClient ClientType = "direct"
	LambdaClient  ClientType = "lambda"
)

// NewRpcProxyClient returns a new RPC client. If proxy type is "lambda", target refers to the lambda function name. If the type is "direct", target refers to the api url.
func NewRpcClient(ctx context.Context, t ClientType, target string) (RpcClient, error) {
	switch t {
	case LambdaClient:
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return nil, err
		}
		l := lambda.NewFromConfig(cfg)
		return NewLambdaRpcClient(l, target), nil
	case DefaultClient:
		return NewDefaultRpcClient(target), nil
	default:
		return nil, fmt.Errorf("unknown rpc client type")
	}
}
