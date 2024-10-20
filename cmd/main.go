package main

import (
	"log"

	"github.com/hvuhsg/gatego"
	"github.com/hvuhsg/gatego/config"
)

const version = "0.0.1"

func main() {
	config, err := config.ParseConfig("config.yaml", version)
	if err != nil {
		log.Fatal(err)
	}

	log.Default().Println("Config loaded successfully")

	server := gatego.New(config, version)

	err = server.Run()
	if err != nil {
		log.Fatalln(err)
	}
}
