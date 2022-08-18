package main

import (
	"github.com/spf13/cobra"

	"github.com/Shell-Company/procurement/cmd/procurement/discord"
	"github.com/Shell-Company/procurement/cmd/procurement/slack"
	"github.com/Shell-Company/procurement/internal/version"
)

func main() {
	rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(discord.NewDiscordCommand())
	rootCmd.AddCommand(slack.NewSlackCommand())
}

var rootCmd = &cobra.Command{
	Use:     "procurement",
	Version: version.Version,
}
