package gmail

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
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
	config, err := google.ConfigFromJSON(credentials, gmail.GmailReadonlyScope, gmail.GmailSendScope, gmail.GmailModifyScope)
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
	if !c.IsAuthenticated() {
		return nil, fmt.Errorf("client not authenticated")
	}

	// Query for unread messages
	query := "is:unread"

	// List messages
	call := c.service.Users.Messages.List("me").Q(query)
	response, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %v", err)
	}

	var messages []*mail.Message

	// Process each message
	for _, msg := range response.Messages {
		// Get full message details
		message, err := c.service.Users.Messages.Get("me", msg.Id).Format("full").Do()
		if err != nil {
			continue // Skip messages that can't be retrieved
		}

		// Parse message headers
		subject := ""
		from := ""
		to := ""
		date := ""

		for _, header := range message.Payload.Headers {
			switch header.Name {
			case "Subject":
				subject = header.Value
			case "From":
				from = header.Value
			case "To":
				to = header.Value
			case "Date":
				date = header.Value
			}
		}

		// Extract body
		body := c.extractBody(message.Payload)

		mailMessage := &mail.Message{
			ID:      message.Id,
			Subject: subject,
			From:    from,
			To:      to,
			Body:    body,
			Date:    date,
		}

		messages = append(messages, mailMessage)
	}

	return messages, nil
}

// extractBody extracts the text body from Gmail message payload
func (c *Client) extractBody(payload *gmail.MessagePart) string {
	if payload == nil {
		return ""
	}

	// If payload has body data, decode it
	if payload.Body != nil && payload.Body.Data != "" {
		decoded, err := base64.URLEncoding.DecodeString(payload.Body.Data)
		if err != nil {
			// If URL encoding fails, try standard base64
			decoded, err = base64.StdEncoding.DecodeString(payload.Body.Data)
			if err != nil {
				// If decoding fails, return the raw data
				return payload.Body.Data
			}
		}
		return string(decoded)
	}

	// If payload has parts, look for text/plain part
	if payload.Parts != nil {
		for _, part := range payload.Parts {
			if part.MimeType == "text/plain" {
				if part.Body != nil && part.Body.Data != "" {
					decoded, err := base64.URLEncoding.DecodeString(part.Body.Data)
					if err != nil {
						// If URL encoding fails, try standard base64
						decoded, err = base64.StdEncoding.DecodeString(part.Body.Data)
						if err != nil {
							// If decoding fails, return the raw data
							return part.Body.Data
						}
					}
					return string(decoded)
				}
			}
		}
	}

	return ""
}

func (c *Client) ParseSubject(subject string) (keyword string, targetName string, ok bool) {
	// Parse email subject format: "keyword - target_name"
	// Example: "重要通知 - 张三" or "urgent - john"

	if subject == "" {
		return "", "", false
	}

	// Find the " - " separator (space-dash-space)
	separator := "-"
	index := strings.Index(subject, separator)
	if index == -1 {
		return "", "", false
	}

	// Extract keyword and target name
	keyword = strings.TrimSpace(subject[:index])
	targetName = strings.TrimSpace(subject[index+len(separator):])

	// Return true if both keyword and targetName are found and not empty
	return keyword, targetName, keyword != "" && targetName != ""
}

func (c *Client) SendForward(original *mail.Message, toEmail string) error {
	if !c.IsAuthenticated() {
		return fmt.Errorf("client not authenticated")
	}

	if original == nil {
		return fmt.Errorf("original message is nil")
	}

	if toEmail == "" {
		return fmt.Errorf("target email is empty")
	}

	// Create email message in RFC 2822 format
	// Note: Gmail API will automatically set the From field to the authenticated user's email
	message := fmt.Sprintf("To: %s\r\n", toEmail)
	message += fmt.Sprintf("Subject: Fwd: %s\r\n", original.Subject)
	message += "\r\n"
	message += "---------- Forwarded message ----------\r\n"
	message += fmt.Sprintf("From: %s\r\n", original.From)
	message += fmt.Sprintf("Date: %s\r\n", original.Date)
	message += fmt.Sprintf("Subject: %s\r\n", original.Subject)
	message += "\r\n"
	message += original.Body

	// Encode message in base64
	encodedMessage := base64.URLEncoding.EncodeToString([]byte(message))

	// Create Gmail message
	gmailMessage := &gmail.Message{
		Raw: encodedMessage,
	}

	// Send the message
	_, err := c.service.Users.Messages.Send("me", gmailMessage).Do()
	if err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}

	return nil
}

func (c *Client) MarkAsRead(messageID string) error {
	if !c.IsAuthenticated() {
		return fmt.Errorf("client not authenticated")
	}

	if messageID == "" {
		return fmt.Errorf("message ID is empty")
	}

	// Create modification request to remove UNREAD label
	modifyRequest := &gmail.ModifyMessageRequest{
		RemoveLabelIds: []string{"UNREAD"},
	}

	// Apply the modification
	_, err := c.service.Users.Messages.Modify("me", messageID, modifyRequest).Do()
	if err != nil {
		return fmt.Errorf("failed to mark message as read: %v", err)
	}

	return nil
}
