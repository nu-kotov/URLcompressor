package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/nu-kotov/URLcompressor/config"
	"github.com/nu-kotov/URLcompressor/internal/app/api/handler"
	"github.com/nu-kotov/URLcompressor/internal/app/logger"
	"github.com/nu-kotov/URLcompressor/internal/app/storage"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {

	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n", buildVersion, buildDate, buildCommit)

	if err := logger.NewLogger("info"); err != nil {
		log.Fatal("Error initialize zap logger: ", err)
	}

	config := config.NewConfig()
	store, err := storage.NewStorage(config)
	if err != nil {
		log.Fatal("Error initialize storage: ", err)
	}

	service := handler.NewService(config, store)
	router := handler.NewRouter(*service)

	defer service.Storage.Close()
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	log.Fatal(http.ListenAndServe(config.RunAddr, router))
}
