package schedule

import (
	"runtime"

	"github.com/helloworldyuhaiyang/mail-handle/internal/mail"
	"github.com/helloworldyuhaiyang/mail-handle/internal/repo"
	"github.com/helloworldyuhaiyang/mail-handle/pkg"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

type SchedulerConfig struct {
	FetchInterval   string   `mapstructure:"fetch_interval"`
	ForwardKeywords []string `mapstructure:"forward_keywords"`
}

type Scheduler struct {
	config *SchedulerConfig
	cron   *cron.Cron

	// 邮件(相关的抽象接口)
	mailService mail.MailService
	// 转发对象查询(抽象接口)
	targetRepo repo.ForwardTargetRepo
}

func NewScheduler(config *SchedulerConfig, mailService mail.MailService, targetRepo repo.ForwardTargetRepo) *Scheduler {
	cron := cron.New(cron.WithSeconds())
	return &Scheduler{
		config:      config,
		mailService: mailService,
		targetRepo:  targetRepo,
		cron:        cron,
	}
}

func (s *Scheduler) Start() error {
	_, err := s.cron.AddFunc(s.config.FetchInterval, s.run)
	if err != nil {
		return err
	}

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
	// 添加 panic 恢复机制，防止程序崩溃
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("Scheduler run() panic recovered: %v", r)
			// 记录堆栈信息
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			logrus.Errorf("Stack trace: %s", buf[:n])
		}
	}()

	// 获取所有未读邮件
	messages, err := s.mailService.FetchUnreadMessages()
	if err != nil {
		logrus.Errorf("Failed to fetch unread emails: %v", err)
		return
	}

	logrus.Infof("Found unread emails number: %d", len(messages))

	for _, msg := range messages {
		// 为每个邮件处理添加单独的 panic 恢复
		func() {
			defer func() {
				if r := recover(); r != nil {
					logrus.Errorf("Panic while processing message %s: %v", msg.ID, r)
				}
			}()

			subject := msg.Subject
			// 抽取关键词与转发对象名
			keyword, targetName, ok := s.mailService.ParseSubject(subject)
			if !ok {
				return // 格式不符，跳过
			}
			if !pkg.StringSliceContains(s.config.ForwardKeywords, keyword) {
				return
			}

			logrus.Infof("Found keyword and target name:keyword: %s, target: %s", keyword, targetName)

			// 通过转发对象名查询邮箱地址
			targetEmail, err := s.targetRepo.FindEmailByName(targetName)
			if err != nil {
				logrus.Errorf("Failed to find target email: %v", err)
				return
			}

			// 构造转发邮件内容
			if err := s.mailService.SendForward(msg, targetEmail); err != nil {
				logrus.Errorf("Failed to send forward: %v", err)
				return
			} else {
				logrus.Infof("Sent forward, subject: %s, id: %s, target: %s", msg.Subject, msg.ID, targetEmail)
			}

			// 标记邮件为已读(转发的邮件)
			if err := s.mailService.MarkAsRead(msg.ID); err != nil {
				logrus.Errorf("Failed to mark email as read: %v", err)
			} else {
				logrus.Infof("Marked email as read, subject: %s, id: %s", msg.Subject, msg.ID)
			}
		}()
	}
}
