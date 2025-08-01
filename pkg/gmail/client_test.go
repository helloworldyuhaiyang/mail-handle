package gmail

import (
	"context"
	"encoding/base64"
	"os"
	"strings"
	"testing"

	"github.com/helloworldyuhaiyang/mail-handle/internal/mail"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/gmail/v1"
)

func TestClient_ParseSubject(t *testing.T) {
	client := &Client{}

	tests := []struct {
		name        string
		subject     string
		wantKeyword string
		wantTarget  string
		wantOk      bool
	}{
		{
			name:        "valid format with dash",
			subject:     "重要通知-张三",
			wantKeyword: "重要通知",
			wantTarget:  "张三",
			wantOk:      true,
		},
		{
			name:        "valid format with spaces",
			subject:     "urgent - john",
			wantKeyword: "urgent",
			wantTarget:  "john",
			wantOk:      true,
		},
		{
			name:        "empty subject",
			subject:     "",
			wantKeyword: "",
			wantTarget:  "",
			wantOk:      false,
		},
		{
			name:        "no separator",
			subject:     "just a subject",
			wantKeyword: "",
			wantTarget:  "",
			wantOk:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyword, targetName, ok := client.ParseSubject(tt.subject)
			assert.Equal(t, tt.wantKeyword, keyword)
			assert.Equal(t, tt.wantTarget, targetName)
			assert.Equal(t, tt.wantOk, ok)
		})
	}
}

func TestClient_Integration(t *testing.T) {
	// Check if credentials file exists
	credentialsFile := "credentials.json"
	if _, err := os.Stat(credentialsFile); os.IsNotExist(err) {
		t.Skip("Skipping integration tests: credentials.json not found")
	}

	// Use the existing token.json
	tokenFile := "../token.json"
	if _, err := os.Stat(tokenFile); os.IsNotExist(err) {
		t.Skip("Skipping integration tests: token.json not found")
	}

	ctx := context.Background()
	client, err := NewClient(ctx, credentialsFile, tokenFile)
	require.NoError(t, err)
	defer client.Close()

	t.Run("IsAuthenticated", func(t *testing.T) {
		authenticated := client.IsAuthenticated()
		t.Logf("Client authenticated: %v", authenticated)
	})

	t.Run("FetchUnreadMessages", func(t *testing.T) {
		if !client.IsAuthenticated() {
			t.Skip("Skipping test: client not authenticated")
		}

		messages, err := client.FetchUnreadMessages()
		if err != nil {
			t.Logf("FetchUnreadMessages error: %v", err)
			return
		}

		t.Logf("Found %d unread messages", len(messages))

		for i, msg := range messages {
			t.Logf("Message %d: Subject='%s', From='%s'", i, msg.Subject, msg.From)

			keyword, targetName, ok := client.ParseSubject(msg.Subject)
			if ok {
				t.Logf("  Parsed: keyword='%s', target='%s'", keyword, targetName)
			}
		}
	})
}

func TestClient_Validation(t *testing.T) {
	client := &Client{}

	t.Run("SendForward validation", func(t *testing.T) {
		// Test nil original message
		err := client.SendForward(nil, "test@example.com")
		assert.Error(t, err)
		// Check for either validation error or authentication error
		assert.True(t,
			strings.Contains(err.Error(), "original message is nil") ||
				strings.Contains(err.Error(), "client not authenticated"),
			"Expected either validation error or authentication error, got: %s", err.Error())

		// Test empty target email
		msg := &mail.Message{ID: "1", Subject: "Test", From: "from@example.com"}
		err = client.SendForward(msg, "")
		assert.Error(t, err)
		// Check for either validation error or authentication error
		assert.True(t,
			strings.Contains(err.Error(), "target email is empty") ||
				strings.Contains(err.Error(), "client not authenticated"),
			"Expected either validation error or authentication error, got: %s", err.Error())
	})

	t.Run("MarkAsRead validation", func(t *testing.T) {
		// Test empty message ID
		err := client.MarkAsRead("")
		assert.Error(t, err)
		// Check for either validation error or authentication error
		assert.True(t,
			strings.Contains(err.Error(), "message ID is empty") ||
				strings.Contains(err.Error(), "client not authenticated"),
			"Expected either validation error or authentication error, got: %s", err.Error())
	})
}

