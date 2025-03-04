#!/bin/sh

set -e

echo "Starting leaked-passwords-checker..."

/app/bin/leaked-passwords-checker --config="/app/configs/leaked-passwords-checker.json"
