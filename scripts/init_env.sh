#!/bin/bash

# Color coding for better visibility
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Initialize variables from existing .env if present
if [ -f .env ]; then
  echo -e "${BLUE}[SETUP_ENV] Loading existing .env file...${NC}"
  export $(grep -v '^#' .env | xargs)
fi

# Load temporary credentials if available
if [ -f db_credentials.tmp ]; then
  echo -e "${BLUE}[SETUP_ENV] Loading temporary credentials...${NC}"
  TEMP_USER=$(grep "Database User:" db_credentials.tmp | awk '{print $3}')
  TEMP_NAME=$(grep "Database Name:" db_credentials.tmp | awk '{print $3}')
  TEMP_PASS=$(grep "Password:" db_credentials.tmp | awk '{print $2}')

  # Update only if temp values exist
  [ -n "$TEMP_USER" ] && DB_USER="$TEMP_USER"
  [ -n "$TEMP_NAME" ] && DB_NAME="$TEMP_NAME"
  [ -n "$TEMP_PASS" ] && DB_PASSWORD="$TEMP_PASS"
fi

# Generate root user key if not defined
if [ -z "$ROOT_USER_KEY" ]; then
  echo -e "${BLUE}[SETUP_ENV] Generating root user key...${NC}"
  ROOT_USER_KEY=$(openssl rand -base64 32)
fi

# Create/update .env file with preserved comments
echo -e "${BLUE}[SETUP_ENV] Creating/updating environment files...${NC}"
{
  echo "#postgres info"
  echo "DB_HOST=${DB_HOST:-localhost}"
  echo "DB_PORT=${DB_PORT:-5434}"
  echo "DB_USER=${DB_USER:-urlapp}"
  echo "DB_PASSWORD=${DB_PASSWORD:-your_db_password}"
  echo "DB_NAME=${DB_NAME:-urlshortener}"
  echo "DB_SSLMODE=${DB_SSLMODE:-disable}"
  echo ""
  echo "# Server Port (default: 8080)"
  echo "SERVER_PORT=${SERVER_PORT:-8080}"
  echo "# required to create the initial root user, use 'openssl rand -base64 32' to generate one"
  echo "ROOT_USER_KEY=${ROOT_USER_KEY}"
  echo "# can be error, warning, info"
  echo "LOG_LEVEL=${LOG_LEVEL:-error}"
  echo "# depends on where you are hosting it, used to filter out for loops"
  echo "PUBLIC_SITE_URL=${PUBLIC_SITE_URL:-example.com}"
  echo "# enable API documentation at the /docs/ endpoint"
  echo "ENABLE_DOCS=${ENABLE_DOCS:-true}"
} >.env.tmp

# Preserve existing comments and structure from .env.example
if [ -f .env.example ]; then
  awk '/^#/ {print} /^[^#]/ {while(getline line < ".env.tmp") print line; close(".env.tmp"); exit}' .env.example >.env
else
  mv .env.tmp .env
fi

# Create .gitignore if missing
if [ ! -f .gitignore ]; then
  echo -e "${BLUE}[SETUP_ENV] Creating .gitignore...${NC}"
  cat >.gitignore <<EOF
# Environment files
.env
.env.*
!.env.example

# Database credentials
db_credentials.tmp

# Compiled binary
bin/
EOF
fi

# Cleanup temporary files
rm -f .env.tmp
if [ -f db_credentials.tmp ]; then
  echo -e "${BLUE}[SETUP_ENV] Removing temporary credentials file...${NC}"
  rm db_credentials.tmp
fi

echo -e "${GREEN}[SETUP_ENV] Environment setup complete!${NC}"
