package main

import (
	"fmt"
	"log"

	"github.com/souvikjs01/go-ecommerce/config"
	"github.com/souvikjs01/go-ecommerce/routes"
)

func main() {
	// setup config
	cfg, err := config.SetConfig()
	if err != nil {
		log.Fatalf("Error in Setting up the Configuration file: %v", err)
	}
	// db Connection
	client, err := config.NewDB(cfg)
	if err != nil {
		log.Fatalf("Error in Setting up the DB connection: %v", err)
	}
	// router
	fmt.Println("okay we are good to go")
	router := routes.SetupRoutes(client)
	router.Run(":8080")
}
