#!/bin/bash

# Color coding for better readability and user experience
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

# Generate a secure random password
RANDOM_PASSWORD=$(openssl rand -base64 32)

# Database configuration
DB_USER="urlapp"
DB_NAME="urlshortener"

# Check if the user or database already exists
USER_EXISTS=$(psql -U postgres -tAc "SELECT 1 FROM pg_roles WHERE rolname='$DB_USER';")
DB_EXISTS=$(psql -U postgres -tAc "SELECT 1 FROM pg_database WHERE datname='$DB_NAME';")

if [[ $USER_EXISTS == "1" || $DB_EXISTS == "1" ]]; then
  echo -e "${RED}User or database already exists.${NC}"
  read -p "Do you want to delete the existing user and database? (Y/N): " DELETE_CONFIRM

  if [[ $DELETE_CONFIRM == "Y" || $DELETE_CONFIRM == "y" ]]; then
    echo -e "${BLUE}Deleting existing user and database...${NC}"

    # Terminate all active connections to the database
    echo -e "${BLUE}Terminating active connections to the database...${NC}"
    psql -U postgres -c "SELECT pg_terminate_backend(pg_stat_activity.pid) FROM pg_stat_activity WHERE pg_stat_activity.datname = '$DB_NAME';"

    # Drop the database if it exists
    echo -e "${BLUE}Dropping database...${NC}"
    psql -U postgres -c "DROP DATABASE IF EXISTS $DB_NAME;"

    # Terminate all active connections to the user
    echo -e "${BLUE}Terminating active connections to the user...${NC}"
    psql -U postgres -c "SELECT pg_terminate_backend(pg_stat_activity.pid) FROM pg_stat_activity WHERE pg_stat_activity.usename = '$DB_USER';"

    # Drop the user if it exists
    echo -e "${BLUE}Dropping user...${NC}"
    psql -U postgres -c "DROP USER IF EXISTS $DB_USER;"

    echo -e "${GREEN}Existing user and database deleted.${NC}"
  else
    echo -e "${RED}Exiting without making changes.${NC}"
    exit 0
  fi
fi

echo -e "${BLUE}Creating new PostgreSQL user and database...${NC}"

# Create the user and database with proper permissions
psql -U postgres <<EOF
-- Create user with generated password
CREATE USER $DB_USER WITH ENCRYPTED PASSWORD '$RANDOM_PASSWORD';

-- Create database with the user as the owner
CREATE DATABASE $DB_NAME OWNER $DB_USER;

-- Grant necessary permissions
GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;

-- Connect to the database to set additional permissions
\c $DB_NAME

-- Grant schema permissions
GRANT ALL ON SCHEMA public TO $DB_USER;
EOF

# Save credentials to a temporary file
echo -e "${GREEN}Creating temporary credentials file...${NC}"
cat >db_credentials.tmp <<EOF
Database User: $DB_USER
Database Name: $DB_NAME
Password: $RANDOM_PASSWORD
EOF

echo -e "${GREEN}Successfully created database user and saved credentials to db_credentials.tmp${NC}"
