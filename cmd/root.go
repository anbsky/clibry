package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "clibry",
	Short: "clibry is a...",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("run ./clibry stream")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
