# XM Companies Service

This repository implements a small, layered Go service for managing company records and authenticating users. It exposes a JSON/REST API via Gin, persists data in MySQL, emits Kafka events for company lifecycle changes. Documentation of the API lives in ./api_documentation, api/openapi.yaml, payload samples in api/sample_payloads and the interactive api/swagger.html.” Information about running the project can be found in the **How to run** section.

## What Was Delivered

- CRUD + auth: Implemented POST/PATCH/DELETE/GET for companies with JWT-protected mutating routes (auth-middleware).
- Data model & validation: Enforces 15-char unique names, optional description, employee counts, types, UUID IDs.
- Eventing: Each successful mutation publishes a Kafka event (publisher injected via service layer).
- Production-ready containerization: Multi-stage Dockerfile builds the API binary; docker-compose.yml brings up API, MySQL, Kafka, and Kafka UI.
- Configuration: Centralized via pkg/config, driven by env vars and prod.env for local overrides.
- Testing: Unit tests for service/handler logic
- Documentation: Human-readable API guide (api_documentation), OpenAPI spec (api/openapi.yaml), browser-ready Swagger UI (api/swagger.html).
- Extras: Request ID middleware for info and error correlation, structured logging with Zap, REST handlers built with Gin, repository abstraction for persistence.

## High-Level Architecture

The architecture splits responsibilities across repositories (storage adapters), services (business logic), and the HTTP transport. The layers depend inward (transport → service → repository).  Each layer exports only what the outer layer needs, depends on the next layer through narrow interfaces making implementations swappable and the codebase easier to test.

- **Entry point** – `cmd/api/main.go` loads configuration, builds the application, and starts the HTTP server.
- **Application wiring** – `internal/platform/app/app.go` assembles shared dependencies (logger, database connections, Kafka publisher) and constructs the HTTP router.
- **Transport layer** – `internal/transport/http` defines Gin handlers and middleware. `router.go` wires routes, `company_handler.go` and `auth_handler.go` translate HTTP concerns into service calls, and middleware handles request IDs plus JWT authentication.
- **Service layer** – `internal/service/company` and `internal/service/user` hold business logic. They validate input, map repository errors into domain errors, and publish domain events when companies change.
- **Repository layer** – `internal/repository` defines interfaces and MySQL-backed implementations for companies and users. Company persistence supports CRUD plus partial updates; user repository retrieves hashed credentials for login.
- **Auth utilities** – `internal/auth` wraps JWT generation and parsing, allowing services and middleware to issue and verify tokens.
- **Events** – `internal/platform/events/kafka` provides a publisher abstraction to emit company events to Kafka. The service layer depends on an interface, so event delivery can be swapped or disabled.
- **Domain models** – `internal/domain` defines JSON-friendly structs shared across layers (`Company`, `PatchCompanyRequest`, `UserLoginRequest`, etc.).

## Future improvements

Below are improvements I couldn’t finish due to time constraints but would prioritize in a production setting.

- Add atomicity between persisting company records and emitting Kafka events. The current flow can commit the database without producing the matching message in case the service crashes. The solution is to adopt the transactional outbox pattern so the service writes events to an outbox MYSQL table in the same transaction. Later, a background process will retrieve these messages and securily publish them to kafka. 
- Increase the test coverage of both of units and integration tests. 
- More detailed observability as currently as Gin’s logger and Uber’s Zap provide basic structured logging in stdout.
- Centralized error handling for better readablity of the happy path. 


## Runtime Configuration

To run the project the user has to run the docker compose command from the top-level directory of the project (see **How to run** section). 

Configuration is centralized in `pkg/config/config.go`. On startup the service reads environment variables, applying sensible defaults when values are missing.

A `prod.env` file beside `docker-compose.yml` can supply overrides for local development. Docker Compose interpolates these values into service definitions, avoiding hard-coded secrets in the YAML. 

Remember to obtain a JWT via `/api/v1/login` before hitting secured routes.


## HTTP API and Documentation

- Routes are namespaced under `/api/v1`.
- Authentication:
  - `/api/v1/login` issues JWTs after validating user credentials.
  - All modifying company endpoints (`POST /companies`, `PATCH`, `DELETE`) require a `Bearer` token.
- Company operations:
  - `GET /api/v1/companies/{uuid}`
  - `POST /api/v1/companies`
  - `PATCH /api/v1/companies/{uuid}` - Allows to update entirely or partialy parts of the company (changing UUID is not allowed)
  - `DELETE /api/v1/companies/{uuid}`
