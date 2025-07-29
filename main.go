package main

import (
	"encoding/json"
	"fmt"
	"log"
	"mcp-server/database"
	"mcp-server/smtp"
	"net/http"
	"time"
)

type FunctionCall struct {
	Function string                 `json:"function"`
	Args     map[string]interface{} `json:"args"`
}

type FunctionResponse struct {
	Result interface{} `json:"result"`
}

func main() {
	// Try to read .env file, but don't fail if it doesn't exist
	// viper.SetConfigFile(".env")
	// if err := viper.ReadInConfig(); err != nil {
	// 	log.Println("No .env file found, using environment variables")
	// }
	
	// Enable automatic environment variable reading
	// viper.AutomaticEnv()

	database.CreateClient()

	http.HandleFunc("/call-function", MCPHandler)
	log.Println("MCP Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func MCPHandler(w http.ResponseWriter, r *http.Request) {
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
