# Go JWT Auth Template

A lightweight Go backend demonstrating JWT authentication, bcrypt password hashing, and modular architecture for learning and reuse in future projects.

---

## Purpose

This template serves as a learning guide for understanding authentication and authorization in Go.

- **Models** – data and database logic  
- **Handlers** – HTTP endpoint logic  
- **Middleware** – route security and validation  
- **Utilities** – shared functions such as JWT and password hashing  

The goal is to make it easy to adapt this structure for future full-stack or API projects.

---

## Project Structure

```
backend/
├── config/         # Database connection and environment loading
├── db/             # SQL schema
├── handlers/       # Authentication and user route logic
├── middleware/     # JWT validation middleware
├── models/         # User model and database operations
├── tests/          # Unit and integration tests
└── utils/          # Hashing, JSON, and JWT helpers
```

---

## Login Flow

1. The client sends login credentials to `/login`.  
2. The password is verified using bcrypt (`utils/hash.go`).  
3. A JWT is generated (`utils/jwt.go`) and returned to the client.  
4. Protected routes require the token in the `Authorization: Bearer <token>` header.  
5. The `AuthMiddleware` validates the JWT and injects the `user_id` into the request context.  
6. Handlers, such as `/me`, extract the `user_id` from context and query the database.

---

## Key Files

| File | Description |
|------|--------------|
| `utils/jwt.go` | Creates and validates JWTs |
| `utils/hash.go` | Provides bcrypt password hashing and verification |
| `middleware/auth.go` | Middleware that validates tokens and secures routes |
| `handlers/auth.go` | Handles user login and JWT issuance |
| `handlers/user.go` | Handles `/me` and user deletion endpoints |
| `models/user.go` | Defines `User` struct and database CRUD functions |
| `config/loadenv.go` | Loads environment variables and test configuration |

---

## To Be Expanded

| Area | Description |
|-------|-------------|
| **Router setup** | Add `routes/` using a router, with or without external dependencies|
| **Frontend integration** | Store JWT on the client side (cookies or local storage) and attach to API requests |
| **Signup route** | Add user registration and validation logic |
| **JWT refresh** | Implement refresh tokens and token expiry handling |
| **Logging and error handling** | Structured logging and unified error responses |
| **Rate limiting and CSRF protection** | Security layers for production readiness |

---
