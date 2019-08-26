package cmd

import (
	"log"
	"strconv"

	"github.com/sayplastic/clibry/client"
	"github.com/spf13/cobra"
)

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "playback a batch of hourly queries",
	Run: func(cmd *cobra.Command, args []string) {
		var (
			n   int
			err error
		)
		if len(args) < 1 {
			n = 0
		} else {
			n, err = strconv.Atoi(args[0])
			if err != nil {
				log.Fatalf("%v is not an integer", args[0])
			}
		}
		client.LaunchClients(n)
	},
}

func init() {
	rootCmd.AddCommand(queryCmd)
}
