package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nu-kotov/URLcompressor/config"
	"github.com/nu-kotov/URLcompressor/internal/app/api/handler"
	"github.com/nu-kotov/URLcompressor/internal/app/logger"
	"github.com/nu-kotov/URLcompressor/internal/app/storage"
	"golang.org/x/crypto/acme/autocert"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {

	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n", buildVersion, buildDate, buildCommit)

	idleConnsClosed := make(chan struct{})
	sigForShutdown := make(chan os.Signal, 1)
	signal.Notify(sigForShutdown, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	if err := logger.NewLogger("info"); err != nil {
		return fmt.Errorf("error initialize zap logger: %w", err)
	}

	config, err := config.NewConfig()
	if err != nil {
		return fmt.Errorf("error initialize config: %w", err)
	}
	var trustedSubnet *net.IPNet
	if config.TrustedSubnet != "" {
		_, subnet, err := net.ParseCIDR(config.TrustedSubnet)
		if err != nil {
			return fmt.Errorf("invalid trusted subnet: %w", err)
		}
		trustedSubnet = subnet
	}
	store, err := storage.NewStorage(*config)
	if err != nil {
		return fmt.Errorf("error initialize storage: %w", err)
	}

	service := handler.NewService(*config, store, trustedSubnet)
	router := handler.NewRouter(*service)

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	server := &http.Server{
		Addr:    config.RunAddr,
		Handler: router,
	}

	go func() error {
		if config.EnableHTTPS {
			manager := &autocert.Manager{
				Cache:      autocert.DirCache("cache-dir"),
				Prompt:     autocert.AcceptTOS,
				HostPolicy: autocert.HostWhitelist("urlcompressor.ru"),
			}
			server.TLSConfig = manager.TLSConfig()
			err = server.ListenAndServeTLS("", "")
			if err != nil && err != http.ErrServerClosed {
				return fmt.Errorf("error ListenAndServeTLS: %w", err)
			}
		} else {
			err = server.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				return fmt.Errorf("error ListenAndServe: %w", err)
			}
		}
		return nil
	}()

	<-sigForShutdown

	logger.Log.Info("shutdown signal received...")

	if err := service.Storage.Close(); err != nil {
		return fmt.Errorf("error closing store: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	} else {
		logger.Log.Info("Server shutdown gracefully")
	}
	close(idleConnsClosed)

	return nil
}
