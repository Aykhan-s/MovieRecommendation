package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "movier",
	Short: "Movie Recommendation System",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() error {
	rootCmd.AddCommand(getDownloadCmd())
	rootCmd.AddCommand(getFilterCmd())
	rootCmd.AddCommand(getServeCmd())
	return rootCmd.Execute()
}
