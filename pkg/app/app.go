package app

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"runtime"
	"runtime/debug"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
)

type App struct {
	Name    string
	Usage   string
	log     *logrus.Entry
	opt     *Options
	app     *cli.App
	vConfig *viper.Viper
}

func NewApp(name, usage string) *App {
	return &App{
		Name:  name,
		Usage: usage,
		log: logrus.WithFields(map[string]interface{}{
			"module": name,
		}),
	}
}

func (a *App) Init(opts ...Option) {
	a.app = cli.NewApp()
	a.app.EnableBashCompletion = true
	a.app.Name = a.Name
	a.app.Usage = a.Usage

	a.opt = NewOptions()
	for _, opt := range opts {
		opt(a.opt)
	}

	a.app.Commands = a.opt.Commands
	a.app.Flags = a.opt.Flags

	a.app.Flags = append(a.app.Flags, &cli.StringFlag{
		Name:        "config",
		Aliases:     []string{"c"},
		Value:       "./config/default.yaml",
		DefaultText: "./config/default.yaml",
		Usage:       "config file",
		Required:    false,
	})
	a.app.Flags = append(a.app.Flags, &cli.StringFlag{
		Name:        "log.level",
		Value:       "info",
		DefaultText: "log level",
		Usage:       "log level",
		Required:    false,
	})
	a.app.Flags = append(a.app.Flags, &cli.StringFlag{
		Name:        "log.format",
		Value:       "text",
		DefaultText: "log format",
		Usage:       "log format",
		Required:    false,
	})
	a.app.Flags = append(a.app.Flags, &cli.StringFlag{
		Name:        "env",
		Value:       "dev",
		DefaultText: "environment",
		Usage:       "environment",
		Required:    false,
	})
}

func (a *App) Start() error {
	if a.app == nil {
		panic("please init app")
	}
	app := a.app
	app.Before = func(c *cli.Context) error {
		// 读配置文件
		a.vConfig = viper.New()
		cf := c.String("config")
		if cf != "" {
			a.opt.ConfigAddress = cf
		}
		a.vConfig.SetConfigFile(a.opt.ConfigAddress)
		if err := a.vConfig.ReadInConfig(); err != nil {
			// 如果找不到配置文件，使用环境变量作为默认配置
			a.log.Warnf("Config file not found: %s, using environment variables", a.opt.ConfigAddress)
		}

		// 支持环境变量覆盖配置
		a.vConfig.AutomaticEnv()
		// 绑定常用 key，支持下划线转小数点
		a.vConfig.SetEnvKeyReplacer(strings.NewReplacer("_", "."))
		// 例如 REDIS_ADDR -> redis.addr

		// 读日志
		logLevel := c.String("log.level")
		logFormat := c.String("log.format")
		InitLogger(logLevel, logFormat)

		return nil
	}
	app.After = func(c *cli.Context) error {
		for _, a := range a.opt.After {
			if err := a(); err != nil {
				return err
			}
		}
		return nil
	}
	err := app.Run(os.Args)
	if err != nil {
		a.log.Fatal(err)
		return err
	}

	return nil
}

func (a *App) Config() *viper.Viper {
	if a.vConfig == nil {
		panic("should start first")
	}
	return a.vConfig
}

func WaitTerminate(stopFunc func(s os.Signal)) {
	signalChan := make(chan os.Signal, 1)

	signal.Notify(signalChan,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	defer func() {
		if e := recover(); e != nil {
			logrus.Errorf("crashed, err: %s stack:%s", e, string(debug.Stack()))
		}
	}()
	recvSignal := <-signalChan
	logrus.Infof("recv signal: %v", recvSignal)
	stopFunc(recvSignal)
}

func InitLogger(level string, format string) {
	parsedLevel, err := logrus.ParseLevel(level)
	if err != nil {
		parsedLevel = logrus.InfoLevel
	}
	logrus.SetLevel(parsedLevel)
	logrus.SetReportCaller(true)
	logrus.SetOutput(os.Stdout)

	if format == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				fileName := fmt.Sprintf("%s:%d", path.Base(f.File), f.Line)
				funcName := path.Base(f.Function)
				return funcName, fileName
			},
		})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
			ForceColors:   true,
			PadLevelText:  true,
		})
	}
}
