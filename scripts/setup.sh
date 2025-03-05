#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

if ! command -v sqlite3 &> /dev/null; then
    echo "sqlite3 is not installed. Please install it to proceed."
    exit 1
fi

# Get the script's directory (./scripts/)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Navigate to the project root
PROJECT_ROOT="$SCRIPT_DIR/.."
cd "$PROJECT_ROOT" || exit 1

mkdir -p bin logs database

if [ ! -f "logs/leaked-passwords-checker.log" ]; then
    echo "Log file not found. Creating empty log file..."
    touch "logs/leaked-passwords-checker.log"
else
    echo "Log file already exists. Skipping log file creation."
fi

if [ ! -f "configs/leaked-passwords-checker.example.json" ]; then
    echo "Config file not found. Copying from example..."
    cp "configs/leaked-passwords-checker.example.json" "configs/leaked-passwords-checker.json"
else
    echo "Config file already exists. Skipping config file creation."
fi

DB_FILE="database/leaked-passwords-checker.db"

if [ ! -f "$DB_FILE" ]; then
    echo "Database file not found. Creating database using setup/database.sql..."
    sqlite3 "$DB_FILE" < "setup/database.sql"

    # Generate a random 20-character alphanumeric API key.
    API_KEY=$(head /dev/urandom | tr -dc 'A-Za-z0-9' | head -c 20)
    sqlite3 "$DB_FILE" "INSERT OR IGNORE INTO api_keys(key_id) VALUES ('$API_KEY');"

    echo "==========================="
    echo "Generated API Key: $API_KEY"
    echo "Generated API Key: $API_KEY"
    echo "Generated API Key: $API_KEY"
    echo "==========================="
    echo "Please use this key as X-API-Key header with your requests"
else
    echo "Database file already exists. Skipping database creation."
fi

echo "Setup was finished successfully."
