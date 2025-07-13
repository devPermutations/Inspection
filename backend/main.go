package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type WorkOrder struct {
	ID        int       `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Date      string    `json:"date"`
	Inspector string    `json:"inspector"`
	Address   string    `json:"address"`
	Floor     string    `json:"floor"`
	Unit      string    `json:"unit"`
	Phone     string    `json:"phone"`
	Room      string    `json:"room"`
	Findings  string    `json:"findings"`
	Signature string    `json:"signature"`
}

func submitHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "https://permutations.app")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var wo WorkOrder
		err := json.NewDecoder(r.Body).Decode(&wo)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		result, err := db.Exec(`INSERT INTO workorders (date, inspector, address, floor, unit, phone, room, findings, signature)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			wo.Date, wo.Inspector, wo.Address, wo.Floor, wo.Unit, wo.Phone, wo.Room, wo.Findings, wo.Signature)
		if err != nil {
			log.Printf("Insert error: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		rows, err := result.RowsAffected()
		if err != nil {
			log.Printf("RowsAffected error: %v", err)
		} else {
			log.Printf("Inserted %d rows", rows)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Submitted successfully"})
	}
}

func retrieveHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		rows, err := db.Query("SELECT id, created_at, date, inspector, address, floor, unit, phone, room, findings, signature FROM workorders ORDER BY created_at DESC")
		if err != nil {
			log.Printf("DB query error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var orders []WorkOrder
		for rows.Next() {
			var wo WorkOrder
			err = rows.Scan(&wo.ID, &wo.CreatedAt, &wo.Date, &wo.Inspector, &wo.Address, &wo.Floor, &wo.Unit, &wo.Phone, &wo.Room, &wo.Findings, &wo.Signature)
			if err != nil {
				log.Printf("Scan error: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			orders = append(orders, wo)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(orders)
	}
}

func main() {
	db := connectDB()
	defer db.Close()

	// Create table if not exists
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS workorders (
		id SERIAL PRIMARY KEY,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		date DATE,
		inspector TEXT,
		address TEXT,
		floor TEXT,
		unit TEXT,
		phone TEXT,
		room TEXT,
		findings TEXT,
		signature TEXT
	)`)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/submit", submitHandler(db))
	http.HandleFunc("/retrieve", retrieveHandler(db))

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func connectDB() *sql.DB {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_NAME"))
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	return db
}
