# Bot Logic Trade - Polyglot Matching Engine Microservices

This project is a scalable, polyglot backend for a trading platform using microservices in Go, Java, Node.js, Rust, and C++. Each language implements a matching engine microservice, with Redis as a shared cache and event-driven stubs for Kafka/RabbitMQ integration.

## Features
- **Polyglot Microservices:** Go, Java, Node.js, Rust, and C++ matching engines, each as a standalone microservice.
- **Unified API:** All services expose `/order`, `/orderbook`, `/health` endpoints for easy orchestration and benchmarking.
- **Concurrency & Memory Management:** Each service uses language-appropriate concurrency controls (mutexes, atomic counters, event loop safety) and documents memory management strategies.
- **Redis Caching:** Orderbooks and order states are managed in Redis for persistence and speed.
- **Event-Driven Stubs:** Stubs for Kafka/RabbitMQ event publishing are included for future extensibility.
- **Docker Compose:** One command launches the entire stack (all services + Redis).
- **Comprehensive Documentation:** Inline comments and endpoint docs in every service.

## Project Structure
- `go/`    — Go matching engine (Gin, go-redis)
- `java/`  — Java Spring Boot matching engine
- `node/`  — Node.js matching engine (Express, redis)
- `rust/`  — Rust matching engine (actix-web, redis)
- `cpp/`   — C++ matching engine (Crow, redis++, nlohmann/json)
- `.env`   — Shared environment variables
- `docker-compose.yml` — Orchestrates all services and Redis

## Endpoints & Ports
| Language | Port  | Endpoints               |
|----------|-------|-------------------------|
| Go       | 8081  | `/order`, `/orderbook`, `/health` |
| Java     | 8082  | `/order`, `/orderbook`, `/health` |
| Node.js  | 8083  | `/order`, `/orderbook`, `/health` |
| Rust     | 8084  | `/order`, `/orderbook`, `/health` |
| C++      | 8085  | `/order`, `/orderbook`, `/health` |

All services use `REDIS_URL=redis://redis:6379` (set by Docker Compose).

## Quick Start (Docker Compose)

1. Ensure Docker is installed.
2. In the project root, run:
   ```sh
   docker-compose up --build
   ```
3. All services and Redis will start. Access endpoints at `localhost:<port>`.

## Usage
- **Add Order:** `POST /order` with JSON body `{ "userId": "...", "type": "buy"|"sell", "price": 100, "amount": 10 }`
- **Get Orderbook:** `GET /orderbook`
- **Health Check:** `GET /health`

## Extensibility
- **Event-Driven:** Stubs for Kafka/RabbitMQ are present for settlement/event workflows.
- **API Gateway:** Add a Node.js (or other) API gateway for unified access and benchmarking.
- **Benchmarking:** All endpoints are aligned for easy performance comparison.
- **Monitoring/Testing:** Add logging, monitoring, or automated tests as needed.

---

See each language folder for details and Docker instructions. Contributions and extensions welcome!
