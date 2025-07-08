# Polling App Backend

A mini polling application backend built with Go, PostgreSQL, and Docker.

## Features

- RESTful API for managing polls
- PostgreSQL database with Docker
- Chi router for HTTP routing
- Clean architecture with repository pattern

## Prerequisites

Before running this application, make sure you have the following installed:

- [Go](https://golang.org/dl/) (version 1.22.2 or later)
- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)


## Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/marko-la/polling.git
cd polling
```

### 2. Start the Database

Start the PostgreSQL database using Docker Compose:

```bash
docker compose up -d
```

This will:
- Pull the PostgreSQL 14.5 Docker image
- Create a database named `polling`
- Initialize the database with tables from `database/create-database.sql`
- Expose the database on port `5432`

### 3. Install Go Dependencies

```bash
go mod download
```

### 4. Run the Application

```bash
go run ./cmd/api
```

The API server will start on `http://localhost:8080`

## Database Configuration

The application connects to PostgreSQL with the following default settings:

- **Host**: localhost
- **Port**: 5432
- **Database**: polling
- **Username**: postgres
- **Password**: postgres

## API Endpoints

The application provides RESTful endpoints for managing polls. Check the `cmd/api/routes.go` file for available routes.
