package main

import (
	"encoding/json"
	"fmt"
	"log"
	"mcp-server/database"
	"mcp-server/smtp"
	"net/http"
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

func main() {
	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error on load config: %s\n", err)
	}

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
		description := call.Args["description"].(string)

		fmt.Println(userEmail, startDateStr, endDateStr, reason, description)

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

		startDate, err := time.Parse("2006-01-02T15:04:05.000Z", startDateStr)
		if err != nil {
			// Try alternative format
			startDate, err = time.Parse("2006-01-02", startDateStr)
			if err != nil {
				fmt.Println("Invalid start_date format")
				http.Error(w, "Invalid start_date format", http.StatusBadRequest)
				return
			}
		}

		// Тохирох интервалыг олох
		var interval database.TimeInterval
		if err := database.DB.Where("begin_date <= ? AND end_date >= ?", startDate, startDate).First(&interval).Error; err != nil {
			fmt.Println("No matching interval found for start_date:", startDate)
			http.Error(w, "No matching interval found for the given start date", http.StatusBadRequest)
			return
		}

		instance := database.Absence{
			CreatedUserID: user.ID,
			StartDate:     startDate,
			Reason:        reason,
			EmployeeID:    user.ID,
			InActiveHours: inActiveHours,
			Status:        "pending",
			LeaderID:      leader.ID,
			IntervalID:    interval.ID,
			Description:   description,
		}

		if err := database.DB.Create(&instance).Error; err != nil {
			fmt.Println("Failed to create absence request")
			http.Error(w, "Failed to create absence request", http.StatusInternalServerError)
			return
		}

		if err := smtp.CreateClient().Send(smtp.EmailInput{
			Template: "request",
			Email:    "tuvshinjargal@fibo.cloud",
			MultiBcc: []string{"tuvshinjargal@fibo.cloud"},
		}, map[string]interface{}{
			"employee_email": "tuvshinjargal@fibo.cloud",
			"start_date":     startDateStr,
			"end_date":       endDateStr,
			"reason":         reason,
		}); err != nil {
			fmt.Println("Failed to send email", err)
			http.Error(w, "Failed to send email", http.StatusInternalServerError)
			return
		}

		// TEAM INTEGRION -> FIBO CLOUD chat ym yvuulna, goy bainadaa

		fmt.Printf("Absence request created successfully with ID: %d\n", instance.ID)
		result = map[string]interface{}{
			"message":    "Absence request created successfully",
			"absence_id": instance.ID,
			"status":     instance.Status,
		}
		
	case "approve_absence":
		absenceID := uint(call.Args["absence_id"].(float64))
		comment := ""
		if call.Args["comment"] != nil {
			comment = call.Args["comment"].(string)
		}

		fmt.Println("approve_absence", absenceID, comment)

		var absence database.Absence
		if err := database.DB.First(&absence, absenceID).Error; err != nil {
			fmt.Println("Absence not found", err)
			http.Error(w, "Absence not found", http.StatusNotFound)
			return
		}

		// Аль хэдийн шийдэгдсэн эсэхийг шалгах
		if absence.Status != "pending" {
			fmt.Println("Absence already processed")
			http.Error(w, "Absence already processed", http.StatusBadRequest)
			return
		}

		// Status-г approved болгох
		absence.Status = "approved"
		absence.UpdatedAt = time.Now()

		if comment != "" {
			fmt.Println("Approval comment:", comment)
		}

		// Database-д хадгалах
		if err := database.DB.Save(&absence).Error; err != nil {
			fmt.Println("Failed to approve absence", err)
			http.Error(w, "Failed to approve absence", http.StatusInternalServerError)
			return
		}

		fmt.Println("Absence approved successfully")
		result = "Absence approved successfully"

	case "reject_absence":
		absenceID := uint(call.Args["absence_id"].(float64))
		comment := ""
		if call.Args["comment"] != nil {
			comment = call.Args["comment"].(string)
		}

		fmt.Println("reject_absence", absenceID, comment)

		var absence database.Absence
		if err := database.DB.First(&absence, absenceID).Error; err != nil {
			fmt.Println("Absence not found", err)
			http.Error(w, "Absence not found", http.StatusNotFound)
			return
		}

		// Аль хэдийн шийдэгдсэн эсэхийг шалгах
		if absence.Status != "pending" {
			fmt.Println("Absence already processed")
			http.Error(w, "Absence already processed", http.StatusBadRequest)
			return
		}

		// Status-г rejected болгох
		absence.Status = "rejected"
		absence.UpdatedAt = time.Now()

		if comment != "" {
			fmt.Println("Rejection comment:", comment)
		}

		// Database-д хадгалах
		if err := database.DB.Save(&absence).Error; err != nil {
			fmt.Println("Failed to reject absence", err)
			http.Error(w, "Failed to reject absence", http.StatusInternalServerError)
			return
		}

		fmt.Println("Absence rejected successfully")
		result = "Absence rejected successfully"

	case "get_time_intervals":
		startDateStr := call.Args["start_date"].(string)
		
		fmt.Println("get_time_intervals", startDateStr)

		var intervals []database.TimeInterval
		if err := database.DB.Where("end_date::date >= date(?)", startDateStr).Find(&intervals).Error; err != nil {
			fmt.Println("Failed to get time intervals", err)
			http.Error(w, "Failed to get time intervals", http.StatusInternalServerError)
			return
		}
		
		result = intervals

	default:
		http.Error(w, "Unknown function", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(FunctionResponse{Result: result})
}

// Чөлөөний хүсэлт зөвшөөрөх/татгалзах параметрууд
type AbsenceApprovalParam struct {
	AbsenceID uint   `json:"absence_id" binding:"required"`
	Comment   string `json:"comment,omitempty"`
}
