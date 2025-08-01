package schedule

import (
	"errors"
	"testing"
	"time"

	"github.com/helloworldyuhaiyang/mail-handle/internal/mail"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMailService is a mock implementation of mail.MailService
type MockMailService struct {
	mock.Mock
}

func (m *MockMailService) FetchUnreadMessages() ([]*mail.Message, error) {
	args := m.Called()
	return args.Get(0).([]*mail.Message), args.Error(1)
}

func (m *MockMailService) ParseSubject(subject string) (keyword string, targetName string, ok bool) {
	args := m.Called(subject)
	return args.String(0), args.String(1), args.Bool(2)
}

func (m *MockMailService) SendForward(original *mail.Message, toEmail string) error {
	args := m.Called(original, toEmail)
	return args.Error(0)
}

func (m *MockMailService) MarkAsRead(messageID string) error {
	args := m.Called(messageID)
	return args.Error(0)
}

// MockForwardTargetRepo is a mock implementation of repo.ForwardTargetRepo
type MockForwardTargetRepo struct {
	mock.Mock
}

func (m *MockForwardTargetRepo) FindEmailByName(targetName string) (email string, err error) {
	args := m.Called(targetName)
	return args.String(0), args.Error(1)
}

func TestNewScheduler(t *testing.T) {
	config := &SchedulerConfig{
		FetchInterval: "*/30 * * * * *",
	}

	mockMailService := &MockMailService{}
	mockTargetRepo := &MockForwardTargetRepo{}

	scheduler := NewScheduler(config, mockMailService, mockTargetRepo)

	assert.NotNil(t, scheduler)
	assert.Equal(t, config, scheduler.config)
	assert.Equal(t, mockMailService, scheduler.mailService)
	assert.Equal(t, mockTargetRepo, scheduler.targetRepo)
	assert.NotNil(t, scheduler.cron)
}

func TestScheduler_Start(t *testing.T) {
	config := &SchedulerConfig{
		FetchInterval: "*/30 * * * * *",
	}

	mockMailService := &MockMailService{}
	mockTargetRepo := &MockForwardTargetRepo{}

	scheduler := NewScheduler(config, mockMailService, mockTargetRepo)

	err := scheduler.Start()
	assert.NoError(t, err)

	// Clean up
	err = scheduler.Stop()
	assert.NoError(t, err)
}

func TestScheduler_Stop(t *testing.T) {
	config := &SchedulerConfig{
		FetchInterval: "*/30 * * * * *",
	}

	mockMailService := &MockMailService{}
	mockTargetRepo := &MockForwardTargetRepo{}

	scheduler := NewScheduler(config, mockMailService, mockTargetRepo)

	// Start first
	err := scheduler.Start()
	assert.NoError(t, err)

	// Then stop
	err = scheduler.Stop()
	assert.NoError(t, err)
}

func TestScheduler_Run_SuccessfulProcessing(t *testing.T) {
	config := &SchedulerConfig{
		FetchInterval: "*/30 * * * * *",
	}

	mockMailService := &MockMailService{}
	mockTargetRepo := &MockForwardTargetRepo{}

	scheduler := NewScheduler(config, mockMailService, mockTargetRepo)

	// Test data
	messages := []*mail.Message{
		{
			ID:      "msg1",
			Subject: "URGENT: Project Update",
			From:    "sender@example.com",
			To:      "me@example.com",
			Body:    "Important project update",
			Date:    "2024-01-01T10:00:00Z",
		},
		{
			ID:      "msg2",
			Subject: "INFO: Weekly Report",
			From:    "reports@example.com",
			To:      "me@example.com",
			Body:    "Weekly report attached",
			Date:    "2024-01-01T11:00:00Z",
		},
	}

	// Setup expectations
	mockMailService.On("FetchUnreadMessages").Return(messages, nil)
	mockMailService.On("ParseSubject", "URGENT: Project Update").Return("URGENT", "Project", true)
	mockMailService.On("ParseSubject", "INFO: Weekly Report").Return("INFO", "Weekly", true)
	mockTargetRepo.On("FindEmailByName", "Project").Return("project@company.com", nil)
	mockTargetRepo.On("FindEmailByName", "Weekly").Return("weekly@company.com", nil)
	mockMailService.On("SendForward", messages[0], "project@company.com").Return(nil)
	mockMailService.On("SendForward", messages[1], "weekly@company.com").Return(nil)
	mockMailService.On("MarkAsRead", "msg1").Return(nil)
	mockMailService.On("MarkAsRead", "msg2").Return(nil)

	// Execute
	scheduler.run()

	// Verify all expectations
	mockMailService.AssertExpectations(t)
	mockTargetRepo.AssertExpectations(t)
}

func TestScheduler_Run_FetchUnreadMessagesError(t *testing.T) {
	config := &SchedulerConfig{
		FetchInterval: "*/30 * * * * *",
	}

	mockMailService := &MockMailService{}
	mockTargetRepo := &MockForwardTargetRepo{}

	scheduler := NewScheduler(config, mockMailService, mockTargetRepo)

	// Setup expectations - return error
	mockMailService.On("FetchUnreadMessages").Return([]*mail.Message{}, errors.New("fetch error"))

	// Execute
	scheduler.run()

	// Verify expectations
	mockMailService.AssertExpectations(t)
	mockTargetRepo.AssertNotCalled(t, "FindEmailByName")
}

func TestScheduler_Run_ParseSubjectFailure(t *testing.T) {
	config := &SchedulerConfig{
		FetchInterval: "*/30 * * * * *",
	}

	mockMailService := &MockMailService{}
	mockTargetRepo := &MockForwardTargetRepo{}

	scheduler := NewScheduler(config, mockMailService, mockTargetRepo)

	// Test data
	messages := []*mail.Message{
		{
			ID:      "msg1",
			Subject: "Invalid Subject Format",
			From:    "sender@example.com",
			To:      "me@example.com",
			Body:    "Invalid format",
			Date:    "2024-01-01T10:00:00Z",
		},
	}

	// Setup expectations - parse subject returns false
	mockMailService.On("FetchUnreadMessages").Return(messages, nil)
	mockMailService.On("ParseSubject", "Invalid Subject Format").Return("", "", false)

	// Execute
	scheduler.run()

	// Verify expectations
	mockMailService.AssertExpectations(t)
	mockTargetRepo.AssertNotCalled(t, "FindEmailByName")
	mockMailService.AssertNotCalled(t, "SendForward")
	mockMailService.AssertNotCalled(t, "MarkAsRead")
}

func TestScheduler_Run_FindEmailByNameError(t *testing.T) {
	config := &SchedulerConfig{
		FetchInterval: "*/30 * * * * *",
	}

	mockMailService := &MockMailService{}
	mockTargetRepo := &MockForwardTargetRepo{}

	scheduler := NewScheduler(config, mockMailService, mockTargetRepo)

	// Test data
	messages := []*mail.Message{
		{
			ID:      "msg1",
			Subject: "URGENT: Project Update",
			From:    "sender@example.com",
			To:      "me@example.com",
			Body:    "Important project update",
			Date:    "2024-01-01T10:00:00Z",
		},
	}

	// Setup expectations
	mockMailService.On("FetchUnreadMessages").Return(messages, nil)
	mockMailService.On("ParseSubject", "URGENT: Project Update").Return("URGENT", "Project", true)
	mockTargetRepo.On("FindEmailByName", "Project").Return("", errors.New("target not found"))

	// Execute
	scheduler.run()

	// Verify expectations
	mockMailService.AssertExpectations(t)
	mockTargetRepo.AssertExpectations(t)
	mockMailService.AssertNotCalled(t, "SendForward")
	mockMailService.AssertNotCalled(t, "MarkAsRead")
}

func TestScheduler_Run_SendForwardError(t *testing.T) {
	config := &SchedulerConfig{
		FetchInterval: "*/30 * * * * *",
	}

	mockMailService := &MockMailService{}
	mockTargetRepo := &MockForwardTargetRepo{}

	scheduler := NewScheduler(config, mockMailService, mockTargetRepo)

	// Test data
	messages := []*mail.Message{
		{
			ID:      "msg1",
			Subject: "URGENT: Project Update",
			From:    "sender@example.com",
			To:      "me@example.com",
			Body:    "Important project update",
			Date:    "2024-01-01T10:00:00Z",
		},
	}

	// Setup expectations
	mockMailService.On("FetchUnreadMessages").Return(messages, nil)
	mockMailService.On("ParseSubject", "URGENT: Project Update").Return("URGENT", "Project", true)
	mockTargetRepo.On("FindEmailByName", "Project").Return("project@company.com", nil)
	mockMailService.On("SendForward", messages[0], "project@company.com").Return(errors.New("send error"))

	// Execute
	scheduler.run()

	// Verify expectations
	mockMailService.AssertExpectations(t)
	mockTargetRepo.AssertExpectations(t)
	mockMailService.AssertNotCalled(t, "MarkAsRead")
}

func TestScheduler_Run_MarkAsReadError(t *testing.T) {
	config := &SchedulerConfig{
		FetchInterval: "*/30 * * * * *",
	}

	mockMailService := &MockMailService{}
	mockTargetRepo := &MockForwardTargetRepo{}

	scheduler := NewScheduler(config, mockMailService, mockTargetRepo)

	// Test data
	messages := []*mail.Message{
		{
			ID:      "msg1",
			Subject: "URGENT: Project Update",
			From:    "sender@example.com",
			To:      "me@example.com",
			Body:    "Important project update",
			Date:    "2024-01-01T10:00:00Z",
		},
	}

	// Setup expectations
	mockMailService.On("FetchUnreadMessages").Return(messages, nil)
	mockMailService.On("ParseSubject", "URGENT: Project Update").Return("URGENT", "Project", true)
	mockTargetRepo.On("FindEmailByName", "Project").Return("project@company.com", nil)
	mockMailService.On("SendForward", messages[0], "project@company.com").Return(nil)
	mockMailService.On("MarkAsRead", "msg1").Return(errors.New("mark as read error"))

	// Execute
	scheduler.run()

	// Verify expectations
	mockMailService.AssertExpectations(t)
	mockTargetRepo.AssertExpectations(t)
}

func TestScheduler_Run_MixedSuccessAndFailure(t *testing.T) {
	config := &SchedulerConfig{
		FetchInterval: "*/30 * * * * *",
	}

	mockMailService := &MockMailService{}
	mockTargetRepo := &MockForwardTargetRepo{}

	scheduler := NewScheduler(config, mockMailService, mockTargetRepo)

	// Test data - mix of valid and invalid messages
	messages := []*mail.Message{
		{
			ID:      "msg1",
			Subject: "URGENT: Project Update",
			From:    "sender@example.com",
			To:      "me@example.com",
			Body:    "Important project update",
			Date:    "2024-01-01T10:00:00Z",
		},
		{
			ID:      "msg2",
			Subject: "Invalid Format",
			From:    "sender@example.com",
			To:      "me@example.com",
			Body:    "Invalid format",
			Date:    "2024-01-01T11:00:00Z",
		},
		{
			ID:      "msg3",
			Subject: "INFO: Weekly Report",
			From:    "reports@example.com",
			To:      "me@example.com",
			Body:    "Weekly report",
			Date:    "2024-01-01T12:00:00Z",
		},
	}

	// Setup expectations
	mockMailService.On("FetchUnreadMessages").Return(messages, nil)
	mockMailService.On("ParseSubject", "URGENT: Project Update").Return("URGENT", "Project", true)
	mockMailService.On("ParseSubject", "Invalid Format").Return("", "", false)
	mockMailService.On("ParseSubject", "INFO: Weekly Report").Return("INFO", "Weekly", true)
	mockTargetRepo.On("FindEmailByName", "Project").Return("project@company.com", nil)
	mockTargetRepo.On("FindEmailByName", "Weekly").Return("weekly@company.com", nil)
	mockMailService.On("SendForward", messages[0], "project@company.com").Return(nil)
	mockMailService.On("SendForward", messages[2], "weekly@company.com").Return(nil)
	mockMailService.On("MarkAsRead", "msg1").Return(nil)
	mockMailService.On("MarkAsRead", "msg3").Return(nil)

	// Execute
	scheduler.run()

	// Verify expectations
	mockMailService.AssertExpectations(t)
	mockTargetRepo.AssertExpectations(t)

	// Verify that FindEmailByName was called only for valid messages
	mockTargetRepo.AssertNumberOfCalls(t, "FindEmailByName", 2)
}

func TestScheduler_Run_EmptyMessages(t *testing.T) {
	config := &SchedulerConfig{
		FetchInterval: "*/30 * * * * *",
	}

	mockMailService := &MockMailService{}
	mockTargetRepo := &MockForwardTargetRepo{}

	scheduler := NewScheduler(config, mockMailService, mockTargetRepo)

	// Setup expectations - return empty messages
	mockMailService.On("FetchUnreadMessages").Return([]*mail.Message{}, nil)

	// Execute
	scheduler.run()

	// Verify expectations
	mockMailService.AssertExpectations(t)
	mockTargetRepo.AssertNotCalled(t, "FindEmailByName")
	mockMailService.AssertNotCalled(t, "SendForward")
	mockMailService.AssertNotCalled(t, "MarkAsRead")
}

func TestScheduler_Run_PanicInFetchUnreadMessages(t *testing.T) {
	config := &SchedulerConfig{
		FetchInterval: "*/30 * * * * *",
	}

	mockMailService := &MockMailService{}
	mockTargetRepo := &MockForwardTargetRepo{}

	scheduler := NewScheduler(config, mockMailService, mockTargetRepo)

	// Setup expectations - FetchUnreadMessages will panic
	mockMailService.On("FetchUnreadMessages").Panic("panic in fetch unread messages")

	// Execute - should not panic due to recovery mechanism
	scheduler.run()

	// Verify expectations
	mockMailService.AssertExpectations(t)
	mockTargetRepo.AssertNotCalled(t, "FindEmailByName")
}

func TestScheduler_Run_PanicInParseSubject(t *testing.T) {
	config := &SchedulerConfig{
		FetchInterval: "*/30 * * * * *",
	}

	mockMailService := &MockMailService{}
	mockTargetRepo := &MockForwardTargetRepo{}

	scheduler := NewScheduler(config, mockMailService, mockTargetRepo)

	// Test data
	messages := []*mail.Message{
		{
			ID:      "msg1",
			Subject: "URGENT: Project Update",
			From:    "sender@example.com",
			To:      "me@example.com",
			Body:    "Important project update",
			Date:    "2024-01-01T10:00:00Z",
		},
		{
			ID:      "msg2",
			Subject: "INFO: Weekly Report",
			From:    "reports@example.com",
			To:      "me@example.com",
			Body:    "Weekly report attached",
			Date:    "2024-01-01T11:00:00Z",
		},
	}

	// Setup expectations - ParseSubject will panic for first message, but second should still be processed
	mockMailService.On("FetchUnreadMessages").Return(messages, nil)
	mockMailService.On("ParseSubject", "URGENT: Project Update").Panic("panic in parse subject")
	mockMailService.On("ParseSubject", "INFO: Weekly Report").Return("INFO", "Weekly", true)
	mockTargetRepo.On("FindEmailByName", "Weekly").Return("weekly@company.com", nil)
	mockMailService.On("SendForward", messages[1], "weekly@company.com").Return(nil)
	mockMailService.On("MarkAsRead", "msg2").Return(nil)

	// Execute - should not panic due to recovery mechanism
	scheduler.run()

	// Verify expectations
	mockMailService.AssertExpectations(t)
	mockTargetRepo.AssertExpectations(t)

	// Verify that FindEmailByName was called only for the second message
	mockTargetRepo.AssertNumberOfCalls(t, "FindEmailByName", 1)
}

func TestScheduler_Run_PanicInFindEmailByName(t *testing.T) {
	config := &SchedulerConfig{
		FetchInterval: "*/30 * * * * *",
	}

	mockMailService := &MockMailService{}
	mockTargetRepo := &MockForwardTargetRepo{}

	scheduler := NewScheduler(config, mockMailService, mockTargetRepo)

	// Test data
	messages := []*mail.Message{
		{
			ID:      "msg1",
			Subject: "URGENT: Project Update",
			From:    "sender@example.com",
			To:      "me@example.com",
			Body:    "Important project update",
			Date:    "2024-01-01T10:00:00Z",
		},
		{
			ID:      "msg2",
			Subject: "INFO: Weekly Report",
			From:    "reports@example.com",
			To:      "me@example.com",
			Body:    "Weekly report attached",
			Date:    "2024-01-01T11:00:00Z",
		},
	}

	// Setup expectations - FindEmailByName will panic for first message, but second should still be processed
	mockMailService.On("FetchUnreadMessages").Return(messages, nil)
	mockMailService.On("ParseSubject", "URGENT: Project Update").Return("URGENT", "Project", true)
	mockMailService.On("ParseSubject", "INFO: Weekly Report").Return("INFO", "Weekly", true)
	mockTargetRepo.On("FindEmailByName", "Project").Panic("panic in find email by name")
	mockTargetRepo.On("FindEmailByName", "Weekly").Return("weekly@company.com", nil)
	mockMailService.On("SendForward", messages[1], "weekly@company.com").Return(nil)
	mockMailService.On("MarkAsRead", "msg2").Return(nil)

	// Execute - should not panic due to recovery mechanism
	scheduler.run()

	// Verify expectations
	mockMailService.AssertExpectations(t)
	mockTargetRepo.AssertExpectations(t)

	// Verify that SendForward was called only for the second message
	mockMailService.AssertNumberOfCalls(t, "SendForward", 1)
}

func TestScheduler_Run_PanicInSendForward(t *testing.T) {
	config := &SchedulerConfig{
		FetchInterval: "*/30 * * * * *",
	}

	mockMailService := &MockMailService{}
	mockTargetRepo := &MockForwardTargetRepo{}

	scheduler := NewScheduler(config, mockMailService, mockTargetRepo)

	// Test data
	messages := []*mail.Message{
		{
			ID:      "msg1",
			Subject: "URGENT: Project Update",
			From:    "sender@example.com",
			To:      "me@example.com",
			Body:    "Important project update",
			Date:    "2024-01-01T10:00:00Z",
		},
		{
			ID:      "msg2",
			Subject: "INFO: Weekly Report",
			From:    "reports@example.com",
			To:      "me@example.com",
			Body:    "Weekly report attached",
			Date:    "2024-01-01T11:00:00Z",
		},
	}

	// Setup expectations - SendForward will panic for first message, but second should still be processed
	mockMailService.On("FetchUnreadMessages").Return(messages, nil)
	mockMailService.On("ParseSubject", "URGENT: Project Update").Return("URGENT", "Project", true)
	mockMailService.On("ParseSubject", "INFO: Weekly Report").Return("INFO", "Weekly", true)
	mockTargetRepo.On("FindEmailByName", "Project").Return("project@company.com", nil)
	mockTargetRepo.On("FindEmailByName", "Weekly").Return("weekly@company.com", nil)
	mockMailService.On("SendForward", messages[0], "project@company.com").Panic("panic in send forward")
	mockMailService.On("SendForward", messages[1], "weekly@company.com").Return(nil)
	mockMailService.On("MarkAsRead", "msg2").Return(nil)

	// Execute - should not panic due to recovery mechanism
	scheduler.run()

	// Verify expectations
	mockMailService.AssertExpectations(t)
	mockTargetRepo.AssertExpectations(t)

	// Verify that MarkAsRead was called only for the second message
	mockMailService.AssertNumberOfCalls(t, "MarkAsRead", 1)
}

func TestScheduler_Run_PanicInMarkAsRead(t *testing.T) {
	config := &SchedulerConfig{
		FetchInterval: "*/30 * * * * *",
	}

	mockMailService := &MockMailService{}
	mockTargetRepo := &MockForwardTargetRepo{}

	scheduler := NewScheduler(config, mockMailService, mockTargetRepo)

	// Test data
	messages := []*mail.Message{
		{
			ID:      "msg1",
			Subject: "URGENT: Project Update",
			From:    "sender@example.com",
			To:      "me@example.com",
			Body:    "Important project update",
			Date:    "2024-01-01T10:00:00Z",
		},
		{
			ID:      "msg2",
			Subject: "INFO: Weekly Report",
			From:    "reports@example.com",
			To:      "me@example.com",
			Body:    "Weekly report attached",
			Date:    "2024-01-01T11:00:00Z",
		},
	}

	// Setup expectations - MarkAsRead will panic for first message, but second should still be processed
	mockMailService.On("FetchUnreadMessages").Return(messages, nil)
	mockMailService.On("ParseSubject", "URGENT: Project Update").Return("URGENT", "Project", true)
	mockMailService.On("ParseSubject", "INFO: Weekly Report").Return("INFO", "Weekly", true)
	mockTargetRepo.On("FindEmailByName", "Project").Return("project@company.com", nil)
	mockTargetRepo.On("FindEmailByName", "Weekly").Return("weekly@company.com", nil)
	mockMailService.On("SendForward", messages[0], "project@company.com").Return(nil)
	mockMailService.On("SendForward", messages[1], "weekly@company.com").Return(nil)
	mockMailService.On("MarkAsRead", "msg1").Panic("panic in mark as read")
	mockMailService.On("MarkAsRead", "msg2").Return(nil)

	// Execute - should not panic due to recovery mechanism
	scheduler.run()

	// Verify expectations
	mockMailService.AssertExpectations(t)
	mockTargetRepo.AssertExpectations(t)

	// Verify that MarkAsRead was called for both messages (first one panicked but second succeeded)
	mockMailService.AssertNumberOfCalls(t, "MarkAsRead", 2)
}

func TestScheduler_Integration_StartStop(t *testing.T) {
	config := &SchedulerConfig{
		FetchInterval: "*/30 * * * * *",
	}

	mockMailService := &MockMailService{}
	mockTargetRepo := &MockForwardTargetRepo{}

	scheduler := NewScheduler(config, mockMailService, mockTargetRepo)

	// Test start
	err := scheduler.Start()
	assert.NoError(t, err)

	// Wait a bit to ensure cron is running
	time.Sleep(100 * time.Millisecond)

	// Test stop
	err = scheduler.Stop()
	assert.NoError(t, err)
}
