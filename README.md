# Bullshit.wtf

A multiplayer social game where players try to fool each other with fake answers to trivia questions.  
Based on the the bullshit.wtf - [GitHub repo](https://github.com/radotzki/bullshit-wtf) with a new UI, more question and many quality of life features.

### Disclaimer:

The base of the project was done via a clean room AI implementation.
Similarly, all later work does not reference the original code from the bullshit.wtf - [GitHub repo](https://github.com/radotzki/bullshit-wtf)

## Development

### Dependencies

To run the full development environment and linting tools, you will need:

- **Node.js** (v18+)
- **Go** (v1.26+)
- **Docker** (for the database)
- **golangci-lint**: Installed via Homebrew on macOS:
  ```bash
  brew install golangci-lint
  ```

### Linting

You can lint both the client (TypeScript/React) and the server (Go) from the project root:

```bash
npm run lint
```

- **Client**: Uses ESLint for React and TypeScript.
- **Server**: Uses `golangci-lint` with the `--fix` flag enabled to automatically resolve common issues.

### Running the App

There are multiple ways to run the application depending on your needs.

#### 1. Local Development (Recommended)

This method runs the database in Docker (via Orbstack/Docker Desktop) while the client and server run on your host machine for fast live-reloading and debugging.

```bash
npm run dev
```

_Behind the scenes, this starts the `db` container, then concurrently runs the Go server on port 8080 and the Vite dev server on port 5173._

#### 2. Partial Docker (Just Database)

If you want to manage the client and server processes manually, you can start only the PostgreSQL database in the background:

```bash
docker compose up -d db
```

#### 3. Backend-in-Docker (Frontend Local)

Useful if you are focusing on frontend work and don't want to set up a Go environment. This runs the Database and the Go Server in Docker, while you run the Vite frontend locally with HMR (Hot Module Replacement).

1. Start the backend services:
   ```bash
   docker compose up -d db server
   ```
2. Start the local frontend:
   ```bash
   cd client && npm run dev
   ```
   _The local Vite server (port 5173) is configured to proxy `/api` and `/ws` requests to the containerized server on port 8080._

#### 4. Full Production-like Stack

This runs the entire application—database, Go server (serving the built frontend), and a Caddy reverse proxy—completely inside Docker. This is closest to how the app runs in production.

```bash
docker compose up --build
```

_The app will be accessible at `http://localhost`. Note that this requires building the frontend and Go binary inside containers, which is slower than local development but ensures a consistent environment._

## Project Structure

- `/client`: React (Vite) frontend application.
- `/server`: Go (Gin/WebSocket) backend API.
- `/server/internal/db/migrations`: SQL migration files for the database schema.
