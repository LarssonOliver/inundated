# 🌊 inundated

> A personal time and task management system.

**inundated** is a lightweight, self-hosted app that helps me track time across 
projects and tasks. It is developed for me, by me. 

---

## 🛠 Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go |
| Frontend | Vue 3 + Pinia + TypeScript + Vite |
| Database | PostgreSQL (or in-memory for dev) |
| API spec | OpenAPI 3 |

---

## 🚀 Getting Started

### Prerequisites

- [Go](https://go.dev/) 
- [Node.js](https://nodejs.org/) 
- [PostgreSQL](https://www.postgresql.org/) (or use the built-in in-memory store for quick local testing)

### Running in development

1. **Clone the repo**

   ```bash
   git clone https://github.com/LarssonOliver/inundated.git
   cd inundated
   ```

2. **Start the backend** (defaults to the in-memory store — no database required for local tinkering)

   ```bash
   make dev
   ```

3. **Start the frontend** (in a separate terminal)

   ```bash
   cd frontend
   npm install
   npm run dev
   ```

   The frontend dev server will open at `http://localhost:5173` and proxy API requests to the backend.

### Connecting to PostgreSQL

Set the `DATABASE_URL` environment variable before starting the server:

```bash
export DATABASE_URL="postgresql://user:password@localhost/inundated"
make dev
```

---

## 🏗 Building for Production

Build a single self-contained binary (frontend assets embedded):

```bash
make build
./bin/inundated
```

Or build and push a multi-arch Docker image using buildx:

```bash
make image-push
```

---

## 🧪 Testing

```bash
# Run all tests (backend + frontend)
make test

# Backend only
make test-backend

# Frontend only
make test-frontend
```

---

## 🔧 Code Generation

The REST API is driven by the OpenAPI spec in `openapi/inundated.yaml`. After editing the spec, regenerate the server stubs and TypeScript client:

```bash
make generate
```

---

## 📄 License

MIT © [Oliver Larsson](https://github.com/LarssonOliver) — see [LICENSE](LICENSE) for details.