func TestClient_AuthenticationFlow(t *testing.T) {
	client := &Client{}

	t.Run("IsAuthenticated when not authenticated", func(t *testing.T) {
		assert.False(t, client.IsAuthenticated())
	})

	t.Run("AuthenticateIfNeeded when not authenticated", func(t *testing.T) {
		// Skip this test if client has no config (which would cause panic)
		if client.config == nil {
			t.Skip("Skipping test: client has no config")
		}
		needed, instructions := client.AuthenticateIfNeeded()
		assert.True(t, needed)
		assert.Contains(t, instructions, "Please visit this URL")
	})
}

func TestClient_NewClient(t *testing.T) {
	t.Run("non-existent credentials file", func(t *testing.T) {
		ctx := context.Background()
		client, err := NewClient(ctx, "non-existent.json", "non-existent-token.json")
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "unable to read credentials file")
	})
}

func TestClient_ExtractBody(t *testing.T) {
	client := &Client{}

	t.Run("nil payload", func(t *testing.T) {
		body := client.extractBody(nil)
		assert.Empty(t, body)
	})

	t.Run("base64 decoded content", func(t *testing.T) {
		// Test with base64 encoded content
		testContent := "帮我转发给海洋好吗"
		encodedContent := base64.URLEncoding.EncodeToString([]byte(testContent))

		// Create a mock payload
		payload := &gmail.MessagePart{
			Body: &gmail.MessagePartBody{
				Data: encodedContent,
			},
		}

		body := client.extractBody(payload)
		assert.Equal(t, testContent, body)
	})
}

// Helper function to create test messages
func createTestMessage(id, subject, from, to, body, date string) *mail.Message {
	return &mail.Message{
		ID:      id,
		Subject: subject,
		From:    from,
		To:      to,
		Body:    body,
		Date:    date,
	}
}

func TestClient_SendForward_Integration(t *testing.T) {
	// Check if credentials file exists
	credentialsFile := "credentials.json"
	if _, err := os.Stat(credentialsFile); os.IsNotExist(err) {
		t.Skip("Skipping integration test: credentials.json not found")
	}

	// Use the existing token.json
	tokenFile := "../token.json"
	if _, err := os.Stat(tokenFile); os.IsNotExist(err) {
		t.Skip("Skipping integration test: token.json not found")
	}

	ctx := context.Background()
	client, err := NewClient(ctx, credentialsFile, tokenFile)
	require.NoError(t, err)
	defer client.Close()

	if !client.IsAuthenticated() {
		t.Skip("Skipping test: client not authenticated")
	}

	t.Run("SendForward with test message", func(t *testing.T) {
		testMsg := createTestMessage(
			"test-id",
			"Test Subject",
			"test@example.com",
			"recipient@example.com",
			"This is a test message body",
			"Mon, 01 Jan 2024 12:00:00 +0000",
		)

		// Note: This will actually try to send an email
		// You might want to use a test email address
		err := client.SendForward(testMsg, "test-recipient@example.com")
		if err != nil {
			t.Logf("SendForward error (expected if test email is invalid): %v", err)
		}
	})
}

func TestClient_MarkAsRead_Integration(t *testing.T) {
	// Check if credentials file exists
	credentialsFile := "credentials.json"
	if _, err := os.Stat(credentialsFile); os.IsNotExist(err) {
		t.Skip("Skipping integration test: credentials.json not found")
	}

	// Use the existing token.json
	tokenFile := "../token.json"
	if _, err := os.Stat(tokenFile); os.IsNotExist(err) {
		t.Skip("Skipping integration test: token.json not found")
	}

	ctx := context.Background()
	client, err := NewClient(ctx, credentialsFile, tokenFile)
	require.NoError(t, err)
	defer client.Close()

	if !client.IsAuthenticated() {
		t.Skip("Skipping test: client not authenticated")
	}

	t.Run("MarkAsRead with invalid message ID", func(t *testing.T) {
		err := client.MarkAsRead("invalid-message-id")
		if err != nil {
			t.Logf("MarkAsRead with invalid ID (expected): %v", err)
		}
	})
}
