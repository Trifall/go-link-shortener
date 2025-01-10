#!/bin/bash
set -e # Exit on error

# Color coding for better readability and user experience
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

chmod +x ./scripts/*.sh

# Get the directory of the running script
SCRIPT_DIR=$(dirname "$(realpath "$0")")

# Set the PGPASSFILE to the repo directory
export PGPASSFILE="$SCRIPT_DIR/../.pgpass"

if [[ ! -f "$PGPASSFILE" ]]; then
  echo -e "${RED}Error: .pgpass does not exist. Exiting...${NC}"
  exit 1
fi

echo "Running create_db_user.sh..."
./scripts/create_db_user.sh

echo "Running init_env.sh..."
./scripts/init_env.sh

echo "All scripts executed successfully!"
