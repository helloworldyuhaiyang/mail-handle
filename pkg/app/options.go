package app

import (
	"context"
	"github.com/urfave/cli/v2"
)

// Options for App
type Options struct {
	// 是否开启pprof
	PProf bool

	// 配置文件
	ConfigAddress string

	// 命令行
	Flags []cli.Flag
	// 子命令
	Commands []*cli.Command

	// Before and After funcs
	Before []func() error
	After  []func() error

	Context context.Context
}

func NewOptions() *Options {
	return &Options{
		PProf:         false,
		ConfigAddress: "",
		Flags:         nil,
		Commands:      nil,
		Before:        nil,
		After:         nil,
		Context:       context.Background(),
	}
}

type Option func(*Options)

func WithCommands(commands []*cli.Command) Option {
	return func(o *Options) {
		o.Commands = commands
	}
}

func WithFlags(flags []cli.Flag) Option {
	return func(o *Options) {
		o.Flags = flags
	}
}

func AddAfter(after func() error) Option {
	return func(o *Options) {
		o.After = append(o.After, after)
	}
}

func AddBefore(before func() error) Option {
	return func(o *Options) {
		o.Before = append(o.Before, before)
	}
}
