#!/bin/bash

# Get the project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

cd infra

docker compose up --build --force-recreate -d
