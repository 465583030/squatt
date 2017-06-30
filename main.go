package main

import (
	"crypto/tls"
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
		log.Info("server starting")

		s := server.NewServer()
		s.SetLogger(log)

		go s.Route()

		if listen := cfg.GetString("listen.tcp"); listen != "" {
			log.Info("starting tcp server", zap.String("address", listen))
			go func() {
				if err := s.ListenAndServe(listen); err != nil {
					log.Fatal("could not start tcp server", zap.Error(err))
				}
			}()
		}

		if listen := cfg.GetString("listen.tls"); listen != "" {
			var tlsConfig tls.Config
			certificate, err := tls.LoadX509KeyPair(cfg.GetString("tls.certificate"), cfg.GetString("tls.key"))
			if err != nil {
				log.Fatal("could not load tls certificate and key", zap.Error(err))
			}
			tlsConfig.Certificates = append(tlsConfig.Certificates, certificate)
			log.Info("starting tls server", zap.String("address", listen))
			go func() {
				if err := s.ListenAndServeTLS(listen, &tlsConfig); err != nil {
					log.Fatal("could not start tls server", zap.Error(err))
				}
			}()
		}

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
	Listen struct {
		TCP   string `name:"tcp" description:"MQTT server TCP listen address"`
		TLS   string `name:"tls" description:"MQTT server TLS listen address"`
		Debug string `name:"debug" description:"Debug server listen address"`
	} `name:"listen"`
	TLS struct {
		Certificate string `name:"certificate" description:"Path to certificate for TLS"`
		Key         string `name:"key" description:"Path to private key for TLS"`
	}
	Debug bool `name:"debug" description:"Debug mode"`
}

func defaults() (defaults squattConfig) {
	defaults.Listen.TCP = ":1883"
	defaults.Listen.Debug = "127.0.0.1:6060"
	defaults.TLS.Certificate = "cert.pem"
	defaults.TLS.Key = "key.pem"
	return
}

func init() {
	cfg = config.Initialize("squatt", defaults())
	cmd.Flags().AddFlagSet(cfg.Flags())
	cobra.OnInitialize(func() {
		if cfg.GetBool("debug") {
			log, _ = zap.NewDevelopment()
		}
		if err := cfg.ReadInConfig(); err != nil {
			log.Info("not using config file", zap.Error(err))
		}
	})
}

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal("failed to run", zap.Error(err))
	}
}
