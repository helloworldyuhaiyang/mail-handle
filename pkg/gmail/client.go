package gmail

import (
	"context"
)

type Client struct {
}

type Email struct {
	ID      string
	Subject string
	From    string
	To      string
	Body    string
	Date    string
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

// 获取所有指定关键字邮件
func (c *Client) GetAllEmail(ctx context.Context, keyword string) ([]Email, error) {
	return nil, nil
}

// 转发邮件
func (c *Client) ForwardEmail(ctx context.Context, emailID string, to string) error {
	return nil
}

// 标记邮件为已读
func (c *Client) MarkEmailAsRead(ctx context.Context, emailID string) error {
	return nil
}
