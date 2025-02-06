# go-link-shortener

A URL shortener written in Go using PostgreSQL. Their is a sibling project at [Link Shortener UI](https://github.com/Trifall/link-shortener-ui).

## Getting Started

### Prerequisites

1. Clone the repository

### Notes

#### Make File

The `makefile` contains all of the various commands for running the server, whether its through docker or locally. Read it for more info

#### Scripts

- The `scripts` folder is for local development only, if you are running through docker, you can ignore this folder.
  - Read the `Running Locally` section for this, it sets up the PostgreSQL database for you.

### Running with Docker (recommended)

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
