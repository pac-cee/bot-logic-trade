# Bot Logic Trade - Polyglot Matching Engine Microservices

This project demonstrates a scalable, polyglot backend for a trading platform using microservices in Go, Java, Node.js, Rust, and C++. Each language implements a matching engine microservice, with Redis as a shared cache and Kafka/RabbitMQ for event-driven communication.

## Structure

- `go/`    — Go implementation (matching engine, Redis, Docker)
- `java/`  — Java Spring Boot implementation
- `node/`  — Node.js implementation (API Gateway and matching engine)
- `rust/`  — Rust implementation
- `cpp/`   — C++ implementation
- `.env`   — Shared environment variables

## Quick Start

1. Copy `.env.example` to `.env` and update credentials.
2. Run the matching engine in your desired language (see each folder for instructions).
3. Use the Node.js API Gateway to benchmark or orchestrate calls.

## Microservices
- Each service exposes `/order`, `/orderbook`, `/health` endpoints.
- Uses Redis for orderbook and Kafka/RabbitMQ for event publishing.

---

See each language folder for details and Docker instructions.
