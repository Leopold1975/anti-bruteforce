package main

import (
	"context"
	"flag"
	"os/signal"
	"syscall"
	"time"

	"github.com/Pos1t1veM1ndset/anti-bruteforce/internal/config"
	"github.com/Pos1t1veM1ndset/anti-bruteforce/internal/redislimiter"
	"github.com/Pos1t1veM1ndset/anti-bruteforce/internal/server"
	"github.com/Pos1t1veM1ndset/anti-bruteforce/pkg/logger"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "configs/config.toml", "Path to configuration file")
}

func main() {
	flag.Parse()

	cfg := *config.NewConfig(configFile)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	logg := logger.New(cfg)
	service := redislimiter.New(cfg)

	server := server.New(service, cfg)
	go func() {
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		logg.Info("redis is shutting down...")
		if err := service.Shutdown(ctx); err != nil {
			logg.Fatal(err)
		}
	}()

	go func() {
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		logg.Info("server is shutting down...")
		if err := server.Stop(ctx); err != nil {
			logg.Error(err)
		}
	}()

	logg.Info("service stated...")
	if err := server.Start(ctx); err != nil {
		logg.Error(err)
		cancel()
	}
}
