package database

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var DB *gorm.DB

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	if value := viper.GetString(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvIntOrDefault returns environment variable as int or default
func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal := viper.GetInt(key); intVal != 0 {
			return intVal
		}
	}
	if intVal := viper.GetInt(key); intVal != 0 {
		return intVal
	}
	return defaultValue
}

// CreateClient initialize databases and tables
func CreateClient() *gorm.DB {
	// Try to load .env file if it exists (for local development)
	viper.SetConfigFile(".env")
	viper.ReadInConfig() // Don't fail if .env doesn't exist
	
	// Get database configuration with Railway environment variables support
	dbHost := getEnvOrDefault("DB_HOST", "localhost")
	dbPort := getEnvIntOrDefault("DB_PORT", 5432)
	dbUser := getEnvOrDefault("DB_USER", "postgres")
	dbName := getEnvOrDefault("DB_NAME", "postgres")
	dbPassword := getEnvOrDefault("DB_PASSWORD", "")
	dbTimezone := getEnvOrDefault("DB_TIMEZONE", "Asia/Ulaanbaatar")
	
	// Check for Railway's DATABASE_URL (Railway PostgreSQL format)
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		fmt.Println("Using Railway DATABASE_URL")
		db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
			PrepareStmt:                              true,
			SkipDefaultTransaction:                   true,
			DisableForeignKeyConstraintWhenMigrating: true,
			NamingStrategy: schema.NamingStrategy{
				TablePrefix: "pmt_",
			},
		})
		
		if err != nil {
			fmt.Printf("Database connection failed with DATABASE_URL: %s\n", err.Error())
			panic(err.Error())
		}
		
		DB = db
		fmt.Println("Successfully connected to database using DATABASE_URL")
		return db
	}
	
	// Print database configuration for debugging
	fmt.Printf("Connecting to database:\n")
	fmt.Printf("Host: %s\n", dbHost)
	fmt.Printf("Port: %d\n", dbPort)
	fmt.Printf("User: %s\n", dbUser)
	fmt.Printf("Database: %s\n", dbName)
	fmt.Printf("Timezone: %s\n", dbTimezone)
	
	// Check if required environment variables are set
	if dbHost == "" || dbUser == "" || dbName == "" {
		panic("Required database environment variables are not set (DB_HOST, DB_USER, DB_NAME or DATABASE_URL)")
	}
	
	connectionString := fmt.Sprintf(
		"host=%s port=%d user=%s dbname=%s password=%s sslmode=require TimeZone=%s",
		dbHost,
		dbPort,
		dbUser,
		dbName,
		dbPassword,
		dbTimezone,
	)
	
	fmt.Println("Attempting database connection...")
	db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{
		PrepareStmt:                              true,
		SkipDefaultTransaction:                   true,
		DisableForeignKeyConstraintWhenMigrating: true,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: "pmt_",
		},
	})

	if err != nil {
		fmt.Println("------------------------------------------")
		fmt.Printf("Database connection failed: %s\n", err.Error())
		fmt.Printf("Host: %s, Port: %d\n", dbHost, dbPort)
		fmt.Println("------------------------------------------")
		panic(err.Error())
	}
	DB = db
	fmt.Println("Successfully connected to database")
	return db
}
