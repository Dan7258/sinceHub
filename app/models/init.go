package models

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dsn := "host=localhost user=admin password=admin dbname=scinceHubDB port=5432 sslmode=disable"
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database")
	}

	fmt.Println("Connected to the database successfully!")

	// Получаем список таблиц
	var tables []string
	result := DB.Raw("SELECT tablename FROM pg_tables WHERE schemaname = 'public'").Scan(&tables)
	if result.Error != nil {
		fmt.Println("Error fetching tables:", result.Error)
		return
	}

	fmt.Println("Tables in the database:")
	for _, table := range tables {
		fmt.Println("-", table)
	}
}
