package auth

import (
	"chat/utils"
	"context"
	"errors"
	"fmt"
	"strings"
)

// WechatAuthClient exchanges a frontend code for WeChat identifiers.
type WechatAuthClient interface {
	ExchangeCode(ctx context.Context, code string) (*WechatAuthResponse, error)
}

// WechatAuthResponse carries WeChat openid/unionid.
type WechatAuthResponse struct {
	OpenID  string
	UnionID string
}

type mockWechatAuthClient struct{}

// ExchangeCode provides a deterministic mock for development/testing.
// TODO: replace with real call to WeChat auth endpoint.
func (m *mockWechatAuthClient) ExchangeCode(ctx context.Context, code string) (*WechatAuthResponse, error) {
	if strings.TrimSpace(code) == "" {
		return nil, errors.New("empty wechat code")
	}

	hash := utils.Sha2Encrypt(code)
	return &WechatAuthResponse{
		OpenID:  fmt.Sprintf("wx_openid_%s", hash[:16]),
		UnionID: fmt.Sprintf("wx_unionid_%s", hash[16:32]),
	}, nil
}

var wechatClient WechatAuthClient = &mockWechatAuthClient{}

// SetWechatAuthClient allows injecting real client in production.
func SetWechatAuthClient(client WechatAuthClient) {
	if client != nil {
		wechatClient = client
	}
}

func getWechatAuthClient() WechatAuthClient {
	return wechatClient
}
