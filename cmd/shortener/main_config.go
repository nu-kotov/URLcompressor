package main

import (
	"flag"
)

var (
	FlagRunAddr string
	FlagBaseURL string
)

func ParseConfigFlags() {
	flag.StringVar(&FlagRunAddr, "run-addr", ":8080", "address and port to run server")
	flag.StringVar(&FlagBaseURL, "base-url", "http://localhost:8080", "default schema, host and port in compressed URL")

	flag.Parse()
}
