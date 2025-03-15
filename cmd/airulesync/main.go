package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/upamune/airulesync/internal/app"
)

var cli struct {
	// Global flags
	Config  string `short:"c" help:"Path to config file" default:".airulesync.yaml"`
	Verbose bool   `short:"v" help:"Enable verbose output"`

	// Commands
	Sync struct {
		DryRun bool `short:"d" help:"Simulate execution without applying changes"`
	} `cmd:"" help:"Synchronize rule files according to configuration"`

	Init struct {
		Dir string `arg:"" optional:"" help:"Directory to scan for rule files"`
	} `cmd:"" help:"Scan directory and generate a configuration file"`

	Version struct{} `cmd:"" help:"Display version information"`
}

func main() {
	ctx := kong.Parse(&cli,
		kong.Name("airulesync"),
		kong.Description("Synchronize AI coding tool rule files across directories"),
		kong.UsageOnError(),
	)

	// Create the application
	application := app.NewApp(cli.Config, cli.Verbose)

	// Execute the appropriate command
	var err error
	switch ctx.Command() {
	case "sync":
		err = application.RunSync(cli.Sync.DryRun)
	case "init":
		err = application.RunInit(cli.Init.Dir)
	case "version":
		err = application.RunVersion()
	}

	// Handle errors
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
