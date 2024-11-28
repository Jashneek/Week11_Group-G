package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

// Response structure for API responses
type Response struct {
	CurrentTime string `json:"current_time"`
}

var db *sql.DB

func main() {
	// Initialize the database connection
	var err error
	dsn := "root:Jashneek@14@tcp(127.0.0.1:3306)/toronto_time" // Replace with your MySQL credentials
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Database is unreachable: %v", err)
	}

	log.Println("Connected to the database successfully.")

	// Set up the router
	r := mux.NewRouter()
	r.HandleFunc("/current-time", getCurrentTime).Methods("GET")
	r.HandleFunc("/all-times", getAllTimes).Methods("GET")

	log.Println("Starting server on port 8080...")
	http.ListenAndServe(":8080", r)
}

// getCurrentTime fetches and returns the current time in Toronto
func getCurrentTime(w http.ResponseWriter, r *http.Request) {
	// Load Toronto timezone
	loc, err := time.LoadLocation("America/Toronto")
	if err != nil {
		http.Error(w, "Error loading timezone", http.StatusInternalServerError)
		log.Printf("Timezone error: %v", err)
		return
	}

	// Get current time in Toronto timezone
	currentTime := time.Now().In(loc)
	response := Response{CurrentTime: currentTime.Format("2006-01-02 15:04:05")}

	// Log the current time to the database
	_, err = db.Exec("INSERT INTO time_log (timestamp) VALUES (?)", currentTime.Format("2006-01-02 15:04:05"))
	if err != nil {
		http.Error(w, "Error logging time to database", http.StatusInternalServerError)
		log.Printf("Database error: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getAllTimes fetches all logged times from the database
func getAllTimes(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, timestamp FROM time_log")
	if err != nil {
		http.Error(w, "Error fetching times from database", http.StatusInternalServerError)
		log.Printf("Database query error: %v", err)
		return
	}
	defer rows.Close()

	var times []Response
	for rows.Next() {
		var id int
		var timestamp string // Use string to scan the raw timestamp value
		if err := rows.Scan(&id, &timestamp); err != nil {
			http.Error(w, "Error reading data", http.StatusInternalServerError)
			log.Printf("Row scan error: %v", err)
			return
		}

		times = append(times, Response{CurrentTime: timestamp})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(times)
}
