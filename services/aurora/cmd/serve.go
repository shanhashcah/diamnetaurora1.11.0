package cmd

import (
	"github.com/spf13/cobra"
	aurora "github.com/diamnet/go/services/aurora/internal"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "run aurora server",
	Long:  "serve initializes then starts the aurora HTTP server",
	Run: func(cmd *cobra.Command, args []string) {
		aurora.NewAppFromFlags(config, flags).Serve()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
