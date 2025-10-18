package main

import (
	"log"

	"mock-cbr/internal/config"
	"mock-cbr/internal/db"
	"mock-cbr/internal/router"
	"mock-cbr/internal/server"
)

// @title MOCK-CBR API
// @version 1.0
// @description Mock сервис для имитации API Центрального Банка России
func main() {

	// config
	config.LoadEnv()
	cfg := config.GetConfig()

	// database
	database, err := db.InitDB(cfg.DatabasePath)

	if err != nil {
		log.Fatalf("Failed to initialize DB: %v", err)
	}
	defer database.Close()

	// router
	r := router.SetupRouter(database)

	//server
	server.Start(cfg, r)

}
