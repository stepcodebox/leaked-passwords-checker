package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Config holds configuration settings.
type Config struct {
	DatabaseFilepath string `json:"database_filepath"`
	LogFilepath      string `json:"log_filepath"`
}

// loadConfig reads the configuration from the specified file.
func loadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

const (
	apiBaseURL  = "https://api.pwnedpasswords.com/range/"
	parallelism = 3           // Number of parallel API requests
	totalRanges = 1024 * 1024 // 1,048,576 possible 5-character prefixes
	retryLimit  = 10
	// Constants for insertion retry logic:
	insertRetryLimit = 100
	insertRetryDelay = 100 * time.Millisecond
	createTableQuery = `
		CREATE TABLE IF NOT EXISTS passwords (
			sha1 TEXT PRIMARY KEY
		);
	`
)

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "./configs/leaked-passwords-checker.json", "Path to configuration file")
	flag.Parse()

	// Load configuration using the provided path
	config, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Open log file and redirect default logger to it
	logFile, err := os.OpenFile(config.LogFilepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Println("Starting downloader...") // to stdout for immediate feedback
	log.Println("Starting download...")

	// Open SQLite database using the config value
	db, err := sql.Open("sqlite3", config.DatabaseFilepath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		log.Fatalf("Failed to set WAL mode: %v", err)
	}

	_, err = db.Exec("PRAGMA busy_timeout=5000;")
	if err != nil {
		log.Fatalf("Failed to set busy_timeout: %v", err)
	}

	// Ensure the table exists
	if _, err := db.Exec(createTableQuery); err != nil {
		log.Fatalf("Failed to set up database schema: %v", err)
	}

	// Process hash ranges in parallel
	wg := sync.WaitGroup{}
	taskCh := make(chan int, totalRanges)

	// Generate all ranges
	go func() {
		for i := 0; i < totalRanges; i++ {
			taskCh <- i
		}
		close(taskCh)
	}()

	// Create worker pool
	for i := 0; i < parallelism; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for rangeID := range taskCh {
				if rangeID%1000 == 0 {
					log.Printf("Processing range %d", rangeID)
				}
				if err := processRange(rangeID, db); err != nil {
					log.Printf("Error processing range %d: %v", rangeID, err)
				}
			}
		}()
	}

	wg.Wait()
	log.Println("Download and processing complete.")
}

// processRange fetches a hash range and stores it in the database.
func processRange(rangeID int, db *sql.DB) error {
	hashPrefix := getHashRange(rangeID)
	url := apiBaseURL + hashPrefix
	var resp *http.Response
	var err error

	// Retry logic with resource cleanup
	for attempt := 1; attempt <= retryLimit; attempt++ {
		resp, err = http.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		log.Printf("Attempt %d failed for range %s: %v", attempt, hashPrefix, err)
		time.Sleep(500 * time.Millisecond)
		if attempt == retryLimit {
			return fmt.Errorf("failed to fetch range after %d attempts: %v", retryLimit, err)
		}
	}
	defer resp.Body.Close()

	// Process and insert into the database
	return processResponse(hashPrefix, resp.Body, db)
}

// processResponse processes the API response and stores hashes in the database.
func processResponse(hashPrefix string, body io.Reader, db *sql.DB) error {
	scanner := bufio.NewScanner(body)
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	stmt, err := tx.Prepare("INSERT OR IGNORE INTO passwords (sha1) VALUES (?)")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			log.Printf("Skipping invalid line: %s", line)
			continue
		}
		fullHash := hashPrefix + parts[0]
		var execErr error
		// Retry insertion if the error indicates the database is locked.
		for i := 0; i < insertRetryLimit; i++ {
			_, execErr = stmt.Exec(fullHash)
			if execErr == nil {
				break
			}
			if strings.Contains(execErr.Error(), "database is locked") {
				time.Sleep(insertRetryDelay)
				continue
			} else {
				break
			}
		}
		if execErr != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert hash %s: %v", fullHash, execErr)
		}
	}
	if err := scanner.Err(); err != nil {
		tx.Rollback()
		return fmt.Errorf("error reading response body: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}
	return nil
}

// getHashRange generates a 5-character uppercase hex string for the given rangeID.
func getHashRange(rangeID int) string {
	return fmt.Sprintf("%05X", rangeID)
}
