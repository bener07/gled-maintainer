# SUGGESTIONS.md — Maintainer Server

## Overview

The maintainer is the central update/maintenance server for the GLED platform. It manages registered equipment servers (clients), distributes software updates, and provides a dashboard for administrators. Below are identified issues and suggestions for building it out properly.

---

## Critical Issues Found

### 1. Hardcoded Credentials in docker-compose.yml
- **Location:** `docker-compose.yml` (lines 25–29, 40, 71–72)
- **Problem:** Database passwords and Twingate tokens are in plaintext in source control.
- **Fix:** Use a `.env` file and reference `${VARIABLE}` in docker-compose. Add `.env` to `.gitignore`. Generate a `.env.example` with empty values.

### 2. No API Authentication
- **Location:** `api/server.go`
- **Problem:** All endpoints (`GET /users`, `POST /users`) are completely open — no token validation, no middleware, no rate limiting.
- **Fix:** Implement JWT or pre-shared secret middleware. Every endpoint must verify an `Authorization: Bearer <token>` header.

### 3. Twingate Tokens in Version Control
- **Location:** `docker-compose.yml` lines 71–72
- **Problem:** Access tokens and refresh tokens are committed — they are considered compromised.
- **Fix:** Revoke all exposed tokens immediately. Store them via secrets management (e.g., Docker secrets or environment variables from a vault).

### 4. SQL Injection Risk
- **Location:** `api/database/db.go`
- **Problem:** `GetQuery()` accepts a raw query string passed from `server.go`. If any user-controlled input reaches this function, it is injectable.
- **Fix:** Always use parameterized queries with the `args ...interface{}` variadic already in the signature. Never build SQL strings with string concatenation.

### 5. connector.go Does Not Compile
- **Location:** `api/Clients/connector.go`
- **Problem:** `NewConnection()` declares both pointer and error return types but only returns `resp.Body` (an `io.ReadCloser`), causing a compilation error.
- **Fix:** Properly implement the HTTP client. Return `(*ClientConfig, error)` with proper error handling.

### 6. Dead Code in server.go
- **Location:** `api/server.go` line 72+
- **Problem:** `r.Run(":8000")` blocks; code after it (`http.Handle()`) is unreachable.
- **Fix:** Remove unreachable code. Use only the Gin router `r.Run()`.

### 7. Frontend in Development Mode in Production
- **Location:** `app/Dockerfile`
- **Problem:** `CMD ["npm", "run", "serve"]` runs a development server with hot reload — not suitable for production.
- **Fix:** Use a multi-stage Docker build: build with `npm run build`, then serve the `dist/` folder with Nginx.

### 8. MySQL 5.7 is End-of-Life
- **Location:** `docker-compose.yml`
- **Problem:** MySQL 5.7 reached EOL in October 2023 — no longer receives security patches.
- **Fix:** Upgrade to MySQL 8.0 (same as the main `projeto/` app).

---

## Architecture Suggestions

### How the Maintenance Pipeline Should Work

```
[Admin Dashboard (Vue)] ──► [Go API] ──► [MySQL DB]
                                │
                         (via Twingate VPN)
                                │
                    [maintainer-client on each equipment server]
```

#### Step 1 — Client Registration
- On first boot, `maintainer-client` connects through the Twingate VPN and registers with the maintainer API.
- Registration sends: machine ID, OS version, installed app version, IP address.
- Maintainer stores this in a `clients` table.

#### Step 2 — Heartbeat / Status Monitoring
- Client sends a heartbeat every 60 seconds.
- Maintainer marks clients as `online`, `offline`, or `unreachable`.
- Dashboard shows a live status grid of all registered machines.

#### Step 3 — Update Distribution
- Admin uploads a new version package via the dashboard.
- Maintainer stores it in DB with: version number, checksum (SHA-256), changelog, target OS.
- Clients poll `GET /updates/latest` periodically (every 5–15 minutes).
- If a newer version is available, client downloads and validates the checksum before installing.

#### Step 4 — Update Confirmation
- After applying the update, client calls `POST /updates/confirm` with result (success/failure).
- Maintainer updates client record in DB.
- Dashboard shows per-client update status.

#### Step 5 — Remote Maintenance Access
- Admin requests maintenance session for a specific client via dashboard.
- Maintainer generates a time-limited access token and sends it to the client.
- Client opens a maintenance channel (e.g., SSH tunnel over Twingate) and reports session start.
- All maintenance actions are logged with timestamps for audit.

---

## Missing API Endpoints to Implement

| Method | Path | Description |
|---|---|---|
| POST | `/clients/register` | Register a new client machine |
| GET | `/clients` | List all registered clients with status |
| GET | `/clients/:id` | Get client details |
| POST | `/clients/:id/heartbeat` | Client heartbeat |
| GET | `/updates/latest` | Get latest update package info |
| POST | `/updates` | Admin uploads a new update |
| POST | `/updates/confirm` | Client reports update result |
| POST | `/maintenance/request` | Admin requests maintenance session |
| POST | `/maintenance/end` | End maintenance session |

---

## Code Quality Improvements

1. **Add dependency injection** — extract DB logic from `server.go` into service/repository packages.
2. **Add structured logging** — use `logrus` or `zerolog` instead of `fmt.Println`.
3. **Add graceful shutdown** — catch `SIGTERM`/`SIGINT` and clean up connections.
4. **Add health check endpoint** — `GET /health` that checks DB connectivity, returns 200/503.
5. **Add Docker health checks** in `docker-compose.yml` for all services.
6. **Add database migrations** — use `golang-migrate` or similar to version the schema.
7. **Add `.env.example`** file with all required environment variables documented.
8. **Use Alpine-based Go images** for smaller Docker images.
