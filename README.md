# Bot Logic Trade - Polyglot Matching Engine Microservices

This project is a scalable, polyglot backend for a trading platform using microservices in Go, Java, Node.js, Rust, and C++. Each language implements a matching engine microservice, with Redis as a shared cache and event-driven stubs for Kafka/RabbitMQ integration.

## Features
- **Polyglot Microservices:** The project demonstrates the implementation of matching engines in multiple programming languages (Go, Java, Node.js, Rust, and C++). Each microservice is self-contained and showcases the idiomatic use of its respective language.
- **Unified API:** All services expose the same set of RESTful endpoints (`/order`, `/orderbook`, `/health`) to ensure consistency and make it easy to integrate or benchmark across different implementations.
- **Concurrency & Memory Management:** Each service employs concurrency mechanisms suitable for its language:
  - **Go:** Uses mutexes and atomic operations for thread safety.
  - **Java:** Relies on synchronized blocks and thread-safe collections.
  - **Node.js:** Leverages its single-threaded event loop for atomicity.
  - **Rust:** Utilizes `Mutex` and other thread-safe primitives.
  - **C++:** Implements mutexes for thread-safe operations.
- **Redis Caching:** Redis is used as a shared in-memory database to store orderbooks and order states. This ensures high performance and persistence across all services.
- **Event-Driven Stubs:** Stubs for Kafka/RabbitMQ are included to facilitate future integration with event-driven architectures, enabling features like settlement workflows or notifications.
- **Docker Compose:** A single `docker-compose.yml` file orchestrates all services and Redis, simplifying deployment and testing.
- **Comprehensive Documentation:** Each service includes inline comments and endpoint documentation to help developers understand the implementation and extend the functionality.

## Project Structure
- `go/`    — Contains the Go implementation of the matching engine, built with the Gin web framework and go-redis library for Redis integration.
- `java/`  — Contains the Java implementation using Spring Boot, with RedisTemplate for Redis operations.
- `node/`  — Contains the Node.js implementation using Express.js and the `redis` library for Redis communication.
- `rust/`  — Contains the Rust implementation using the Actix-web framework and the `redis` crate for Redis integration.
- `cpp/`   — Contains the C++ implementation using the Crow web framework, redis++ library for Redis, and nlohmann/json for JSON serialization.
- `.env`   — A shared environment file for configuration, such as the Redis URL.
- `docker-compose.yml` — A Docker Compose file to launch all services and Redis in a single command.

## Endpoints & Ports
| Language | Port  | Endpoints               |
|----------|-------|-------------------------|
| Go       | 8081  | `/order`, `/orderbook`, `/health` |
| Java     | 8082  | `/order`, `/orderbook`, `/health` |
| Node.js  | 8083  | `/order`, `/orderbook`, `/health` |
| Rust     | 8084  | `/order`, `/orderbook`, `/health` |
| C++      | 8085  | `/order`, `/orderbook`, `/health` |

### Additional Notes
- **Scalability:** Each microservice is designed to be horizontally scalable. Redis acts as a central cache, ensuring data consistency across instances.
- **Extensibility:** The project is structured to allow easy addition of new features, such as integrating with external APIs or adding new microservices in other languages.
- **Benchmarking:** The unified API and consistent endpoints make it easy to compare the performance of different implementations under similar workloads.
- **Quick Start:** Developers can quickly spin up the entire stack using Docker Compose, making it ideal for testing and development environments.

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
