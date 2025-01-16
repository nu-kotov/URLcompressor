package main

import (
	"log"
	"net/http"

	"github.com/nu-kotov/URLcompressor/api/handler"
	"github.com/nu-kotov/URLcompressor/config"
	"github.com/nu-kotov/URLcompressor/internal/app/logger"
)

func main() {
	if err := logger.InitLogger("info"); err != nil {
		log.Fatal("Error initialize zap logger: ", err)
	}
	config := config.ParseConfig()
	service := handler.InitService(config)
	router := NewRouter(service)

	log.Fatal(http.ListenAndServe(config.RunAddr, router))
}
