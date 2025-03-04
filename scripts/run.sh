#!/bin/bash

# Get the project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

./bin/leaked-passwords-checker --config="./configs/leaked-passwords-checker.json"
