package main

import (
	"fmt"
	"log"
	"os"

	"cms/config"
	"cms/internal/seed"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run cmd/seed/main.go [seed|clean|reset]")
		fmt.Println("\nCommands:")
		fmt.Println("  seed  - Populate database with test data")
		fmt.Println("  clean - Delete all test data")
		fmt.Println("  reset - Clean and reseed database")
		os.Exit(1)
	}

	cfg := config.LoadConfig()
	db := config.NewDatabase(cfg.DatabaseURL, cfg.Environment == "production")

	command := os.Args[1]

	switch command {
	case "seed":
		log.Println("Seeding database with test data...")
		if err := seed.SeedDatabase(db); err != nil {
			log.Fatalf("❌ Seeding failed: %v\n", err)
		}
	case "clean":
		log.Println("Cleaning database...")
		if err := seed.CleanDatabase(db); err != nil {
			log.Fatalf("❌ Cleaning failed: %v\n", err)
		}
		log.Println("✅ Database cleaned successfully")
	case "reset":
		log.Println("Resetting database...")
		if err := seed.CleanDatabase(db); err != nil {
			log.Fatalf("❌ Cleaning failed: %v\n", err)
		}
		if err := seed.SeedDatabase(db); err != nil {
			log.Fatalf("❌ Seeding failed: %v\n", err)
		}
	default:
		fmt.Printf("❌ Unknown command: %s\n", command)
		os.Exit(1)
	}
}
