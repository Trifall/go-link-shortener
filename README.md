# go-link-shortener

A URL shortener written in Go using PostgreSQL.

## Getting Started

### Prerequisites

1. Clone the repository
2. Install [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/)

### Running with Docker

1. Run `make docker-build` to build the Docker image
2. Run `make docker-run` to start the application and database
3. The application will be available at `http://localhost:8080` (or the port you set in .env), the API docs will be available at `http://localhost:8080/docs/` (or the port you set in .env)

### Running Locally

1. Install and Update [Go](https://go.dev/doc/install)
2. Install and run [PostgreSQL](https://www.postgresql.org/download/)
3. If you don't want to type your postgres user password every time, create a `.pgpass` file in the project directory with the following content (or see .pgpass.example):

```txt
localhost:5432:*:postgres:your_postgres_password
```

4. Run `make setup` to setup the project
5. Remove the `.pgpass` file, if you created it, after setup for security reasons
6. Run `make run` to start the server or `make build` to build the binary
7. You can use `make help` for more commands

See .env.example for environment variables.
