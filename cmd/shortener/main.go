package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime/pprof"

	"github.com/nu-kotov/URLcompressor/config"
	"github.com/nu-kotov/URLcompressor/internal/app/api/handler"
	"github.com/nu-kotov/URLcompressor/internal/app/logger"
	"github.com/nu-kotov/URLcompressor/internal/app/storage"
)

func main() {
	memProfileFile, err := os.Create("../../profile/result.pprof")
	if err != nil {
		log.Fatalf("could not create memory profile file: %v", err)
	}
	defer memProfileFile.Close()

	if err := pprof.WriteHeapProfile(memProfileFile); err != nil {
		log.Fatalf("could not write heap profile: %v", err)
	}

	if err := logger.NewLogger("info"); err != nil {
		log.Fatal("Error initialize zap logger: ", err)
	}

	config := config.NewConfig()
	store, err := storage.NewStorage(config)
	if err != nil {
		log.Fatal("Error initialize storage: ", err)
	}

	service := handler.NewService(config, store)
	router := NewRouter(*service)

	defer service.Storage.Close()
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	log.Fatal(http.ListenAndServe(config.RunAddr, router))
}
