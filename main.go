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

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./database/leaked-passwords-checker.db")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	http.HandleFunc("/v1/password/check", checkPasswordHandler)

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func checkPasswordHandler(w http.ResponseWriter, r *http.Request) {
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

	hasBeenLeaked, err := isPasswordLeaked(sha1Hash)
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

func calculateSHA1(password string) string {
	hash := sha1.New()
	hash.Write([]byte(password))
	return strings.ToUpper(hex.EncodeToString(hash.Sum(nil)))
}

func isPasswordLeaked(sha1 string) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM passwords WHERE sha1 = ?)", sha1).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
