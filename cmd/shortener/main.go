package main

import (
	"context"
	"fmt"
	"log"
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
	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {

	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n", buildVersion, buildDate, buildCommit)

	idleConnsClosed := make(chan struct{})
	sigForShutdown := make(chan os.Signal, 1)
	signal.Notify(sigForShutdown, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	if err := logger.NewLogger("info"); err != nil {
		log.Fatal("Error initialize zap logger: ", err)
	}

	config, err := config.NewConfig()
	if err != nil {
		log.Fatal("Error initialize config: ", err)
	}
	store, err := storage.NewStorage(*config)
	if err != nil {
		log.Fatal("Error initialize storage: ", err)
	}

	service := handler.NewService(*config, store)
	router := handler.NewRouter(*service)

	defer service.Storage.Close()
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	server := &http.Server{
		Addr:    config.RunAddr,
		Handler: router,
	}

	go func() {
		<-sigForShutdown

		logger.Log.Info("shutdown signal received...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		defer cancel()

		if err := service.Storage.Close(); err != nil {
			logger.Log.Error("Error closing store", zap.Error(err))
		}

		if err := server.Shutdown(ctx); err != nil {
			logger.Log.Error("Server forced to shutdown", zap.Error(err))
		} else {
			logger.Log.Info("Server shutdown gracefully")
		}
		close(idleConnsClosed)
	}()

	if config.EnableHTTPS {
		manager := &autocert.Manager{
			Cache:      autocert.DirCache("cache-dir"),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist("urlcompressor.ru"),
		}
		server.TLSConfig = manager.TLSConfig()
		log.Fatal(server.ListenAndServeTLS("", ""))
	} else {
		log.Fatal(server.ListenAndServe())
	}
}
