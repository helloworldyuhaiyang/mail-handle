package schedule

import (
	"github.com/helloworldyuhaiyang/mail-handle/internal/mail"
	"github.com/helloworldyuhaiyang/mail-handle/internal/repo"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

type Scheduler struct {
	cron *cron.Cron

	// 邮件(相关的抽象接口)
	mailService mail.MailService
	// 转发对象查询(抽象接口)
	targetRepo repo.ForwardTargetRepo
}

func NewScheduler(mailService mail.MailService, targetRepo repo.ForwardTargetRepo) *Scheduler {
	cron := cron.New(cron.WithSeconds())
	return &Scheduler{
		mailService: mailService,
		targetRepo:  targetRepo,
		cron:        cron,
	}
}

func (s *Scheduler) Start() error {
	s.cron.AddFunc("@every 1m", s.run)
	s.cron.Start()
	return nil
}

func (s *Scheduler) Stop() error {
	ctx := s.cron.Stop()
	<-ctx.Done()
	return nil
}

// 定时任务
// 获取所有未读邮件
// 抽取关键词与转发对象名
// 通过转发对象名查询邮箱地址
// 构造转发邮件内容
// 使用 Gmail API 进行转发
// 标记邮件为已读(转发的邮件)
func (s *Scheduler) run() {
	// 获取所有未读邮件
	messages, err := s.mailService.FetchUnreadMessages()
	if err != nil {
		logrus.Errorf("Failed to fetch unread emails: %v", err)
		return
	}

	for _, msg := range messages {
		subject := msg.Subject
		// 抽取关键词与转发对象名
		keyword, targetName, ok := s.mailService.ParseSubject(subject)
		if !ok {
			continue // 格式不符，跳过
		}

		logrus.Infof("Found keyword and target name: %s, %s", keyword, targetName)

		// 通过转发对象名查询邮箱地址
		targetEmail, err := s.targetRepo.FindEmailByName(targetName)
		if err != nil {
			logrus.Errorf("Failed to find target email: %v", err)
			continue
		}

		// 构造转发邮件内容
		if err := s.mailService.SendForward(msg, targetEmail); err != nil {
			logrus.Errorf("Failed to send forward: %v", err)
			continue
		}

		// 标记邮件为已读(转发的邮件)
		if err := s.mailService.MarkAsRead(msg.ID); err != nil {
			logrus.Errorf("Failed to mark email as read: %v", err)
		}
	}
}
