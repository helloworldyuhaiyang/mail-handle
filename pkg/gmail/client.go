package gmail

import (
	"context"

	"github.com/helloworldyuhaiyang/mail-handle/internal/mail"
)

type Client struct {
}

func NewClient(ctx context.Context, credentialsFile, tokenFile string) (*Client, error) {
	return &Client{}, nil
}

func (c *Client) Close() error {
	return nil
}

// 获取授权链接
func (c *Client) GetAuthURL() string {
	return ""
}

// 完成 oauth2.0
func (c *Client) Callback(code string) error {
	return nil
}

func (c *Client) FetchUnreadMessages() ([]*mail.Message, error) {
	panic("not implemented") // TODO: Implement
}

func (c *Client) ParseSubject(subject string) (keyword string, targetName string, ok bool) {
	panic("not implemented") // TODO: Implement
}

func (c *Client) SendForward(original *mail.Message, toEmail string) error {
	panic("not implemented") // TODO: Implement
}

func (c *Client) MarkAsRead(messageID string) error {
	panic("not implemented") // TODO: Implement
}
