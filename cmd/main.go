package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/helloworldyuhaiyang/mail-handle/internal/api"
	"github.com/helloworldyuhaiyang/mail-handle/internal/db"
	"github.com/helloworldyuhaiyang/mail-handle/internal/schedule"
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
			{
				Name:  "auth",
				Usage: "Authenticate with Gmail OAuth2",
				Action: func(c *cli.Context) error {
					return AuthenticateGmail(c, mailApp)
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
	gDatabase, err := initDatabase(mailApp)
	if err != nil {
		logrus.Error("init database failed", err)
		return err
	}
	defer func() {
		if err := gDatabase.Close(); err != nil {
			logrus.Errorf("failed to close database: %v", err)
		}
	}()

	// 初始化 gmail client
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
	apiServer := api.NewApiServer(mailApp.Config().GetString("server.addr"), gDatabase, gmailClient)
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

	// db 初始化
	forwardTargetsRepo := db.NewForwardTargetsRepo(gDatabase.DB)

	// 启动定时任务
	schedulerConfig := schedule.SchedulerConfig{}
	err = mailApp.Config().UnmarshalKey("scheduler", &schedulerConfig)
	if err != nil {
		logrus.Errorf("failed to unmarshal scheduler config: %v", err)
	}

	// 打印配置信息用于调试
	logrus.Infof("Scheduler config: %+v", schedulerConfig)

	scheduler := schedule.NewScheduler(&schedulerConfig, gmailClient, forwardTargetsRepo)
	if err := scheduler.Start(); err != nil {
		logrus.Errorf("failed to start scheduler: %v", err)
		return err
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

// AuthenticateGmail handles the Gmail OAuth2 authentication flow
func AuthenticateGmail(c *cli.Context, mailApp *app.App) error {
	var gmailConfig GmailConfig
	err := mailApp.Config().UnmarshalKey("gmail", &gmailConfig)
	if err != nil {
		return fmt.Errorf("failed to unmarshal gmail config: %v", err)
	}

	gmailClient, err := gmail.NewClient(context.Background(), gmailConfig.CredentialsFile, gmailConfig.TokenFile)
	if err != nil {
		return fmt.Errorf("failed to initialize Gmail client: %v", err)
	}
	defer gmailClient.Close()

	// Check if already authenticated
	if gmailClient.IsAuthenticated() {
		logrus.Info("Gmail client is already authenticated")
		return nil
	}

	// Get authorization URL
	authURL := gmailClient.GetAuthURL()
	logrus.Info("Please visit this URL to authorize the application:")
	logrus.Info(authURL)
	logrus.Info("")
	logrus.Info("After authorization, you will receive an authorization code.")
	logrus.Info("Please enter the authorization code:")

	// Read authorization code from stdin
	var authCode string
	fmt.Print("Authorization code: ")
	fmt.Scanln(&authCode)

	if authCode == "" {
		return fmt.Errorf("authorization code is required")
	}

	// Exchange authorization code for token
	err = gmailClient.Callback(authCode)
	if err != nil {
		return fmt.Errorf("failed to exchange authorization code: %v", err)
	}

	logrus.Info("Authentication successful! Token has been saved.")
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

	// Check if authentication is needed
	if needsAuth, authURL := gmailClient.AuthenticateIfNeeded(); needsAuth {
		logrus.Warn("Gmail client needs authentication")
		logrus.Info(authURL)
		logrus.Info("After authorization, you can restart the application")
	} else {
		logrus.Info("Gmail client is authenticated")
	}

	logrus.Info("gmail client initialized successfully")
	return gmailClient, nil
}
