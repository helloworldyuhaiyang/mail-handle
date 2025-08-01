package gmail

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"

	"github.com/helloworldyuhaiyang/mail-handle/internal/mail"
)

type Client struct {
	service         *gmail.Service
	config          *oauth2.Config
	tokenFile       string
	credentialsFile string
}

// Token represents the OAuth2 token structure
type Token struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
}

func NewClient(ctx context.Context, credentialsFile, tokenFile string) (*Client, error) {
	// Read credentials file
	credentials, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials file: %v", err)
	}

	// Create OAuth2 config
	config, err := google.ConfigFromJSON(credentials, gmail.GmailReadonlyScope, gmail.GmailSendScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}

	client := &Client{
		config:          config,
		tokenFile:       tokenFile,
		credentialsFile: credentialsFile,
	}

	// Try to load existing token
	token, err := client.loadToken()
	if err != nil {
		// Token file doesn't exist or is invalid, will need to authenticate
		return client, nil
	}

	// Create Gmail service with token
	service, err := gmail.NewService(ctx, option.WithTokenSource(config.TokenSource(ctx, token)))
	if err != nil {
		return nil, fmt.Errorf("unable to create Gmail service: %v", err)
	}

	client.service = service
	return client, nil
}

func (c *Client) Close() error {
	return nil
}

// GetAuthURL returns the authorization URL for OAuth2 flow
func (c *Client) GetAuthURL() string {
	return c.config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
}

// Callback handles the OAuth2 callback with authorization code
func (c *Client) Callback(code string) error {
	ctx := context.Background()

	// Exchange authorization code for token
	token, err := c.config.Exchange(ctx, code)
	if err != nil {
		return fmt.Errorf("unable to retrieve token from web: %v", err)
	}

	// Save token to file
	if err := c.saveToken(token); err != nil {
		return fmt.Errorf("unable to save token: %v", err)
	}

	// Create Gmail service with new token
	service, err := gmail.NewService(ctx, option.WithTokenSource(c.config.TokenSource(ctx, token)))
	if err != nil {
		return fmt.Errorf("unable to create Gmail service: %v", err)
	}

	c.service = service
	return nil
}

// loadToken loads token from file
func (c *Client) loadToken() (*oauth2.Token, error) {
	file, err := os.Open(c.tokenFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var token Token
	if err := json.NewDecoder(file).Decode(&token); err != nil {
		return nil, err
	}

	return &oauth2.Token{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	}, nil
}

// saveToken saves token to file
func (c *Client) saveToken(token *oauth2.Token) error {
	file, err := os.OpenFile(c.tokenFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	tokenData := Token{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	}

	return json.NewEncoder(file).Encode(tokenData)
}

// IsAuthenticated checks if the client is authenticated
func (c *Client) IsAuthenticated() bool {
	return c.service != nil
}

// AuthenticateIfNeeded checks if authentication is needed and provides instructions
func (c *Client) AuthenticateIfNeeded() (bool, string) {
	if c.IsAuthenticated() {
		return false, ""
	}

	authURL := c.GetAuthURL()
	return true, fmt.Sprintf("Please visit this URL to authorize the application: %s", authURL)
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
