package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	_ "modernc.org/sqlite"
)

type Job struct {
	ID        int
	Title     string
	Company   string
	Salary    int
	CreatedAt time.Time
}

type JobsResponse struct {
	Items       []Job
	NextAfterID *int
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite", "jobs.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS jobs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		company TEXT NOT NULL,
		salary INTEGER NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`)
	if err != nil {
		log.Fatal(err)
	}

	db.Exec(`DELETE FROM jobs;`)
	_, err = db.Exec(`
	INSERT INTO jobs (title, company, salary) VALUES
	('Go Developer', 'Kolesa', 615000),
	('Frontend Developer', 'Kolesa', 550000),
	('Backend Developer', 'Google', 800000),
	('Data Engineer', 'Kolesa', 700000),
	('QA Tester', 'Meta', 400000);
	`)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/jobs", getJobsHandler)

	fmt.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getJobsHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	company := r.URL.Query().Get("company")
	afterIDStr := r.URL.Query().Get("after_id")
	limitStr := r.URL.Query().Get("limit")

	limit := 10
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}

	where := "WHERE 1=1"
	args := []interface{}{}

	if company != "" {
		where += " AND company = ?"
		args = append(args, company)
	}

	if afterIDStr != "" {
		afterID, err := strconv.Atoi(afterIDStr)
		if err != nil {
			http.Error(w, "invalid after_id", http.StatusBadRequest)
			return
		}

		var cursorTime time.Time
		row := db.QueryRow("SELECT created_at FROM jobs WHERE id = ?", afterID)
		if err := row.Scan(&cursorTime); err != nil {
			http.Error(w, "after_id not found", http.StatusBadRequest)
			return
		}

		where += " AND (created_at < ? OR (created_at = ? AND id < ?))"
		args = append(args, cursorTime, cursorTime, afterID)
	}

	query := fmt.Sprintf(`
		SELECT id, title, company, salary, created_at
		FROM jobs
		%s
		ORDER BY created_at DESC, id DESC
		LIMIT ?;
	`, where)
	args = append(args, limit)

	queryStart := time.Now()
	rows, err := db.Query(query, args...)
	queryDuration := time.Since(queryStart)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var jobs []Job
	for rows.Next() {
		var j Job
		if err := rows.Scan(&j.ID, &j.Title, &j.Company, &j.Salary, &j.CreatedAt); err != nil {
			log.Println("Scan error:", err)
			continue
		}
		jobs = append(jobs, j)
	}

	var nextAfterID *int
	if len(jobs) > 0 {
		lastID := jobs[len(jobs)-1].ID
		nextAfterID = &lastID
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Query-Time", queryDuration.String())

	resp := JobsResponse{
		Items:       jobs,
		NextAfterID: nextAfterID,
	}

	json.NewEncoder(w).Encode(resp)
	log.Printf("Handled /jobs in %v (SQL time: %v)\n", time.Since(start), queryDuration)
}
