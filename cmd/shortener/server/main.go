package main

import (
	"log"
	"net/http"

	"github.com/nu-kotov/URLcompressor/api/handler"
	"github.com/nu-kotov/URLcompressor/config"
)

func main() {
	config := config.ParseConfig()
	service := handler.InitService(config)
	router := NewRouter(service)

	log.Fatal(http.ListenAndServe(config.RunAddr, router))
}
