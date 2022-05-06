package main

import (
	"Mining-Profitability/pkg/appcontext"
	"Mining-Profitability/pkg/applog"
	"Mining-Profitability/pkg/config"
	"Mining-Profitability/pkg/miningprofitability"
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	shutdownTimeout = 5 * time.Second
)

func main() {
	configFile := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.New(*configFile)
	if err != nil {
		log.Fatalf("error getting the config: %s", err)
	}

	logger := applog.New(cfg)
	router := http.NewServeMux()
	server := &http.Server{Addr: cfg.Address, Handler: router}

	appContext, appCtxCancel, err := appcontext.New(cfg, logger)
	if err != nil {
		logger.Fatalf("error creating the app context: %s", err)
	}

	router.Handle("/mining-calc", miningprofitability.NewHandler(appContext))

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGKILL)

	shutdownComplete := make(chan struct{}, 1)

	go func() {
		<-signals
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		logger.Info("shutting down the server")
		if err := server.Shutdown(ctx); err != nil {
			logger.Fatalf("error shutting down the server: %s", err)
		}
		close(shutdownComplete)
	}()

	logger.Infof("starting the server on %s", cfg.Address)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		logger.Fatalf("error starting the server: %s", err)
	}
	appCtxCancel()

	<-shutdownComplete
	logger.Info("server shutdown successfully")
}
