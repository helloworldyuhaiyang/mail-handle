package main

import (
	"fmt"
	"os"

	"github.com/helloworldyuhaiyang/mail-handle/pkg/app"
	"github.com/urfave/cli/v2"
)

func main() {
	mailApp := app.NewApp("mail-handle", "handle mail from gmail")
	mailApp.Init(
		app.WithCommands([]*cli.Command{
			{
				Name: "run",
				Action: func(c *cli.Context) error {
					fmt.Println("run")
					app.WaitTerminate(func(s os.Signal) {
						fmt.Println("process receive signal", s)
					})

					fmt.Println("process exit")
					return nil
				},
			},
		}),
	)
	mailApp.Start()
}
