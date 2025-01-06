package main

import (
	"context"
	"log"
	monitor2 "monitor2/src"
	"monitor2/src/alerts"
	database "monitor2/src/db"
	"os"
	"strconv"

	"github.com/rs/zerolog"
)

var ctx context.Context
var app monitor2.App

func main() {
	ctx = context.Background()
	logger := zerolog.New(os.Stderr).
    With().
    Timestamp().
    Caller().
    Logger()
	log.SetFlags(0)
	log.SetOutput(logger)

	err := database.Init()
	if err != nil {
		log.Fatalf("Could not start db: %+v", err)
	}

  ngrok_url := os.Getenv("NGROK_URL")
	if len(ngrok_url) == 0 {
		log.Printf("NGROK_URL is empty")
		return
	}

	repos_url := os.Getenv("REPOS_PATH")
	if len(repos_url) == 0 {
		log.Printf("REPOS_PATH is empty")
		return
	}

	port := os.Getenv("API_PORT")
	if len(port) == 0 {
		log.Fatalf("API_PORT is empty.")
	}
	_, err = strconv.Atoi(port)
	if err != nil {
		log.Fatalf("API_PORT is not an integer: %+v", err)
	}

	alerts.Init()

	addr := "0.0.0.0" + ":" + port
	go app.Init(addr)

	done := make(chan bool)
	go monitor2.StartScheduler()
	<-done
}
