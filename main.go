package main

import (
	"encoding/json"
	"fmt"
	"log"
	"mcp-server/database"
	"mcp-server/smtp"
	"net/http"
	"os"
	"time"

	"github.com/spf13/viper"
)

type FunctionCall struct {
	Function string                 `json:"function"`
	Args     map[string]interface{} `json:"args"`
}

type FunctionResponse struct {
	Result interface{} `json:"result"`
}

var dbInitialized = false

func main() {
	log.Println("Starting MCP Server......")
	
	// Configure viper to read from environment variables
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	
	// Also try to read from .env file if it exists (for local development)
	if _, err := os.Stat(".env"); err == nil {
		viper.SetConfigFile(".env")
		if err := viper.ReadInConfig(); err != nil {
			log.Printf("Warning: Error reading .env file: %s\n", err)
		}
	}

	// Set default values for required environment variables
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", 5432)
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_NAME", "mcp_db")
	viper.SetDefault("DB_PASSWORD", "")
	viper.SetDefault("DB_TIMEZONE", "Asia/Ulaanbaatar")

	// Add health check endpoint first
	http.HandleFunc("/", healthCheckHandler)
	http.HandleFunc("/call-function", MCPHandler)
	
	// Start HTTP server in a goroutine
	go func() {
		log.Println("MCP Server listening on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Initialize database in background
	go func() {
		log.Println("Initializing database connection...")
		initDatabase()
	}()

	// Keep the main goroutine alive
	select {}
}

func initDatabase() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Database initialization failed: %v", r)
			log.Println("Server will continue without database connection")
			return
		}
	}()
	
	// Log environment variables for debugging
	log.Printf("DB_HOST: %s", viper.GetString("DB_HOST"))
	log.Printf("DB_PORT: %d", viper.GetInt("DB_PORT"))
	log.Printf("DB_USER: %s", viper.GetString("DB_USER"))
	log.Printf("DB_NAME: %s", viper.GetString("DB_NAME"))
	log.Printf("DB_TIMEZONE: %s", viper.GetString("DB_TIMEZONE"))
	
	database.CreateClient()
	dbInitialized = true
	log.Println("Database connection established successfully")
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Health check request from %s", r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	status := map[string]interface{}{
		"status": "healthy",
		"message": "MCP Server is running",
		"database": dbInitialized,
	}
	
	json.NewEncoder(w).Encode(status)
}

func MCPHandler(w http.ResponseWriter, r *http.Request) {
	if !dbInitialized {
		http.Error(w, "Database not ready", http.StatusServiceUnavailable)
		return
	}

	var call FunctionCall
	if err := json.NewDecoder(r.Body).Decode(&call); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var result interface{}
	switch call.Function {
	case "get_teams":
		fmt.Println("get_teams")
		var teams []database.Team
		database.DB.Find(&teams)
		result = teams

	case "get_users":
		type UserResponse struct {
			ID        uint   `json:"id"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Email     string `json:"email"`
		}

		fmt.Println("get_users")
		var users []UserResponse
		database.DB.Model(&database.User{}).Select("id, first_name, last_name, email").Find(&users)
		result = users

	case "create_absence_request":
		userEmail := call.Args["user_email"].(string)
		startDateStr := call.Args["start_date"].(string)
		endDateStr := call.Args["end_date"].(string)
		reason := call.Args["reason"].(string)
		inActiveHours := call.Args["in_active_hours"].(float64)

		fmt.Println(userEmail, startDateStr, endDateStr, reason)

		var user database.User
		if err := database.DB.Preload("Team").Preload("Role").Where("email = ?", userEmail).First(&user).Error; err != nil {
			fmt.Println("User not found", err)
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		leader, err := GetUserLeader(database.DB, user, *user.Role)
		if err != nil || leader == nil {
			fmt.Println("Leader not found")
			http.Error(w, "Leader not found", http.StatusNotFound)
			return
		}

		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			fmt.Println("Invalid start_date format")
			http.Error(w, "Invalid start_date format", http.StatusBadRequest)
			return
		}

		instance := database.Absence{
			CreatedUserID: user.ID,
			StartDate:     startDate,
			Reason:        reason,
			EmployeeID:    user.ID,
			InActiveHours: inActiveHours,
			Status:        "pending",
			LeaderID:      1,
		}

		if err := database.DB.Create(&instance).Error; err != nil {
			fmt.Println("Failed to create absence request")
			http.Error(w, "Failed to create absence request", http.StatusInternalServerError)
			return
		}

		if err := smtp.CreateClient().Send(smtp.EmailInput{
			Template: "request",
			Email:    "darkhanbayar@fibo.cloud",
			MultiBcc: []string{"darkhanbayar@fibo.cloud"},
		}, map[string]interface{}{
			"employee_email": "darkhanbayar@fibo.cloud",
			"start_date":     startDateStr,
			"end_date":       endDateStr,
			"reason":         reason,
		}); err != nil {
			fmt.Println("Failed to send email", err)
			http.Error(w, "Failed to send email", http.StatusInternalServerError)
			return
		}

		// TEAM INTEGRION -> FIBO CLOUD chat ym yvuulna

		fmt.Println("DONE")
		result = "DONE"
	default:
		http.Error(w, "Unknown function", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(FunctionResponse{Result: result})
}
