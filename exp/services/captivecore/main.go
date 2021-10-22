package main

import (
	"fmt"
	"go/types"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/diamnet/go/exp/services/captivecore/internal"
	"github.com/diamnet/go/ingest/ledgerbackend"
	"github.com/diamnet/go/network"
	"github.com/diamnet/go/support/config"
	supporthttp "github.com/diamnet/go/support/http"
	supportlog "github.com/diamnet/go/support/log"
)

func main() {
	var port int
	var networkPassphrase, binaryPath, configPath string
	var historyArchiveURLs []string
	var logLevel logrus.Level
	logger := supportlog.New()

	configOpts := config.ConfigOptions{
		{
			Name:        "port",
			Usage:       "Port to listen and serve on",
			OptType:     types.Int,
			ConfigKey:   &port,
			FlagDefault: 8000,
			Required:    true,
		},
		{
			Name:        "network-passphrase",
			Usage:       "Network passphrase of the Diamnet network transactions should be signed for",
			OptType:     types.String,
			ConfigKey:   &networkPassphrase,
			FlagDefault: network.TestNetworkPassphrase,
			Required:    true,
		},
		&config.ConfigOption{
			Name:        "diamnet-core-binary-path",
			OptType:     types.String,
			FlagDefault: "",
			Required:    true,
			Usage:       "path to diamnet core binary",
			ConfigKey:   &binaryPath,
		},
		&config.ConfigOption{
			Name:        "diamnet-core-config-path",
			OptType:     types.String,
			FlagDefault: "",
			Required:    false,
			Usage:       "path to diamnet core config file",
			ConfigKey:   &configPath,
		},
		&config.ConfigOption{
			Name:        "history-archive-urls",
			ConfigKey:   &historyArchiveURLs,
			OptType:     types.String,
			Required:    true,
			FlagDefault: "",
			CustomSetValue: func(co *config.ConfigOption) {
				stringOfUrls := viper.GetString(co.Name)
				urlStrings := strings.Split(stringOfUrls, ",")

				*(co.ConfigKey.(*[]string)) = urlStrings
			},
			Usage: "comma-separated list of diamnet history archives to connect with",
		},
		&config.ConfigOption{
			Name:        "log-level",
			ConfigKey:   &logLevel,
			OptType:     types.String,
			FlagDefault: "info",
			CustomSetValue: func(co *config.ConfigOption) {
				ll, err := logrus.ParseLevel(viper.GetString(co.Name))
				if err != nil {
					logger.Fatalf("Could not parse log-level: %v", viper.GetString(co.Name))
				}
				*(co.ConfigKey.(*logrus.Level)) = ll
			},
			Usage: "minimum log severity (debug, info, warn, error) to log",
		},
	}
	cmd := &cobra.Command{
		Use:   "captivecore",
		Short: "Run the remote captive core server",
		Run: func(_ *cobra.Command, _ []string) {
			configOpts.Require()
			configOpts.SetValues()
			logger.Level = logLevel

			core, err := ledgerbackend.NewCaptive(binaryPath, configPath, networkPassphrase, historyArchiveURLs)
			if err != nil {
				logger.WithError(err).Fatal("Could not create captive core instance")
			}
			api := internal.NewCaptiveCoreAPI(core, logger)

			supporthttp.Run(supporthttp.Config{
				ListenAddr: fmt.Sprintf(":%d", port),
				Handler:    internal.Handler(api),
				OnStarting: func() {
					logger.Infof("Starting Captive Core server on %v", port)
				},
				OnStopping: func() {
					api.Shutdown()
				},
			})
		},
	}

	if err := configOpts.Init(cmd); err != nil {
		logger.WithError(err).Fatal("could not parse config options")
	}

	if err := cmd.Execute(); err != nil {
		logger.WithError(err).Fatal("could not run")
	}
}
