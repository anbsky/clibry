package cmd

import (
	"log"
	"strconv"

	"github.com/sayplastic/clibry/client"
	"github.com/spf13/cobra"
)

var streamCmd = &cobra.Command{
	Use:   "stream",
	Short: "issue a batch of stream playback requests",
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
		client.LaunchStreams(n)
	},
}

func init() {
	rootCmd.AddCommand(streamCmd)
}
