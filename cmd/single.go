package cmd

import (
	"log"
	"strconv"

	"github.com/sayplastic/clibry/client"
	"github.com/spf13/cobra"
)

var singleCmd = &cobra.Command{
	Use:   "single",
	Short: "issue a single query",
	Run: func(cmd *cobra.Command, args []string) {
		var (
			n   int
			err error
		)
		if len(args) < 2 {
			n = 0
		} else {
			n, err = strconv.Atoi(args[0])
			if err != nil {
				log.Fatalf("%v is not an integer", args[0])
			}
		}
		client.LaunchQuery(n, args[1])
	},
}

func init() {
	rootCmd.AddCommand(singleCmd)
}
