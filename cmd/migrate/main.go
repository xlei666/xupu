package main

import (
	"log"
	"time"

	"github.com/xlei/xupu/internal/models"
	"github.com/xlei/xupu/pkg/db"
)

func main() {
	log.Println("Starting migration tool...")

	// Force wait for DB
	time.Sleep(1 * time.Second)

	database := db.Get()
	pg, ok := database.(*db.PostgresDatabase)
	if !ok {
		log.Fatal("Not using Postgres database")
	}

	gormDB := pg.GetDB()

	log.Println("Migrating PromptTemplate...")
	if err := gormDB.AutoMigrate(&models.PromptTemplate{}); err != nil {
		log.Fatalf("Failed to migrate PromptTemplate: %v", err)
	}

	log.Println("Migrating NarrativeTemplate...")
	if err := gormDB.AutoMigrate(&models.NarrativeTemplate{}); err != nil {
		log.Fatalf("Failed to migrate NarrativeTemplate: %v", err)
	}

	log.Println("Migrating SysConfig...")
	if err := gormDB.AutoMigrate(&models.SysConfig{}); err != nil {
		log.Fatalf("Failed to migrate SysConfig: %v", err)
	}

	// Check if table exists
	if !gormDB.Migrator().HasTable(&models.PromptTemplate{}) {
		log.Fatal("Assertion failed: Table prompt_templates does not exist after migration!")
	}

	log.Println("Migration completed successfully!")
}
