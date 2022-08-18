package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"

	"github.com/Shell-Company/procurement/internal/corporate"
)

type LambdaRpcRequest struct {
	Command   corporate.RpcCommand `json:"command"`
	Principal corporate.Principal  `json:"principal"`
}

func (r *LambdaRpcRequest) MarhsalJSON() ([]byte, error) {
	if r.Command == "" {
		return nil, fmt.Errorf("missing command in rpc request")
	}
	return json.Marshal(r)
}

type LambdaRpc struct {
	client       *lambda.Client
	functionName string
}

func NewLambdaRpcClient(client *lambda.Client, functionName string) *LambdaRpc {
	return &LambdaRpc{
		client:       client,
		functionName: functionName,
	}
}

func (p *LambdaRpc) Call(ctx context.Context, command corporate.RpcCommand, principal corporate.Principal) (string, error) {
	req := &LambdaRpcRequest{Command: command, Principal: principal}
	b, err := req.MarhsalJSON()
	if err != nil {
		return "", fmt.Errorf("error marshalling rpc payload, %s", err.Error())
	}

	resp, err := p.client.Invoke(ctx, &lambda.InvokeInput{
		FunctionName:   &p.functionName,
		InvocationType: types.InvocationTypeRequestResponse,
		Payload:        b,
	})
	if err != nil {
		return "", fmt.Errorf("error calling rpc, %s", err.Error())
	}

	// the response is quoted over the wire because the handler returns a bare
	// string. just a SILLY bit of bullshit that amazon does
	r, err := strconv.Unquote(string(resp.Payload))
	if err != nil {
		return "", fmt.Errorf("format error from rpc client lol, %s", err.Error())
	}
	return r, nil
}
