# Copilot Instructions for mealplan

## Big Picture Architecture
- **Backend (Go):** RESTful API for meal planning, recipes, shopping lists, pantry, tags, and household management. See `mealplan.go` for routing and service boundaries. API endpoints are documented in `openapi.yaml` and served at `/docs`.
- **Frontend (React + Vite):** Located in `frontend/`, communicates with backend via REST API. Data models/types are mirrored in `frontend/src/api.ts`.
- **Database:** Managed via SQL migrations in `migrations/` using Goose. Models are defined in `models/` and accessed via SQLX.
- **Auth:** Uses Clerk for authentication. Backend expects JWT in `Authorization` header. See `api/auth.go` and frontend Clerk integration.

## Developer Workflows
- **Backend Build:**
  - Build with Docker: `docker build .`
  - Local run: `go run mealplan.go`
  - Migrations: `goose -dir migrations postgres up`
- **Frontend Build:**
  - Dev server: `npm run dev` (in `frontend/`)
  - Build: `npm run build`
  - Lint: `npm run lint`
- **Testing:**
  - Backend: Run Go tests with `go test ./...`
  - Frontend: No test setup detected; add if needed.
- **Deployment:**
  - Uses Fly.io (`fly.toml`). Release runs DB migrations before app start.

## Project-Specific Conventions
- **API Data Models:** Types/interfaces in `frontend/src/api.ts` should match backend models and OpenAPI spec.
- **Caching:** Frontend uses `axios-cache-interceptor` for API calls. Cache IDs (e.g., `recipes-list`) are used for invalidation.
- **Slug Uniqueness:** Backend ensures unique slugs for meals/recipes (see `models/meals.go`).
- **Auth Context:** Backend handlers expect user/household context from middleware (`AuthCtx`, `DbCtx`, `IdCtx`).
- **Error Handling:** Use `api.ErrorResponse` for consistent error output.

## Integration Points
- **Clerk:** Both backend and frontend integrate with Clerk for user management. Frontend passes JWT to backend.
- **Tailwind + daisyUI:** Frontend uses Tailwind CSS v4 and daisyUI v5 for UI components. See `tailwind.config.js`.
- **OpenAPI:** API contract in `openapi.yaml`. Served at `/openapi.yaml` and `/docs` (Redoc viewer).

## Key Files & Directories
- `mealplan.go`: Main backend entry, routing, middleware
- `api/`: Backend route handlers
- `models/`: Backend data models
- `migrations/`: SQL migration scripts
- `frontend/src/api.ts`: Frontend API types and calls
- `frontend/src/components/`: React components
- `openapi.yaml`: API contract
- `Dockerfile`, `fly.toml`: Build/deploy config

## Example Patterns
- **Adding a new API resource:**
  1. Define model in `models/`
  2. Add handler in `api/`
  3. Register route in `mealplan.go`
  4. Update OpenAPI spec
  5. Add frontend type and API call in `frontend/src/api.ts`
  6. Use in React component
- **Frontend API call with cache:**
  ```ts
  export const getMeals = () => apiClient.get<Meal[]>("/meals", { id: MEALS_LIST_ID, cache: {} });
  ```

---
If any section is unclear or missing, please provide feedback to improve these instructions.
