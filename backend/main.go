package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

type WorkOrder struct {
	Date      string `json:"date"`
	Inspector string `json:"inspector"`
	Address   string `json:"address"`
	Floor     string `json:"floor"`
	Unit      string `json:"unit"`
	Phone     string `json:"phone"`
	Room      string `json:"room"`
	Findings  string `json:"findings"`
	Signature string `json:"signature"`
}

func submitHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var wo WorkOrder
		err := json.NewDecoder(r.Body).Decode(&wo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err = db.Exec(`INSERT INTO workorders (date, inspector, address, floor, unit, phone, room, findings, signature)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			wo.Date, wo.Inspector, wo.Address, wo.Floor, wo.Unit, wo.Phone, wo.Room, wo.Findings, wo.Signature)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Submitted successfully")
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
