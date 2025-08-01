package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/helloworldyuhaiyang/mail-handle/internal"
	"github.com/helloworldyuhaiyang/mail-handle/pkg/app"
	"github.com/helloworldyuhaiyang/mail-handle/pkg/data"
	"github.com/helloworldyuhaiyang/mail-handle/pkg/gmail"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	mailApp := app.NewApp("mail-handle", "handle mail from gmail")
	mailApp.Init(
		app.WithCommands([]*cli.Command{
			{
				Name: "run",
				Action: func(c *cli.Context) error {
					return RunServer(c, mailApp)
				},
			},
		}),
	)
	err := mailApp.Start()
	if err != nil {
		logrus.Error("start mail-handle failed", err)
	}
}

func RunServer(c *cli.Context, mailApp *app.App) error {
	// 初始化数据库
	db, err := initDatabase(mailApp)
	if err != nil {
		logrus.Error("init database failed", err)
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			logrus.Errorf("failed to close database: %v", err)
		}
	}()

	// 初始化 Gmail client
	gmailClient, err := initGmailClient(mailApp)
	if err != nil {
		logrus.Error("init gmail client failed", err)
		return err
	}
	defer func() {
		if err := gmailClient.Close(); err != nil {
			logrus.Errorf("failed to close Gmail client: %v", err)
		}
	}()

	// 启动 api server
	apiServer := internal.NewApiServer(mailApp.Config().GetString("server.addr"), db, gmailClient)
	if err := apiServer.Start(); err != nil {
		logrus.Errorf("api server error: %v", err)
	} else {
		logrus.Info("api server started successfully")
	}
	defer func() {
		if err := apiServer.Stop(30 * time.Second); err != nil {
			logrus.Errorf("failed to stop server gracefully: %v", err)
		}
	}()

	// 启动定时任务
	scheduler := internal.NewScheduler()
	if err := scheduler.Start(); err != nil {
		logrus.Errorf("failed to start scheduler: %v", err)
	} else {
		logrus.Info("scheduler started successfully")
	}
	defer func() {
		if err := scheduler.Stop(); err != nil {
			logrus.Errorf("failed to stop scheduler: %v", err)
		}
	}()

	// 等待信号, 优雅关闭
	app.WaitTerminate(func(s os.Signal) {
		logrus.Infof("received signal %v, shutting down gracefully", s)
	})

	logrus.Info("server shutdown complete")
	return nil
}

type DBConfig struct {
	DSN string `mapstructure:"dsn"`
}

type GmailConfig struct {
	CredentialsFile string `mapstructure:"credentials_file"`
	TokenFile       string `mapstructure:"token_file"`
}

func initDatabase(mailApp *app.App) (*data.DB, error) {
	db, err := data.NewDB(mailApp.Config().GetString("database.dsn"))
	if err != nil {
		return nil, err
	}
	logrus.Info("database initialized successfully")
	return db, nil
}

// initGmailClient initializes the Gmail client
func initGmailClient(mailApp *app.App) (*gmail.Client, error) {
	var gmailConfig GmailConfig
	err := mailApp.Config().UnmarshalKey("gmail", &gmailConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal gmail config: %v", err)
	}

	gmailClient, err := gmail.NewClient(context.Background(), gmailConfig.CredentialsFile, gmailConfig.TokenFile)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Gmail client: %v", err)
	}

	logrus.Info("gmail client initialized successfully")
	return gmailClient, nil
}
