package internal

import "fmt"

var (
	AppName   = "unknown"
	Version   = "unknown"
	Commit    = "unknown"
	BuildDate = "unknown"
)

func ShowInfo() {
	fmt.Println("==============================================")
	fmt.Println("App:", AppName)
	fmt.Println("Version", Version)
	fmt.Println("Commit:", Commit)
	fmt.Println("BuildDate:", BuildDate)
	fmt.Println("==============================================")
}
