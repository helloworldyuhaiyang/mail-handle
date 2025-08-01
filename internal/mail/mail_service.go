package mail

// 邮件信息
type Message struct {
	ID      string
	Subject string
	From    string
	To      string
	Body    string
	Date    string
}

type MailService interface {
	FetchUnreadMessages() ([]*Message, error)
	ParseSubject(subject string) (keyword string, targetName string, ok bool)
	SendForward(original *Message, toEmail string) error
	MarkAsRead(messageID string) error
}
