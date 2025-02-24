#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

# Get the script's directory (./scripts/)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Navigate to the project root
PROJECT_ROOT="$SCRIPT_DIR/.."
cd "$PROJECT_ROOT" || exit 1

echo "Formatting Go code..."
go fmt ./...

echo "Code formatting complete!"
