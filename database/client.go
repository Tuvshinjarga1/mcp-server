package database

import (
	"fmt"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var DB *gorm.DB

// CreateClient initialize databases and tables
func CreateClient() *gorm.DB {
	// Print database configuration for debugging
	dbHost := viper.GetString("DB_HOST")
	dbPort := viper.GetInt("DB_PORT")
	dbUser := viper.GetString("DB_USER")
	dbName := viper.GetString("DB_NAME")
	dbTimezone := viper.GetString("DB_TIMEZONE")
	
	fmt.Printf("Connecting to database:\n")
	fmt.Printf("Host: %s\n", dbHost)
	fmt.Printf("Port: %d\n", dbPort)
	fmt.Printf("User: %s\n", dbUser)
	fmt.Printf("Database: %s\n", dbName)
	fmt.Printf("Timezone: %s\n", dbTimezone)
	
	// Check if required environment variables are set
	if dbHost == "" || dbUser == "" || dbName == "" {
		panic("Required database environment variables are not set (DB_HOST, DB_USER, DB_NAME)")
	}
	
	connectionString := fmt.Sprintf(
		"host=%s port=%d user=%s dbname=%s password=%s sslmode=disable TimeZone=%s",
		dbHost,
		dbPort,
		dbUser,
		dbName,
		viper.GetString("DB_PASSWORD"),
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
	return db
}
