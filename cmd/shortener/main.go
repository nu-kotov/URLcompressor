package main

import (
	"log"
	"net/http"

	"github.com/nu-kotov/URLcompressor/config"
	"github.com/nu-kotov/URLcompressor/internal/app/api/handler"
	"github.com/nu-kotov/URLcompressor/internal/app/logger"
)

func main() {
	if err := logger.InitLogger("info"); err != nil {
		log.Fatal("Error initialize zap logger: ", err)
	}
	config := config.ParseConfig()
	service, err := handler.InitService(config)

	if err != nil {
		log.Fatal("Error initialize service: ", err)
	}
	router := NewRouter(*service)

	defer service.DBStorage.Close()
	log.Fatal(http.ListenAndServe(config.RunAddr, router))
}
