# go-link-shortener

A URL shortener written in Go using PostgreSQL.

## Getting Started

### Prerequisites

1. Clone the repository
2. Install and Update [Go](https://go.dev/doc/install)
3. Install and run [PostgreSQL](https://www.postgresql.org/download/)
4. If you don't want to type your postgres user password every time, create a `.pgpass` file in the project directory with the following content (or see .pgpass.example):

```txt
localhost:5432:*:postgres:your_postgres_password
```

### Compilation and Setup

1. Run `make setup` to setup the project
2. Remove the `.pgpass` file, if you created it, after setup for security reasons
3. Run `make run` to start the server or `make build` to build the binary
4. You can use `make help` for more commands

See .env.example for environment variables.
