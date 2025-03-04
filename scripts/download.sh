#!/bin/bash

# Get the project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

./bin/downloader --config="./configs/leaked-passwords-checker.json"
