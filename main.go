package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// SlackRequestBody structure to match the expected JSON
type SlackRequestBody struct {
	Text string `json:"text"`
}

type Record struct {
	UUID      string
	DateAdded time.Time
}

func executeSQLFile(db *sql.DB, filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	sql := string(content)
	_, err = db.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}

// getPendingRecords queries the database for records with a status of "PENDING"
func getPendingRecords(db *sql.DB) ([]Record, error) {
	var records []Record

	sqlStatement := `SELECT uuid, date_added FROM test_slack_data WHERE status = 'PENDING' AND date_added > NOW() - INTERVAL '1 hour'`
	rows, err := db.Query(sqlStatement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var r Record
		if err := rows.Scan(&r.UUID, &r.DateAdded); err != nil {
			return nil, err
		}
		records = append(records, r)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

// SendSlackNotification sends a notification to a Slack channel via webhook
func SendSlackNotification(webhookURL string, msg string) error {
	slackBody, _ := json.Marshal(SlackRequestBody{Text: msg})
	req, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewBuffer(slackBody))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	if buf.String() != "ok" {
		log.Fatalf("Non-ok response returned from Slack")
	}

	return nil
}

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	// Get environment variables
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	slackWebhookUrl := os.Getenv("SLACK_WEBHOOK_URL")
	// Connect to the PostgreSQL database
	// PostgreSQL connection string
	psqlConnString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// Connect to the PostgreSQL database
	db, err := sql.Open("postgres", psqlConnString)
	if err != nil {
		log.Fatalf("Failed to open a DB connection: %v", err)
	}
	defer db.Close()
	// Check init db first
	err = executeSQLFile(db, "sql/init-db.sql")
	if err != nil {
		fmt.Println(err)
	}
	// Query the database
	records, err := getPendingRecords(db)
	if err != nil {
		log.Fatal(err)
	}

	// Send each record to Slack
	for _, record := range records {
		message := fmt.Sprintf("UUID: %s, Date Added: %s", record.UUID, record.DateAdded)
		err := SendSlackNotification(slackWebhookUrl, message)
		if err != nil {
			log.Printf("Error sending notification to Slack: %v", err)
		}
	}

	fmt.Printf("Message(s) successfully sent to channel\n")
}
