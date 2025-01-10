#!/bin/bash

# Color coding for better visibility
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

# Check if credentials file exists
if [ ! -f "db_credentials.tmp" ]; then
  echo -e "${RED}[SETUP_ENV] Error: Database credentials not found. Exiting process.${NC}"
  exit 1
fi

# Read credentials from temporary file
DB_USER=$(grep "Database User:" db_credentials.tmp | cut -d' ' -f3)
DB_PASSWORD=$(grep "Password:" db_credentials.tmp | cut -d' ' -f2)
DB_NAME=$(grep "Database Name:" db_credentials.tmp | cut -d' ' -f3)
ROOT_USER_KEY=$(openssl rand -base64 32)

# Create .env and .env.example files
echo -e "${BLUE}Creating environment files...${NC}"

# Create .env file with actual credentials
cat >.env <<EOF
DB_HOST=localhost
DB_PORT=5432
DB_USER=$DB_USER
DB_PASSWORD=$DB_PASSWORD
DB_NAME=$DB_NAME
DB_SSLMODE=disable
ROOT_USER_KEY=$ROOT_USER_KEY
EOF

# Create .env.example without sensitive data
cat >.env.example <<EOF
DB_HOST=localhost
DB_PORT=5432
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_NAME=your_db_name
DB_SSLMODE=disable
EOF

# Create .gitignore if it doesn't exist
if [ ! -f ".gitignore" ]; then
  echo -e "${BLUE}Creating .gitignore file...${NC}"
  cat >.gitignore <<EOF
# Environment files
.env
.env.*
!.env.example

# Database credentials
scripts/db_credentials.tmp

# Compiled binary
bin/
EOF
fi

echo -e "${BLUE}Removing temporary credentials file...${NC}"
# Remove credentials file
rm db_credentials.tmp
echo -e "${GREEN}Successfully removed temporary credentials file${NC}"

echo -e "${GREEN}Environment setup complete!${NC}"