- Health probe at `/api/v1/healthz`.

More details:
- Human-friendly overview: `api_documentation`
- OpenAPI definition: `api/openapi.yaml`
- Interactive Swagger UI: open `api/swagger.html` in a browser

## Data Storage

- **MySQL** is used as the persistent storage
- It contains a signle database with two tables, namely users and companies. (see the folder ./scirpts/mysql for the schema)
- They both come withe preloaded data.
- The users table holds the users information including their hashed password with bcrypt. 
- `internal/repository/company/mysql` persists companies, handling uniqueness validation and partial updates. 
- `internal/repository/user/mysql` returns users with hashed passwords for authentication.

## Messaging

- Company lifecycle events are published to Kafka via `internal/platform/events/kafka`. The publisher is constructed in `app.New` and injected into the company service. If Kafka is unavailable, publication errors are logged without failing API requests.

## How to run

The API sevice is dockerized using a multi-stage Dockerfile. The build stage pulls the desired Go toolchain, copies all the necessary files and builds the module. Subsequently, the
runtime starts from a slim base, copies the compiled binary and any required assets, sets the service port, and defines the entrypoint to run the API server.

The Kafka container includes a single replica KRaft based verison that with a topic created during the docker compose phase. In order to create the topic a specialized container (kafka-init)
is utilized. To visualize the contents a Kafka-UI container is included as well.  

THe database container includes a MySQL. During docker compose, two tables are creaeted and are prepopulated.

1. **Prerequisites**
   - Go 1.25+
   - Docker + Docker Compose (installed separately)
2. **Start dependencies**
   1. Start the docker daemon (i.e. either via docker desktop or cli)
   2. Run the following command: 

   ```sh
   docker compose --env-file ./prod.env up --build -d 
   ```
   - The prod.env is used to adjust the host ports of the API server and mysql containers.
   - The stack provisions containers for the API, MySQL, a single replica Kafka with KRaft, a topic initializer, and Kafka UI. 
   - It will take some time to startup depending on the hardware
   - For observation of the Kafka topic, the user can use the Kafka UI (by default running on http://localhost:8082) 
   - To check the contents of the database, the user can use the MySQL workbench to connect  (by default on http://localhost:3307, user:xm, password:xmpass)
3. **Stop dependencies**
   ```sh
   docker compose down
   ```
   -  This terminates all the service containers 
4. **Access the API**
   - Base URL: `http://localhost${HTTP_PORT_HOST}/api/v1/` (defaults to `8081`)για
   - Happy-path create: curl -X POST http://localhost:8081/api/v1/companies with a valid payload and Authorization: Bearer <token>; expect 201 and JSON payload.
   - During the initialization of the MySQL database the users database is populated with a user whose credentials are email:"john_doe@example.com" and password: "12345678"
   - Fetch & verify: call GET /api/v1/companies/{uuid} (using the ID from create) and confirm fields match.
   - Patch workflow: PATCH /api/v1/companies/{uuid} changing one field (passed in the body); expect 201 and follow with a GET to confirm update persisted.
   - Delete workflow: DELETE /api/v1/companies/{uuid}; expect 200, then GET again to ensure 404.
   - Validation checks: repeat POST with malformed JSON to see 400, without token to see 401, and with duplicate name to trigger 409.
   - The required payloads can be seen from the swagger-ui html file.
   - (a user may use MySQL workbench to access the database with having to create a new company first)
5. **Run unit tests**
   ```sh
   go test ./...
   ```
   This is run from the top-level directory of the project.

## Project Layout

```
.
├── api/                     # OpenAPI spec + Swagger UI assets
├── cmd/api/                 # Application entry point
├── internal/
│   ├── auth/                # JWT helpers
│   ├── domain/              # Shared domain types
│   ├── platform/
│   │   ├── app/             # Application assembly
│   │   └── events/kafka/    # Kafka publisher
│   ├── repository/          # Data access interfaces + implementations
│   ├── service/             # Business logic
│   └── transport/http/      # Gin router, handlers, middleware
├── pkg/config/              # Configuration loader
├── scripts/                 # Support scripts (MySQL initialization)
```

## Deployment Notes

- The provided `docker-compose.yml` is suited for local or demo environments. For production, supply secrets via a secure secrets manager and let infrastructure provide Kafka/MySQL endpoints.
- Health checks: `/healthz` is safe for container orchestration probes.
- Observability: Gin’s logger and Uber’s Zap provide basic structured logging; you can configure sinks by adjusting the logger setup in `app.New`.

---
