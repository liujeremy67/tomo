# Running Tests

This guide explains how to run integration tests for the backend using the test database defined in `.env.test`.

---

## Start the Test Database

Start the PostgreSQL container for tests:

`docker-compose -f docker-compose.test.yml up -d`

This starts a separate test database on the port defined in `.env.test`.

---

## Setting Test Environment

Create a `.env.test` file from the `tests/` folder.

Local example with dummy credentials:

```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASS=password
DB_NAME=auth_template

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
JWT_EXPIRY_MINUTES=60

# Google OAuth Configuration
GOOGLE_CLIENT_ID=your-google-client-id.apps.googleusercontent.com

# Application Environment
APP_ENV=development
```

---

## Importing Environment Variables

Tests require the environment variables from `.env.test`. You can skip manual imports and let `test-main.go` load them. Functions defined in `setup.go`.

---

## Run All Tests

From the `backend/` folder:

`go test ./tests/... -v`

- `-v` enables verbose output.
- Tests will automatically connect to the test database.

---

## Clean Up Between Runs

Tests truncate tables before each run, so they remain isolated. After finishing tests, stop and remove the test database container:

`docker-compose -f docker-compose.test.yml down -v`

This removes the test container and associated volumes.

---

## Notes

- The `RunSchema` function automatically loads `db/schema.sql` into the test database before tests start.  
- All tests under `tests/` will use this test database; do not point them to development database.