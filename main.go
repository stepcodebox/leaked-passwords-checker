package main

import (
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type Config struct {
	DatabaseFilepath string `json:"database_filepath"`
	LogFilepath      string `json:"log_filepath"`
}

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

type CheckPasswordRequest struct {
	Password string `json:"password"`
}

type CheckPasswordResponse struct {
	Result        string `json:"result"`
	HasBeenLeaked bool   `json:"has_been_leaked"`
}

type App struct {
	DB         *sql.DB
	StmtCheck  *sql.Stmt
	StmtAPIKey *sql.Stmt
}

// Constructor
func NewApp(db *sql.DB) (*App, error) {
	stmtCheck, err := db.Prepare("SELECT EXISTS(SELECT 1 FROM passwords WHERE sha1 = ?)")
	if err != nil {
		return nil, err
	}

	stmtAPIKey, err := db.Prepare("SELECT EXISTS(SELECT 1 FROM api_keys WHERE key_id = ?)")
	if err != nil {
		stmtCheck.Close()
		return nil, err
	}

	return &App{
		DB:         db,
		StmtCheck:  stmtCheck,
		StmtAPIKey: stmtAPIKey,
	}, nil
}

// Destructor
func (a *App) Close() {
	a.StmtCheck.Close()
	a.StmtAPIKey.Close()
	a.DB.Close()
}

func (a *App) checkPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	var req CheckPasswordRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Password == "" {
		log.Printf("Error decoding request: %+v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	sha1Hash := calculateSHA1(req.Password)

	hasBeenLeaked, err := a.isPasswordLeaked(sha1Hash)
	if err != nil {
		log.Printf("Error checking password: %+v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := CheckPasswordResponse{
		Result:        "success",
		HasBeenLeaked: hasBeenLeaked,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Error encoding response: %+v", err)
	}
}

func (a *App) isPasswordLeaked(sha1 string) (bool, error) {
	var exists bool
	err := a.StmtCheck.QueryRow(sha1).Scan(&exists)
	return exists, err
}

func calculateSHA1(password string) string {
	hash := sha1.New()
	hash.Write([]byte(password))
	return strings.ToUpper(hex.EncodeToString(hash.Sum(nil)))
}

// API key check
func (a *App) apiKeyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-API-Key")
		if key == "" {
			http.Error(w, "API key required", http.StatusUnauthorized)
			return
		}

		var exists bool
		err := a.StmtAPIKey.QueryRow(key).Scan(&exists)
		if err != nil || !exists {
			http.Error(w, "Invalid API key", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	// Load configuration
	config, err := loadConfig("configs/leaked-passwords-checker.json")
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

	// Use the database filepath from the config
	db, err := sql.Open("sqlite3", config.DatabaseFilepath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialise the application
	app, err := NewApp(db)
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}
	defer app.Close()

	// Setup HTTP handler with API key middleware
	http.Handle("/v1/password/check", app.apiKeyMiddleware(http.HandlerFunc(app.checkPasswordHandler)))

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
