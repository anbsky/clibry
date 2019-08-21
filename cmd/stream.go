package cmd

import (
	"github.com/sayplastic/clibry/client"
	"github.com/spf13/cobra"
)

var streamCmd = &cobra.Command{
	Use:   "stream",
	Short: "playback a stream of hourly requests",
	Run: func(cmd *cobra.Command, args []string) {
		client.LaunchClients(1)
	},
}

func init() {
	rootCmd.AddCommand(streamCmd)
}
