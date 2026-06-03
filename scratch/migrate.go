package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/refynehq/refyne-backend/internal/database/migrations"
	"github.com/refynehq/refyne-backend/pkg/logging"
)

func main() {
	if err := godotenv.Load("../.env.dev"); err != nil {
		log.Println("No ../.env.dev file found")
	}
	logging.Initialize()
	
	migrations.Force(20)

	err := migrations.MigrateUp()
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Migrations completed successfully")
}
