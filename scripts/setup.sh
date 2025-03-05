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
    touch "logs/leaked-passwords-checker.log"
else
    echo "Log file already exists. Skipping log file creation."
fi

if [ ! -f "configs/leaked-passwords-checker.example.json" ]; then
    cp "configs/leaked-passwords-checker.example.json" "configs/leaked-passwords-checker.json"
else
    echo "Config file already exists. Skipping config file creation."
fi

cp "/home/stephen/Projects/Leaked Passwords Checker/Git Repositories/leaked-passwords-checker/configs/leaked-passwords-checker.example.json"

if [ ! -f "database/leaked-passwords-checker.db" ]; then
    sqlite3 "database/leaked-passwords-checker.db" < "setup/database.sql"
else
    echo "Database file already exists. Skipping database creation."
fi

echo "Setup was finished successfully."
