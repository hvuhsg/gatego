package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/hvuhsg/gatego"
	"github.com/hvuhsg/gatego/config"
)

const version = "0.0.1"

func main() {
	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	config, err := config.ParseConfig("config.yaml", version)
	if err != nil {
		log.Fatal(err)
	}

	log.Default().Println("Config loaded successfully")

	server := gatego.New(ctx, config, version)

	err = server.Run()
	if err != nil {
		log.Fatalln(err)
	}
}
