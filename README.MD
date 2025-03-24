# Calendar App

A modern calendar application with user authentication, event management, and holiday tracking.

## Setup Instructions

### Prerequisites

- Go 1.23 or higher
- Node.js and npm
- Docker and Docker Compose

### Required Go Tools

```bash
# Install templ for HTML templating
go install github.com/a-h/templ/cmd/templ@latest

# Install goose for database migrations
go install github.com/pressly/goose/v3/cmd/goose@latest
```

### Development

#### CSS Build

TailwindCSS is used for styling:

```bash
# Initial build or after adding new styles
npx tailwindcss -i app/assets/app.css -o ./public/assets/styles.css

# Watch mode for development
npx tailwindcss -i app/assets/app.css -o ./public/assets/styles.css --watch
```

#### Authentication

The application uses an authentication system that can be accessed in handlers via:

```go
userID := kit.Auth().(auth.Auth).UserID
```

## Deployment

### Docker Compose

Deploy the application using Docker Compose:

```bash
# Start the application
sudo docker-compose up -d


# Update from Git with a clean rebuild
docker compose down && docker compose build --no-cache && docker compose up -d
```

### Cloudflare Tunnels

To expose the application through Cloudflare Tunnels:

1. Create a Docker network:
   ```bash
   sudo docker network create app-network
   ```

2. Run Cloudflare tunnel:
   ```bash
   sudo docker run -d --name cloudflared --network app-network \
     cloudflare/cloudflared:latest tunnel --no-autoupdate run --token YOUR_TOKEN_HERE
   ```

3. Make sure your docker-compose.yaml uses the same network.

## TODO List

- [ ] Allow deletion and editing of calendar entries in Calendar view
- [ ] Fix Juhannus holiday calculation
- [ ] Fix duplicate email signup error