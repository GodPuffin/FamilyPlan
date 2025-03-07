# Family Plan Management Application

This is a web application for managing shared family plans and subscriptions, built with HTMX, Go, and PocketBase for a modern, interactive user experience with minimal JavaScript.

## Features

- Create and manage family plans for shared subscriptions
- Invite members using unique join codes
- Approve or reject join requests
- Track monthly costs and membership details
- Owner controls for updating plan details and managing members
- Server-side rendering with Go templates
- HTMX for dynamic content without writing custom JavaScript
- PocketBase for database and authentication

## Prerequisites

- Go 1.21 or higher
- Git

## Getting Started

1. Clone this repository:
   ```
   git clone https://github.com/yourusername/family-plan-manager.git
   cd family-plan-manager
   ```

2. Install dependencies:
   ```
   go mod tidy
   ```

3. Run the application:
   ```
   go run main.go
   ```
   
   Alternatively, use the Makefile:
   ```
   make run
   ```

4. Open your browser and navigate to:
   ```
   http://localhost:8090
   ```

5. Access the PocketBase Admin UI:
   ```
   http://localhost:8090/_/
   ```

## Development

For development, you can use [Air](https://github.com/cosmtrek/air) for hot reloading:

```
# Install Air
go install github.com/cosmtrek/air@latest

# Run with Air
air
```

The project includes an `.air.toml` configuration file for Air.

## Setting up the Database

When you first run the application, you'll need to set up the PocketBase database:

1. Navigate to http://localhost:8090/_/
2. Create an admin account
   
## Project Structure

- `main.go` - Main application entry point
- `models.go` - Data models and structures
- `routes.go` - HTTP route definitions
- `auth_handlers.go` - Authentication-related handlers
- `plan_handlers.go` - Plan management handlers
- `plan_actions.go` - Business logic for plan operations
- `init_db.go` - Database initialization
- `utils.go` - Utility functions
- `template_renderer.go` - Template rendering utilities
- `templates/` - HTML templates for the web interface
- `static/` - Static assets (CSS, JS, images)
- `migrations/` - Database migration files
- `pb_data/` - PocketBase data directory (created automatically)

## Deployment

This application is currently deployed at [familyplanmanager.xyz](https://familyplanmanager.xyz) using a DigitalOcean Droplet with the following configuration:

- Ubuntu 24.04 LTS
- Nginx as a reverse proxy
- SSL certificates from Let's Encrypt
- UFW firewall for security
- Systemd for service management

### Common Deployment Notes

- The application runs on port 8090 by default
- When deploying, ensure IPv4 binding with `--http=0.0.0.0:8090`
- For security, set up a firewall allowing only necessary ports (SSH, HTTP, HTTPS)
- Use a reverse proxy (like Nginx) for SSL termination and to route traffic to the application

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Technologies Used

- [Go](https://golang.org/)
- [HTMX](https://htmx.org/)
- [PocketBase](https://pocketbase.io/)
- [Tailwind CSS](https://tailwindcss.com/)
- [Nginx](https://nginx.org/)
- [Certbot/Let's Encrypt](https://certbot.eff.org/)

## License

MIT 
