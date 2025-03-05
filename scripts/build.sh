#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

# Get the script's directory (./scripts/)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Navigate to the project root
PROJECT_ROOT="$SCRIPT_DIR/.."
cd "$PROJECT_ROOT" || exit 1

# Create the output directory if it doesn't exist
mkdir -p "./bin"

rm "./bin/leaked-passwords-checker" "./bin/downloader"

echo "Building leaked-passwords-checker..."

# Build the project
CGO_ENABLED=1 go build -o ./bin/leaked-passwords-checker .

# Check if the build was successful
if [ $? -ne 0 ]; then
  echo "Build failed (leaked-passwords-checker). Exiting."
  exit 1
fi

echo "Building downloader..."

CGO_ENABLED=1 go build -o ./bin/downloader ./tools/downloader.go

# Check if the build was successful
if [ $? -ne 0 ]; then
  echo "Build failed (downloader). Exiting."
  exit 1
fi

echo "Build successful!"
