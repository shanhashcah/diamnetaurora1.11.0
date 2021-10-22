package cmd

import (
	"fmt"
	stdLog "log"
	"os"

	"github.com/spf13/cobra"
	aurora "github.com/diamnet/go/services/aurora/internal"
)

var (
	config, flags = aurora.Flags()

	rootCmd = &cobra.Command{
		Use:   "aurora",
		Short: "client-facing api server for the diamnet network",
		Long:  "client-facing api server for the diamnet network. It acts as the interface between Diamnet Core and applications that want to access the Diamnet network. It allows you to submit transactions to the network, check the status of accounts, subscribe to event streams and more.",
		Run: func(cmd *cobra.Command, args []string) {
			aurora.NewAppFromFlags(config, flags).Serve()
		},
	}
)

func init() {
	err := flags.Init(rootCmd)
	if err != nil {
		stdLog.Fatal(err.Error())
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
