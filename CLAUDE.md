# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Family Plan Manager is a Go web application for managing shared subscription plans (e.g., Netflix, Spotify) among family members and friends. It uses PocketBase for database/authentication, server-side rendering with Go templates, and HTMX for dynamic interactions without custom JavaScript.

**Live site**: https://familyplanmanager.xyz

## Development Commands

### Running the Application

```bash
# Standard run
go run main.go
# or
make run

# Development with hot reload (recommended)
air
# or
make dev
```

The app runs on http://localhost:8090 by default.
PocketBase Admin UI is at http://localhost:8090/_/

### Testing and Linting

```bash
# Run all tests
make test
# or
go test -v ./...

# Run tests with coverage
make test-coverage

# Lint the code
make lint
# This runs go vet and golangci-lint if available
```

### Building

```bash
make build    # Builds to ./app
make clean    # Removes build artifacts and pb_data
make deps     # Run go mod tidy
```

## Architecture

### PocketBase Integration

The application is built on top of PocketBase, which provides:
- SQLite database with automatic migrations
- Built-in authentication system (users collection)
- HTTP routing via Echo framework
- Data access layer (DAO)

**Key pattern**: The app extends PocketBase by registering custom routes and hooks rather than building from scratch.

### Database Collections

The schema is managed through migrations in `migrations/` directory:

1. **users** (PocketBase built-in auth collection)
   - username, password (hashed), name, tokenKey
   - tokenKey stores custom session tokens for cookie-based auth

2. **family_plans**
   - Core subscription plan data
   - Fields: name, description, cost, individual_cost, owner (relation to users), join_code (unique)
   - join_code is used as the URL path for accessing plans

3. **memberships**
   - Links users to plans they're members of
   - Fields: plan_id, user_id, leave_requested (bool), date_ended, is_artificial (bool), name (for artificial members)
   - Soft deletes: date_ended is set instead of deleting the record
   - Supports "artificial members" for tracking non-registered users

4. **join_requests**
   - Pending requests to join a plan
   - Fields: plan_id, user_id
   - Deleted after approval/denial

5. **payments**
   - Payment tracking for members
   - Fields: plan_id, user_id, amount, date, status (pending/approved/rejected), notes, for_month
   - Status workflow: pending → approved/rejected
   - for_month allows attributing payments to specific billing periods

### Request Flow

```
main.go
  ↓
setupRoutes (routes.go) - registers all HTTP routes
  ↓
Auth middleware (auth_handlers.go) - validates cookie and sets session context
  ↓
Handler functions (auth_handlers.go, plan_handlers.go)
  ↓
Business logic (plan_actions.go) - complex operations like balance calculations
  ↓
PocketBase DAO - database operations
```

### Authentication System

Custom cookie-based auth instead of PocketBase's default JWT:
- Login generates a random token stored in user's tokenKey field
- Cookie contains the token value
- Middleware looks up user by tokenKey on each request
- Session data stored in Echo context as SessionData struct

### Balance Calculation Logic

The balance system (in `calculateMemberBalance` function in plan_actions.go) is complex:
- Calculates member balance by month-by-month analysis
- Accounts for when members join/leave mid-month (they pay for the full month)
- Divides plan cost by active member count for each month
- Tracks payments by month when for_month is specified
- Positive balance = member has overpaid, negative = owes money
- Members can't leave until balance is >= 0

### Template Rendering

- Templates stored in `templates/` directory as embedded FS
- Uses Go html/template with custom template functions (defined in template_renderer.go)
- Layout pattern: layout.html wraps page-specific templates
- HTMX attributes in templates enable dynamic updates

### Artificial Members Feature

Allows plan owners to add non-registered "placeholder" members:
- Creates membership with is_artificial=true and auto-generated user_id
- Tracks payments and balances like real members
- Can be "transferred" to a real user when they join (preserving payment history)
- Useful for tracking family members who haven't registered yet

## Important Patterns

### PocketBase Record Operations

```go
// Finding records
collection, _ := app.Dao().FindCollectionByNameOrId("collection_name")
record, _ := app.Dao().FindRecordById(collection.Id, recordId)
records, _ := app.Dao().FindRecordsByFilter(collection.Id, "field = 'value'", "", limit, offset)

// Creating records
newRecord := models.NewRecord(collection)
newRecord.Set("field", value)
app.Dao().SaveRecord(newRecord)

// Updating records
record.Set("field", newValue)
app.Dao().SaveRecord(record)
```

### Transaction Pattern

Use `app.Dao().RunInTransaction()` for operations that must succeed/fail together (see handleDeletePlan):

```go
err = app.Dao().RunInTransaction(func(txDao *daos.Dao) error {
    // All database operations here use txDao
    return nil
})
```

### Migration Pattern

Migrations in `migrations/` use init() to self-register. Each has an up and down function. The app runs migrations automatically on startup (see migratecmd.Config in main.go).

## File Organization

- `main.go` - Entry point, PocketBase initialization, serves static files
- `routes.go` - Route registration
- `auth_handlers.go` - Authentication routes and middleware
- `plan_handlers.go` - Plan viewing/listing handlers
- `plan_actions.go` - Complex business logic (join/leave/payments/balance calculations)
- `models.go` - Go structs for data models
- `utils.go` - Utility functions
- `template_renderer.go` - Custom template renderer and functions
- `init_db.go` - Legacy manual collection initialization (no longer used with migrations)
- `migrations/*.go` - Database schema migrations
- `templates/*.html` - HTML templates
- `static/` - CSS, JS, images

## Deployment Details

The app is deployed on DigitalOcean with Nginx reverse proxy:
- Runs as systemd service
- Nginx handles SSL termination (Let's Encrypt)
- Uses --http=0.0.0.0:8090 for IPv4 binding
- UFW firewall configured for SSH/HTTP/HTTPS only

## Code Style Notes

- PocketBase records use relation fields that return slices (e.g., ownerSlice := plan.GetStringSlice("owner"))
- Owner is always the first element of the owner relation slice
- Use -1 as limit parameter to FindRecordsByFilter to get all records
- Date fields use PocketBase's DateTime type (record.GetDateTime("field"))
- Always check if user is plan owner before allowing destructive operations
