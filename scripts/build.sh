#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

# Get the script's directory (./scripts/)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Navigate to the project root
PROJECT_ROOT="$SCRIPT_DIR/.."
cd "$PROJECT_ROOT" || exit 1

# Define variables
PROJECT_NAME="leaked-passwords-checker"
DOWNLOADER_NAME="downloader"
OUTPUT_DIR="./bin"
OUTPUT_MAIN_FILE="$OUTPUT_DIR/$PROJECT_NAME"
OUTPUT_DOWNLOADER_FILE="$OUTPUT_DIR/$DOWNLOADER_NAME"

# Create the output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

echo "Building $PROJECT_NAME..."

# Build the project
go build -o "$OUTPUT_MAIN_FILE" .

# Check if the build was successful
if [ $? -ne 0 ]; then
  echo "Build failed (main). Exiting."
  exit 1
fi

echo "Building $DOWNLOADER_NAME..."

# Build the downloader
go build -o "$OUTPUT_DOWNLOADER_FILE" .

# Check if the build was successful
if [ $? -ne 0 ]; then
  echo "Build failed (downloader). Exiting."
  exit 1
fi

echo "Build successful!"
