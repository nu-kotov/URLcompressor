package main

import (
	"log"
	"net/http"

	"github.com/nu-kotov/URLcompressor/config"
	"github.com/nu-kotov/URLcompressor/internal/app/api/handler"
	"github.com/nu-kotov/URLcompressor/internal/app/logger"
	"github.com/nu-kotov/URLcompressor/internal/app/storage"
)

func main() {
	if err := logger.InitLogger("info"); err != nil {
		log.Fatal("Error initialize zap logger: ", err)
	}

	config := config.ParseConfig()
	store, err := storage.NewStorage(config)
	if err != nil {
		log.Fatal("Error initialize storage: ", err)
	}

	service := handler.InitService(config, store)
	router := NewRouter(*service)

	defer service.Storage.Close()

	log.Fatal(http.ListenAndServe(config.RunAddr, router))
}
