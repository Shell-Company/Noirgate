package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Shell-Company/procurement/internal/corporate"
)

type RpcClient interface {
	Call(ctx context.Context, command corporate.RpcCommand, principal corporate.Principal) (string, error)
}

type HttpRpcClient struct {
	ApiUrl string
}

func NewDefaultRpcClient(apiUrl string) *HttpRpcClient {
	return &HttpRpcClient{
		ApiUrl: apiUrl,
	}
}

type rpcPayload struct {
	Command corporate.RpcCommand `json:"command"`
}

func (c *HttpRpcClient) Call(ctx context.Context, command corporate.RpcCommand, principal corporate.Principal) (string, error) {
	p := rpcPayload{Command: command}
	data, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest(http.MethodPost, c.ApiUrl, bytes.NewBuffer(data))
	if err != nil {
		panic(err)
	}
	req = req.WithContext(ctx)

	switch principal.Type {
	case corporate.DiscordPrincipal:
		req.Header.Set("X-Discord-UserId", principal.Id)
	case corporate.SlackPrincipal:
		req.Header.Set("X-Slack-UserId", principal.Id)
	default:
		return "", fmt.Errorf("invalid principal type, %s", principal.Type)
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	buf := strings.Builder{}
	io.Copy(&buf, resp.Body)
	text := buf.String()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("http error (status %s, %s)", resp.Status, text)
	}

	if text == "" {
		return "", fmt.Errorf("empty response from api")
	}

	return text, nil
}
