# go-link-shortener

A URL shortener written in Go using PostgreSQL. There is a sibling project at [Link Shortener UI](https://github.com/Trifall/link-shortener-ui).

## Getting Started

### Prerequisites

1. Clone the repository or if using Docker, see [Running with Docker (recommended, DockerHub)](#running-with-docker-recommended-dockerhub)

### Notes

#### Make File

The `makefile` contains all of the various commands for running the server, whether its through docker or locally. Read it for more info

#### Scripts

- The `scripts` folder is for local development only, if you are running through docker, you can ignore this folder.
  - Read the `Running Locally` section for this, it sets up the PostgreSQL database for you.

#### Environment Variables

- `PUBLIC_SITE_URL`: This is the public URL of the app, it is used to avoid redirect loops.
- `ENABLE_DOCS`: This is a boolean that enables or disables the API docs. If set to 'false', it will allow you to use `/docs` as a valid shortened route.
- `ROOT_USER_KEY`: This is used to create the root user.
- All of the other variables are required for the database connection.

### Running with Docker (recommended, DockerHub)

DockerHub Link: [https://hub.docker.com/r/jerrent/go-link-shortener](https://hub.docker.com/r/jerrent/go-link-shortener)

1. Create `docker-compose.yml`

```yaml
services:
  db:
    image: postgres:16-alpine
    container_name: go-link-shortener-db # name whatever you want
    environment:
      POSTGRES_USER: ${DB_USER:-urlapp} # define here or env. Default is urlapp
      POSTGRES_PASSWORD: ${DB_PASSWORD:-DEFINE_ME_IN_ENV} # define here or env. No default, you must set this.
      POSTGRES_DB: ${DB_NAME:-urlshortener} # define here or env. Default is urlshortener
      LANG: en_US.UTF-8
      LC_CTYPE: en_US.UTF-8
      LC_COLLATE: en_US.UTF-8
      PGTZ: 'America/New_York' # define timezone here. Default is America/New_York
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: [ "CMD", "pg_isready", "-U", "${DB_USER}", "-d", "${DB_NAME}", "-p", "5432" ]
      interval: 10s
      timeout: 5s
      retries: 10
      start_period: 20s
    networks:
      - app-network

  app:
    deploy:
      resources:
        limits:
          memory: 512M # can remove or change, it sets the memory limit for the container
    image: jerrent/go-link-shortener:latest
    container_name: go-link-shortener # name whatever you want
    environment:
      DB_HOST: db
      DB_PORT: ${DB_PORT:-5432} # define here or env. Default is 5432
      DB_USER: ${DB_USER:-urlapp} # define here or env. Default is urlapp
      DB_PASSWORD: ${DB_PASSWORD:?DEFINE_ME_IN_ENV} # define here or env. No default, you must set this.
      DB_NAME: ${DB_NAME:-urlshortener} # define here or env. Default is urlshortener
      DB_SSLMODE: disable
      ROOT_USER_KEY: ${ROOT_USER_KEY:?DEFINE_ME_IN_ENV} # define here or env. No default, you must set this. You can use `openssl rand -base64 32` to generate one if you need to.
      LOG_LEVEL: ${LOG_LEVEL:-info} # define here or env. Default is info
      PUBLIC_SITE_URL: ${PUBLIC_SITE_URL:-http://localhost:8080} # define here or env. Default is http://localhost:8080
      ENABLE_DOCS: ${ENABLE_DOCS:-true} # define here or env. Default is true
      SERVER_PORT: ${SERVER_PORT:-8080} # define here or env. Default is 8080
    depends_on:
      db:
        condition: service_healthy
    networks:
      - app-network
    ports:
      - "${SERVER_PORT:-8080}:${SERVER_PORT:-8080}" # define here or env. Default is 8080
    restart: unless-stopped

networks:
  app-network:
    driver: bridge

volumes:
  postgres_data:

```

2. You can either use a .env in the same directory as the docker-compose.yml or you can set the variables yourself manually in the docker-compose.yml. See .env.example for the .env format.

3. Run `docker compose up -d` to start the application and database.

### Running with Docker (manual)

0. Install [Docker](https://docs.docker.com/get-docker/) and [Docker Compose Plugin](https://docs.docker.com/compose/install/)
1. Populate the .env file with your own information (see .env.example for the format)
2. Run `make docker-build` to build the Docker image
3. Run `make docker-run` to start the application and database
4. The application will be available at `http://localhost:8080` (or the port you set in .env), the API docs will be available at `http://localhost:8080/docs/` (or the port you set in .env)

- You can use the commands inside the `makefile` to rebuild and restart the docker instance with the `docker-*` commands.
  - `make help` will show you all of the commands as well.

### Running Locally

1. Install and Update [Go](https://go.dev/doc/install)
2. Install and run [PostgreSQL](https://www.postgresql.org/download/)
3. Populate the .env file with your own information (see .env.example for the format)
4. If you don't want to type your postgres user password every time, create a `.pgpass` file in the project directory with the following content (or see .pgpass.example):

```txt
localhost:5432:*:postgres:your_postgres_password
```

4. Run `make setup` to setup the project
5. Remove the `.pgpass` file, if you created it, after setup for security reasons
6. Run `make run` to start the server or `make build` to build the binary
7. You can use `make help` for more commands

See .env.example for environment variables.
