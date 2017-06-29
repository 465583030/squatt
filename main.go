package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/htdvisser/pkg/config"
	"github.com/htdvisser/squatt/server"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var log, _ = zap.NewProduction()

var cfg *config.Config

var cmd = &cobra.Command{
	Use:   "squatt",
	Short: "The SQuaTT MQTT Server",
	Run: func(cmd *cobra.Command, args []string) {
		defer func() { log.Info("server stopped") }()

		listen := cfg.GetString("listen")

		s := server.NewServer()
		s.SetLogger(log)

		go s.Route()

		go func() {
			if err := s.ListenAndServe(listen); err != nil {
				log.Fatal("Could not start server", zap.Error(err))
			}
		}()

		log.Info("server started", zap.String("listen", listen))

		if cfg.GetBool("debug") {
			go func() {
				if err := http.ListenAndServe(cfg.GetString("listen-debug"), nil); err != nil {
					log.Fatal("Could not start debug server", zap.Error(err))
				}
			}()
		}

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		signal := (<-sigChan).String()
		log.Info("signal received", zap.String("signal", signal))
	},
}

type squattConfig struct {
	Listen      string `name:"listen" description:"MQTT server listen address"`
	Debug       bool   `name:"debug" description:"Debug mode"`
	ListenDebug string `name:"listen-debug" description:"Debug server listen address"`
}

func init() {
	cfg = config.Initialize("squatt", squattConfig{
		Listen:      ":1883",
		ListenDebug: "127.0.0.1:6060",
	})
	cmd.Flags().AddFlagSet(cfg.Flags())
	cobra.OnInitialize(func() {
		if cfg.GetBool("debug") {
			log, _ = zap.NewDevelopment()
		}
		if err := cfg.ReadInConfig(); err != nil {
			log.Info("Not using config file", zap.Error(err))
		}
	})
}

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal("failed to run", zap.Error(err))
	}
}
