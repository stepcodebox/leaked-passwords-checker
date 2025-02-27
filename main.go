package main

import (
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

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

// A method on App
func (a *App) checkPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	var req CheckPasswordRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Password == "" {
		fmt.Printf("%+v\n", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	sha1Hash := calculateSHA1(req.Password)

	hasBeenLeaked, err := a.isPasswordLeaked(sha1Hash)
	if err != nil {
		fmt.Printf("%+v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := CheckPasswordResponse{
		Result:        "success",
		HasBeenLeaked: hasBeenLeaked,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
	var err error
	db, err := sql.Open("sqlite3", "./database/leaked-passwords-checker.db")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Use App's constructor
	app, err := NewApp(db)
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}
	defer app.Close()

	http.Handle("/v1/password/check", app.apiKeyMiddleware(http.HandlerFunc(app.checkPasswordHandler)))

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
